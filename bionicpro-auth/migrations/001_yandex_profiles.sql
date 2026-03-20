CREATE TABLE yandex_profiles (
     id BIGSERIAL PRIMARY KEY,
     user_id VARCHAR(255),
     external_id VARCHAR(255) UNIQUE,
     login VARCHAR(255),
     email VARCHAR(255),
     first_name VARCHAR(255),
     last_name VARCHAR(255),
     raw_profile JSONB,
     consent_granted BOOLEAN,
     consent_at TIMESTAMP,
     created_at TIMESTAMP DEFAULT now()
);