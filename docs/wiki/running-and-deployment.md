# 运行与部署

## 目录

1. [环境要求](#1-环境要求)
2. [环境变量参考](#2-环境变量参考)
3. [本地开发](#3-本地开发)
4. [生产部署](#4-生产部署)
5. [质量检查](#5-质量检查)
6. [Docker 部署](#6-docker-部署)
7. [常见问题](#7-常见问题)

---

## 1. 环境要求

### 运行时依赖

| 组件    | 版本要求 | 说明           |
| ------- | -------- | -------------- |
| Go      | >= 1.25  | 后端编译运行   |
| Redis   | >= 5.x   | 缓存和速率限制 |
| Node.js | >= 20    | 前端构建       |
| npm     | >= 10    | 前端包管理     |

### 可选依赖

| 组件          | 用途             |
| ------------- | ---------------- |
| S3 兼容存储   | 上传文件存储后端 |
| WebDAV 服务器 | 上传文件存储后端 |

---

## 2. 环境变量参考

**文件**: [.env.example](file:///d:/Works/MyProject/OmePic/.env.example)

### 启动环境变量

后端启动环境变量保留无法在打开 SQLite 前读取或属于启动安全密钥的配置：

| 变量                 | 默认值                     | 说明                               |
| -------------------- | -------------------------- | ---------------------------------- |
| `HTTP_ADDR`          | `:8080`                    | HTTP 监听地址                      |
| `DATABASE_PATH`      | `data/omepic.db`           | SQLite 数据库文件路径              |
| `REDIS_URL`          | `redis://localhost:6379/0` | Redis 连接 URL                     |
| `UID_PREFIX`         | `omeo_`                    | UID 明文前缀（尾部下划线会规范化） |
| `UID_ENCRYPTION_KEY` | `change-me-uid-secret`     | UID XOR 加密密钥                   |
| `JWT_SECRET`         | `change-me-too`            | JWT 签名密钥                       |

存储配置、公开访问基准 URL、上传策略、AVIF 参数、维护模式、限流、管理员密码和 Cloudflare 图片 URL 清理配置均保存在 SQLite。Cloudflare API Token 是 admin-only runtime secret：SQLite 保存真实值，后台读取时返回遮罩值。首次登录或首次修改密码时，如果尚无密码哈希，程序会写入默认 `admin123` 的 bcrypt 哈希；登录后可在管理端设置页修改密码。

---

## 3. 本地开发

### 3.1 后端开发

```powershell
# 1. 配置环境变量（PowerShell）
$env:HTTP_ADDR=":8080"
$env:DATABASE_PATH="data/omepic.db"
$env:REDIS_URL="redis://localhost:6379/0"
$env:UID_PREFIX="omeo_"
$env:UID_ENCRYPTION_KEY="dev-uid-secret"
$env:JWT_SECRET="dev-jwt-secret"

# 或者从 .env.example 复制并加载

# 2. 启动 Redis
# Windows: 使用 WSL 或 Docker
docker run -d -p 6379:6379 redis:7-alpine

# 3. 启动后端
cd backend
go mod tidy
go run ./cmd/server
```

后端将在 `http://localhost:8080` 启动。首次启动会自动创建 SQLite 数据库并建表。

### 3.2 前端开发

```powershell
cd frontend
npm install
npm run dev
```

前端将在 `http://localhost:3000` 启动。开发/构建如需覆盖 API 基准地址，使用当前 Vite 配置约定；生产单端口部署默认走同源相对路径。

---

## 4. 生产部署

### 4.1 单端口部署（推荐）

前端静态导出并嵌入后端，统一端口提供服务：

```powershell
# 1. 构建前端并复制到后端
cd frontend
npm run build:backend
# 这会生成 frontend/out/ 并复制到 backend/web/

# 2. 构建后端
cd ../backend
go build -o server.exe ./cmd/server

# 3. 运行
./server.exe
```

此时：

- 访问 `http://host:8080/` → 前端首页
- 访问 `http://host:8080/admin/dashboard` → 管理后台
- 访问 `http://host:8080/health` → API 健康检查
- 访问 `http://host:8080/v1/image` → API 上传

### 4.2 分离部署

适合开发或需要独立扩展的场景：

```powershell
# 终端 1: 后端
cd backend
go build -o server.exe ./cmd/server
./server.exe

# 终端 2: 前端
cd frontend
npm run dev
```

前端开发服务器请求代理到后端。

---

## 5. 质量检查

### 后端

```powershell
cd backend

# 运行所有测试
go test ./...

# 编译检查
go build ./cmd/server

# Vet 检查
go vet ./...
```

### 前端

```powershell
cd frontend

# ESLint
npm run lint

# TypeScript 类型检查
npm run typecheck

# 单元测试
npm run test

# 构建
npm run build

# 构建（含复制到 backend/web/）
npm run build:backend
```

---

## 6. Docker 部署

### 使用 Docker Compose

创建 `docker-compose.yml`（位于项目根目录）：

```yaml
version: "3.8"

services:
  redis:
    image: redis:7-alpine
    restart: unless-stopped
    volumes:
      - redis-data:/data

  app:
    build: .
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - HTTP_ADDR=:8080
      - DATABASE_PATH=/data/omepic.db
      - REDIS_URL=redis://redis:6379/0
      - UID_PREFIX=omeo_
      - UID_ENCRYPTION_KEY=secure-uid-key
      - JWT_SECRET=secure-jwt-secret
      # Cloudflare image URL cache purge is configured in the admin runtime settings UI.
    volumes:
      - app-data:/data
      - app-images:/data/images
    depends_on:
      - redis

volumes:
  redis-data:
  app-data:
  app-images:
```

创建 `Dockerfile`（位于项目根目录）：

```dockerfile
# 构建阶段 1: 前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build:backend

# 构建阶段 2: 后端
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend-builder /app/backend/web/ ./web/
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

# 运行阶段
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=backend-builder /server /server
EXPOSE 8080
CMD ["/server"]
```

然后运行：

```powershell
docker-compose up -d
```

---

## 7. 常见问题

### Q: 启动时报数据库迁移错误

如果存在旧版 SQLite 文件（包含已删除的字段如 `original_filename`），删除数据库文件重新创建即可：

```powershell
rm backend/data/omepic.db
go run ./cmd/server
```

### Q: Redis 连接失败

确保 Redis 已启动并可访问。默认连接 `redis://localhost:6379/0`。

### Q: 图片上传后返回 503

检查：

1. 存储后端配置是否正确（文件系统权限？S3 凭证？）
2. `data/images/` 目录是否存在且可写
3. Redis 是否正常运行

### Q: 生产环境中如何配置 HTTPS

推荐使用反向代理（如 Nginx、Caddy）：

```
server {
    listen 443 ssl;
    server_name your.domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

公开访问基准 URL 不再通过环境变量设置。登录管理端后进入运行设置，将 Public URL 设置为 `https://your.domain.com`。当前启动配置不读取可信代理 CIDR 或真实 IP header 环境变量，默认不信任转发头。

### Q: 如何切换默认存储后端

1. 登录管理后台
2. 进入 Settings → Storage
3. 创建或选择目标存储实例
4. 设置为默认存储

存储配置保存在 SQLite `storage_configs`，不再通过环境变量切换。

### Q: 前端白屏/路由不工作

确保：

1. 生产模式下 `backend/web/` 目录存在（执行了 `npm run build:backend`）
2. 开发模式下确认前端 API 基准地址指向正在运行的后端；生产单端口部署默认使用相对路径

### Q: 如何备份数据

需要备份的文件和数据库：

1. SQLite 数据库文件（默认 `data/omepic.db`）
2. 本地图片文件（默认 `data/images/`）
3. S3/WebDAV 上的图片文件（如使用远程存储）

Redis 中的数据可在重启后自动预热，无需单独备份。
