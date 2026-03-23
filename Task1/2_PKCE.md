# BionicPRO — задача 2 (PKCE)

## Цель

Усилить безопасность OAuth 2.0 Authorization Code Flow за счет внедрения механизма PKCE (Proof Key for Code Exchange).

---

## Основная идея

PKCE защищает процесс обмена `authorization code` на токены от атак перехвата.

Даже если злоумышленник получит `authorization code`, он не сможет обменять его на `access_token`, так как не знает `code_verifier`.

---

## Изменения в архитектуре

PKCE внедряется в поток:

Frontend → bionicpro-auth → Keycloak

При этом:
- frontend не получает access_token и refresh_token;
- bionicpro-auth выступает как OAuth-клиент;
- PKCE реализуется на стороне backend.

---

## Настроики Keycloak

Для клиента (frontend или bionicpro-auth):

- Standard Flow: включен
- Implicit Flow: отключен
- Direct Access Grants: отключен
- PKCE: включен
- PKCE Method: S256

---

## Реализация PKCE

### Шаг 1. Генерация параметров

bionicpro-auth генерирует:
- code_verifier (случаиная строка)
- code_challenge = BASE64URL(SHA256(code_verifier))

---

### Шаг 2. Запрос авторизации

Пользователь перенаправляется в Keycloak:

/authorize?
- response_type=code
- client_id=...
- redirect_uri=...
- code_challenge=...
- code_challenge_method=S256

---

### Шаг 3. Callback

Keycloak возвращает:
- authorization code

в bionicpro-auth

---

### Шаг 4. Обмен на токены

bionicpro-auth отправляет запрос в Keycloak:

/token:
- code
- code_verifier

Если verifier совпадает — выдаются токены.

---

## Безопасность

PKCE защищает от:
- перехвата authorization code
- повторного использования code
- атак через redirect

---

## Роль PKCE в архитектуре BionicPRO

PKCE дополняет архитектуру:

- токены не передаются во frontend
- используются только server-side
- минимизируется риск утечки

---

## Вывод

Внедрение PKCE позволяет:
- защитить Authorization Code Flow
- исключить использование перехваченного кода
- повысить безопасность всеи системы аутентификации
