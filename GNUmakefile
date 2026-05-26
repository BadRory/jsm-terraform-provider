default: build

.PHONY: build install test testacc fmt lint generate docs

build:
	go build -v ./...

install: build
	go install -v ./...

test:
	go test -v -cover -timeout=120s -parallel=4 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout=300s -parallel=4 ./...

fmt:
	gofmt -s -w .

lint:
	golangci-lint run

generate:
	go generate ./...

tidy:
	go mod tidy

# Build for release (requires goreleaser)
release:
	goreleaser release --clean
