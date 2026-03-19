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

## Шпаргалка  -
```
docker compose exec keycloak /opt/keycloak/bin/kcadm.sh config credentials \
--server http://localhost:8080 \
--realm master \
--user admin \
--password admin
```
```
docker compose exec keycloak /opt/keycloak/bin/kcadm.sh update realms/master -s sslRequired=NONE
```
```
docker compose exec keycloak /opt/keycloak/bin/kcadm.sh update realms/reports-realm -s sslRequired=NONE
```
http://localhost:8085 
Логин: cn=admin,dc=example,dc=com
пароль: admin


reports-realm → User Federation → Add provider → ldap
Connection
Connection URL: ldap://openldap:389
Bind DN: cn=admin,dc=example,dc=com
Bind Credential: admin
Users
Users DN: ou=People,dc=example,dc=com
Username LDAP attribute: uid
RDN LDAP attribute: uid
UUID LDAP attribute: entryUUID
User Object Classes: inetOrgPerson
Other
Edit Mode: READ_ONLY
Import Users: ON
Sync Registrations: OFF

Synchronize all users

Добавляем роли (Group Mapper)
LDAP provider → Mappers → Create

Name: ldap-groups
Mapper Type: group-ldap-mapper

LDAP Groups DN: ou=Groups,dc=example,dc=com
Group Name LDAP Attribute: cn
Membership LDAP Attribute: member
Membership Attribute Type: DN

Group Object Classes: groupOfNames

User Groups Retrieve Strategy: LOAD_GROUPS_BY_MEMBER_ATTRIBUTE
Mode: READ_ONLY