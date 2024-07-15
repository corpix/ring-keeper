.DEFAULT_GOAL = all

ifeq ($(GENERATOR_VERBOSE),y)
	verbose = "--verbose"
endif

test_arguments = $(verbose) --config config.json

.PHONY: all
all: lint test

.PHONY: lint
lint:
	go vet ./...

.PHONY: test
test:
	go test ./...
	bash -ec 'cd test && bash ./generate && go run ../main.go $(test_arguments) --dry-run'
	bash -ec 'cd test && go run ../main.go $(test_arguments)'
	bash -ec 'cd test && go run ../main.go $(test_arguments) --dry-run'
	ls -la test/subject
