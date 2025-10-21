# 🚀 智能自动下载工具

基于 Wails + Playwright-Go 开发的智能网站下载工具，具备强大的反检测能力。

## ✨ 核心特性

- 🛡️ **反检测技术**：绕过网站的 F12 禁用和自动化检测
- 🔍 **智能链接提取**：自动识别和提取下载链接
- 📄 **自动翻页**：支持多页面批量抓取
- ⬇️ **批量下载**：高效的文件下载管理
- 🎨 **现代界面**：直观的 Vue3 用户界面

## 🚀 环境安装

### 1. 安装 Go 环境
```bash
# 下载 Go 1.21+ 版本
wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz
tar -xzf go1.25.1.linux-amd64.tar.gz -C ~/go/

# 设置环境变量（添加到 ~/.bashrc）
export PATH="$HOME/go/go1.25.1/bin:$HOME/go/bin:$PATH"
export GOROOT="$HOME/go/go1.25.1"
export GOPATH="$HOME/go"
```

### 2. 安装 Wails
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 3. 安装 Playwright 浏览器
```bash
# 进入项目目录
cd inspdinfo

# 更新依赖
go mod tidy

# 安装 Playwright 浏览器（使用正确版本）
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install
```

## ⚡ 一键安装

```bash
# 一键安装所有依赖
chmod +x quick-start.sh
./quick-start.sh

# 启动应用
./run.sh
```

## 🔧 手动安装

### 4. 启动应用
```bash
# 使用启动脚本
./run.sh

# 或手动启动
export PATH="$HOME/go/go1.25.1/bin:$HOME/go/bin:$PATH"
wails dev -tags webkit2_41
```

### 5. 访问应用
- **应用界面**: http://localhost:34115
- **前端开发**: http://localhost:5173/

## 🔧 故障排除

### Playwright 驱动问题
```bash
# 如果出现 "please install the driver" 错误
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install
```

### 浏览器不弹出
```bash
# 检查浏览器安装
ls ~/.cache/ms-playwright/

# 重新安装 Chromium
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install chromium
```

### 端口占用
```bash
# 检查端口占用
lsof -i :34115
lsof -i :5173

# 杀死占用进程
kill -9 <PID>
```

## 📖 文档指南

- **安装指南**: [INSTALL.md](./INSTALL.md) - 详细的环境搭建步骤
- **使用手册**: [USAGE.md](./USAGE.md) - 功能使用说明

## 🛠️ 技术栈

- **后端**：Go + Wails v2
- **前端**：Vue 3 + TypeScript
- **浏览器**：Playwright-Go
- **反检测**：自定义 JavaScript 注入

## ⚠️ 免责声明

本工具仅供学习和合法用途使用，请遵守相关法律法规和网站使用条款。