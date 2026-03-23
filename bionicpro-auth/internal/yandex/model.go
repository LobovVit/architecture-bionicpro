package yandex

type Profile struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DefaultEmail    string `json:"default_email"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	DisplayName     string `json:"display_name"`
	DefaultAvatarID string `json:"default_avatar_id"`
}
