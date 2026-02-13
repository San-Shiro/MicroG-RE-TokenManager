// Package login provides WebView-based Google login with full Android spoofing.
// Uses the webview/webview_go library (WebView2 on Windows, WebKitGTK on Linux).
package login

import (
	"fmt"
	"strings"

	"github.com/nicksrandall/gauth/internal/config"
)

// BuildLoginURL constructs the EmbeddedSetup URL with proper parameters.
func BuildLoginURL() string {
	return "https://accounts.google.com/EmbeddedSetup" +
		"?source=android" +
		"&xoauth_display_name=Android+Device" +
		"&lang=en" +
		"&cc=us" +
		"&langCountry=en_us" +
		"&hl=en-US" +
		"&tmpl=new_account"
}

// BuildJSBridge returns the JavaScript that creates the complete `window.mm` bridge.
// This must be injected BEFORE any page scripts run (via w.Init()).
// All method signatures match MicroG's LoginActivity.JsBridge exactly.
func BuildJSBridge(cfg *config.Config) string {
	androidIDHex := cfg.AndroidID
	gmsVersion := 224714044
	sdkVersion := cfg.Device.SDKVersion

	return fmt.Sprintf(`
// ====== NAVIGATOR SPOOFING ======
// Override browser fingerprint to look like Android WebView
Object.defineProperty(navigator, 'userAgent', {
    get: function() { return %q; }
});
Object.defineProperty(navigator, 'platform', {
    get: function() { return 'Linux armv8l'; }
});
Object.defineProperty(navigator, 'appVersion', {
    get: function() { return '5.0 (Linux; Android %d; %s Build/%s; wv) AppleWebKit/537.36'; }
});
Object.defineProperty(navigator, 'vendor', {
    get: function() { return 'Google Inc.'; }
});
// Fake touch support (Android device)
Object.defineProperty(navigator, 'maxTouchPoints', {
    get: function() { return 5; }
});

// ====== GOOGLE GMS JAVASCRIPT BRIDGE ======
// Google's EmbeddedSetup page calls these methods to verify it's running
// inside a real Google Play Services WebView. Every method must exist.
window.mm = {
    // === CRITICAL: Device identity ===
    getAndroidId: function() {
        console.log('[gauth] mm.getAndroidId called');
        return '%s';
    },

    getBuildVersionSdk: function() { return %d; },

    getPlayServicesVersionCode: function() { return %d; },

    getAuthModuleVersionCode: function() { return %d; },

    // === Account management ===
    getAccounts: function() {
        console.log('[gauth] mm.getAccounts called');
        return '[]';
    },

    getAllowedDomains: function() { return '[]'; },

    getDeviceDataVersionInfo: function() { return 1; },

    getDeviceContactsCount: function() { return -1; },

    getFactoryResetChallenges: function() { return '[]'; },

    // === Phone/SIM info (we have none) ===
    getPhoneNumber: function() { return null; },
    getSimSerial: function() { return null; },
    getSimState: function() { return 1; },
    hasPhoneNumber: function() { return false; },
    hasTelephony: function() { return false; },
    fetchVerifiedPhoneNumber: function() { return null; },

    // === User info ===
    isUserOwner: function() { return true; },

    // === UI control (must exist, most are no-ops) ===
    showView: function() {
        console.log('[gauth] mm.showView called');
        document.title = 'GAUTH_SHOW';
    },

    closeView: function() {
        console.log('[gauth] mm.closeView called');
        // Extract oauth_token from cookies and signal to Go
        var cookies = document.cookie;
        document.title = 'GAUTH_CLOSE:' + cookies;
    },

    hideKeyboard: function() {},
    showKeyboard: function() {},

    setBackButtonEnabled: function(b) {},
    setPrimaryActionEnabled: function(b) {},
    setPrimaryActionLabel: function(str, i) {},
    setSecondaryActionEnabled: function(b) {},
    setSecondaryActionLabel: function(str, i) {},
    setAllActionsEnabled: function(b) {},

    setAccountIdentifier: function(name) {
        console.log('[gauth] mm.setAccountIdentifier: ' + name);
    },

    setNewAccountCreated: function() {
        console.log('[gauth] mm.setNewAccountCreated');
    },

    // === Auth flow stubs ===
    addAccount: function(json) {
        console.log('[gauth] mm.addAccount');
    },

    attemptLogin: function(name, pass) {
        console.log('[gauth] mm.attemptLogin');
    },

    skipLogin: function() {
        console.log('[gauth] mm.skipLogin');
        document.title = 'GAUTH_CANCEL';
    },

    goBack: function() {
        console.log('[gauth] mm.goBack');
    },

    log: function(s) {
        console.log('[gauth] mm.log: ' + s);
    },

    // === Misc stubs ===
    backupSyncOptIn: function(name) {},
    clearOldLoginAttempts: function() {},
    notifyOnTermsOfServiceAccepted: function() {},
    fetchIIDToken: function(entity) {},
    startAfw: function() {},
    launchEmergencyDialer: function() {}
};

console.log('[gauth] Android WebView bridge injected successfully');
`,
		cfg.UserAgent(),
		sdkVersion, cfg.Device.Model, cfg.Device.BuildID,
		androidIDHex,
		sdkVersion,
		gmsVersion,
		gmsVersion,
	)
}

// ExtractOAuthToken parses the oauth_token from a cookie string.
func ExtractOAuthToken(cookies string) string {
	for _, cookie := range strings.Split(cookies, ";") {
		cookie = strings.TrimSpace(cookie)
		if strings.HasPrefix(cookie, "oauth_token=") {
			parts := strings.SplitN(cookie, "=", 2)
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}
	return ""
}

// IsCloseSignal checks if a URL/title indicates login is complete.
func IsCloseSignal(urlOrTitle string) bool {
	return strings.Contains(urlOrTitle, "#close") ||
		strings.HasPrefix(urlOrTitle, "GAUTH_CLOSE:") ||
		strings.Contains(urlOrTitle, "accounts.google.com/signin/continue")
}
