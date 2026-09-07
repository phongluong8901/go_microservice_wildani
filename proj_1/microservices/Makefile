DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet
AUTH_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet_auth

postgres:
	docker run --name mysql-db -e MYSQL_ROOT_PASSWORD=secret -e MYSQL_DATABASE=gowallet -p 3306:3306 -d mysql:latest

createdb:
	docker exec -it gowallet-mysql-ms mysql -u root -prootpassword -e "CREATE DATABASE IF NOT EXISTS gowallet; CREATE DATABASE IF NOT EXISTS gowallet_auth;"

dropdb:
	docker exec -it gowallet-mysql-ms mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS gowallet; DROP DATABASE IF EXISTS gowallet_auth;"

migrateup:
	migrate -path db/migrations -database "$(DB_URL)" up

migratedown:
	migrate -path db/migrations -database "$(DB_URL)" down

migratedrop:
	migrate -path db/migrations -database "$(DB_URL)" drop

db_reset: migratedrop migrateup

# === Auth Service Specific Commands ===
create-auth-db:
	docker exec -it gowallet-mysql-ms mysql -u root -prootpassword -e "CREATE DATABASE IF NOT EXISTS gowallet_auth;"

drop-auth-db:
	docker exec -it gowallet-mysql-ms mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS gowallet_auth;"

migrateup-auth-service:
	migrate -path auth-service/db/migrations -database "$(AUTH_DB_URL)" up

migratedown-auth-service:
	migrate -path auth-service/db/migrations -database "$(AUTH_DB_URL)" down

migratedrop-auth-service:
	migrate -path auth-service/db/migrations -database "$(AUTH_DB_URL)" drop

auth_db_reset: migratedrop-auth-service migrateup-auth-service

.PHONY: postgres createdb dropdb migrateup migratedown migratedrop db_reset create-auth-db drop-auth-db migrateup-auth-service migratedown-auth-service migratedrop-auth-service auth_db_reset