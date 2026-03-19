package main

import (
	"log"
	"net/http"

	"bionicpro-auth/internal/config"
	"bionicpro-auth/internal/cryptoenc"
	"bionicpro-auth/internal/httpapi"
	"bionicpro-auth/internal/oidc"
	"bionicpro-auth/internal/session"
)

func main() {
	cfg := config.Load()

	store := session.NewStore()
	crypto := cryptoenc.NewAESGCM(cfg.EncryptionKey)
	oidcClient := oidc.NewClient(cfg)

	srv := httpapi.NewServer(cfg, store, oidcClient, crypto)

	addr := ":" + cfg.Port
	log.Printf("bionicpro-auth listening on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      srv.Handler(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Fatal(server.ListenAndServe())
}
