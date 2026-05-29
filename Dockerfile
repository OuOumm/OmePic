# Stage 1: Build frontend static files, copy into backend/web/
FROM node:24-alpine AS frontend-build
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build:backend

# Stage 2: Build Go backend (needs libavif for AVIF conversion)
FROM golang:1.25-alpine AS backend-build
RUN apk add --no-cache libavif-dev
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-build /app/backend/web ./web/
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -trimpath -o /server ./cmd/server/

# Stage 3: Minimal runtime (~12 MB)
FROM alpine:3.21
RUN apk add --no-cache ca-certificates libavif
COPY --from=backend-build /server /usr/local/bin/server
COPY --from=backend-build /app/web /opt/omepic/web
WORKDIR /opt/omepic
EXPOSE 8080
ENTRYPOINT ["server"]