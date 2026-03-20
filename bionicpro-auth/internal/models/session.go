package models

import "time"

type Session struct {
	SessionID             string    `json:"session_id"`
	UserID                string    `json:"user_id"`
	Username              string    `json:"username"`
	Roles                 []string  `json:"roles"`
	AccessToken           string    `json:"-"`
	EncryptedRefreshToken string    `json:"-"`
	AccessTokenExpiresAt  time.Time `json:"-"`
	RefreshTokenExpiresAt time.Time `json:"-"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	IDToken               string    `json:"-"`
}

type PendingAuth struct {
	State        string
	Nonce        string
	CodeVerifier string
	CreatedAt    time.Time
}
