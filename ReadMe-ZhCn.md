# bing-bulk-image-downloader

*其他语言阅读: [English](README.md) | [简体中文](ReadMe-ZhCn.md)*

## 项目介绍

**bing-bulk-image-downloader** 是一个命令行（CLI）工具，旨在从必应（Bing）图像搜索中批量下载图片。该工具利用并发 worker 高效地根据搜索查询词下载指定数量的图片。

## 技术栈清单

以下是本项目采用的核心技术组件及技术栈梳理：

### 前端
- **框架/库：** 无。这是一个纯命令行应用程序，没有图形用户界面（GUI）或 Web 前端。

### 后端及运行环境
- **语言：** Go (Golang)
- **版本：** `>= 1.15`
- **核心作用：** 处理核心业务逻辑，包括并发 HTTP 请求、URL 解析、文件 I/O 操作，以及使用标准库包（如 `sync`、`net/http` 等）进行进程同步。

### 基础设施 / 数据库 / 中间件
- **数据库：** 无。操作在内存中执行，文件直接保存到本地磁盘。
- **中间件：** 无。使用 Go 标准库的 `net/http` 客户端，无需依赖第三方中间件。

### 工具链与构建工具
- **构建工具：** Go Modules (`go mod`)
- **版本：** Go `1.15` 及以上
- **核心作用：** 依赖管理（尽管目前仅依赖标准库）以及构建协调。

## 环境依赖要求

构建或运行此项目需要以下环境依赖：

- **Go 语言版本：** `1.15` 或更高版本

*注：本项目无任何外部第三方 Go 包依赖，避免了依赖版本冲突的问题。*

## 本地部署与启动步骤

您可以在多种操作系统上运行或构建该项目。请确保您的机器上已安装 Go 环境。

### 安装步骤

```bash
# 克隆仓库或使用 go get 命令
go get github.com/mattn/bing-bulk-image-downloader
```

### 从源码构建

导航至项目根目录，运行以下命令构建可执行文件：

```bash
go build -o bing-bulk-image-downloader main.go
```

### 运行工具

构建完成后，您可以执行该命令。请提供所需的下载图片数量（`-n`）、输出目录（`-o`）以及搜索查询词。

**Windows:**
```cmd
bing-bulk-image-downloader.exe -n 10 -o .\output\ golang gopher
```

**macOS / Linux:**
```bash
./bing-bulk-image-downloader -n 10 -o ./output/ golang gopher
```

**命令行参数说明：**
- `-n`: 计划下载的图片数量（默认值：100）
- `-o`: 输出目录路径（默认值：`.`）
- `-s`: 是否开启安全搜索（布尔值，默认值：true）

## 项目结构说明

```text
.
├── .git/                 # Git 版本控制目录
├── .github/              # GitHub Actions 工作流与配置文件
├── .gitignore            # Git 忽略文件配置
├── README.md             # 项目说明文档（英文）
├── ReadMe-ZhCn.md        # 项目说明文档（简体中文）
├── go.mod                # Go 模块定义文件（指定了 Go 1.15 版本）
└── main.go               # 应用程序主要源代码
```

## 开发规范

- **代码风格：** 遵循 Go 语言的标准格式化规范。在提交代码前，请运行 `gofmt -w .`。
- **并发处理：** 应用程序使用了 Go 协程 (`go worker(...)`)、`sync.WaitGroup` 以及 `sync.Mutex`。在修改共享资源时需确保线程安全，尤其是通过 `sync/atomic` 处理的 `count` 变量。
- **错误处理：** 使用 `log` 包合理地记录错误日志，确保个别图片下载失败时应用程序能够优雅降级或跳过，而不是直接崩溃。
- **文件操作：** 程序会先创建临时文件，随后移动到目标路径。请确保文件句柄被正确关闭以防止内存泄漏和文件锁问题，特别是在跨不同操作系统时。

## 常见问题排查

### Q: 执行 `go get` 命令失败或超时怎么办？
**A:** 请确保您的网络能够正常访问 `github.com`。如果您所在的地区访问 Go 模块受限，请适当配置您的 `GOPROXY` 环境变量。例如执行命令：`go env -w GOPROXY=https://goproxy.cn,direct`。

### Q: 下载失败并提示包含 "Rename Fail" 的错误信息？
**A:** 请检查您对指定的输出目录（`-o` 参数）是否具有写权限。该程序通过 `ioutil.TempDir` 创建临时目录，然后移动文件。虽然在 `moveFile()` 中通过复制文件的方式处理了跨设备链接（不同分区/驱动器间移动），但文件系统权限必须有效。

### Q: 为何我输入的搜索查询未能下载任何图片？
**A:** 脚本通过正则表达式解析必应的搜索结果。如果必应更改了其 HTML 结构或 `murl` 属性的格式，则 `main.go` 中的正则表达式（"regexp.MustCompile(`murl&quot;:&quot;(.*?)&quot;`)"）可能需要同步更新。此外，也请确认您的网络连接是否畅通。

---

## 许可证
MIT License

## 作者
Yasuhiro Matsumoto (a.k.a. mattn)
