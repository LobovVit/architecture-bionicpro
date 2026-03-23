from __future__ import annotations

import os
from datetime import datetime, timedelta
from io import StringIO

import pandas as pd
import psycopg2
from airflow import DAG
from airflow.operators.python import PythonOperator
from clickhouse_driver import Client as ClickHouseClient

DAG_ID = "build_user_reports_mart"

CRM_DSN = os.environ.get("CRM_DSN", "postgresql://bionicpro:bionicpro@profiledb:5432/bionicpro")
TELEMETRY_DSN = os.environ.get("TELEMETRY_DSN", "postgresql://telemetry:telemetry@telemetrydb:5432/telemetry")
CLICKHOUSE_HOST = os.environ.get("CLICKHOUSE_HOST", "clickhouse")
CLICKHOUSE_PORT = int(os.environ.get("CLICKHOUSE_PORT", "9000"))
CLICKHOUSE_DB = os.environ.get("CLICKHOUSE_DB", "report_mart")
CLICKHOUSE_USER = os.environ.get("CLICKHOUSE_USER", "report_user")
CLICKHOUSE_PASSWORD = os.environ.get("CLICKHOUSE_PASSWORD", "report_pass")
ROLLING_DAYS = int(os.environ.get("REPORTS_ROLLING_DAYS", "7"))


def ch() -> ClickHouseClient:
    return ClickHouseClient(
        host=CLICKHOUSE_HOST,
        port=CLICKHOUSE_PORT,
        database=CLICKHOUSE_DB,
        user=CLICKHOUSE_USER,
        password=CLICKHOUSE_PASSWORD,
    )


def extract_crm(**context):
    sql = '''
          SELECT
              yp.user_id,
              yp.external_id AS crm_customer_id,
              COALESCE(NULLIF(TRIM(yp.display_name), ''), CONCAT_WS(' ', yp.first_name, yp.last_name)) AS full_name,
              yp.email,
              'prosthesis-001'::text AS prosthesis_id,
              'BP-X1'::text AS prosthesis_model
          FROM yandex_profiles yp
          WHERE yp.consent_granted = true \
          '''
    conn = psycopg2.connect(CRM_DSN)
    try:
        df = pd.read_sql(sql, conn)
    finally:
        conn.close()

    context["ti"].xcom_push(key="crm_json", value=df.to_json(orient="records"))


def extract_telemetry(**context):
    sql = f'''
        SELECT
            date_trunc('day', event_ts)::date AS report_date,
            user_id,
            prosthesis_id,
            max(calibration_state) AS calibration_state,
            sum(movements_count)::bigint AS daily_movements,
            avg(load_value)::float8 AS avg_load,
            sum(battery_cycles_delta)::bigint AS battery_cycles,
            avg(battery_level)::float8 AS avg_battery_level,
            sum(CASE WHEN has_alert THEN 1 ELSE 0 END)::bigint AS alerts_count,
            count(*)::bigint AS telemetry_events_count
        FROM telemetry_events
        WHERE event_ts >= now() - interval '{ROLLING_DAYS} day'
        GROUP BY 1, 2, 3
        ORDER BY 1, 2, 3
    '''
    conn = psycopg2.connect(TELEMETRY_DSN)
    try:
        df = pd.read_sql(sql, conn)
    finally:
        conn.close()

    context["ti"].xcom_push(
        key="telemetry_json",
        value=df.to_json(orient="records", date_format="iso"),
    )


def load_crm_stage(**context):
    raw = context["ti"].xcom_pull(key="crm_json", task_ids="extract_crm")
    if not raw:
        return

    df = pd.read_json(StringIO(raw))
    client = ch()
    client.execute("TRUNCATE TABLE report_mart.crm_stage")

    if not df.empty:
        df = df[
            [
                "user_id",
                "crm_customer_id",
                "full_name",
                "email",
                "prosthesis_id",
                "prosthesis_model",
            ]
        ]

        # nullable поля
        for col in ["full_name", "email"]:
            df[col] = df[col].where(pd.notnull(df[col]), None)

        # string поля обязаны быть строками
        for col in ["user_id", "crm_customer_id", "prosthesis_id", "prosthesis_model"]:
            df[col] = df[col].astype(str)

        # если full_name/email не null, тоже делаем строками
        for col in ["full_name", "email"]:
            df[col] = df[col].apply(lambda v: None if v is None else str(v))

        rows = [tuple(row) for row in df.itertuples(index=False, name=None)]

        client.execute(
            """
            INSERT INTO report_mart.crm_stage
            (user_id, crm_customer_id, full_name, email, prosthesis_id, prosthesis_model)
            VALUES
            """,
            rows,
        )


