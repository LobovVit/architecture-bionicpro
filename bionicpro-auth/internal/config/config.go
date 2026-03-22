package config

import (
	"encoding/base64"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port                string
	BaseURL             string
	FrontendURL         string
	APIBaseURL          string
	KeycloakInternalURL string
	KeycloakPublicURL   string
	KeycloakRealm       string
	ClientID            string
	ClientSecret        string
	RedirectURI         string
	CookieName          string
	CookieSecure        bool
	CookieDomain        string
	SessionTTL          time.Duration
	AccessTokenLeeway   time.Duration
	EncryptionKey       []byte
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	ProfileDBDSN        string
	KeycloakBrokerAlias string
	YandexUserInfoURL   string
	ReportsServiceURL   string
}

func Load() Config {
	keyB64 := getEnv("SESSION_ENCRYPTION_KEY_B64", "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=") // demo only
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil || len(key) != 32 {
		log.Fatalf("SESSION_ENCRYPTION_KEY_B64 must be base64-encoded 32-byte key")
	}

	return Config{
		Port:                getEnv("PORT", "8081"),
		BaseURL:             getEnv("BASE_URL", "http://localhost:8081"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:3000"),
		APIBaseURL:          getEnv("API_BASE_URL", "http://backend:8082"),
		KeycloakInternalURL: getEnv("KEYCLOAK_INTERNAL_URL", "http://keycloak:8080"),
		KeycloakPublicURL:   getEnv("KEYCLOAK_PUBLIC_URL", "http://localhost:8080"),
		KeycloakRealm:       getEnv("KEYCLOAK_REALM", "reports-realm"),
		ClientID:            getEnv("CLIENT_ID", "bionicpro-auth"),
		ClientSecret:        getEnv("CLIENT_SECRET", "change-me"),
		RedirectURI:         getEnv("REDIRECT_URI", "http://localhost:8081/auth/callback"),
		CookieName:          getEnv("COOKIE_NAME", "bp_session"),
		CookieSecure:        strings.EqualFold(getEnv("COOKIE_SECURE", "false"), "true"),
		CookieDomain:        getEnv("COOKIE_DOMAIN", ""),
		SessionTTL:          getDurationEnv("SESSION_TTL", 30*time.Minute),
		AccessTokenLeeway:   getDurationEnv("ACCESS_TOKEN_LEEWAY", 10*time.Second),
		EncryptionKey:       key,
		ReadTimeout:         getDurationEnv("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:        getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:         getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		ProfileDBDSN:        getEnv("PROFILE_DB_DSN", "postgres://bionicpro:bionicpro@profile_db:5432/bionicpro?sslmode=disable"),
		KeycloakBrokerAlias: getEnv("KEYCLOAK_BROKER_ALIAS", "yandex"),
		YandexUserInfoURL:   getEnv("YANDEX_USERINFO_URL", "https://login.yandex.ru/info"),
		ReportsServiceURL:   getEnv("REPORTS_SERVICE_URL", "http://localhost:8090"),
	}
}

func (c Config) PublicRealmURL() string {
	return strings.TrimRight(c.KeycloakPublicURL, "/") + "/realms/" + c.KeycloakRealm
}

func (c Config) InternalRealmURL() string {
	return strings.TrimRight(c.KeycloakInternalURL, "/") + "/realms/" + c.KeycloakRealm
}

func (c Config) AuthURL() string {
	return c.PublicRealmURL() + "/protocol/openid-connect/auth"
}

func (c Config) TokenURL() string {
	return c.InternalRealmURL() + "/protocol/openid-connect/token"
}

func (c Config) LogoutURL() string {
	return c.PublicRealmURL() + "/protocol/openid-connect/logout"
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Fatalf("%s invalid duration: %v", key, err)
	}
	return d
}
