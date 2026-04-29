# ystools

个人常用的一些工具包，包含图片处理和视频下载等实用工具。

## 📦 工具列表

### 1. m3u8down - m3u8视频下载器

m3u8格式视频流下载工具，支持多并发下载。

**功能特点：**
- 支持AES加密的m3u8视频下载
- 多并发下载（4个并发）
- 实时显示下载进度
- 支持自定义输出文件名

**使用方法：**

```bash
./out/m3u8down -url <m3u8_url> -name <output.mp4>
```

**示例：**

```bash
./out/m3u8down -url https://example.com/video.m3u8 -name video.mp4
```

### 2. sketch - 图片转线稿工具

将图片转换为线稿风格，支持批量处理。

**功能特点：**
- 图片灰度化处理
- 边缘检测和线稿提取
- 多协程并发处理（4个协程）
- 支持PNG格式输出

**使用方法：**

```bash
./out/sketch -dir <图片目录路径>
```

**示例：**

```bash
./out/sketch -dir /Users/yangsen/Pictures
```

### 3. lines - 图片九宫格画线工具

将图片目录中的图画上九宫格，便于绘画新手临摹。

**功能特点：**
- 图片九宫格展示
- 并发处理（4个工作线程）
- 实时处理进度显示
- 自动识别JPG/PNG格式

**使用方法：**

```bash
./out/draw -folder <图片目录路径>
```

**示例：**

```bash
./out/draw -folder ./images
```

## 🚀 快速开始

### 环境要求

- Go 1.24 或更高版本

### 编译所有工具

```bash
make
```

或

```bash
make build
```

### 清理生成的可执行文件

```bash
make clean
```

## 📂 项目结构

```
ystools/
├── Makefile              # 编译脚本
├── go.mod                # Go模块定义
├── go.sum                # 依赖锁定
├── LICENSE               # MIT许可证
├── README.md             # 项目说明文档
├── m3u8down/             # m3u8视频下载器
│   └── main.go
├── sketch/               # 图片转线稿工具
│   └── main.go
└── lines/                 # 图片九宫格画线工具
    ├── main.go
    └── tools/
        └── get_file.go
```

## 🔧 编译说明

所有工具编译后统一输出到 `out/` 目录：

- `out/m3u8down` - m3u8视频下载器
- `out/sketch` - 图片转线稿工具
- `out/lines` - 图片九宫格画线工具

## 📝 依赖说明

项目使用以下Go依赖：

- `github.com/disintegration/imaging` - 图像处理库（用于sketch工具）

依赖已在 `go.mod` 中声明，使用 `make build` 编译时会自动下载。

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件
