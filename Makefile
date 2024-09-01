startdb:
	docker run --rm --name sso-postgres -v sso_postgres_data:/var/lib/postgresql/data -d -p 5432:5432 \
	--env-file .env postgres:15-alpine

stopdb:
	docker stop sso-postgres

migrate:
	docker run -v ./migrations:/migrations migrate/migrate \
    -path=/migrations -database "postgres://sso_user:1234@172.17.0.2:5432/sso?sslmode=disable" $(direction)

makemigrations:
	docker run  -v ./migrations:/migrations migrate/migrate create -ext=".sql" -seq -dir="./migrations" $(name)

migrate-test:
	docker run -v ./tests/migrations:/tests/migrations migrate/migrate \
    -path=/migrations -database "postgres://sso_testuser:test@localhost:5432/sso?sslmode=disable" $(direction)

run:
	go run ./cmd/sso -config=./config/local.yaml