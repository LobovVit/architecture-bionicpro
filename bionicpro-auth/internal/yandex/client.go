package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

func GetYandexToken(ctx context.Context, keycloakURL, realm, accessToken string) (string, error) {
	url := fmt.Sprintf("%s/realms/%s/broker/yandex/token", keycloakURL, realm)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.AccessToken, nil
}

func FetchProfile(token string) (*Profile, error) {
	req, err := http.NewRequest("GET", "https://login.yandex.ru/info", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "OAuth "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	return &profile, nil
}
