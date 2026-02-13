// Package checkin performs device registration with Google.
package checkin

import (
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/nicksrandall/gauth/internal/config"
	"github.com/nicksrandall/gauth/internal/proto"
)

const checkinURL = "https://android.clients.google.com/checkin"

// Result contains the check-in response.
type Result struct {
	AndroidID     uint64
	SecurityToken uint64
}

// Checkin performs device registration and returns a GSF ID + security token.
func Checkin(cfg *config.Config) (*Result, error) {
	msg := buildCheckinRequest(cfg)

	// Encode to protobuf
	encoded, err := proto.Encode(msg, proto.CheckinRequestSchema)
	if err != nil {
		return nil, fmt.Errorf("encode checkin request: %w", err)
	}

	// GZip compress
	var body strings.Builder
	gz := gzip.NewWriter(&body)
	if _, err := gz.Write(encoded); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	// POST request
	req, err := http.NewRequest("POST", checkinURL, strings.NewReader(body.String()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-protobuffer")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("User-Agent", "Android-Checkin/2.0 (vbox86p JLS36G); gzip")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("checkin request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("checkin failed: status %d: %s", resp.StatusCode, string(respBody))
	}

	// Decompress response
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	respBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Decode protobuf response
	decoded, err := proto.DecodeMessage(respBytes)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	result := &Result{}

	// Field 7 = androidId (fixed64)
	if v, ok := decoded["7"]; ok {
		switch val := v.(type) {
		case uint64:
			result.AndroidID = val
		case int64:
			result.AndroidID = uint64(val)
		}
	}

	// Field 8 = securityToken (fixed64)
	if v, ok := decoded["8"]; ok {
		switch val := v.(type) {
		case uint64:
			result.SecurityToken = val
		case int64:
			result.SecurityToken = uint64(val)
		}
	}

	if result.AndroidID == 0 {
		return nil, fmt.Errorf("checkin response missing androidId")
	}

	return result, nil
}

func buildCheckinRequest(cfg *config.Config) map[string]interface{} {
	now := time.Now()

	return map[string]interface{}{
		"2": int64(0), // androidId = 0 for first check-in
		"4": map[string]interface{}{ // Checkin
			"1": map[string]interface{}{ // Build
				"1":  cfg.Device.Fingerprint,
				"2":  cfg.Device.Hardware,
				"3":  cfg.Device.Brand,
				"5":  cfg.Device.Bootloader,
				"6":  "android-google",
				"7":  cfg.Device.BuildTime,
				"9":  cfg.Device.Device,
				"10": cfg.Device.SDKVersion,
				"11": cfg.Device.Model,
				"12": cfg.Device.Manufacturer,
				"13": cfg.Device.Product,
				"14": false, // otaInstalled
			},
			"2": int64(0), // lastCheckinMs
			"3": []interface{}{ // Event (repeated)
				map[string]interface{}{
					"1": "event_log_start",
					"3": now.UnixMilli(),
				},
			},
			"8": "WIFI::",
			"9": 0, // userNumber
		},
		"6":  "en_US",
		"7":  rand.Int63(),      // loggingId
		"11": []interface{}{""}, // accountCookie (empty on first)
		"12": time.Now().Location().String(),
		"14": 3,                                         // version
		"15": []interface{}{"71Q6Rn2DDZl1zPDVaaeEHItd"}, // otaCert
		"18": map[string]interface{}{ // DeviceConfig
			"1":  3,      // touchScreen
			"2":  1,      // keyboardType
			"3":  1,      // navigation
			"4":  2,      // screenLayout
			"5":  false,  // hasHardKeyboard
			"6":  false,  // hasFiveWayNavigation
			"7":  420,    // densityDpi
			"8":  196608, // glEsVersion (3.0)
			"11": []interface{}{"arm64-v8a", "armeabi-v7a", "armeabi"},
			"12": 1080, // widthPixels
			"13": 2400, // heightPixels
		},
		"20": 0, // fragment (0 for first check-in)
	}
}
