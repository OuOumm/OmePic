FROM node:24-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build:backend

FROM golang:1.25-alpine AS backend-build
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-build /app/backend/web ./web/
RUN CGO_ENABLED=1 go build -o /omepic ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates libavif
COPY --from=backend-build /omepic /usr/local/bin/omepic
COPY --from=backend-build /app/backend/web /opt/omepic/web

WORKDIR /opt/omepic
EXPOSE 8080

ENTRYPOINT ["omepic"]