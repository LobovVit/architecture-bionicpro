CREATE TABLE IF NOT EXISTS yandex_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL DEFAULT 'yandex',
    external_id VARCHAR(255) NOT NULL,
    login VARCHAR(255),
    email VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    display_name VARCHAR(255),
    avatar_id VARCHAR(255),
    raw_profile JSONB NOT NULL,
    consent_granted BOOLEAN NOT NULL DEFAULT false,
    consent_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT uq_yandex_profiles_external UNIQUE (provider, external_id),
    CONSTRAINT uq_yandex_profiles_user UNIQUE (user_id, provider)
);
