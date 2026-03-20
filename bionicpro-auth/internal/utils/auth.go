package utils

import (
	"net/http"
	"strings"
)

func ExtractAccessToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		return ""
	}

	return parts[1]
}

func ExtractUserID(r *http.Request) string {
	return "user-id-from-token"
}
