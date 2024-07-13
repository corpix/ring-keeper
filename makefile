.DEFAULT_GOAL = all

.PHONY: all
all: lint test

.PHONY: lint
lint:
	go vet ./...

.PHONY: test
test:
	go test ./...
	bash -ec 'cd test && bash ./generate && go run ../main.go --verbose --config config.json --dry-run'
	bash -ec 'cd test && go run ../main.go --verbose --config config.json'
	bash -ec 'cd test && go run ../main.go --verbose --config config.json --dry-run'
	ls -la test/subject
