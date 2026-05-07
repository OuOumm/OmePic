import type { Language } from "@/types";

type TranslationMap = Record<string, string>;

const en: TranslationMap = {
  // Nav
  "nav.upload": "Upload",
  "nav.history": "History",
  "nav.api": "API",
  "nav.admin": "Admin",

  // Upload page
  "upload.title": "Upload Image",
  "upload.dropzone": "Click or drag & drop an image here",
  "upload.supportedFormats": "Supports AVIF, PNG, JPG, JPEG, GIF, WEBP, BMP",
  "upload.chooseFile": "Choose File",
  "upload.urlLabel": "Image URL",
  "upload.urlPlaceholder": "https://example.com/image.jpg",
  "upload.urlUpload": "Upload from URL",
  "upload.uploading": "Uploading...",
  "upload.success": "Upload successful!",
  "upload.duplicate": "Duplicate image detected (already uploaded before)",
  "upload.error": "Upload failed",
  "upload.recentTitle": "Recent Uploads",
  "upload.noRecent": "No recent uploads",
  "upload.progress": "Uploading: {pct}%",
  "upload.pasting": "Paste an image here (Ctrl+V)",
  "upload.pasteHint": "Or press Ctrl+V to paste an image",
  "upload.noClipboard": "No image found in clipboard",
  "upload.multiUploading": "Uploading {current}/{total} files",
  "upload.multiSuccess": "{count} files uploaded",
  "upload.multiPartial": "{success}/{total} files uploaded",
  "upload.invalidUrl": "Invalid URL. Must start with http:// or https://",
  "upload.urlDownloadFail": "Failed to download image from URL",
  "upload.urlNotImage": "URL does not point to a supported image type",
  "upload.urlSuccess": "Image downloaded from URL, starting upload...",
  "upload.storageLabel": "Storage Location",

  // History page
  "history.title": "Upload History",
  "history.count": "{count} records",
  "history.clear": "Clear History",
  "history.clearConfirm": "Clear all local upload history?",
  "history.cleared": "History cleared",
  "history.empty": "No upload history",
  "history.delete": "Delete",
  "history.deleteConfirm": "Delete this image?",
  "history.deleted": "Image deleted",
  "history.deleteError": "Failed to delete image",
  "history.viewPreview": "Open preview",

  // API page
  "api.title": "API Examples",

  // Common
  "common.copyUrl": "Copy URL",
  "common.copyMarkdown": "Copy Markdown",
  "common.copyBBCode": "Copy BBCode",
  "common.copied": "Copied!",
  "common.close": "Close",
  "common.token": "Client Token",
  "common.copyToken": "Copy Token",
  "common.download": "Download",
  "common.success": "Success",
  "common.error": "Error",
  "common.loading": "Loading...",
  "common.save": "Save",
  "common.cancel": "Cancel",
  "common.edit": "Edit",
  "common.create": "Create",
  "common.delete": "Delete",
  "common.confirm": "Confirm",
  "common.search": "Search",
  "common.refresh": "Refresh",
  "common.language": "Language",
  "common.theme": "Theme",
  "common.themeLight": "Light",
  "common.themeDark": "Dark",
  "common.themeSystem": "System",
  "common.storage": "Storage",
  "common.default": "Default",
  "common.openPreview": "Open Preview of {title}",

  // Image info
  "image.uid": "UID",
  "image.size": "Size",
  "image.type": "Type",
  "image.created": "Created",
  "image.storageKey": "Storage Key",
  "image.storageBackend": "Storage Backend",
  "image.token": "Token",
  "image.md5": "MD5",
  "image.duplicate": "Duplicate",
  "image.filename": "Filename",

  // Admin
  "admin.login": "Admin Login",
  "admin.password": "Password",
  "admin.loginBtn": "Login",
  "admin.loggingIn": "Logging in...",
  "admin.loginSuccess": "Login successful",
  "admin.loginError": "Login failed",
  "admin.logout": "Logout",
  "admin.sessionExpired": "Session expired, please login again",
  "admin.sidebarStatus": "Status",
  "admin.sidebarImages": "Images",
  "admin.sidebarSettings": "Settings",

  // Admin status
  "admin.statusTitle": "System Status",
  "admin.totalImages": "Total Images",
  "admin.totalSize": "Total Size",
  "admin.todayUploads": "Today's Uploads",
  "admin.uniqueTokens": "Unique Tokens",

  // Admin images
  "admin.imagesTitle": "Image Management",
  "admin.imagesSearch": "Search by UID...",
  "admin.imagesTotal": "{total} images",
  "admin.imagesSelected": "{count} selected",
  "admin.imagesSelectAll": "Select Page",
  "admin.imagesDeselectAll": "Deselect Page",
  "admin.imagesDelete": "Delete Selected",
  "admin.imagesDeleteConfirm": "Delete {count} selected images?",
  "admin.imagesDeleted": "Deleted {count} images",
  "admin.imagesDeleteError": "Failed to delete images",
  "admin.imagesGridView": "Grid",
  "admin.imagesListView": "List",
  "admin.imagesPrev": "Previous",
  "admin.imagesNext": "Next",
  "admin.imagesPage": "Page {page}",
  "admin.imagesViewPreview": "View Preview",

  // Admin settings
  "admin.settingsTitle": "Storage Settings",
  "admin.settingsNew": "New Instance",
  "admin.settingsName": "Name",
  "admin.settingsBackend": "Backend",
  "admin.settingsKey": "Storage Key",
  "admin.settingsDefault": "Default",
  "admin.settingsSetDefault": "Set as Default",
  "admin.settingsDelete": "Delete Instance",
  "admin.settingsDeleteConfirm": "Delete storage instance \"{name}\"?",
  "admin.settingsDeleted": "Instance deleted",
  "admin.settingsSaved": "Settings saved",
  "admin.settingsCreated": "Instance created",
  "admin.settingsLocalPath": "Local Path",
  "admin.settingsS3Endpoint": "Endpoint",
  "admin.settingsS3Region": "Region",
  "admin.settingsS3Bucket": "Bucket",
  "admin.settingsS3AccessKey": "Access Key",
  "admin.settingsS3SecretKey": "Secret Key",
  "admin.settingsS3SSL": "Use SSL",
  "admin.settingsS3PathStyle": "Force Path Style",
  "admin.settingsWebdavUrl": "URL",
  "admin.settingsWebdavUser": "User",
  "admin.settingsWebdavPassword": "Password",
};

