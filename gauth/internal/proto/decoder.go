package proto

import (
	"encoding/binary"
	"fmt"
	"math"
)

// DecodeMessage decodes protobuf bytes into a map[string]interface{}.
// Values are decoded as: varint→int64, fixed64→uint64, fixed32→uint32, bytes→[]byte or nested message.
func DecodeMessage(data []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	pos := 0

	for pos < len(data) {
		// Read tag
		tag, n := binary.Uvarint(data[pos:])
		if n <= 0 {
			return nil, fmt.Errorf("failed to read tag at pos %d", pos)
		}
		pos += n

		fieldNum := int(tag >> 3)
		wireType := WireType(tag & 0x7)
		key := fmt.Sprintf("%d", fieldNum)

		switch wireType {
		case WireVarint:
			val, n := binary.Uvarint(data[pos:])
			if n <= 0 {
				return nil, fmt.Errorf("failed to read varint for field %d", fieldNum)
			}
			pos += n
			addToResult(result, key, int64(val))

		case WireFixed64:
			if pos+8 > len(data) {
				return nil, fmt.Errorf("not enough data for fixed64 field %d", fieldNum)
			}
			val := binary.LittleEndian.Uint64(data[pos : pos+8])
			pos += 8
			addToResult(result, key, val)

		case WireBytes:
			length, n := binary.Uvarint(data[pos:])
			if n <= 0 {
				return nil, fmt.Errorf("failed to read length for field %d", fieldNum)
			}
			pos += n
			if pos+int(length) > len(data) {
				return nil, fmt.Errorf("not enough data for bytes field %d", fieldNum)
			}
			payload := data[pos : pos+int(length)]
			pos += int(length)

			// Try to decode as nested message
			nested, err := DecodeMessage(payload)
			if err == nil && len(nested) > 0 {
				addToResult(result, key, nested)
			} else {
				// Try as string (if it looks like valid UTF-8 text)
				if isLikelyString(payload) {
					addToResult(result, key, string(payload))
				} else {
					addToResult(result, key, payload)
				}
			}

		case WireFixed32:
			if pos+4 > len(data) {
				return nil, fmt.Errorf("not enough data for fixed32 field %d", fieldNum)
			}
			val := binary.LittleEndian.Uint32(data[pos : pos+4])
			pos += 4
			// Could be float32 or uint32
			addToResult(result, key, math.Float32frombits(val))

		default:
			return nil, fmt.Errorf("unknown wire type %d for field %d", wireType, fieldNum)
		}
	}

	return result, nil
}

func addToResult(result map[string]interface{}, key string, value interface{}) {
	if existing, ok := result[key]; ok {
		// Convert to repeated field
		switch v := existing.(type) {
		case []interface{}:
			result[key] = append(v, value)
		default:
			result[key] = []interface{}{v, value}
		}
	} else {
		result[key] = value
	}
}

func isLikelyString(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	for _, b := range data {
		if b < 0x20 && b != '\n' && b != '\r' && b != '\t' {
			return false
		}
	}
	return true
}
