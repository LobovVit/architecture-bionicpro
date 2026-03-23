# BionicPRO — задача 3 

## 1. Цель
Интеграция LDAP (OpenLDAP) с Keycloak

## 2. Архитектура
Frontend → bionicpro-auth → Keycloak → LDAP

## 3. LDAP
- Base DN: dc=example,dc=com
- Users: ou=People
- Groups: ou=Groups

## 4. Keycloak LDAP Federation
- URL: ldap://openldap:389
- Bind DN: cn=admin,dc=example,dc=com
- Import Users: ON
- Edit Mode: READ_ONLY

## 5. Group Mapping
LDAP → Keycloak roles:
- user → user
- administrator → administrator
- prothetic_user → prothetic_user

## 6. Проверка
- Login через Keycloak
- Проверка ролей через /auth/me

## 7. Результат
- Централизованная IAM
- Поддержка разных стран
- Унифицированные роли
