# fn-music-dl

飞牛 FNOS 原生音乐搜索与下载工具。

基于 [go-music-dl](https://github.com/guohuiyuan/go-music-dl) 的核心能力，以 Native 应用方式运行在飞牛 FNOS 系统上，不需要 Docker。

## 功能

- **多平台搜索和下载**: 网易云、QQ、酷狗、酷我、咪咕、Bilibili、汽水音乐、Apple Music 等 13 个平台
- **响应式 UI**: 手机和电脑自适应布局
- **歌单管理**: 本地收藏歌单，跨平台聚合收藏
- **本地音乐管理**: 扫码本地下载目录，上传、删除管理
- **元数据内嵌**: 下载时自动嵌入封面和歌词（需 FFmpeg）
- **无损音乐**: 支持部分平台 FLAC 无损下载
- **音频试听**: 内置播放器

## 项目结构

```
fn-music-dl/
├── backend/           # Go 后端 (Gin HTTP server)
│   ├── main.go        # 入口: Unix Socket / TCP 模式
│   ├── pkg/           # 核心逻辑 (工厂、下载、配置、媒体工具)
│   └── api/           # REST API 处理器
├── frontend/          # React + TypeScript SPA
│   └── src/
│       ├── components/ # 前端组件 (搜索、歌曲列表、播放器等)
│       └── pages/      # 页面 (搜索、歌单、本地、设置)
├── package/           # FNOS 应用打包目录
│   ├── manifest       # 应用元数据
│   ├── config/        # 权限、资源共享定义
│   ├── cmd/           # 生命周期脚本 (启动/停止/配置)
│   ├── app/ui/        # 桌面入口配置
│   ├── wizard/        # 安装/配置向导
│   └── ...
├── scripts/           # 构建脚本
└── .github/workflows/ # CI workflow
```

## 构建

### 前置要求

- Go 1.25+
- Node.js 22+
- fnpack (FNOS 打包工具)

### 本地构建

```bash
# 生成图标
bash scripts/generate-icons.sh

# 构建
bash scripts/build.sh
```

产物为 `package/` 目录下的 `.fpk` 文件。

### 仅编译后端二进制

```bash
cd backend
go mod download
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o music-dl .
```

### 构建前端

```bash
cd frontend
npm install
npm run build
```

### 远程构建（GitHub Actions）

推送到 GitHub 后，Actions 会自动构建 arm64/amd64 二进制和 FPK 包。创建 tag 时会自动发布 Release。

## 安装

1. 在飞牛 FNOS 应用中心，点击手动安装，选择 `.fpk` 包
2. 或使用命令行：
   ```bash
   appcenter-cli install-fpk music-dl.fpk
   ```

安装向导会要求设置下载目录（建议 `/vol1/1000/Music` 等用户可访问路径）。

## API

后端提供 REST API，可通过 `/app/music-dl/api/` 访问：

| 端点 | 方法 | 说明 |
|------|------|------|
| `/search` | GET | 搜索歌曲/歌单/专辑 |
| `/parse` | GET | 解析分享链接 |
| `/download` | POST | 下载单曲或批量下载 |
| `/stream` | GET | 试听音频流 |
| `/playlists` | GET/POST | 歌单列表/创建 |
| `/playlists/:id` | GET/DELETE | 歌单详情/删除 |
| `/local/music` | GET | 本地音乐列表 |
| `/local/upload` | POST | 上传音乐 |
| `/settings` | GET/POST | 设置读取/保存 |
| `/cookies` | GET/POST | Cookie 管理 |
| `/downloads` | GET | 下载记录 |
| `/sources` | GET | 可用音乐源列表 |

## 配置

- **下载目录**: 安装向导设置，或通过应用设置修改
- **文件名模板**: 支持 `{name}` `{artist}` `{album}` `{source}` `{id}` `{ext}`
- **环境变量**: `MUSIC_DL_CONFIG_DB` 覆盖 SQLite 路径

## License

AGPL-3.0
