package client

import "net/http"

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func Login() error {
	return nil
}

func startLocalServer(res *AuthResponse) error {
	server := http.Server{
		Addr: "localhost:9417",
	}
	server.ListenAndServe()

	return nil
}