def load_telemetry_stage(**context):
    raw = context["ti"].xcom_pull(key="telemetry_json", task_ids="extract_telemetry")
    if not raw:
        return

    df = pd.read_json(StringIO(raw))
    client = ch()
    client.execute("TRUNCATE TABLE report_mart.telemetry_stage")

    if not df.empty:
        df["report_date"] = pd.to_datetime(df["report_date"]).dt.date

        for col in ["daily_movements", "battery_cycles", "alerts_count", "telemetry_events_count"]:
            df[col] = df[col].fillna(0).astype(int)

        for col in ["avg_load", "avg_battery_level"]:
            df[col] = df[col].fillna(0.0).astype(float)

        for col in ["user_id", "prosthesis_id"]:
            df[col] = df[col].astype(str)

        df["calibration_state"] = df["calibration_state"].where(pd.notnull(df["calibration_state"]), None)
        df["calibration_state"] = df["calibration_state"].apply(lambda v: None if v is None else str(v))

        df = df[
            [
                "report_date",
                "user_id",
                "prosthesis_id",
                "calibration_state",
                "daily_movements",
                "avg_load",
                "battery_cycles",
                "avg_battery_level",
                "alerts_count",
                "telemetry_events_count",
            ]
        ]

        rows = [tuple(row) for row in df.itertuples(index=False, name=None)]

        client.execute(
            """
            INSERT INTO report_mart.telemetry_stage
            (
                report_date,
                user_id,
                prosthesis_id,
                calibration_state,
                daily_movements,
                avg_load,
                battery_cycles,
                avg_battery_level,
                alerts_count,
                telemetry_events_count
            )
            VALUES
            """,
            rows,
        )


def build_user_reports_mart():
    client = ch()
    client.execute("TRUNCATE TABLE report_mart.user_reports")
    client.execute(
        '''
        INSERT INTO report_mart.user_reports
        SELECT
            t.report_date,
            t.user_id,
            c.crm_customer_id,
            c.full_name,
            c.email,
            t.prosthesis_id,
            c.prosthesis_model,
            t.calibration_state,
            toUInt32(t.daily_movements),
            toFloat64(t.avg_load),
            toUInt32(t.battery_cycles),
            toFloat64(t.avg_battery_level),
            toUInt32(t.alerts_count),
            toUInt32(t.telemetry_events_count),
            now(),
            now()
        FROM report_mart.telemetry_stage t
                 INNER JOIN report_mart.crm_stage c
                            ON c.user_id = t.user_id
        '''
    )


def update_watermark():
    client = ch()
    now = datetime.utcnow()
    client.execute(
        '''
        INSERT INTO report_mart.etl_watermark (pipeline_name, loaded_until, updated_at)
        VALUES
        ''',
        [(DAG_ID, now, now)],
    )


default_args = {
    "owner": "bionicpro",
    "depends_on_past": False,
    "retries": 1,
    "retry_delay": timedelta(minutes=5),
}

with DAG(
        dag_id=DAG_ID,
        default_args=default_args,
        start_date=datetime(2026, 3, 20),
        schedule="0 * * * *",
        catchup=False,
        tags=["bionicpro", "reports", "etl"],
) as dag:
    extract_crm_task = PythonOperator(task_id="extract_crm", python_callable=extract_crm)
    extract_telemetry_task = PythonOperator(task_id="extract_telemetry", python_callable=extract_telemetry)
    load_crm_stage_task = PythonOperator(task_id="load_crm_stage", python_callable=load_crm_stage)
    load_telemetry_stage_task = PythonOperator(task_id="load_telemetry_stage", python_callable=load_telemetry_stage)
    build_mart_task = PythonOperator(task_id="build_user_reports_mart", python_callable=build_user_reports_mart)
    update_watermark_task = PythonOperator(task_id="update_watermark", python_callable=update_watermark)

    extract_crm_task >> load_crm_stage_task
    extract_telemetry_task >> load_telemetry_stage_task
    [load_crm_stage_task, load_telemetry_stage_task] >> build_mart_task >> update_watermark_task
