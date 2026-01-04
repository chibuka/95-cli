package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/chibuka/95-cli/internal/config"
	"github.com/pkg/browser"
)

type AuthRequest struct {
	Otp string `json:"otp"`
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	UserId       int    `json:"userId"`
	Username     string `json:"username"`
	Email        string `json:"email"`
}

func Login() error {
	codeChan := make(chan string)

	// Get API URL from environment or use default
	apiURL := getAPIURL()

	// Start local server
	fmt.Println("Starting local server on port 9417...")
	err := startLocalServer(codeChan, apiURL)
	if err != nil {
		return err
	}

	// Open the browser to CLI login endpoint
	fmt.Printf("Opening browser for GitHub authentication at %s...\n", apiURL)
	err = browser.OpenURL(fmt.Sprintf("%s/oauth2/cli-login", apiURL))
	if err != nil {
		return err
	}

	// Race: Web POST vs Manual paste
	go func() {
		fmt.Println("\nIf browser doesn't auto-submit, paste your code here:")
		var code string
		_, err := fmt.Scanln(&code)
		if err != nil {
			// If scan fails (e.g. EOF), we just return from the goroutine
			return
		}
		codeChan <- code
	}()

	fmt.Println("Waiting for OTP code...")
	otp := <-codeChan
	fmt.Println("✓ OTP received!")

	auth, err := LoginWithCode(otp, apiURL)
	if err != nil {
		return err
	}

	cfg := config.Config{
		APIUrl:       apiURL,
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		UserId:       auth.UserId,
		Username:     auth.Username,
	}

	err = cfg.Save()
	if err != nil {
		return err
	}
	return nil
}

func getAPIURL() string {
	if apiURL := os.Getenv("API_URL"); apiURL != "" {
		return apiURL
	}
	if os.Getenv("DEV_MODE") == "true" {
		return config.LocalAPIURL
	}
	return config.DefaultAPIURL
}

func startLocalServer(codeChan chan string, apiURL string) error {
	server := http.Server{
		Addr: "localhost:9417",
	}

	http.HandleFunc("/submit", handleSubmit(codeChan, apiURL))

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Local server error: %v", err)
		}
	}()

	return nil
}

func handleSubmit(codeChan chan string, apiURL string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Printf("📥 Received %s request to /submit\n", req.Method)

		res.Header().Set("Access-Control-Allow-Origin", apiURL)
		res.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		res.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if req.Method == "OPTIONS" {
			res.WriteHeader(http.StatusOK)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Println("❌ Error reading request body:", err)
			http.Error(res, "Couldn't read request body", http.StatusInternalServerError)
			return
		}

		otp := strings.TrimSpace(string(body))
		if otp == "" {
			fmt.Println("❌ Empty OTP code")
			http.Error(res, "Empty OTP code", http.StatusBadRequest)
			return
		}

		fmt.Println("✅ Sending OTP to channel...")
		codeChan <- otp

		_, _ = res.Write([]byte("Success! You can close this window."))
	}
}

func LoginWithCode(otp string, apiURL string) (*AuthResponse, error) {
	reqBody := AuthRequest{Otp: otp}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	loginURL := fmt.Sprintf("%s/api/auth/otp/login", apiURL)
	res, err := http.Post(loginURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("login failed: %d - %s", res.StatusCode, body)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var authResponse AuthResponse
	if err := json.Unmarshal(body, &authResponse); err != nil {
		return nil, err
	}

	return &authResponse, nil
}

func RefreshToken(refreshToken string, apiURL string) (*AuthResponse, error) {
	reqBody := map[string]string{"refreshToken": refreshToken}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	refreshURL := fmt.Sprintf("%s/api/auth/refresh", apiURL)
	res, err := http.Post(refreshURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call refresh endpoint: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: %d - %s", res.StatusCode, string(body))
	}

	var authResponse AuthResponse
	if err := json.Unmarshal(body, &authResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refresh response: %w", err)
	}

	return &authResponse, nil
}

func Logout(accessToken string, apiURL string) error {
	logoutURL := fmt.Sprintf("%s/api/auth/logout", apiURL)
	req, err := http.NewRequest("POST", logoutURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create logout request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call logout endpoint: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode != 200 && res.StatusCode != 204 {
		return fmt.Errorf("logout failed: %d - %s", res.StatusCode, string(body))
	}

	return nil
}
