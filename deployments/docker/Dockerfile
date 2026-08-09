# syntax=docker/dockerfile:1

FROM node:24-alpine AS frontend-build
WORKDIR /src/frontend
COPY VERSION /src/VERSION
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.26-alpine AS backend-build
WORKDIR /src/backend
COPY VERSION /src/VERSION
COPY backend/go.mod ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-build /src/backend/internal/webui/dist ./internal/webui/dist
RUN VERSION="$(cat /src/VERSION)" && \
    CGO_ENABLED=0 go test ./... && \
    CGO_ENABLED=0 go build -tags production -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/draftmeld ./cmd/draftmeld && \
    mkdir -p /out/data

FROM debian:bookworm-slim AS runtime
RUN apt-get update && \
    apt-get install --yes --no-install-recommends ca-certificates poppler-utils tesseract-ocr tesseract-ocr-eng && \
    rm -rf /var/lib/apt/lists/* && \
    groupadd --system --gid 65532 draftmeld && \
    useradd --system --uid 65532 --gid draftmeld --home-dir /nonexistent --shell /usr/sbin/nologin draftmeld && \
    mkdir -p /data && \
    chown draftmeld:draftmeld /data
COPY --from=backend-build /out/draftmeld /draftmeld
VOLUME ["/data"]
EXPOSE 8080
ENV DRAFTMELD_ADDRESS=:8080
ENV DRAFTMELD_DATA_DIR=/data
USER draftmeld:draftmeld
ENTRYPOINT ["/draftmeld"]
