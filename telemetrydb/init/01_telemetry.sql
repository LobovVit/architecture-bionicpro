CREATE TABLE IF NOT EXISTS telemetry_events (
                                                id BIGSERIAL PRIMARY KEY,
                                                event_ts TIMESTAMP NOT NULL,
                                                user_id TEXT NOT NULL,
                                                prosthesis_id TEXT NOT NULL,
                                                movements_count INT NOT NULL DEFAULT 0,
                                                load_value NUMERIC(10,2) NOT NULL DEFAULT 0,
    battery_cycles_delta INT NOT NULL DEFAULT 0,
    battery_level NUMERIC(10,2) NOT NULL DEFAULT 0,
    has_alert BOOLEAN NOT NULL DEFAULT FALSE,
    calibration_state TEXT NOT NULL DEFAULT 'OK'
    );

INSERT INTO telemetry_events (
    event_ts,
    user_id,
    prosthesis_id,
    movements_count,
    load_value,
    battery_cycles_delta,
    battery_level,
    has_alert,
    calibration_state
) VALUES
      ('2026-03-14 09:10:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 120, 34.5, 1, 91.0, FALSE, 'OK'),
      ('2026-03-14 12:45:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 180, 36.2, 0, 88.0, FALSE, 'OK'),
      ('2026-03-15 10:00:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 210, 40.1, 1, 84.0, FALSE, 'OK'),
      ('2026-03-15 18:20:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 90, 30.0, 0, 80.0, TRUE, 'WARN'),
      ('2026-03-16 08:00:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 240, 41.8, 1, 78.0, FALSE, 'OK'),
      ('2026-03-16 14:30:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 150, 35.0, 0, 75.0, FALSE, 'OK'),
      ('2026-03-17 09:15:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 300, 44.2, 1, 72.0, FALSE, 'OK'),
      ('2026-03-17 19:00:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 130, 31.4, 0, 69.0, FALSE, 'OK'),
      ('2026-03-18 11:40:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 260, 42.5, 1, 65.0, FALSE, 'OK'),
      ('2026-03-18 17:10:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 110, 29.9, 0, 62.0, TRUE, 'WARN'),
      ('2026-03-19 09:00:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 280, 43.1, 1, 58.0, FALSE, 'OK'),
      ('2026-03-19 16:45:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 160, 33.7, 0, 55.0, FALSE, 'OK'),
      ('2026-03-20 10:30:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 220, 39.8, 1, 51.0, FALSE, 'OK'),
      ('2026-03-20 18:00:00', 'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd', 'prosthesis-001', 140, 28.6, 0, 48.0, FALSE, 'OK');