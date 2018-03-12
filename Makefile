.PHONY: install
install:
	@go install

.PHONY: test
test:
	go test -cover ./...

.PHONY: test-race
test-race:
	go test -race -cover ./...

.PHONY: lint
lint: install
	@gometalinter ./...
