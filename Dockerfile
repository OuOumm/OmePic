# Stage 1: Build frontend static files, copy into backend/web/
FROM node:24-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build:backend

# Stage 2: Build Go backend (nodynamic: pure-Go AVIF, zero C deps)
FROM golang:1.25-alpine AS backend-build
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-build /app/backend/web ./web/
RUN CGO_ENABLED=0 go build -tags nodynamic -ldflags="-s -w" -trimpath -o /server ./cmd/server/

# Stage 3: Minimal runtime — scratch, ~0 MB overhead
FROM scratch
COPY --from=backend-build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend-build /server /opt/omepic/server
COPY --from=backend-build /app/web /opt/omepic/web
WORKDIR /opt/omepic
EXPOSE 8080
ENTRYPOINT ["./server"]