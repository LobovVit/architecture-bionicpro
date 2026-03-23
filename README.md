### Чистим: 
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
```
docker compose up -d --build
```
### Дозапускаем 
```
docker compose up -d bionicpro-auth airflow-scheduler airflow-webserver airflow-init
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
### Ходим смотрим
```
http://localhost:8085
cn=admin,dc=example,dc=com / admin
```
```
http://localhost:8080  admin / admin
```
```
http://localhost:3000` jane.smith / password
```

### Atirflow 
```
http://localhost:8089    admin / admin
``` 

### Minio
```
http://localhost:9010/  minioadmin / minioadmin
```