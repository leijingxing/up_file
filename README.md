# Go 文件上传服务器

一个基于Go语言开发的轻量级文件上传服务器，支持多项目、多目录的文件管理，具有安全验证和文件时间戳查询功能。

## 功能特性

- 🔐 **安全验证**: 支持自定义Header密钥验证
- 📁 **多项目管理**: 支持按项目和目录组织文件
- 🛡️ **安全防护**: 文件名安全过滤，防止路径遍历攻击
- ⏰ **时间戳查询**: 支持查询目录下最新文件的修改时间
- 🌐 **跨平台**: 支持Windows、Linux、macOS等操作系统
- 📊 **详细日志**: 完整的操作日志记录，便于调试和监控

## 系统要求

- Go 1.16 或更高版本
- 网络访问权限（用于监听端口）

## 安装和运行

### 1. 编译项目

```bash
go build -o gofileserver main.go
```

### 2. 设置环境变量（可选）

```bash
# Windows
set UPLOAD_DIR=D:\uploads

# Linux/macOS
export UPLOAD_DIR=/var/uploads
```

如果不设置 `UPLOAD_DIR` 环境变量，程序将默认在可执行文件所在目录下创建 `uploads` 文件夹。

### 3. 运行服务器

```bash
./gofileserver
```

服务器将在 `0.0.0.0:8080` 端口启动。

## API 文档

### 1. 健康检查

**端点**: `GET /`

**描述**: 检查服务器是否正常运行

**响应**:
```
file_upload.rs
```

### 2. 文件上传

**端点**: `POST /upload`

**描述**: 上传文件到指定项目和目录

**请求参数**:
- `filename`: 文件名（必需）
- `proj`: 项目名称（必需）
- `dir`: 目录名称（必需）

**请求头**:
- `HYZH-KEY`: 验证密钥（可选，默认值为 "HYzh221015"）

**请求体**: multipart/form-data 格式的文件数据

**示例**:
```bash
curl -X POST "http://localhost:8080/upload?filename=test.txt&proj=myproject&dir=docs" \
  -H "HYZH-KEY: HYzh221015" \
  -F "file=@/path/to/test.txt"
```

**成功响应**:
```
上传成功
```

### 3. 查询最新文件时间

**端点**: `GET /latest`

**描述**: 查询指定目录下最新文件的修改时间（以秒为单位）

**请求参数**:
- `path`: 目录路径（必需）

**示例**:
```bash
curl "http://localhost:8080/latest?path=myproject/docs"
```

**响应**: 返回最新文件的修改时间（秒数）

## 目录结构

```
uploads/
├── project1/
│   ├── docs/
│   │   ├── file1.txt
│   │   └── file2.pdf
│   └── images/
│       └── image1.jpg
└── project2/
    └── data/
        └── data.csv
```

## 安全特性

### 文件名安全过滤

程序会自动过滤文件名中的危险字符：
- 移除 `..` 防止路径遍历
- 移除 `/` 和 `\` 防止目录操作

### 访问控制

- 支持自定义Header密钥验证
- 默认密钥：`HYzh221015`
- 未授权访问会被记录到日志中

## 日志记录

程序会记录以下操作：
- 服务器启动信息
- 文件上传操作
- 目录创建操作
- 错误信息和异常情况
- 未授权访问尝试

日志输出到标准输出，可以通过重定向保存到文件：

```bash
./gofileserver > app.log 2>&1
```

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `UPLOAD_DIR` | 文件上传根目录 | `./uploads` |

### 服务器配置

- **监听地址**: `0.0.0.0:8080`
- **文件权限**: `0755`（目录）
- **最大文件大小**: 无限制（受系统内存限制）

## 故障排除

### 常见问题

1. **端口被占用**
   ```
   Fatal: Server failed to start: listen tcp 0.0.0.0:8080: bind: address already in use
   ```
   解决方案：更改端口或停止占用端口的程序

2. **权限不足**
   ```
   Fatal: Error creating directory: permission denied
   ```
   解决方案：确保程序有创建目录的权限

3. **目录不存在**
   ```
   Error reading directory: no such file or directory
   ```
   解决方案：确保查询的目录路径存在

### 调试模式

启用详细日志输出：
```bash
./gofileserver 2>&1 | tee debug.log
```

## 开发说明

### 项目结构

```
up_file/
├── main.go          # 主程序文件
├── go.mod           # Go模块文件
├── gofileserver     # 编译后的可执行文件
└── README.md        # 项目文档
```

### 主要函数

- `sanitizeFilename()`: 文件名安全过滤
- `uploadHandler()`: 文件上传处理
- `latestHandler()`: 最新文件时间查询
- `aliveHandler()`: 健康检查

## 许可证

本项目采用 MIT 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 更新日志

### v1.0.0
- 初始版本发布
- 支持文件上传和下载
- 支持多项目管理
- 添加安全验证功能
- 支持时间戳查询 