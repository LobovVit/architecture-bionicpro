FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git tzdata

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/reports-service ./cmd/server

FROM alpine:3.20

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app && \
    apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/reports-service /app/reports-service

USER app

EXPOSE 8090

ENV PORT=8090

ENTRYPOINT ["/app/reports-service"]
