BINARY_NAME := kafka-api

LOCAL_BIN := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/.bin

GOIMPORTS_VERSION=v0.8.0
SQLBOILER_VERSION=v4.16.2
MIGRATE_VERSION=v4.15.2
URLENC_VERSION=v1.1.1
GRPCURL_VERSION=v1.8.7
version := "0.1.0"

export PATH := ${LOCAL_BIN}:$(PATH)

tools: go-imports database-tools
database-tools: sqlboiler-driver-install sqlboiler-install golang-migrate-install urlenc-install
go-imports:
	GOBIN=$(LOCAL_BIN) go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
sqlboiler-install:
	GOBIN=$(LOCAL_BIN) go install \
        github.com/volatiletech/sqlboiler/v4@$(SQLBOILER_VERSION)
sqlboiler-driver-install:
	GOBIN=$(LOCAL_BIN) go install \
        github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@$(SQLBOILER_VERSION)
urlenc-install:
	GOBIN=$(LOCAL_BIN) go install \
        github.com/bww/urlencode/cmd/urlenc@$(URLENC_VERSION)
golang-migrate-install:
	GOBIN=$(LOCAL_BIN) go install \
        -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)


install-libs:
	brew install golang-migrate

create-migration-file:
	migrate create -ext sql -dir internal/db/migrations -seq topic

migration_up:
	export POSTGRESQL_URL=postgres://${PGUSER}:${PGPASSWORD | urlenc enc}@${PGHOST}:${PGPORT}/${PGDATABASE}?sslmode=disable; \
	migrate -path internal/db/migrations/ -database "$$POSTGRESQL_URL" -verbose up

migrate_down:
	export POSTGRESQL_URL=postgres://${PGUSER}:${PGPASSWORD | urlenc enc}@${PGHOST}:${PGPORT}/${PGDATABASE}?sslmode=disable; \
	migrate -path internal/db/migrations/ -database "$$POSTGRESQL_URL" -verbose down

set-env:
	export CONFIG_FILE_PATH=./local;
	export OTEL_SERVICE_NAME=kafka-janitor-stage;

test:
	go test -failfast -v -vet=off -race ./...

generate-sql-models:
	PSQL_HOST=${PGHOST} PSQL_PORT=${PGPORT} PSQL_DBNAME=${PGDATABASE} PSQL_USER=${PGUSER} PSQL_PASS=${PGPASSWORD} \
		sqlboiler psql

sonar-scanner-install:
	brew install sonar-scanner

start-sonar-local:
	docker run -d --name sonarqube -p 9000:9000 sonarqube

sonar-test-coverage:
	go test -failfast -v -vet=off -race ./... -coverprofile=./.bin/coverage.out

sonar-test:
	go test -failfast -v -vet=off -race ./... -json > ./.bin/report.json

sonar-local: sonar-test sonar-test-coverage
	sonar-scanner -Dsonar.working.directory=.bin/sonar  -Dproject.settings=./local-sonar.properties -Dsonar.projectVersion="$(version)"

swag-init:
	go generate ./...

build: swag-init
	go build -o myapp main.go

run: build
	go run main.go

