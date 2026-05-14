# OmePic

OmePic 是一个图床服务：Go 后端、SQLite 持久化、Redis 缓存、SvelteKit 前端，支持本地/S3/WebDAV 运行时存储配置和 AVIF 输出。

## 环境变量

启动环境变量只保留不可运行时修改或启动必需的 6 项，示例见 `.env.example`：

```env
HTTP_ADDR=:8080
DATABASE_PATH=data/omepic.db
REDIS_URL=redis://localhost:6379/0
UID_PREFIX=omeo_
UID_ENCRYPTION_KEY=change-me-uid-secret
JWT_SECRET=change-me-too
```

存储配置、公开访问基准 URL、上传策略、维护模式、限流和管理员密码均保存在 SQLite。首次登录会在 `config.admin_password_hash` 写入默认管理员密码 `admin123` 的 bcrypt 哈希；登录后可在管理端设置页修改密码。

## 本地运行

```powershell
cd backend
go run ./cmd/server
```

前端开发：

```powershell
cd frontend
npm install
npm run dev
```

单端口部署构建：

```powershell
cd frontend
npm run build:backend
```

主要管理接口包括 `POST /admin/login`、`PUT /admin/password`、`GET|PUT /admin/system-settings` 和存储配置相关 `/admin/config/*`。
