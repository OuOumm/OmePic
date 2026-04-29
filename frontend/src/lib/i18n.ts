import { languageToLocale, type Language } from "@/types/preferences";

const en = {
  common: {
    upload: "Upload",
    history: "History",
    api: "API",
    admin: "Admin",
    duplicate: "Duplicate",
    clientToken: "Client token",
    preparingToken: "Preparing token...",
    uid: "UID",
    type: "Type",
    size: "Size",
    created: "Created",
    uploaded: "Uploaded",
    token: "Token",
    md5: "MD5",
    backend: "Backend",
    select: "Select",
    previous: "Previous",
    next: "Next",
    loading: "Loading...",
    refresh: "Refresh",
    save: "Save",
    search: "Search",
    total: "Total",
    storage: "Storage",
    preview: "Preview",
    closePreview: "Close preview",
    previewMetadata: "Image metadata",
    openPreview: (title: string) => `Open preview for ${title}`,
    skipToContent: "Skip to main content",
    copyTargets: {
      url: "URL",
      markdown: "Markdown",
      bbcode: "BBCode"
    },
    items: (count: number) => `${count} items`,
    totalResults: (count: number) => `Total results: ${count}`
  },
  header: {
    tagline: "Image Hosting Service",
    languageLabel: "Language",
    themeLabel: "Theme",
    navLabel: "Primary navigation",
    openMenu: "Open menu",
    closeMenu: "Close menu",
    languages: {
      zh: "中文",
      en: "English"
    },
    themes: {
      light: "Light",
      dark: "Dark",
      system: "System"
    }
  },
  copyButton: {
    copy: (label: string) => `Copy ${label}`,
    copied: (label: string) => `${label} copied`
  },
  upload: {
    eyebrow: "Upload Flow",
    title: "Fast local image hosting with real deduplication",
    description:
      "OmePic generates a client token on first visit, converts raster uploads to AVIF with progress, and remembers your last ten results in IndexedDB.",
    recentUploads: "Recent uploads",
    status: "Upload status",
    statusIdle: "Pick a file to start.",
    statusSuccess: "Upload completed and stored as AVIF.",
    statusUploading: (progress: number) => `Uploading ${progress}%`,
    latestResult: "Latest result",
    previewAlt: "Uploaded preview",
    uploadCompleteToast: "Upload complete",
    duplicateUploadToast: "Duplicate upload linked",
    localHistorySaveFailed: "Upload succeeded, but local history could not be updated.",
    uploadFailed: "Upload failed",
    storageTarget: "Storage target",
    backendDefaultStorage: "Backend default storage",
    defaultStorageOption: (name: string, backend: string) => `Use backend default: ${name} (${backend})`,
    storageOption: (name: string, backend: string) => `${name} (${backend})`,
    storageOptionDefault: (name: string, backend: string) => `${name} (${backend}) - default`,
    storageSelectionHint: (name: string, backend: string) => `Uploads will target ${name} using ${backend}.`,
    storageOptionsFailed: "Failed to load storage options",
    dropTitle: "Drop an image here",
    dropDescription:
      "Supports PNG, JPG, JPEG, GIF, WEBP, BMP up to 20 MB. Uploads are converted to AVIF before storage, and duplicate original uploads reuse the same stored AVIF file.",
    fileInputLabel: "Choose an image file to upload",
    chooseFile: "Choose file",
    browseLocally: "Browse locally",
    emptyRecent: "No uploads yet."
  },
  historyPage: {
    eyebrow: "IndexedDB",
    title: "Local upload history",
    description:
      "Deleting here removes the logical UID only. Physical files stay in storage until a later cleanup flow removes orphaned assets.",
    clear: "Clear local history",
    empty: "No local history yet.",
    deleteUid: "Delete UID",
    deleteEnabledTitle: "Delete logical UID; physical file cleanup is deferred",
    deleteDisabledTitle: "Only the owning token can delete this logical UID",
    deleteSuccessToast: "Logical UID deleted; physical file cleanup is deferred",
    deleteFailed: "Delete failed"
  },
  apiPage: {
    eyebrow: "API",
    title: "Upload and delete endpoints",
    uploadDescription:
      "Accepts multipart/form-data with field name file and optional storage_key. Returns UID, a public .avif URL, Markdown, BBCode, MIME type, size, created time, duplicate flag, storage key, and storage backend.",
    storageOptionsDescription:
      "Returns safe public storage choices with storage key, display name, backend, and default marker. Credentials, paths, buckets, and secret values are never included.",
    deleteDescription:
      "Deletes the logical UID when the current token owns it. The backend keeps the physical file in place and treats zero-reference objects as deferred cleanup assets for a later maintenance flow."
  },
  admin: {
    checkingSession: "Checking admin session...",
    shellEyebrow: "Dashboard",
    shellTitle: "Admin controls",
    signOut: "Sign out",
    nav: {
      status: "Status",
      images: "Images",
      settings: "Settings"
    },
    loginEyebrow: "Admin",
    loginTitle: "Sign in to dashboard",
    password: "Password",
    passwordPlaceholder: "Enter admin password",
    signingIn: "Signing in...",
    signIn: "Sign in",
    loginSuccessToast: "Admin session created",
    loginFailed: "Login failed",
    dashboardEyebrow: "Status",
    dashboardTitle: "System overview",
    statusLoading: "Loading status...",
    statusLoadFailed: "Failed to load status",
    stats: {
      totalImages: "Total images",
      storageSize: "Storage size",
      todaysUploads: "Today's uploads",
      uniqueTokens: "Unique tokens"
    },
    imageManagementTitle: "Image management",
    imageManagementDescription:
      "Search by UID, token, IP, or MD5. Deletes here remove logical UIDs only and leave physical-file cleanup to a later orphan sweep.",
    deleteSelected: "Delete selected UIDs",
    deleteSelectedTitle: "Delete selected logical UIDs; physical file cleanup is deferred",
    deleteSelectedSuccessToast: "Logical UIDs deleted; physical file cleanup is deferred",
    loadImagesFailed: "Failed to load images",
    batchDeleteFailed: "Batch delete failed",
    loadingImages: "Loading images...",
    noImagesFound: "No images match the current search.",
    gridView: "Grid view",
    listView: "List view",
    selectVisible: "Select visible",
    deselectVisible: "Deselect visible",
    imagePreviewAlt: (uid: string) => `Uploaded image ${uid}`,
    searchInputLabel: "Search images",
    selectImage: (uid: string) => `Select image ${uid}`,
    searchPlaceholder: "Search",
    table: {
      select: "Select",
      uid: "UID",
      type: "Type",
      size: "Size",
      token: "Token",
      md5: "MD5",
      storageKey: "Storage key",
      backend: "Backend",
      created: "Created"
    },
    settingsLoading: "Loading config...",
    settingsTitle: "Storage instances",
    settingsDescription: "Manage a runtime catalog of named storage instances and switch the default target for new uploads.",
    configUpdatedToast: "Config updated",
    configUpdateFailed: "Config update failed",
    storageCreateSuccessToast: "Storage instance created",
    storageUpdateSuccessToast: "Storage instance updated",
    storageDeleteSuccessToast: "Storage instance deleted",
    storageDeleteFailed: "Storage delete failed",
    defaultStorageUpdatedToast: "Default storage updated",
    defaultStorageUpdateFailed: "Default storage update failed",
    saveSettings: "Save settings",
    createStorageSubmit: "Create instance",
    createStorageInstance: "New storage instance",
    createStorageTitle: "Create storage instance",
    createStorageDescription: "Add a named runtime storage target. The generated storage key stays stable after creation.",
    editStorageTitle: "Edit storage instance",
    editStorageDescription: "Update the selected storage instance. Existing images keep resolving through its stored storage key.",
    storageBackendLockedHint:
      "Backend type stays locked after creation so existing images keep resolving through the same storage key.",
    makeDefault: "Make default",
    deleteStorageInstance: "Delete instance",
    defaultBadge: "Default",
    defaultDeleteBlockedHint: "Default storage instance cannot be deleted",
    saving: "Saving...",
    backends: {
      local: "Local filesystem",
      s3: "S3 compatible",
      webdav: "WebDAV"
    },
    fields: {
      storageName: "Storage name",
      storageKey: "Storage key",
      storageBackend: "Storage backend",
      localStoragePath: "Local storage path",
      s3Endpoint: "S3 endpoint",
      s3Region: "S3 region",
      s3Bucket: "S3 bucket",
      s3AccessKey: "S3 access key",
      s3SecretKey: "S3 secret key",
      webdavUrl: "WebDAV URL",
      webdavUser: "WebDAV user",
      webdavPassword: "WebDAV password"
    },
    toggles: {
      s3UseSsl: "Use SSL for S3",
      s3ForcePathStyle: "Force path-style S3 requests"
    }
  }
};

