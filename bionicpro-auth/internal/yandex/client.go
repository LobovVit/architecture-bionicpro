package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 15 * time.Second}}
}

func (c *Client) GetBrokerToken(ctx context.Context, keycloakBaseURL, realm, providerAlias, accessToken string) (string, error) {
	url := fmt.Sprintf("%s/realms/%s/broker/%s/token", strings.TrimRight(keycloakBaseURL, "/"), realm, providerAlias)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("broker token endpoint %d: %s", resp.StatusCode, string(body))
	}

	var data TokenResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	if data.AccessToken == "" {
		return "", fmt.Errorf("empty yandex access token returned by broker")
	}
	return data.AccessToken, nil
}

func (c *Client) FetchProfile(ctx context.Context, token string) (*Profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://login.yandex.ru/info", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "OAuth "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yandex info endpoint %d: %s", resp.StatusCode, string(body))
	}

	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}
