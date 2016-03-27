.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build:
	go build ./...

.PHONY: test
test: build
	go test -cover ./...

.PHONY: coverage
coverage: build
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: install_deps
install_deps:
	go get ./...
