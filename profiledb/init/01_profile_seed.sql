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
    raw_profile JSONB NOT NULL DEFAULT '{}'::jsonb,
    consent_granted BOOLEAN NOT NULL DEFAULT false,
    consent_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT uq_yandex_profiles_external UNIQUE (provider, external_id),
    CONSTRAINT uq_yandex_profiles_user UNIQUE (user_id, provider)
    );

INSERT INTO yandex_profiles (
    user_id,
    provider,
    external_id,
    login,
    email,
    first_name,
    last_name,
    display_name,
    avatar_id,
    raw_profile,
    consent_granted,
    consent_at
) VALUES (
             'dec21cbf-4e40-4c9d-8c9e-5e50e6785dcd',
             'yandex',
             '305239105',
             'lobov.qwedsa',
             'lobov.qwedsa@yandex.ru',
             'Виталий',
             'Лобов',
             'Vitalii Lobov',
             '51381/test-avatar',
             '{"seed": true}'::jsonb,
             true,
             now()
         )
    ON CONFLICT (provider, external_id) DO UPDATE SET
    login = EXCLUDED.login,
               email = EXCLUDED.email,
               first_name = EXCLUDED.first_name,
               last_name = EXCLUDED.last_name,
               display_name = EXCLUDED.display_name,
               avatar_id = EXCLUDED.avatar_id,
               raw_profile = EXCLUDED.raw_profile,
               consent_granted = true,
               consent_at = now(),
               updated_at = now();