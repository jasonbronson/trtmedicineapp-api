package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/jasonbronson/go-gin-boilerplate/config"
)

type GoogleTokenInfo struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Audience      string `json:"aud"`
	ExpiresIn     string `json:"expires_in"`
}

func VerifyGoogleIDToken(idToken string) (*GoogleTokenInfo, error) {
	if config.Cfg.GoogleClientID == "" {
		return nil, errors.New("GOOGLE_CLIENT_ID is required")
	}

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid google token")
	}

	var tokenInfo GoogleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, err
	}
	if tokenInfo.Audience != config.Cfg.GoogleClientID {
		return nil, errors.New("google token audience mismatch")
	}
	if tokenInfo.Subject == "" || tokenInfo.Email == "" || tokenInfo.EmailVerified != "true" {
		return nil, errors.New("google token is missing verified email")
	}
	return &tokenInfo, nil
}
