DOCKER_PLATFORM=linux/arm64
TESTCOVERAGE_THRESHOLD=0

# Build info
CGO_ENABLED=0

.PHONY: all
all: tidy build unit-test lint

.PHONY: tidy
tidy:
	echo "Tidy dependency modules"
	go mod tidy

.PHONY: download
tidy:
	echo "Download dependency modules"
	go mod download

.PHONY: build
build:
	echo "Building go binaries for service"
	go build -o proxy-generic ./cmd/generic/cmd.go
	go build -o proxy-mysql ./cmd/mysql/cmd.go
	go build -o proxy-postgres ./cmd/postgres/cmd.go

.PHONY: unit-test
unit-test: 
	echo "Running unit tests for service"
	go test ./... -skip ./test/... -cover -coverprofile coverage.out -covermode count
	go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/[%]//g' | awk 'current=$$1; {if (current < ${TESTCOVERAGE_THRESHOLD}) {print "\033[31mTest coverage is " current " which is below threshold\033[0m"; exit 1} else {print "\033[32mTest coverage is above threshold\033[0m"}}'

.PHONY: test-coverage-report
test-coverage-report: 
	go test ./... -skip ./test/... -cover -coverprofile coverage.out -covermode count
	go tool cover -html=coverage.out

.PHONY: lint-install
lint-install: 
	echo "Installing golangci-lint"
	brew install golangci-lint
	brew upgrade golangci-lint	

.PHONY: lint
lint:
	echo "Running checks for service"
	golangci-lint run ./...

.PHONY: sec-install
sec-install: 
	echo "Installing gosec"
	go install github.com/securego/gosec/v2/cmd/gosec@latest

.PHONY: sec
sec: 
	echo "Security scanning for service"
	gosec --quiet ./...

.PHONY: update-dependencies
update-dependencies:
	echo "Updating dependencies"
	go get -t -u ./...
	go mod tidy

.PHONY: run-generic
run-generic:
	echo "Running service" && \
    export DRY_RUN=true && \
    export LOG_LEVEL=debug && \
    export LOG_FORMAT=pretty && \
	go run ${BUILD_FLAGS} ./cmd/generic/cmd.go

.PHONY: docker-build-generic
docker-build-generic:
	docker buildx build --platform=${DOCKER_PLATFORM} -f ./build/Dockerfile --build-arg PROXYVARIANT=generic -t serverless-proxy-generic:latest . 

.PHONY: docker-build-mysql
docker-build-generic-mysql:
	docker buildx build --platform=${DOCKER_PLATFORM} -f ./build/Dockerfile --build-arg PROXYVARIANT=mysql -t serverless-proxy-mysql:latest . 

.PHONY: docker-build-postgres
docker-build-generic-postgres:
	docker buildx build --platform=${DOCKER_PLATFORM} -f ./build/Dockerfile --build-arg PROXYVARIANT=postgres -t serverless-proxy-postgres:latest . 