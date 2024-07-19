build-docker:
	docker compose build

run:
	docker compose down
	docker compose build
	rm -rf ./geth-data
	cp -R ./geth-data-archive ./geth-data
	docker compose up	-d
	docker compose logs -f replayor

initialize:
	mkdir secret
	openssl rand -hex 32 > secret/jwt

build:
	go build ./cmd/replayor/main.go
.PHONY: build

test:
	go test ./...
.PHONY: test

vet:
	go vet ./...
.PHONY: vet

fmt:
	gofmt -s -w .
.PHONY: fmt

check: lint fmt vet test build build-docker
.PHONY: check

lint:
	golangci-lint run -E goimports,sqlclosecheck,bodyclose,asciicheck,misspell,errorlint --timeout 5m -e "errors.As" -e "errors.Is" ./...
.PHONY: lint