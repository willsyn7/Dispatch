.PHONY: api web test lint

api:
	cd apps/api && go run ./cmd/api

web:
	cd apps/web && npm start

test:
	cd apps/api && go test ./...

lint:
	cd apps/api && gofmt -w $$(find . -name '*.go' -type f)

