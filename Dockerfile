# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG LDFLAGS
ARG COMPILER_FLAGS
ARG APP_VERSION
ARG APP_VERSION_COMMIT
ARG APP_ENV_PREFIX
ARG TARGETOS
ARG TARGETARCH
ARG GO_RACE_DETECTOR_ENABLE
ARG CGO_ENABLE

RUN apk add --no-cache git bash make

WORKDIR /app

COPY ./ ./

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH \
    LDFLAGS=$LDFLAGS \
    COMPILER_FLAGS=$COMPILER_FLAGS \
    APP_VERSION=${APP_VERSION} \
    APP_VERSION_COMMIT=${APP_VERSION_COMMIT} \
    APP_ENV_PREFIX=${APP_ENV_PREFIX} \
    GO_RACE_DETECTOR_ENABLE=${GO_RACE_DETECTOR_ENABLE} \
    CGO_ENABLE=${CGO_ENABLE} \
    BIN_OUTPUT_DIR=./bin \
    make build

# Final stage
FROM alpine:3.22

ARG APP_NAME

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/UTC /etc/localtime && \
    echo "UTC" > /etc/timezone && \
    mkdir -p /opt/bin/

COPY --from=builder /app/bin/${APP_NAME} /opt/bin/
RUN chmod +x /opt/bin/${APP_NAME}

RUN echo "#!/bin/sh" > /entrypoint.sh && \
    echo "set -e" >> /entrypoint.sh && \
    echo "exec /opt/bin/${APP_NAME} \$@" >> /entrypoint.sh && \
  chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
