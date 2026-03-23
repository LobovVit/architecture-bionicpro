##  Как запустить и проверить (команды ниже): 
1. Чистим (если уже запускали и хотим с чистого листа)
2. Запускаем (Дозапускаем если нужно)
3. Ходим смотрим (ссылки ниже)
4. Для отчета под конкретного пользователя проливаем на БД telemetrydb скрипт  - [01_telemetry.sql](telemetrydb/init/01_telemetry.sql) 
с указанием UUID нужного пользователя
5. Atirflow DAG запускать в ручную (в веб интерфейсе)


## Чистим: 
```
docker compose down -v
```
```
rm -rf postgres-keycloak-data
rm -rf ldap/ldap-data ldap/slapd-data
rm -rf postgres-bionicpro-data
rm -rf postgres-telemetry-data
rm -rf postgres-airflow-data
rm -rf clickhouse/data
rm -rf nginx/cache
rm -rf minio-data
```
## Запускаем
```
docker compose up -d --build
```
### Дозапускаем (если что то не стартануло)
```
docker compose up -d bionicpro-auth airflow-scheduler airflow-webserver airflow-init kafka-connect-init
```
```
curl http://localhost:8083/connectors/crm-profiledb-connector/status  -- проверяем - чаще всего не отрабатывает создание коннектора и не стартует airflow
```
### Отключаем ssh (если хотим в keycloak http://localhost:8080/)
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
## Ходим смотрим
### LDAP админка
```
http://localhost:8085
cn=admin,dc=example,dc=com / admin
```
### keycloak админка
```
http://localhost:8080  admin / admin
```
### frontend
``` 
http://localhost:3000` jane.smith / password
```

### Atirflow 
```
Atirflow админка
http://localhost:8089    admin / admin
``` 

### Minio
```
Minio админка
http://localhost:9010/  minioadmin / minioadmin
```
