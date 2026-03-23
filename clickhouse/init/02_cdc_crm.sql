CREATE DATABASE IF NOT EXISTS report_mart;

CREATE TABLE IF NOT EXISTS report_mart.crm_profiles_queue
(
    user_id String,
    provider String,
    external_id String,
    login Nullable(String),
    email Nullable(String),
    first_name Nullable(String),
    last_name Nullable(String),
    display_name Nullable(String),
    avatar_id Nullable(String),
    consent_granted UInt8,
    consent_at Nullable(String),
    updated_at Nullable(String),
    __deleted Nullable(String)
    )
    ENGINE = Kafka
    SETTINGS
    kafka_broker_list = 'kafka:29092',
    kafka_topic_list = 'crm.public.yandex_profiles',
    kafka_group_name = 'clickhouse-crm-profiles',
    kafka_format = 'JSONEachRow',
    kafka_num_consumers = 1,
    kafka_handle_error_mode = 'stream';

CREATE TABLE IF NOT EXISTS report_mart.crm_profiles_current
(
    user_id String,
    crm_customer_id String,
    full_name Nullable(String),
    email Nullable(String),
    prosthesis_id String,
    prosthesis_model String,
    consent_granted UInt8,
    deleted UInt8,
    updated_at DateTime
    )
    ENGINE = ReplacingMergeTree(updated_at)
    ORDER BY (user_id);

CREATE MATERIALIZED VIEW IF NOT EXISTS report_mart.crm_profiles_mv
TO report_mart.crm_profiles_current
AS
SELECT
    user_id,
    external_id AS crm_customer_id,
    nullIf(
            trim(BOTH ' ' FROM coalesce(display_name, concat(coalesce(first_name, ''), ' ', coalesce(last_name, '')))),
            ''
    ) AS full_name,
    email,
    'prosthesis-001' AS prosthesis_id,
    'BP-X1' AS prosthesis_model,
    consent_granted,
    if(__deleted = 'true', 1, 0) AS deleted,
    now() AS updated_at
FROM report_mart.crm_profiles_queue;

CREATE TABLE IF NOT EXISTS report_mart.user_reports_v2
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
