# syntax=docker/dockerfile:1.7
ARG BUILDER_IMAGE=golang:1.23.2-alpine
ARG BASE_IMAGE=alpine:3.20

# ---------- Stage 1: Builder (Compilers and build-time caches live ONLY here) ----------
FROM ${BUILDER_IMAGE} AS builder

WORKDIR /build

COPY go.mod ./
RUN if [ -f go.sum ]; then cp go.sum .; fi
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download || true

COPY . .

ARG GIT_SHA=unknown
ARG BUILD_DATE=unknown
ARG VERSION=1.0.0

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${GIT_SHA} -X main.date=${BUILD_DATE}" \
      -o /build/bin/platform-svc ./src/main.go

# ---------- Stage 2: Runtime (Hardened minimal non-root image) ----------
FROM ${BASE_IMAGE} AS runtime

ARG UID=10001
ARG GID=10001

ENV HOME=/tmp

COPY --from=builder --chown=${UID}:${GID} /build/bin/platform-svc /app/platform-svc

ARG GIT_SHA=unknown
ARG BUILD_DATE=unknown
LABEL org.opencontainers.image.source="https://github.com/knowledge-graph-platform/knowledge-graph-platform" \
      org.opencontainers.image.revision="${GIT_SHA}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="knowledge-graph-platform"

USER ${UID}:${GID}
WORKDIR /app
EXPOSE 8000 9090

ENTRYPOINT ["/app/platform-svc"]
CMD ["server"]
