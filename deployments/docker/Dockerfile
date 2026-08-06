# syntax=docker/dockerfile:1

FROM node:24-alpine AS frontend-build
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.26-alpine AS backend-build
WORKDIR /src/backend
COPY backend/go.mod ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-build /src/backend/internal/webui/dist ./internal/webui/dist
ARG VERSION=dev
RUN CGO_ENABLED=0 go test ./... && \
    CGO_ENABLED=0 go build -tags production -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/draftmeld ./cmd/draftmeld

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
COPY --from=backend-build /out/draftmeld /draftmeld
VOLUME ["/data"]
EXPOSE 8080
ENV DRAFTMELD_ADDRESS=:8080
ENTRYPOINT ["/draftmeld"]
