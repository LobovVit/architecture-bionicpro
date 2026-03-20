package yandex

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"your_project/internal/utils"
)

type Handler struct {
	DB          *sql.DB
	KeycloakURL string
	Realm       string
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	accessToken := utils.ExtractAccessToken(r)
	if accessToken == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	yToken, err := GetYandexToken(r.Context(), h.KeycloakURL, h.Realm, accessToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	profile, err := FetchProfile(yToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(profile)
}

func (h *Handler) Consent(w http.ResponseWriter, r *http.Request) {
	accessToken := utils.ExtractAccessToken(r)
	if accessToken == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	yToken, err := GetYandexToken(r.Context(), h.KeycloakURL, h.Realm, accessToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	profile, err := FetchProfile(yToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	userID := utils.ExtractUserID(r)

	raw, _ := json.Marshal(profile)

	_, err = h.DB.Exec(`
		INSERT INTO yandex_profiles 
		(user_id, external_id, login, email, first_name, last_name, raw_profile, consent_granted, consent_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,true,now())
		ON CONFLICT (external_id) DO UPDATE SET
			login = EXCLUDED.login,
			email = EXCLUDED.email,
			raw_profile = EXCLUDED.raw_profile,
			consent_granted = true,
			consent_at = now()
	`,
		userID,
		profile.ID,
		profile.Login,
		profile.DefaultEmail,
		profile.FirstName,
		profile.LastName,
		raw,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}
