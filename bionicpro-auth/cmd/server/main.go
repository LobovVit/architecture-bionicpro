package main

import (
	"log"
	"net/http"

	"bionicpro-auth/internal/config"
	"bionicpro-auth/internal/cryptoenc"
	appdb "bionicpro-auth/internal/db"
	"bionicpro-auth/internal/httpapi"
	"bionicpro-auth/internal/oidc"
	"bionicpro-auth/internal/session"
)

func main() {
	cfg := config.Load()

	profileDB, err := appdb.Open(cfg.ProfileDBDSN)
	if err != nil {
		log.Fatalf("open profile db: %v", err)
	}
	defer profileDB.Close()
	if err := appdb.Migrate(profileDB); err != nil {
		log.Fatalf("migrate profile db: %v", err)
	}

	store := session.NewStore()
	crypto := cryptoenc.NewAESGCM(cfg.EncryptionKey)
	oidcClient := oidc.NewClient(cfg)

	srv := httpapi.NewServer(cfg, store, oidcClient, crypto, profileDB)

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
