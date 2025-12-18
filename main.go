package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const totpPeriod = 30 // seconds

// Response represents the API response
type Response struct {
	Success   bool   `json:"success"`
	Code      string `json:"code,omitempty"`
	Remaining int    `json:"remaining,omitempty"` // seconds remaining
	Error     string `json:"error,omitempty"`
}

// normalizeSecret removes spaces and converts to uppercase
func normalizeSecret(secret string) string {
	secret = strings.ReplaceAll(secret, " ", "")
	return strings.ToUpper(secret)
}

// generateTOTP generates a TOTP code from a base32 secret
func generateTOTP(secret string) (string, int, error) {
	// Normalize the secret
	secret = normalizeSecret(secret)

	// Decode base32 secret
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", 0, fmt.Errorf("invalid secret key: %v", err)
	}

	// Calculate time counter
	now := time.Now().Unix()
	counter := uint64(now / totpPeriod)
	remaining := totpPeriod - int(now%totpPeriod)

	// Convert counter to bytes (big-endian)
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)

	// Generate HMAC-SHA1
	h := hmac.New(sha1.New, key)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	// Generate OTP
	otp := truncated % 1000000
	code := fmt.Sprintf("%06d", otp)

	return code, remaining, nil
}

// handleTOTP handles the /api/totp endpoint (GET and POST)
func handleTOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var secret string

	if r.Method == http.MethodPost {
		// POST: Get secret from JSON body
		var req struct {
			Secret string `json:"secret"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				Success: false,
				Error:   "invalid JSON body",
			})
			return
		}
		secret = req.Secret
	} else {
		// GET: Get secret from query parameter
		secret = r.URL.Query().Get("secret")
	}

	if secret == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   "missing secret parameter",
		})
		return
	}

	// Generate TOTP
	code, remaining, err := generateTOTP(secret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(Response{
		Success:   true,
		Code:      code,
		Remaining: remaining,
	})
}

// handleHealth handles health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleRobots serves robots.txt to block all bots
func handleRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive, nosnippet, noimageindex")
	w.Write([]byte("User-agent: *\nDisallow: /\n"))
}

func main() {
	// Use all available CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Create a new ServeMux for routing
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/api/totp", handleTOTP)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/robots.txt", handleRobots)

	// Server configuration for high performance
	server := &http.Server{
		Addr:           ":7842",
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	log.Printf("2FA API Server starting on :7842 (using %d CPU cores)", runtime.NumCPU())
	log.Println("Endpoints:")
	log.Println("  GET  /api/totp?secret=YOUR_SECRET")
	log.Println("  POST /api/totp {\"secret\":\"YOUR_SECRET\"}")
	log.Println("  GET  /health")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
