package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	// Start local server
	fmt.Println("Starting local server on port 9417...")
	err := startLocalServer(codeChan)
	if err != nil {
		return err
	}

	// Open the browser to CLI login endpoint (sets session flag, then redirects to GitHub OAuth)
	fmt.Println("Opening browser for GitHub authentication...")
	err = browser.OpenURL("http://localhost:8080/oauth2/cli-login")
	if err != nil {
		return err
	}

	// Race: Web POST vs Manual paste
	go func() {
		// prompt the user to paste the code manually
		fmt.Println("\nIf browser doesn't auto-submit, paste your code here:")
		var code string
		fmt.Scanln(&code)
		codeChan <- code
	}()

	// waits for whichever completes first
	fmt.Println("Waiting for OTP code...")
	otp := <-codeChan
	fmt.Println("✓ OTP received!")

	auth, err := LoginWithCode(otp)
	if err != nil {
		return err
	}

	cfg := config.Config{
		APIUrl:       "http://localhost:8080",
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

func startLocalServer(codeChan chan string) error {
	server := http.Server{
		Addr: "localhost:9417",
	}

	http.HandleFunc("/submit", handleSubmit(codeChan))
	go server.ListenAndServe()

	return nil
}

func handleSubmit(codeChan chan string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Printf("📥 Received %s request to /submit\n", req.Method)

		// Add CORS headers to allow browser requests
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		res.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS request
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
		fmt.Printf("📝 Received OTP: %s\n", otp)

		if otp == "" {
			fmt.Println("❌ Empty OTP code")
			http.Error(res, "Empty OTP code", http.StatusBadRequest)
			return
		}

		// Send to channel (this is how Login() receives it!)
		fmt.Println("✅ Sending OTP to channel...")
		codeChan <- otp

		res.Write([]byte("Success! You can close this window."))
	}
}

func LoginWithCode(otp string) (*AuthResponse, error) {
	reqBody := AuthRequest{Otp: otp}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// TODO: we need to make this dynamic (in prod this should be "api.95ninefive.dev")
	res, err := http.Post("http://localhost:8080/api/auth/otp/login", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

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

// RefreshToken exchanges a refresh token for a new access token
func RefreshToken(refreshToken string) (*AuthResponse, error) {
	reqBody := map[string]string{"refreshToken": refreshToken}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	res, err := http.Post("http://localhost:8080/api/auth/refresh", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call refresh endpoint: %w", err)
	}
	defer res.Body.Close()

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

func Logout(accessToken string) error {
	req, err := http.NewRequest("POST", "http://localhost:8080/api/auth/logout", nil)
	if err != nil {
		return fmt.Errorf("failed to create logout request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call logout endpoint: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode != 200 && res.StatusCode != 204 {
		return fmt.Errorf("logout failed: %d - %s", res.StatusCode, string(body))
	}

	return nil
}
