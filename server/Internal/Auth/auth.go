package auth

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func RefreshUserToken(ctx context.Context, refresh_token string, refresh_token_endpoint, clientID, clientSecret string) {
	log.Println(clientID)
	log.Println(clientSecret)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refresh_token)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, refresh_token_endpoint, strings.NewReader(data.Encode()))

	if err != nil {
		// return "", fmt.Errorf("Failed to create request: %w", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// cookie := &http.Cookie{
	// 	Name:  "refresh_token",
	// 	Value: refresh_token,
	// }
	// req.AddCookie(cookie)
	// req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		// return "", fmt.Errorf("Failed to request token refresh: %w", err)
		return
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	log.Println("response status: ", res.Status)
	log.Println("Data: ", string(body))

	// return string(body)
}
