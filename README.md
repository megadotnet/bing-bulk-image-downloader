# bing-bulk-image-downloader

*Read this in other languages: [English](README.md) | [简体中文](ReadMe-ZhCn.md)*

## Project Introduction

**bing-bulk-image-downloader** is a command-line interface (CLI) tool designed to download images in bulk from Bing image search. The tool makes use of concurrent workers to efficiently download a specified number of images based on a search query.

## Tech Stack

The following outlines the core technical components and stack used in this project:

### Frontend
- **Framework/Library:** None. This is a pure CLI application without a graphical user interface (GUI) or web frontend.

### Backend & Runtime Environment
- **Language:** Go (Golang)
- **Version:** `>= 1.15`
- **Role:** Handles core logic including concurrent HTTP requests, URL parsing, file I/O operations, and process synchronization using standard library packages (`sync`, `net/http`, etc.).

### Infrastructure / Database / Middleware
- **Database:** None. Operations are performed in-memory and files are saved directly to the local disk.
- **Middleware:** None. It uses Go's standard `net/http` client without third-party middleware.

### Toolchain & Build Tools
- **Build Tool:** Go Modules (`go mod`)
- **Version:** Go `1.15` and above
- **Role:** Dependency management (although currently relying solely on standard libraries) and build coordination.

## Environment Dependencies

To build or run this project, the following dependencies are required:

- **Go version:** `1.15` or higher

*Note: There are no external third-party Go packages required, which avoids dependency version conflicts.*

## Local Deployment and Run Instructions

You can run or build the project on various operating systems. Ensure Go is installed on your machine.

### Installation

```bash
# Clone the repository or use go get
go get github.com/mattn/bing-bulk-image-downloader
```

### Building from Source

Navigate to the project root directory and run the following command to build the executable:

```bash
go build -o bing-bulk-image-downloader main.go
```

### Running the Tool

After building, you can execute the command. Provide the desired number of images (`-n`), output directory (`-o`), and the search query terms.

**Windows:**
```cmd
bing-bulk-image-downloader.exe -n 10 -o .\output\ golang gopher
```

**macOS / Linux:**
```bash
./bing-bulk-image-downloader -n 10 -o ./output/ golang gopher
```

**Command-line Arguments:**
- `-n`: Number of images to download (default: 100)
- `-o`: Output directory path (default: `.`)
- `-s`: Safe search enabled (boolean, default: true)

## Project Structure

```text
.
├── .git/                 # Git version control directory
├── .github/              # GitHub Actions workflows and configurations
├── .gitignore            # Files to ignore in Git
├── README.md             # Project documentation (English)
├── ReadMe-ZhCn.md        # Project documentation (Simplified Chinese)
├── go.mod                # Go modules definition file (specifies Go version 1.15)
└── main.go               # Main application source code
```

## Development Guidelines

- **Code Style:** Follow standard Go formatting guidelines. Run `gofmt -w .` before committing changes.
- **Concurrency Handling:** The application uses Go routines (`go worker(...)`), `sync.WaitGroup`, and `sync.Mutex`. Ensure thread safety when modifying shared resources, especially the `count` variable handled via `sync/atomic`.
- **Error Handling:** Log errors appropriately using the `log` package and ensure graceful degradation or skipping of failed individual downloads rather than crashing the application.
- **File Operations:** Temporary files are created and then moved to the destination. Ensure proper closure of file handles to prevent memory leaks and file lock issues, particularly across different operating systems.

## Troubleshooting

### Q: Command `go get` fails or times out.
**A:** Ensure your network allows access to `github.com`. If you are in a region with restricted access to Go modules, configure your `GOPROXY` environment variable appropriately. For example: `go env -w GOPROXY=https://goproxy.cn,direct`.

### Q: Download fails with error related to "Rename Fail".
**A:** Ensure you have write permissions to the specified output directory (`-o`). The application creates a temporary directory using `ioutil.TempDir` and then moves files. Cross-device linking (moving between different partitions/drives) is handled via file copy in `moveFile()`, but permissions must be valid.

### Q: No images are being downloaded for my query.
**A:** The script parses Bing search results via a regular expression. If Bing changes its HTML structure or the `murl` attribute format, the regex in `main.go` ("regexp.MustCompile(`murl&quot;:&quot;(.*?)&quot;`)") might need an update. Also, ensure you have an active internet connection.

---

## License
MIT

## Author
Yasuhiro Matsumoto (a.k.a. mattn)
