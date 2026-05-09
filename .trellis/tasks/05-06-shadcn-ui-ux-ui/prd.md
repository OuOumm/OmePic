# 使用 shadcn/ui 重构整个前端 UX-UI 界面

## Goal

使用 shadcn/ui 组件体系全面重构 OmePic 图床前端，抛弃现有自定义设计，打造精美、现代、具有 Glassmorphism 2.0（毛玻璃+深度层次）风格的界面。目标是让图床工具看起来专业、精致，同时保持良好的可用性和无障碍访问。

## What I already know

* 当前是 Next.js 16 + React 19 + Tailwind CSS 3.4
* 已有 shadcn/ui New York 风格基础配置（components.json）
* 已有 12 个基础 UI 组件（Button、Card、Dialog、DropdownMenu、Input、Label、Select、Separator、Table、Tabs、Badge、Textarea）
* 使用 Zustand 做状态管理，react-hot-toast 做通知
* 6 个页面：首页上传、历史记录、API 文档、管理后台（仪表盘/图片管理/设置）
* 当前自定义了玻璃态面板（glass-panel）、渐变文字等 CSS 工具类
* 图标库为 lucide-react
* 支持中英文双语、亮暗主题

## Research References

* [`research/design-trends.md`](research/design-trends.md) — 2026 设计趋势：Glassmorphism 2.0 是主流，Neumorphism 仅适合局部点缀；图片托管站点通用模式分析
* [`research/shadcn-ecosystem.md`](research/shadcn-ecosystem.md) — shadcn/ui v4 注册表 54 个组件、Aceternity UI 106 个组件、Magic UI 74 个组件；组件优先级推荐

## Design Direction (from research)

**主风格：Dark Glassmorphism（深色毛玻璃+深度层次）**
- 深色背景（slate-900/950）+ 半透明毛玻璃面板
- violet-cyan 渐变光晕作为背景氛围
- `backdrop-blur-xl`（20-24px 模糊）+ 半透明背景 + 渐变边框
- 亮色模式作为辅助（也采用毛玻璃，但透明度更高、背景更亮）
- 类似 macOS Liquid Glass 的生产级质感

**辅助点缀：Soft Spatial UI**
- 柔和阴影、浮动元素、充裕留白
- 微妙的悬浮提升效果
- Bento Grid 布局用于管理后台仪表盘

**色彩方案：深色为主，violet 主色调**
- 默认深色主题，亮色为可选切换
- violet（#7C3AED）作为 primary 保持不变
- 深色背景 slate-900/950，亮色背景 slate-50
- 渐变光晕：violet → cyan → blue 微妙的 aurora 背景效果

## Requirements

### 设计系统层面
* 统一采用 Glassmorphism 2.0 设计语言
* 所有卡片/面板使用毛玻璃效果（`backdrop-blur` + 半透明背景 + 渐变边框）
* 深色模式使用 Dark Glassmorphism 风格
* 增加微交互和过渡动画（丰富级别：页面过渡、悬浮效果、进度动画、成功庆祝、骨架屏脉冲、按钮波纹）
* 动画使用纯 CSS animation + Tailwind transition，不引入 Framer Motion（保持轻量）
* 统一间距、圆角、阴影系统

### 组件层面
* 新增 shadcn/ui 组件：Skeleton、Progress、Tooltip、Sheet、ScrollArea、AlertDialog、Toggle、Popover、Collapsible
* 用 shadcn Sonner 替换 react-hot-toast
* 重构所有自定义组件使用 shadcn/ui 基础组件
* 动画使用纯 CSS animation + Tailwind transition，不引入 Framer Motion（保持轻量）

### 页面层面（保持功能不变，仅重构 UI）
* **首页上传**：毛玻璃上传区域 + 渐变背景 + 上传进度动画 + 成功庆祝效果
* **历史记录**：毛玻璃图片卡片 + 骨架屏加载 + 悬浮预览
* **API 文档**：毛玻璃代码块 + 更好的排版
* **管理后台**：毛玻璃侧边栏 + Bento Grid 仪表盘 + 数据表格增强

### 保持不变的
* 所有现有功能完整
* 路由结构不变
* 数据流/状态管理不变（Zustand + IndexedDB）
* API 调用层不变
* 中英文双语支持
* 亮暗主题支持
* 响应式设计

## Acceptance Criteria

* [ ] 所有页面视觉风格统一为 Glassmorphism 2.0
* [ ] 新增 shadcn/ui 组件正常工作
* [ ] Sonner toast 替换 react-hot-toast 完成
* [ ] 亮暗主题下毛玻璃效果均正常
* [ ] 移动端响应式正常
* [ ] 所有现有功能正常运作
* [ ] 无 TypeScript 类型错误
* [ ] 无 ESLint 错误
* [ ] 动画流畅（60fps），移动端不卡顿

## Decision (ADR-lite)

**Context**: 需要在 Glassmorphism 和 Neumorphism 之间选择主导设计风格，以及亮色/暗色主题策略
**Decision**: 
1. 采用 Dark Glassmorphism 作为主风格（默认深色主题），亮色模式作为辅助
2. Neumorphism 仅用于局部点缀（如开关、滑块等次要控件）
3. 丰富动画方案：纯 CSS animation + Tailwind transition，不引入 Framer Motion
**Consequences**:
- 毛玻璃效果需要有趣的背景才能展现（需添加渐变/纹理背景）
- `backdrop-filter` 创建独立合成层，需控制每视口玻璃元素数量（≤5个）以保证滚动性能
- 图片密集型页面（首页、历史记录）需特别注意性能

## Out of Scope (explicit)

* 后端 API 修改
* 新增功能
* 路由结构调整
* 数据流/状态管理重构
* Tailwind CSS v3 → v4 升级（风险过大，单独评估）
* Framer Motion 引入（可选，如不使用 Aceternity/Magic UI 组件则不需要）
* Aceternity UI / Magic UI 第三方组件（保持纯 shadcn/ui 官方组件）

## Technical Notes

* 前端目录：`D:\Works\MyProject\OmePic\frontend`
* 当前 shadcn/ui 配置：New York 风格，slate 基调，CSS 变量模式
* Tailwind CSS 3.4.17（不升级到 v4，避免兼容性风险）
* 当前 CSS 变量使用 RGB 值，shadcn v4 使用 HSL — 保持 RGB 不变
* 现有 `.glass-panel` 类只是带样式的卡片（无 backdrop-blur），需要重写为真正的毛玻璃效果
* 毛玻璃效果需要页面有背景内容才能显示 — 需在 body/html 层级添加渐变/纹理背景
* 每视口玻璃元素限制 3-5 个以保证图片密集页面的滚动性能