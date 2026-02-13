// gauth — Standalone Google auth tool with device check-in, WebView login, and token server.
//
// Usage:
//
//	gauth login        Open WebView to sign in with Google
//	gauth token        Show stored master token and account info
//	gauth fetch <scope> Fetch a service token (photos, youtube, gmail, drive, or custom scope)
//	gauth checkin      Force device check-in (get new GSF ID)
//	gauth serve [port] Start HTTP token server (default: 8080)
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/nicksrandall/gauth/internal/auth"
	"github.com/nicksrandall/gauth/internal/checkin"
	"github.com/nicksrandall/gauth/internal/config"
	"github.com/nicksrandall/gauth/internal/login"
	"github.com/nicksrandall/gauth/internal/server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.Load()
	cmd := os.Args[1]

	switch cmd {
	case "login":
		cmdLogin(cfg)
	case "token":
		cmdToken(cfg)
	case "fetch":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gauth fetch <scope>")
			fmt.Println("Examples: gauth fetch photos")
			fmt.Println("          gauth fetch \"oauth2:https://www.googleapis.com/auth/youtube\"")
			os.Exit(1)
		}
		cmdFetch(cfg, os.Args[2])
	case "checkin":
		cmdCheckin(cfg)
	case "serve":
		port := 8080
		if len(os.Args) >= 3 {
			p, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Printf("Invalid port: %s\n", os.Args[2])
				os.Exit(1)
			}
			port = p
		}
		cmdServe(cfg, port)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func cmdLogin(cfg *config.Config) {
	// Step 1: Check-in (if not done)
	if !cfg.HasRegistration() {
		fmt.Println("📡 Step 1/3: Device check-in...")
		if err := doCheckin(cfg); err != nil {
			log.Fatalf("Check-in failed: %v", err)
		}
		fmt.Printf("   ✅ GSF ID: %s\n", cfg.AndroidID)
	} else {
		fmt.Printf("📡 Device already registered (GSF ID: %s)\n", cfg.AndroidID)
	}

	// Step 2: WebView login
	fmt.Println("🌐 Step 2/3: Opening Google sign-in...")
	result, err := login.RunWebViewLogin(cfg)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Printf("   ✅ OAuth token received (length: %d)\n", len(result.OAuthToken))

	// Step 3: Exchange for master token
	fmt.Println("🔑 Step 3/3: Exchanging for master token...")
	resp, err := auth.ExchangeOAuthForMaster(cfg, result.OAuthToken)
	if err != nil {
		log.Fatalf("Token exchange failed: %v", err)
	}

	if resp.Token == "" && resp.Auth == "" {
		log.Fatalf("No token in response. Error: %s", resp.Error)
	}

	// Save master token
	masterToken := resp.Token
	if masterToken == "" {
		masterToken = resp.Auth
	}
	cfg.MasterToken = masterToken
	cfg.Email = resp.Email

	if err := cfg.Save(); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ Login successful!")
	fmt.Printf("   Email: %s\n", cfg.Email)
	fmt.Printf("   Master Token: %s...%s\n", masterToken[:10], masterToken[len(masterToken)-5:])
	fmt.Println()
	fmt.Println("You can now use:")
	fmt.Println("  gauth fetch photos    — Get Google Photos token")
	fmt.Println("  gauth serve           — Start token server")
}

func cmdToken(cfg *config.Config) {
	if !cfg.HasMasterToken() {
		fmt.Println("❌ Not logged in. Run: gauth login")
		os.Exit(1)
	}
	fmt.Printf("Email:        %s\n", cfg.Email)
	fmt.Printf("Android ID:   %s\n", cfg.AndroidID)
	fmt.Printf("Master Token: %s\n", cfg.MasterToken)
	fmt.Printf("Device:       %s (%s)\n", cfg.Device.Model, cfg.Device.Fingerprint)
}

func cmdFetch(cfg *config.Config, scope string) {
	if !cfg.HasMasterToken() {
		fmt.Println("❌ Not logged in. Run: gauth login")
		os.Exit(1)
	}

	// Resolve shortcut
	appPkg := "com.google.android.gms"
	appSig := auth.GoogleSig
	if app, ok := auth.KnownApps[scope]; ok {
		fmt.Printf("📱 Fetching token for %s (%s)...\n", scope, app.Package)
		appPkg = app.Package
		scope = app.Scope
	} else {
		fmt.Printf("📱 Fetching token for scope: %s...\n", scope)
	}

	resp, err := auth.FetchServiceToken(cfg, scope, appPkg, appSig)
	if err != nil {
		log.Fatalf("Token fetch failed: %v", err)
	}

	if resp.Auth == "" {
		log.Fatalf("Empty token response. Error: %s", resp.Error)
	}

	tokenType := "Unknown"
	if len(resp.Auth) > 7 && resp.Auth[:7] == "aas_et/" {
		tokenType = "AES (1P)"
	} else if len(resp.Auth) > 5 && resp.Auth[:5] == "ya29." {
		tokenType = "OAuth2"
	}

	fmt.Printf("✅ Token (%s):\n%s\n", tokenType, resp.Auth)
}

func cmdCheckin(cfg *config.Config) {
	fmt.Println("📡 Performing device check-in...")
	if err := doCheckin(cfg); err != nil {
		log.Fatalf("Check-in failed: %v", err)
	}
	fmt.Printf("✅ GSF ID:         %s\n", cfg.AndroidID)
	fmt.Printf("   Security Token: %s\n", cfg.SecurityToken)
}

func cmdServe(cfg *config.Config, port int) {
	if !cfg.HasMasterToken() {
		fmt.Println("⚠️  Warning: Not logged in. Token endpoints will fail until you run: gauth login")
	}
	if err := server.Start(cfg, port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func doCheckin(cfg *config.Config) error {
	result, err := checkin.Checkin(cfg)
	if err != nil {
		return err
	}
	cfg.AndroidID = fmt.Sprintf("%x", result.AndroidID)
	cfg.SecurityToken = fmt.Sprintf("%d", result.SecurityToken)
	return cfg.Save()
}

func printUsage() {
	fmt.Println("gauth — Google Auth Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gauth login          Sign in with Google (opens WebView)")
	fmt.Println("  gauth token          Show stored account info and master token")
	fmt.Println("  gauth fetch <scope>  Fetch a service token")
	fmt.Println("  gauth checkin        Force device check-in")
	fmt.Println("  gauth serve [port]   Start HTTP token server (default: 8080)")
	fmt.Println()
	fmt.Println("Scope shortcuts: photos, youtube, gmail, drive, calendar")
	fmt.Println("Custom scope:    gauth fetch \"oauth2:https://...\"")
}
