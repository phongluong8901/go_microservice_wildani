DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet
AUTH_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet_auth
USER_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet_user
WALLET_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet_wallet
TRANSACTION_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet_transaction
PAYMENT_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet_payment
LEDGER_DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet_ledger

mysql-container:
	docker run --name gowallet-mysql-ms -e MYSQL_ROOT_PASSWORD=rootpassword -e MYSQL_DATABASE=gowallet -p 3306:3306 -d mysql:latest

createdb:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "CREATE DATABASE IF NOT EXISTS gowallet; CREATE DATABASE IF NOT EXISTS gowallet_auth; CREATE DATABASE IF NOT EXISTS gowallet_user; CREATE DATABASE IF NOT EXISTS gowallet_wallet; CREATE DATABASE IF NOT EXISTS gowallet_transaction; CREATE DATABASE IF NOT EXISTS gowallet_payment; CREATE DATABASE IF NOT EXISTS gowallet_ledger;"

dropdb:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS gowallet; DROP DATABASE IF EXISTS gowallet_auth; DROP DATABASE IF EXISTS gowallet_user; DROP DATABASE IF EXISTS gowallet_wallet; DROP DATABASE IF EXISTS gowallet_transaction; DROP DATABASE IF EXISTS gowallet_payment; DROP DATABASE IF EXISTS gowallet_ledger;"

migrateup:
	migrate -path db/migrations -database "$(DB_URL)" up

migratedown:
	migrate -path db/migrations -database "$(DB_URL)" down

migratedrop:
	migrate -path db/migrations -database "$(DB_URL)" drop

db_reset: migratedrop migrateup

# === Auth Service Specific Commands ===
create-auth-db:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "CREATE DATABASE IF NOT EXISTS gowallet_auth;"

drop-auth-db:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS gowallet_auth;"

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

# === Wallet Service Specific Commands ===
create-wallet-db:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "CREATE DATABASE IF NOT EXISTS gowallet_wallet;"

drop-wallet-db:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS gowallet_wallet;"

migrateup-wallet-service:
	migrate -path wallet-service/db/migrations -database "$(WALLET_DB_URL)" up

migratedown-wallet-service:
	migrate -path wallet-service/db/migrations -database "$(WALLET_DB_URL)" down

migratedrop-wallet-service:
	migrate -path wallet-service/db/migrations -database "$(WALLET_DB_URL)" drop

wallet_db_reset: migratedrop-wallet-service migrateup-wallet-service

# === Transaction Service Specific Commands ===
create-transaction-db:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "CREATE DATABASE IF NOT EXISTS gowallet_transaction;"

drop-transaction-db:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS gowallet_transaction;"

migrateup-transaction-service:
	migrate -path transaction-service/db/migrations -database "$(TRANSACTION_DB_URL)" up

migratedown-transaction-service:
	migrate -path transaction-service/db/migrations -database "$(TRANSACTION_DB_URL)" down

migratedrop-transaction-service:
	migrate -path transaction-service/db/migrations -database "$(TRANSACTION_DB_URL)" drop

transaction_db_reset: migratedrop-transaction-service migrateup-transaction-service

# === Payment Service Specific Commands ===
create-payment-db:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "CREATE DATABASE IF NOT EXISTS gowallet_payment;"

drop-payment-db:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS gowallet_payment;"

migrateup-payment-service:
	migrate -path payment-service/db/migrations -database "$(PAYMENT_DB_URL)" up

migratedown-payment-service:
	migrate -path payment-service/db/migrations -database "$(PAYMENT_DB_URL)" down

migratedrop-payment-service:
	migrate -path payment-service/db/migrations -database "$(PAYMENT_DB_URL)" drop

payment_db_reset: migratedrop-payment-service migrateup-payment-service

# === Ledger Service Specific Commands ===
create-ledger-db:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "CREATE DATABASE IF NOT EXISTS gowallet_ledger;"

drop-ledger-db:
	docker exec -i gowallet-mysql-ms mysql -u root -prootpassword -e "DROP DATABASE IF EXISTS gowallet_ledger;"

migrateup-ledger-service:
	migrate -path ledger-service/db/migrations -database "$(LEDGER_DB_URL)" up

migratedown-ledger-service:
	migrate -path ledger-service/db/migrations -database "$(LEDGER_DB_URL)" down

migratedrop-ledger-service:
	migrate -path ledger-service/db/migrations -database "$(LEDGER_DB_URL)" drop

ledger_db_reset: migratedrop-ledger-service migrateup-ledger-service

# === Run All Migrations Command ===
migrateup-all: createdb migrateup-auth-service migrateup-user-service migrateup-wallet-service migrateup-transaction-service migrateup-payment-service migrateup-ledger-service
	@echo "All database migrations executed successfully!"

# === Run Services Locally ===
run-user:
	cd user-service && go run main.go

run-auth:
	cd auth-service && go run main.go

run-wallet:
	cd wallet-service && go run main.go

run-transaction:
	cd transaction-service && go run main.go

run-payment:
	cd payment-service && go run main.go

run-ledger:
	cd ledger-service && go run main.go

.PHONY: mysql-container createdb dropdb migrateup migratedown migratedrop db_reset create-auth-db drop-auth-db migrateup-auth-service migratedown-auth-service migratedrop-auth-service auth_db_reset migrateup-user-service migratedown-user-service migratedrop-user-service user_db_reset create-wallet-db drop-wallet-db migrateup-wallet-service migratedown-wallet-service migratedrop-wallet-service wallet_db_reset create-transaction-db drop-transaction-db migrateup-transaction-service migratedown-transaction-service migratedrop-transaction-service transaction_db_reset create-payment-db drop-payment-db migrateup-payment-service migratedown-payment-service migratedrop-payment-service payment_db_reset create-ledger-db drop-ledger-db migrateup-ledger-service migratedown-ledger-service migratedrop-ledger-service ledger_db_reset migrateup-all run-user run-auth run-wallet run-transaction run-payment run-ledger