DB_URL=mysql://gowallet_user:gowallet_password@tcp(localhost:3306)/gowallet

postgres:
	docker run --name mysql-db -e MYSQL_ROOT_PASSWORD=secret -e MYSQL_DATABASE=gowallet -p 3306:3306 -d mysql:latest

createdb:
	docker exec -it mysql-db mysql -u root -psecret -e "CREATE DATABASE IF NOT EXISTS gowallet;"

dropdb:
	docker exec -it mysql-db mysql -u root -psecret -e "DROP DATABASE IF EXISTS gowallet;"

migrateup:
	migrate -path db/migrations -database "$(DB_URL)" up

migratedown:
	migrate -path db/migrations -database "$(DB_URL)" down

migratedrop:
	migrate -path db/migrations -database "$(DB_URL)" drop

db_reset: migratedrop migrateup

.PHONY: postgres createdb dropdb migrateup migratedown migratedrop db_reset