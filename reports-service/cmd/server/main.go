package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Port          string
	ClickHouseDSN string
	PipelineName  string

	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3Region    string
	S3UseSSL    bool
	CDNBaseURL  string
}

type Server struct {
	cfg Config
	db  *sql.DB
	s3  *minio.Client
	mux *http.ServeMux
}

type ReportItem struct {
	ReportDate           string  `json:"report_date"`
	ProsthesisID         string  `json:"prosthesis_id"`
	ProsthesisModel      string  `json:"prosthesis_model"`
	CalibrationState     string  `json:"calibration_state"`
	DailyMovements       uint32  `json:"daily_movements"`
	AvgLoad              float64 `json:"avg_load"`
	BatteryCycles        uint32  `json:"battery_cycles"`
	AvgBatteryLevel      float64 `json:"avg_battery_level"`
	AlertsCount          uint32  `json:"alerts_count"`
	TelemetryEventsCount uint32  `json:"telemetry_events_count"`
}

type CachedReportResponse struct {
	UserID      string       `json:"user_id"`
	DateFrom    string       `json:"date_from"`
	DateTo      string       `json:"date_to"`
	LoadedUntil string       `json:"loaded_until"`
	GeneratedAt string       `json:"generated_at"`
	Items       []ReportItem `json:"items"`
}

type CDNResponse struct {
	Status      string `json:"status"`
	CDNURL      string `json:"cdn_url"`
	ObjectKey   string `json:"object_key"`
	LoadedUntil string `json:"loaded_until"`
}

func main() {
	cfg := Config{
		Port:          getEnv("PORT", "8090"),
		ClickHouseDSN: getEnv("CLICKHOUSE_DSN", "clickhouse://report_user:report_pass@clickhouse:9000/report_mart"),
		PipelineName:  getEnv("ETL_PIPELINE_NAME", "build_user_reports_mart"),
		S3Endpoint:    getEnv("S3_ENDPOINT", "minio:9000"),
		S3AccessKey:   getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:   getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:      getEnv("S3_BUCKET", "reports"),
		S3Region:      getEnv("S3_REGION", "us-east-1"),
		S3UseSSL:      strings.EqualFold(getEnv("S3_USE_SSL", "false"), "true"),
		CDNBaseURL:    strings.TrimRight(getEnv("CDN_BASE_URL", "http://localhost:8088"), "/"),
	}

	db, err := sql.Open("clickhouse", cfg.ClickHouseDSN)
	if err != nil {
		log.Fatalf("open clickhouse: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping clickhouse: %v", err)
	}

	s3Client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
		Region: cfg.S3Region,
	})
	if err != nil {
		log.Fatalf("init s3 client: %v", err)
	}

	s := &Server{
		cfg: cfg,
		db:  db,
		s3:  s3Client,
		mux: http.NewServeMux(),
	}
	s.routes()

	log.Printf("reports-service listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, s.mux))
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	s.mux.HandleFunc("/reports/me", s.handleReportsMeCached)
}

func (s *Server) handleReportsMeCached(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if userID == "" {
		http.Error(w, "missing X-User-ID", http.StatusUnauthorized)
		return
	}

	dateFrom := strings.TrimSpace(r.URL.Query().Get("date_from"))
	dateTo := strings.TrimSpace(r.URL.Query().Get("date_to"))
	if dateFrom == "" || dateTo == "" {
		http.Error(w, "date_from and date_to are required", http.StatusBadRequest)
		return
	}

	from, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		http.Error(w, "invalid date_from", http.StatusBadRequest)
		return
	}
	to, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		http.Error(w, "invalid date_to", http.StatusBadRequest)
		return
	}
	if to.Before(from) {
		http.Error(w, "date_to must be >= date_from", http.StatusBadRequest)
		return
	}

	loadedUntil, err := s.getLoadedUntil(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"error":   "report_period_not_ready",
				"message": "etl watermark not found",
			})
			return
		}
		http.Error(w, "failed to read watermark: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if to.After(loadedUntil) {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":   "report_period_not_ready",
			"message": fmt.Sprintf("requested period exceeds loaded_until=%s", loadedUntil.UTC().Format(time.RFC3339)),
		})
		return
	}

	objectKey := buildReportObjectKey(userID, dateFrom, dateTo, loadedUntil)
	cdnURL := buildCDNURL(s.cfg.CDNBaseURL, objectKey)

	exists, err := s.objectExists(r.Context(), objectKey)
	if err != nil {
		http.Error(w, "failed to check s3 object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		writeJSON(w, http.StatusOK, CDNResponse{
			Status:      "cached",
			CDNURL:      cdnURL,
			ObjectKey:   objectKey,
			LoadedUntil: loadedUntil.UTC().Format(time.RFC3339),
		})
		return
	}

	items, err := s.queryReportItems(r.Context(), userID, dateFrom, dateTo)
	if err != nil {
		http.Error(w, "query report failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	payload, err := json.MarshalIndent(CachedReportResponse{
		UserID:      userID,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		LoadedUntil: loadedUntil.UTC().Format(time.RFC3339),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Items:       items,
	}, "", "  ")
	if err != nil {
		http.Error(w, "failed to serialize report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.uploadReportJSON(r.Context(), objectKey, payload); err != nil {
		http.Error(w, "failed to upload report to s3: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, CDNResponse{
		Status:      "generated",
		CDNURL:      cdnURL,
		ObjectKey:   objectKey,
		LoadedUntil: loadedUntil.UTC().Format(time.RFC3339),
	})
}

