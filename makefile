.DEFAULT_GOAL = all

.PHONY: all
all: lint test

.PHONY: lint
lint:
	go vet ./...

.PHONY: test
test:
	go test ./...
	bash -ec 'cd test && bash ./generate && go run ../main.go --config config.json --dry-run'
