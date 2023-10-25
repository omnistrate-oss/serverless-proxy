# Stage 1: Build
FROM golang:1.20.7-bullseye as builder

ARG TARGETOS
ARG TARGETARCH

# Set the Current Working Directory inside the container
WORKDIR /go/src/pgproxy

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN echo Fetching project dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go mod tidy

RUN echo Building and installing the pgproxy service
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build  \
      -o /go/bin/proxyd /go/src/pgproxy/cmd

RUN ls -lrt /go/bin/proxyd

# Stage 2: Run
FROM debian:bullseye AS final

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends netcat ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /go/bin/proxyd /proxy/proxyd

# Expose port 3001 to the outside
EXPOSE 30001 30000 30002

# Command to run the executable
ENTRYPOINT ["/proxy/proxyd"]