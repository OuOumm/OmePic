# Stage 1: Build frontend static files, copy into backend/web/
FROM node:24-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build:backend

# Stage 2: Build Go backend (lilliput requires CGO + C libraries)
FROM golang:1.25-alpine AS backend-build
ARG TARGETARCH
RUN apk add --no-cache build-base
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
RUN set -eux; \
    moddir="$(go list -m -f '{{.Dir}}' github.com/discord/lilliput)"; \
    case "${TARGETARCH:-amd64}" in \
      amd64) dep_arch="amd64" ;; \
      arm64) dep_arch="aarch64" ;; \
      *) echo "unsupported TARGETARCH=${TARGETARCH}"; exit 1 ;; \
    esac; \
    inc="$moddir/deps/linux/$dep_arch/include"; \
    chmod -R u+w "$inc"; \
    ln -sf libpng16/png.h "$inc/png.h"; \
    ln -sf libpng16/pngconf.h "$inc/pngconf.h"; \
    ln -sf libpng16/pnglibconf.h "$inc/pnglibconf.h"
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