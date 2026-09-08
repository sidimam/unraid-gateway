# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build
ARG TARGETOS TARGETARCH VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X github.com/sidimam/unraid-gateway/internal/server.Version=${VERSION}" \
    -o /out/unraid-gateway ./cmd/unraid-gateway

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl
COPY --from=build /out/unraid-gateway /usr/local/bin/unraid-gateway
# Console walkthrough: `docker exec -it … sh` (Unraid's Console button) sources $ENV and gets the `gw` helper.
COPY console/gw /usr/local/bin/gw
COPY console/profile.sh /etc/unraid-gateway-profile.sh
ENV ENV=/etc/unraid-gateway-profile.sh
# Unraid's default "nobody:users" so files created through the gateway get
# the same ownership as files created via SMB.
USER 99:100
ENV LISTEN_ADDR=:8484 DATA_ROOT=/data
EXPOSE 8484
VOLUME ["/data"]
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s \
  CMD curl -fsS http://127.0.0.1:8484/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/unraid-gateway"]
