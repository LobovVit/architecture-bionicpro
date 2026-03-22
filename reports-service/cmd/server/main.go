package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func main() {
	ClickHouseDSN := getEnv("CLICKHOUSE_DSN", "clickhouse://report_user:report_pass@clickhouse:9000/report_mart")
	db, _ := sql.Open("clickhouse", ClickHouseDSN)

	http.HandleFunc("/reports/me", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		dateFrom := r.URL.Query().Get("date_from")
		dateTo := r.URL.Query().Get("date_to")
		if dateFrom == "" || dateTo == "" {
			http.Error(w, "date_from and date_to are required", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(`
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
		FROM report_mart.user_reports
		WHERE user_id = ? AND report_date BETWEEN ? AND ?
		ORDER BY report_date DESC, prosthesis_id
	`, userID, dateFrom, dateTo)
		if err != nil {
			http.Error(w, "query report failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

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

		type ReportResponse struct {
			UserID      string       `json:"user_id"`
			DateFrom    string       `json:"date_from"`
			DateTo      string       `json:"date_to"`
			GeneratedAt string       `json:"generated_at"`
			Items       []ReportItem `json:"items"`
		}

		var items []ReportItem

		for rows.Next() {
			var item ReportItem
			var reportDate string

			err := rows.Scan(
				&reportDate,
				&item.ProsthesisID,
				&item.ProsthesisModel,
				&item.CalibrationState,
				&item.DailyMovements,
				&item.AvgLoad,
				&item.BatteryCycles,
				&item.AvgBatteryLevel,
				&item.AlertsCount,
				&item.TelemetryEventsCount,
			)
			if err != nil {
				http.Error(w, "scan report failed: "+err.Error(), http.StatusInternalServerError)
				return
			}

			item.ReportDate = reportDate
			items = append(items, item)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "rows iteration failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ReportResponse{
			UserID:      userID,
			DateFrom:    dateFrom,
			DateTo:      dateTo,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			Items:       items,
		})
	})

	http.ListenAndServe(":8090", nil)
}