const zh: typeof en = {
  common: {
    upload: "上传",
    history: "历史",
    api: "API",
    admin: "管理",
    duplicate: "重复",
    clientToken: "客户端令牌",
    preparingToken: "正在准备令牌...",
    uid: "UID",
    type: "类型",
    size: "大小",
    created: "创建时间",
    uploaded: "上传时间",
    token: "令牌",
    md5: "MD5",
    backend: "后端",
    select: "选择",
    previous: "上一页",
    next: "下一页",
    loading: "加载中...",
    refresh: "刷新",
    save: "保存",
    search: "搜索",
    total: "总数",
    storage: "存储",
    preview: "预览",
    closePreview: "关闭预览",
    previewMetadata: "图片元数据",
    openPreview: (title: string) => `打开 ${title} 的预览`,
    skipToContent: "跳至主要内容",
    copyTargets: {
      url: "URL",
      markdown: "Markdown",
      bbcode: "BBCode"
    },
    items: (count: number) => `${count} 项`,
    totalResults: (count: number) => `总结果数：${count}`
  },
  header: {
    tagline: "图床服务",
    languageLabel: "语言",
    themeLabel: "主题",
    navLabel: "主导航",
    openMenu: "打开菜单",
    closeMenu: "关闭菜单",
    languages: {
      zh: "中文",
      en: "English"
    },
    themes: {
      light: "浅色",
      dark: "深色",
      system: "跟随系统"
    }
  },
  copyButton: {
    copy: (label: string) => `复制${label}`,
    copied: (label: string) => `${label}已复制`
  },
  upload: {
    eyebrow: "上传流程",
    title: "快速本地图床，支持真实去重",
    description:
      "OmePic 会在首次访问时生成客户端令牌，将光栅图片带进度转换为 AVIF，并在 IndexedDB 中保留最近十条结果。",
    recentUploads: "最近上传",
    status: "上传状态",
    statusIdle: "选择一个文件即可开始。",
    statusSuccess: "上传完成，已保存为 AVIF。",
    statusUploading: (progress: number) => `正在上传 ${progress}%`,
    latestResult: "最新结果",
    previewAlt: "上传预览",
    uploadCompleteToast: "上传完成",
    duplicateUploadToast: "已关联到重复上传",
    localHistorySaveFailed: "上传已成功，但本地历史记录未能更新。",
    uploadFailed: "上传失败",
    storageTarget: "目标存储",
    backendDefaultStorage: "后端默认存储",
    defaultStorageOption: (name: string, backend: string) => `使用后端默认：${name}（${backend}）`,
    storageOption: (name: string, backend: string) => `${name}（${backend}）`,
    storageOptionDefault: (name: string, backend: string) => `${name}（${backend}）- 默认`,
    storageSelectionHint: (name: string, backend: string) => `上传将写入 ${name}，后端为 ${backend}。`,
    storageOptionsFailed: "加载存储选项失败",
    dropTitle: "将图片拖到这里",
    dropDescription:
      "支持 PNG、JPG、JPEG、GIF、WEBP、BMP，最大 20 MB。上传前会转换为 AVIF，重复的原始上传会复用同一个已存储的 AVIF 文件。",
    fileInputLabel: "选择要上传的图片文件",
    chooseFile: "选择文件",
    browseLocally: "本地浏览",
    emptyRecent: "还没有上传记录。"
  },
  historyPage: {
    eyebrow: "IndexedDB",
    title: "本地上传历史",
    description:
      "在这里删除的只是逻辑 UID。物理文件仍会保留在存储中，直到后续清理流程移除孤儿资源。",
    clear: "清空本地历史",
    empty: "还没有本地历史记录。",
    deleteUid: "删除 UID",
    deleteEnabledTitle: "删除逻辑 UID；物理文件清理由后续流程处理",
    deleteDisabledTitle: "只有拥有该逻辑 UID 的令牌才能删除",
    deleteSuccessToast: "逻辑 UID 已删除；物理文件清理由后续流程处理",
    deleteFailed: "删除失败"
  },
  apiPage: {
    eyebrow: "API",
    title: "上传与删除接口",
    uploadDescription:
      "接受字段名为 file 的 multipart/form-data，也可附带可选 storage_key。返回 UID、公开的 .avif URL、Markdown、BBCode、MIME 类型、大小、创建时间、重复标记、存储键和存储后端。",
    storageOptionsDescription:
      "返回可公开展示的存储选项：存储键、显示名称、后端和默认标记。凭据、路径、存储桶和密钥值不会出现在响应中。",
    deleteDescription:
      "当当前令牌拥有该逻辑 UID 时可删除。后端会保留物理文件，并将零引用对象视为等待后续维护流程清理的资源。"
  },
  admin: {
    checkingSession: "正在检查管理员会话...",
    shellEyebrow: "控制台",
    shellTitle: "管理控制",
    signOut: "退出登录",
    nav: {
      status: "状态",
      images: "图片",
      settings: "设置"
    },
    loginEyebrow: "管理",
    loginTitle: "登录控制台",
    password: "密码",
    passwordPlaceholder: "输入管理员密码",
    signingIn: "正在登录...",
    signIn: "登录",
    loginSuccessToast: "管理员会话已创建",
    loginFailed: "登录失败",
    dashboardEyebrow: "状态",
    dashboardTitle: "系统概览",
    statusLoading: "正在加载状态...",
    statusLoadFailed: "加载状态失败",
    stats: {
      totalImages: "图片总数",
      storageSize: "存储大小",
      todaysUploads: "今日上传",
      uniqueTokens: "唯一令牌数"
    },
    imageManagementTitle: "图片管理",
    imageManagementDescription:
      "可按 UID、令牌、IP 或 MD5 搜索。此处删除的只是逻辑 UID，物理文件会留给后续孤儿清理流程处理。",
    deleteSelected: "删除所选 UID",
    deleteSelectedTitle: "删除所选逻辑 UID；物理文件清理由后续流程处理",
    deleteSelectedSuccessToast: "逻辑 UID 已删除；物理文件清理由后续流程处理",
    loadImagesFailed: "加载图片失败",
    batchDeleteFailed: "批量删除失败",
    loadingImages: "正在加载图片...",
    noImagesFound: "没有匹配当前搜索条件的图片。",
    gridView: "网格视图",
    listView: "列表视图",
    selectVisible: "选择当前页",
    deselectVisible: "取消选择当前页",
    imagePreviewAlt: (uid: string) => `已上传图片 ${uid}`,
    searchInputLabel: "搜索图片",
    selectImage: (uid: string) => `选择图片 ${uid}`,
    searchPlaceholder: "搜索",
    table: {
      select: "选择",
      uid: "UID",
      type: "类型",
      size: "大小",
      token: "令牌",
      md5: "MD5",
      storageKey: "存储键",
      backend: "后端",
      created: "创建时间"
    },
    settingsLoading: "正在加载配置...",
    settingsTitle: "存储实例",
    settingsDescription: "管理具名运行时存储实例目录，并切换新上传默认写入的实例。",
    configUpdatedToast: "配置已更新",
    configUpdateFailed: "配置更新失败",
    storageCreateSuccessToast: "存储实例已创建",
    storageUpdateSuccessToast: "存储实例已更新",
    storageDeleteSuccessToast: "存储实例已删除",
    storageDeleteFailed: "删除存储实例失败",
    defaultStorageUpdatedToast: "默认存储已更新",
    defaultStorageUpdateFailed: "更新默认存储失败",
    saveSettings: "保存设置",
    createStorageSubmit: "创建实例",
    createStorageInstance: "新建存储实例",
    createStorageTitle: "创建存储实例",
    createStorageDescription: "新增一个具名运行时存储目标。创建后生成的 storage key 会保持稳定。",
    editStorageTitle: "编辑存储实例",
    editStorageDescription: "更新当前选中的存储实例。已有图片仍会通过其已保存的 storage key 解析。",
    storageBackendLockedHint: "实例创建后会锁定后端类型，避免已有图片通过同一个 storage key 解析到错误的 provider。",
    makeDefault: "设为默认",
    deleteStorageInstance: "删除实例",
    defaultBadge: "默认",
    defaultDeleteBlockedHint: "默认存储实例不能删除",
    saving: "正在保存...",
    backends: {
      local: "本地文件系统",
      s3: "S3 兼容存储",
      webdav: "WebDAV"
    },
    fields: {
      storageName: "存储名称",
      storageKey: "存储键",
      storageBackend: "存储后端",
      localStoragePath: "本地存储路径",
      s3Endpoint: "S3 端点",
      s3Region: "S3 区域",
      s3Bucket: "S3 存储桶",
      s3AccessKey: "S3 Access Key",
      s3SecretKey: "S3 Secret Key",
      webdavUrl: "WebDAV URL",
      webdavUser: "WebDAV 用户",
      webdavPassword: "WebDAV 密码"
    },
    toggles: {
      s3UseSsl: "S3 使用 SSL",
      s3ForcePathStyle: "强制使用 path-style S3 请求"
    }
  }
};

export type TranslationDictionary = typeof en;

export const dictionaries: Record<Language, TranslationDictionary> = {
  en,
  zh
};

export function getDictionary(language: Language) {
  return dictionaries[language];
}

export function getLocaleForLanguage(language: Language) {
  return languageToLocale(language);
}
