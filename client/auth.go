package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	err := startLocalServer(codeChan)
	if err != nil {
		return err
	}

	// Open the browser
	err = browser.OpenURL("http://localhost:8080/oauth2/authorization/github")
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
	otp := <-codeChan

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
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Couldn't read request body", http.StatusInternalServerError)
			return
		}

		otp := string(body)
		if otp == "" {
			http.Error(res, "Empty OTP code", http.StatusBadRequest)
			return
		}

		// Send to channel (this is how Login() receives it!)
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
