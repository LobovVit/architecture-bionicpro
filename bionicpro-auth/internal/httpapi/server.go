package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"bionicpro-auth/internal/config"
	"bionicpro-auth/internal/cryptoenc"
	"bionicpro-auth/internal/models"
	"bionicpro-auth/internal/oidc"
	"bionicpro-auth/internal/session"
)

type Server struct {
	cfg       config.Config
	store     *session.Store
	oidc      *oidc.Client
	crypto    *cryptoenc.AESGCM
	mux       *http.ServeMux
	cookieOpt CookieOptions
}

type CookieOptions struct {
	Name     string
	Secure   bool
	Domain   string
	SameSite http.SameSite
}

func NewServer(cfg config.Config, store *session.Store, oidcClient *oidc.Client, crypto *cryptoenc.AESGCM) *Server {
	s := &Server{
		cfg:    cfg,
		store:  store,
		oidc:   oidcClient,
		crypto: crypto,
		cookieOpt: CookieOptions{
			Name:     cfg.CookieName,
			Secure:   cfg.CookieSecure,
			Domain:   cfg.CookieDomain,
			SameSite: http.SameSiteLaxMode,
		},
		mux: http.NewServeMux(),
	}

	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return withCORS(s.mux, s.cfg.FrontendURL)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/auth/login", s.handleLogin)
	s.mux.HandleFunc("/auth/callback", s.handleCallback)
	s.mux.HandleFunc("/auth/me", s.handleMe)
	s.mux.HandleFunc("/auth/logout", s.handleLogout)
	s.mux.HandleFunc("/api/reports", s.handleReports)
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	state := randString(32)
	nonce := randString(24)
	codeVerifier := randString(64)
	codeChallenge := pkceChallenge(codeVerifier)

	s.store.SavePending(models.PendingAuth{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		CreatedAt:    time.Now().UTC(),
	})

	authURL, err := url.Parse(oidc.AuthURL(s.cfg))
	if err != nil {
		http.Error(w, "failed to build auth url", http.StatusInternalServerError)
		return
	}

	q := authURL.Query()
	q.Set("client_id", s.cfg.ClientID)
	q.Set("redirect_uri", s.cfg.RedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	authURL.RawQuery = q.Encode()

	http.Redirect(w, r, authURL.String(), http.StatusFound)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		http.Error(w, "missing code or state", http.StatusBadRequest)
		return
	}

	pending, err := s.store.GetPending(state)
	if err != nil {
		http.Error(w, "invalid state", http.StatusUnauthorized)
		return
	}
	s.store.DeletePending(state)

	tokenResp, err := s.oidc.ExchangeCode(code, pending.CodeVerifier)
	if err != nil {
		http.Error(w, "code exchange failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	claims, err := oidc.ParseClaimsUnverified(tokenResp.AccessToken)
	if err != nil {
		http.Error(w, "failed to parse token claims", http.StatusUnauthorized)
		return
	}

	userID, _ := claims["sub"].(string)
	username, _ := claims["preferred_username"].(string)
	if username == "" {
		username, _ = claims["email"].(string)
	}

	roles := extractRoles(claims)

	encRefresh, err := s.crypto.Encrypt(tokenResp.RefreshToken)
	if err != nil {
		http.Error(w, "failed to encrypt refresh token", http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	sessionID := randString(32)

	s.store.SaveSession(models.Session{
		SessionID:             sessionID,
		UserID:                userID,
		Username:              username,
		Roles:                 roles,
		AccessToken:           tokenResp.AccessToken,
		EncryptedRefreshToken: encRefresh,
		AccessTokenExpiresAt:  oidc.AccessExpiry(now, tokenResp),
		RefreshTokenExpiresAt: oidc.RefreshExpiry(now, tokenResp),
		CreatedAt:             now,
		UpdatedAt:             now,
	})

	s.setSessionCookie(w, sessionID)
	http.Redirect(w, r, s.cfg.FrontendURL, http.StatusFound)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	sess, newSessionID, ok := s.authenticateAndMaybeRotate(w, r)
	if !ok {
		return
	}

	resp := map[string]any{
		"authenticated": true,
		"username":      sess.Username,
		"roles":         sess.Roles,
		"session_id":    newSessionID,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if c, err := r.Cookie(s.cookieOpt.Name); err == nil {
		s.store.DeleteSession(c.Value)
	}
	s.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleReports(w http.ResponseWriter, r *http.Request) {
	sess, newSessionID, ok := s.authenticateAndMaybeRotate(w, r)
	if !ok {
		return
	}

	resp := map[string]any{
		"owner":            sess.Username,
		"report_generated": time.Now().UTC().Format(time.RFC3339),
		"prosthesis_ids":   []string{"prosthesis-001"},
		"summary": map[string]any{
			"daily_movements":   1842,
			"battery_cycles":    3,
			"calibration_state": "OK",
		},
		"security": map[string]any{
			"token_on_client": false,
			"session_rotated": true,
			"new_session_id":  newSessionID,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) authenticateAndMaybeRotate(w http.ResponseWriter, r *http.Request) (models.Session, string, bool) {
	c, err := r.Cookie(s.cookieOpt.Name)
	if err != nil || c.Value == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return models.Session{}, "", false
	}

	sess, err := s.store.GetSession(c.Value)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return models.Session{}, "", false
	}

	now := time.Now().UTC()

	if now.After(sess.RefreshTokenExpiresAt) {
		s.store.DeleteSession(sess.SessionID)
		s.clearSessionCookie(w)
		http.Error(w, "session expired", http.StatusUnauthorized)
		return models.Session{}, "", false
	}

	if now.After(sess.AccessTokenExpiresAt.Add(-s.cfg.AccessTokenLeeway)) {
		refreshToken, err := s.crypto.Decrypt(sess.EncryptedRefreshToken)
		if err != nil {
			log.Printf("decrypt refresh token failed: %v", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return models.Session{}, "", false
		}

		tokenResp, err := s.oidc.Refresh(refreshToken)
		if err != nil {
			log.Printf("refresh failed: %v", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return models.Session{}, "", false
		}

		encRefresh, err := s.crypto.Encrypt(tokenResp.RefreshToken)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return models.Session{}, "", false
		}

		sess.AccessToken = tokenResp.AccessToken
		sess.EncryptedRefreshToken = encRefresh
		sess.AccessTokenExpiresAt = oidc.AccessExpiry(now, tokenResp)
		sess.RefreshTokenExpiresAt = oidc.RefreshExpiry(now, tokenResp)
		sess.UpdatedAt = now
		s.store.SaveSession(sess)
	}

	newSessionID := randString(32)
	rotated, err := s.store.RotateSession(sess.SessionID, newSessionID)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return models.Session{}, "", false
	}
	s.setSessionCookie(w, newSessionID)

	return rotated, newSessionID, true
}

func (s *Server) setSessionCookie(w http.ResponseWriter, sessionID string) {
	c := &http.Cookie{
		Name:     s.cookieOpt.Name,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cookieOpt.Secure,
		SameSite: s.cookieOpt.SameSite,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
	}
	if s.cookieOpt.Domain != "" {
		c.Domain = s.cookieOpt.Domain
	}
	http.SetCookie(w, c)
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:     s.cookieOpt.Name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.cookieOpt.Secure,
		SameSite: s.cookieOpt.SameSite,
		MaxAge:   -1,
	}
	if s.cookieOpt.Domain != "" {
		c.Domain = s.cookieOpt.Domain
	}
	http.SetCookie(w, c)
}

func extractRoles(claims map[string]any) []string {
	var roles []string

	realmAccess, ok := claims["realm_access"].(map[string]any)
	if ok {
		rawRoles, ok := realmAccess["roles"].([]any)
		if ok {
			for _, r := range rawRoles {
				if rs, ok := r.(string); ok {
					roles = append(roles, rs)
				}
			}
		}
	}

	return roles
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func withCORS(next http.Handler, frontendURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if frontendURL != "" {
			w.Header().Set("Access-Control-Allow-Origin", frontendURL)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func randString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