func (s *Server) getLoadedUntil(ctx context.Context) (time.Time, error) {
	var loadedUntil time.Time
	err := s.db.QueryRowContext(ctx, `
		SELECT loaded_until
		FROM report_mart.etl_watermark
		WHERE pipeline_name = ?
		ORDER BY updated_at DESC
		LIMIT 1
	`, s.cfg.PipelineName).Scan(&loadedUntil)
	return loadedUntil, err
}

func (s *Server) queryReportItems(ctx context.Context, userID, dateFrom, dateTo string) ([]ReportItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			report_date,
			prosthesis_id,
			prosthesis_model,
			calibration_state,
			daily_movements,
			avg_load,
			battery_cycles,
			avg_battery_level,
			alerts_count,
			telemetry_events_count
		FROM report_mart.user_reports_v2 --task4 debezium sinc
		WHERE user_id = ? AND report_date BETWEEN ? AND ?
		ORDER BY report_date DESC, prosthesis_id
	`, userID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ReportItem, 0)
	for rows.Next() {
		var (
			reportDate       time.Time
			prosthesisID     string
			prosthesisModel  sql.NullString
			calibrationState sql.NullString
			dailyMovements   uint32
			avgLoad          float64
			batteryCycles    uint32
			avgBatteryLevel  float64
			alertsCount      uint32
			telemetryCount   uint32
		)
		if err := rows.Scan(
			&reportDate,
			&prosthesisID,
			&prosthesisModel,
			&calibrationState,
			&dailyMovements,
			&avgLoad,
			&batteryCycles,
			&avgBatteryLevel,
			&alertsCount,
			&telemetryCount,
		); err != nil {
			return nil, err
		}
		items = append(items, ReportItem{
			ReportDate:           reportDate.Format("2006-01-02"),
			ProsthesisID:         prosthesisID,
			ProsthesisModel:      nullString(prosthesisModel),
			CalibrationState:     nullString(calibrationState),
			DailyMovements:       dailyMovements,
			AvgLoad:              avgLoad,
			BatteryCycles:        batteryCycles,
			AvgBatteryLevel:      avgBatteryLevel,
			AlertsCount:          alertsCount,
			TelemetryEventsCount: telemetryCount,
		})
	}
	return items, rows.Err()
}

func (s *Server) objectExists(ctx context.Context, objectKey string) (bool, error) {
	_, err := s.s3.StatObject(ctx, s.cfg.S3Bucket, objectKey, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}
	var resp minio.ErrorResponse
	if errors.As(err, &resp) && (resp.Code == "NoSuchKey" || resp.Code == "NoSuchBucket" || resp.StatusCode == 404) {
		return false, nil
	}
	return false, err
}

func (s *Server) uploadReportJSON(ctx context.Context, objectKey string, payload []byte) error {
	_, err := s.s3.PutObject(ctx, s.cfg.S3Bucket, objectKey, bytes.NewReader(payload), int64(len(payload)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	return err
}

func buildReportObjectKey(userID, dateFrom, dateTo string, loadedUntil time.Time) string {
	safeLoaded := strings.ReplaceAll(loadedUntil.UTC().Format(time.RFC3339), ":", "-")
	return path.Join(userID, fmt.Sprintf("%s_%s", dateFrom, dateTo), safeLoaded, "report.json")
}

func buildCDNURL(baseURL, objectKey string) string {
	return strings.TrimRight(baseURL, "/") + "/reports/" + strings.ReplaceAll(objectKey, " ", "%20")
}

func nullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