const zh: TranslationMap = {
  // Nav
  "nav.upload": "上传",
  "nav.history": "历史",
  "nav.api": "API",
  "nav.admin": "管理",

  // Upload page
  "upload.title": "上传图片",
  "upload.dropzone": "点击或拖拽图片到此处",
  "upload.supportedFormats": "支持 AVIF、PNG、JPG、JPEG、GIF、WEBP、BMP",
  "upload.chooseFile": "选择文件",
  "upload.urlLabel": "图片 URL",
  "upload.urlPlaceholder": "https://example.com/image.jpg",
  "upload.urlUpload": "从 URL 上传",
  "upload.uploading": "正在上传...",
  "upload.success": "上传成功！",
  "upload.duplicate": "检测到重复图片（已上传过）",
  "upload.error": "上传失败",
  "upload.recentTitle": "最近上传",
  "upload.noRecent": "暂无上传记录",
  "upload.progress": "上传中：{pct}%",
  "upload.pasting": "在此处粘贴图片（Ctrl+V）",
  "upload.pasteHint": "或按 Ctrl+V 粘贴图片",
  "upload.noClipboard": "剪贴板中没有图片",
  "upload.multiUploading": "正在上传 {current}/{total} 个文件",
  "upload.multiSuccess": "{count} 个文件上传成功",
  "upload.multiPartial": "{success}/{total} 个文件上传成功",
  "upload.invalidUrl": "无效的 URL，必须以 http:// 或 https:// 开头",
  "upload.urlDownloadFail": "从 URL 下载图片失败",
  "upload.urlNotImage": "URL 指向的文件不是支持的图片格式",
  "upload.urlSuccess": "已从 URL 下载图片，开始上传...",
  "upload.storageLabel": "存储位置",

  // History page
  "history.title": "上传历史",
  "history.count": "{count} 条记录",
  "history.clear": "清空历史",
  "history.clearConfirm": "确认清空所有本地上传历史？",
  "history.cleared": "历史已清空",
  "history.empty": "暂无上传历史",
  "history.delete": "删除",
  "history.deleteConfirm": "确认删除该图片？",
  "history.deleted": "图片已删除",
  "history.deleteError": "删除失败",
  "history.viewPreview": "打开预览",

  // API page
  "api.title": "API 示例",

  // Common
  "common.copyUrl": "复制 URL",
  "common.copyMarkdown": "复制 Markdown",
  "common.copyBBCode": "复制 BBCode",
  "common.copied": "已复制！",
  "common.close": "关闭",
  "common.token": "客户端 Token",
  "common.copyToken": "复制 Token",
  "common.download": "下载",
  "common.success": "操作成功",
  "common.error": "操作失败",
  "common.loading": "加载中...",
  "common.save": "保存",
  "common.cancel": "取消",
  "common.edit": "编辑",
  "common.create": "新建",
  "common.delete": "删除",
  "common.confirm": "确认",
  "common.search": "搜索",
  "common.refresh": "刷新",
  "common.language": "语言",
  "common.theme": "主题",
  "common.themeLight": "浅色",
  "common.themeDark": "深色",
  "common.themeSystem": "跟随系统",
  "common.storage": "存储",
  "common.default": "默认",
  "common.openPreview": "打开 {title} 的预览",

  // Image info
  "image.uid": "UID",
  "image.size": "大小",
  "image.type": "类型",
  "image.created": "创建时间",
  "image.storageKey": "存储标识",
  "image.storageBackend": "存储后端",
  "image.token": "Token",
  "image.md5": "MD5",
  "image.duplicate": "重复",
  "image.filename": "文件名",

  // Admin
  "admin.login": "管理员登录",
  "admin.password": "密码",
  "admin.loginBtn": "登录",
  "admin.loggingIn": "登录中...",
  "admin.loginSuccess": "登录成功",
  "admin.loginError": "登录失败",
  "admin.logout": "退出登录",
  "admin.sessionExpired": "会话已过期，请重新登录",
  "admin.sidebarStatus": "状态",
  "admin.sidebarImages": "图片",
  "admin.sidebarSettings": "设置",

  // Admin status
  "admin.statusTitle": "系统状态",
  "admin.totalImages": "图片总数",
  "admin.totalSize": "总存储大小",
  "admin.todayUploads": "今日上传",
  "admin.uniqueTokens": "独立 Token 数",

  // Admin images
  "admin.imagesTitle": "图片管理",
  "admin.imagesSearch": "按 UID 搜索...",
  "admin.imagesTotal": "共 {total} 张",
  "admin.imagesSelected": "已选 {count} 张",
  "admin.imagesSelectAll": "选择当前页",
  "admin.imagesDeselectAll": "取消当前页选择",
  "admin.imagesDelete": "删除选中",
  "admin.imagesDeleteConfirm": "确认删除 {count} 张选中的图片？",
  "admin.imagesDeleted": "已删除 {count} 张图片",
  "admin.imagesDeleteError": "删除图片失败",
  "admin.imagesGridView": "网格",
  "admin.imagesListView": "列表",
  "admin.imagesPrev": "上一页",
  "admin.imagesNext": "下一页",
  "admin.imagesPage": "第 {page} 页",
  "admin.imagesViewPreview": "查看",

  // Admin settings
  "admin.settingsTitle": "存储设置",
  "admin.settingsNew": "新建实例",
  "admin.settingsName": "名称",
  "admin.settingsBackend": "存储后端",
  "admin.settingsKey": "存储标识",
  "admin.settingsDefault": "默认",
  "admin.settingsSetDefault": "设为默认",
  "admin.settingsDelete": "删除实例",
  "admin.settingsDeleteConfirm": "确认删除存储实例 \"{name}\"？",
  "admin.settingsDeleted": "实例已删除",
  "admin.settingsSaved": "设置已保存",
  "admin.settingsCreated": "实例已创建",
  "admin.settingsLocalPath": "本地路径",
  "admin.settingsS3Endpoint": "Endpoint",
  "admin.settingsS3Region": "区域",
  "admin.settingsS3Bucket": "Bucket",
  "admin.settingsS3AccessKey": "Access Key",
  "admin.settingsS3SecretKey": "Secret Key",
  "admin.settingsS3SSL": "使用 SSL",
  "admin.settingsS3PathStyle": "Force Path Style",
  "admin.settingsWebdavUrl": "URL",
  "admin.settingsWebdavUser": "用户",
  "admin.settingsWebdavPassword": "密码",
};

const translations: Record<Language, TranslationMap> = { en, zh };

// Detect browser language
export function detectLanguage(): Language {
  if (typeof navigator === "undefined") return "en";
  const lang = navigator.language.toLowerCase();
  if (lang.startsWith("zh")) return "zh";
  return "en";
}

export function t(lang: Language, key: string, params?: Record<string, string | number>): string {
  const map = translations[lang] ?? translations.en;
  let text = map[key] ?? translations.en[key] ?? key;
  if (params) {
    Object.entries(params).forEach(([k, v]) => {
      text = text.replace(`{${k}}`, String(v));
    });
  }
  return text;
}
