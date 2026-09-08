DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet
AUTH_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet_auth
USER_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet_user
WALLET_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet
TRANSACTION_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet_transaction

mysql-container:
	docker run --name gowallet-mysql-ms -e MYSQL_ROOT_PASSWORD=rootpassword -e MYSQL_DATABASE=gowallet -p 3306:3306 -d mysql:latest

createdb:
	docker exec -it gowallet-mysql-ms mysql -u root -prootpassword -e "CREATE DATABASE IF NOT EXISTS gowallet; CREATE DATABASE IF NOT EXISTS gowallet_auth; CREATE DATABASE IF NOT EXISTS gowallet_user; CREATE DATABASE IF NOT EXISTS gowallet_transaction;"

dropdb:
	docker exec -it gowallet-mysql-ms mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS gowallet; DROP DATABASE IF EXISTS gowallet_auth; DROP DATABASE IF EXISTS gowallet_user; DROP DATABASE IF EXISTS gowallet_transaction;"

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

# === User Service Specific Commands ===
migrateup-user-service:
	migrate -path user-service/db/migrations -database "$(USER_DB_URL)" up

migratedown-user-service:
	migrate -path user-service/db/migrations -database "$(USER_DB_URL)" down

migratedrop-user-service:
	migrate -path user-service/db/migrations -database "$(USER_DB_URL)" drop

user_db_reset: migratedrop-user-service migrateup-user-service

# === Transaction Service Specific Commands ===
create-transaction-db:
	docker exec -it gowallet-mysql-ms mysql -u root -prootpassword -e "CREATE DATABASE IF NOT EXISTS gowallet_transaction;"

drop-transaction-db:
	docker exec -it gowallet-mysql-ms mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS gowallet_transaction;"

migrateup-transaction-service:
	migrate -path transaction-service/db/migrations -database "$(TRANSACTION_DB_URL)" up

migratedown-transaction-service:
	migrate -path transaction-service/db/migrations -database "$(TRANSACTION_DB_URL)" down

migratedrop-transaction-service:
	migrate -path transaction-service/db/migrations -database "$(TRANSACTION_DB_URL)" drop

transaction_db_reset: migratedrop-transaction-service migrateup-transaction-service

# === Run Services Locally ===
run-user:
	cd user-service && go run main.go

run-auth:
	cd auth-service && go run main.go

run-transaction:
	cd transaction-service && go run main.go

.PHONY: mysql-container createdb dropdb migrateup migratedown migratedrop db_reset create-auth-db drop-auth-db migrateup-auth-service migratedown-auth-service migratedrop-auth-service auth_db_reset migrateup-user-service migratedown-user-service migratedrop-user-service user_db_reset create-transaction-db drop-transaction-db migrateup-transaction-service migratedown-transaction-service migratedrop-transaction-service transaction_db_reset run-user run-auth run-transaction