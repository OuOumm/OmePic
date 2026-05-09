# 根据FRONTEND_FEATURES_PROMPT.md重建OmePic图床前端

## Goal

根据项目根目录的 `FRONTEND_FEATURES_PROMPT.md` 完整重建 OmePic 图床前端，支持图片上传（文件/拖拽/粘贴/URL）、本地上传历史、API示例页、管理后台（状态/图片管理/存储设置）、中英文切换、主题切换，构建为静态HTML产物由Gin后端直接serve。

## What I already know

* 需求文档：`FRONTEND_FEATURES_PROMPT.md` 已详细描述所有功能需求
* 原前端技术栈（已删除）：Next.js 16.2.4 + React 19 + TypeScript + Tailwind CSS + shadcn/ui (new-york) + Zustand + react-hot-toast + Lucide icons
* 后端：Go/Gin，静态文件目录 `backend/web/`，已有旧版构建产物
* 构建目标：静态HTML导出（`output: "export"`），单端口同域部署
* API基础路径：开发环境 `http://localhost:8080`，生产环境相对路径
* 项目规范：`.trellis/spec/frontend/` 下有完整的组件/状态/目录/类型规范
* 原前端目录结构：`frontend/src/app/`（路由）、`frontend/src/features/`（功能组件）、`frontend/src/components/ui/`（shadcn/ui）、`frontend/src/components/shared/`（共享组件）、`frontend/src/stores/`（Zustand）、`frontend/src/lib/`（工具库）、`frontend/src/types/`（类型定义）

## Decision (ADR-lite)

**Context**: 用户提到"Nuxt UI+next"，需确认前端框架
**Decision**: 使用 Next.js 16 + shadcn/ui (React)，与原项目技术栈一致
**Consequences**: 可复用 `.trellis/spec/frontend/` 现有规范，社区生态成熟，静态导出支持好

## Requirements (evolving)

* 上传首页：文件选择/拖拽/粘贴/URL四种上传方式，上传进度，结果展示，最近10条记录
* 本地历史页：IndexedDB读取，预览，复制，按token删除，清空
* API示例页：三个核心接口的curl示例
* 管理后台：登录、状态概览、图片管理（搜索/分页/网格列表切换/批量删除）、存储设置（local/S3/WebDAV CRUD）
* 全局偏好：中英文、浅色/深色/系统主题、客户端token、持久化

## Acceptance Criteria (evolving)

* [ ] 首页可通过文件、拖拽、粘贴、URL四种方式上传图片
* [ ] 上传时有进度，成功后有最近上传、复制链接和预览
* [ ] 用户可选择公开存储目标
* [ ] 历史页能查看、清空、预览、复制、按token删除
* [ ] API页展示三个核心接口示例
* [ ] 管理员可登录、验证会话、退出
* [ ] 状态页展示系统统计
* [ ] 图片管理页支持搜索、分页、网格/列表、选择、批量删除、预览
* [ ] 设置页支持local/S3/WebDAV存储实例的创建、编辑、删除和默认切换
* [ ] 支持中英文、浅色/深色/系统主题
* [ ] 所有失败路径有用户可见错误反馈
* [ ] 构建为静态HTML，由Gin单端口serve

## Out of Scope (explicit)

* 后端代码修改
* E2E测试
* PWA/Service Worker
* 多用户/账号体系
* 大文件分片上传（>100MB由后端限制）
* 并发多文件上传（保持单文件依次上传）
* 移动端原生App

## Implementation Plan

鉴于需求文档极其详尽，采用**一次性完整实现**策略（非分PR），按以下顺序构建文件：

1. **项目脚手架**：package.json, tsconfig, tailwind, postcss, next.config, eslint, components.json
2. **基础层**：types/, lib/ (api, indexeddb, i18n, preferences), stores/
3. **UI原语**：components/ui/ (shadcn/ui primitives: Button, Card, Input, etc.)
4. **共享组件**：components/shared/ (AppHeader, ImageLightbox, ImgStyleImageCard, etc.)
5. **功能页面**：features/upload/, features/history/, features/admin/
6. **路由页面**：app/ (layout, page, history, api, admin)
7. **构建验证**：npm run typecheck, npm run lint, npm run build, 产物复制到 backend/web/

## Technical Notes

* 原 `frontend/` 目录已被删除（git status显示所有文件为 D）
* `backend/web/` 仍有旧构建产物
* `.trellis/spec/frontend/` 规范文档完整，应作为实现参考
* 原 `package.json` 依赖：next 16.2.4, react 19, zustand 5, react-hot-toast 2, lucide-react 0.468, tailwindcss 3.4, class-variance-authority 0.7, clsx 2.1, tailwind-merge 3.3
