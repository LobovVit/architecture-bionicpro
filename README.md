# bionicpro-auth package layout

- `bionicpro-auth/cmd/server/main.go` — application entrypoint
- `bionicpro-auth/internal/config` — env config
- `bionicpro-auth/internal/cryptoenc` — AES-GCM encryption for refresh token
- `bionicpro-auth/internal/models` — session/state structs
- `bionicpro-auth/internal/oidc` — token exchange and refresh with Keycloak
- `bionicpro-auth/internal/session` — in-memory session/state store and rotation
- `bionicpro-auth/internal/httpapi` — HTTP handlers, PKCE, cookie session flow

`Dockerfile` is inside `bionicpro-auth/`.
`docker-compose.yaml` is at project root.
