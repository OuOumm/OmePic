# Stage 1: Build frontend static files, copy into backend/web/
FROM node:24-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build:backend

# Stage 2: Build Go backend (lilliput requires CGO + C libraries)
FROM golang:1.25-alpine AS backend-build
RUN apk add --no-cache \
    libavif-dev libwebp-dev libjpeg-turbo-dev \
    libpng16-dev giflib-dev pkgconf build-base && \
    ln -sf /usr/include/libpng16/png.h /usr/include/png.h && \
    ln -sf /usr/include/libpng16/pngconf.h /usr/include/pngconf.h && \
    ln -sf /usr/include/libpng16/pnglibconf.h /usr/include/pnglibconf.h
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-build /app/backend/web ./web/
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -trimpath -o /server ./cmd/server/

# Stage 3: Runtime with C libraries (~30 MB)
FROM alpine:3.21
RUN apk add --no-cache ca-certificates \
    libavif libwebp libjpeg-turbo libpng giflib
COPY --from=backend-build /server /opt/omepic/server
COPY --from=backend-build /app/web /opt/omepic/web
WORKDIR /opt/omepic
EXPOSE 8080
ENTRYPOINT ["./server"]