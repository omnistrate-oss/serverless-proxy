GIT_USER?=$(shell gh api user -q ".login") # gets current user using github cli if the variable is not already set
GIT_TOKEN?=$(shell gh config get -h github.com oauth_token) # gets current user using github cli if the variable is not already set
DOCKER_PLATFORM=linux/amd64
TESTCOVERAGE_THRESHOLD=0
REPO_ROOT=$(shell git rev-parse --show-toplevel)

# Build info
CGO_ENABLED=0
GOPRIVATE=github.com/omnistrate

.PHONY: all
all: tidy build sec

.PHONY: docker-build-and-push
docker-build-and-push:
	docker buildx build --platform=linux/amd64,linux/arm64 -f cmd/build/Dockerfile -t omnistrate/observability:latest . --push

.PHONY: tidy
tidy:
	echo "Tidy dependency modules"
	go mod tidy

.PHONY: build
build:
	echo "Building go binaries for service"
	go build -o proxyd ./cmd/cmd.go

.PHONY: sec-install
sec-install:
	echo "Installing gosec"
	go install github.com/securego/gosec/v2/cmd/gosec@latest

.PHONY: sec
sec:
	echo "Security scanning for service"
	gosec --quiet ./...

.PHONY: install-dependencies
install-dependencies: lint-install sec-install

.PHONY: update-dependencies
update-dependencies:
	echo "Updating dependencies"
	go get -t -u ./...
	go mod tidy

.PHONY: run
run:
	echo "Running service" && \
    export PG_USER=postgres && \
    export PG_PASSWORD=XXXX && \
    export LOG_LEVEL=debug && \
    export LOG_FORMAT=pretty && \
	go run ${BUILD_FLAGS} cmd/cmd.go