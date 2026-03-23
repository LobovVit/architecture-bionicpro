CREATE DATABASE IF NOT EXISTS report_mart;

CREATE TABLE IF NOT EXISTS report_mart.crm_stage
(
    user_id String,
    crm_customer_id String,
    full_name Nullable(String),
    email Nullable(String),
    prosthesis_id String,
    prosthesis_model String
    )
    ENGINE = MergeTree
    ORDER BY (user_id);

CREATE TABLE IF NOT EXISTS report_mart.telemetry_stage
(
    report_date Date,
    user_id String,
    prosthesis_id String,
    calibration_state Nullable(String),
    daily_movements UInt32,
    avg_load Float64,
    battery_cycles UInt32,
    avg_battery_level Float64,
    alerts_count UInt32,
    telemetry_events_count UInt32
    )
    ENGINE = MergeTree
    ORDER BY (user_id, report_date, prosthesis_id);

CREATE TABLE IF NOT EXISTS report_mart.user_reports
(
    report_date Date,
    user_id String,
    crm_customer_id Nullable(String),
    full_name Nullable(String),
    email Nullable(String),
    prosthesis_id String,
    prosthesis_model Nullable(String),
    calibration_state Nullable(String),
    daily_movements UInt32,
    avg_load Float64,
    battery_cycles UInt32,
    avg_battery_level Float64,
    alerts_count UInt32,
    telemetry_events_count UInt32,
    created_at DateTime,
    updated_at DateTime
    )
    ENGINE = MergeTree
    ORDER BY (user_id, report_date, prosthesis_id);

CREATE TABLE IF NOT EXISTS report_mart.etl_watermark
(
    pipeline_name String,
    loaded_until DateTime,
    updated_at DateTime
)
    ENGINE = MergeTree
    ORDER BY (pipeline_name);
