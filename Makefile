.PHONY: build, test, docker-build, run, lint, generate, docker-run docker-stop client

LDFLAGS=-ldflags="-s -w"

build:
	mkdir -p bin
	go build $(LDFLAGS) -o bin/rates-service ./cmd/service/main.go

client:
	mkdir -p bin
	go build $(LDFLAGS) -o bin/testing-client ./cmd/testing_client/main.go

test:
	DB_MIGRATIONS_PATH=$(shell pwd)/repository/db/migrations go test ./...

clean:
	rm -rf bin/

docker-build:
	docker build -t app .

run: build
	./bin/rates-service

lint:
	golangci-lint run

docker-run:
	docker-compose up -d --build

docker-stop:
	docker-compose down
