# Airflow for BionicPRO reports

## Included files
- `Dockerfile` - custom Airflow image with PostgreSQL + ClickHouse Python drivers
- `requirements.txt` - Python dependencies for DAG execution
- `dags/build_user_reports_mart.py` - ready ETL DAG

## What the DAG does
1. Reads consented users from CRM/profile PostgreSQL (`yandex_profiles`).
2. Aggregates telemetry from PostgreSQL source table `telemetry_events`.
3. Loads both datasets into ClickHouse staging tables.
4. Builds the ready-to-read OLAP mart `report_mart.user_reports`.
5. Updates `report_mart.etl_watermark`.

## Compose patch
Use this image in `airflow-init`, `airflow-webserver`, and `airflow-scheduler`:

```yaml
build:
  context: ./airflow
  dockerfile: Dockerfile
```

And pass these environment variables:

```yaml
CRM_DSN: postgresql://bionicpro:bionicpro@profiledb:5432/bionicpro
TELEMETRY_DSN: postgresql://telemetry:telemetry@telemetrydb:5432/telemetry
CLICKHOUSE_HOST: clickhouse
CLICKHOUSE_PORT: "9000"
CLICKHOUSE_DB: report_mart
CLICKHOUSE_USER: report_user
CLICKHOUSE_PASSWORD: report_pass
REPORTS_ROLLING_DAYS: "7"
```
