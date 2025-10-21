# 🛠️ 详细安装指南

## 系统要求

- Linux (Ubuntu 20.04+)
- 至少 2GB 内存
- 500MB 磁盘空间

## 完整安装步骤

### 步骤 1: 安装 Go 环境

```bash
# 创建 Go 目录
mkdir -p ~/go

# 下载 Go 1.25.1
cd ~/go
wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz
tar -xzf go1.25.1.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH="$HOME/go/go1.25.1/bin:$HOME/go/bin:$PATH"' >> ~/.bashrc
echo 'export GOROOT="$HOME/go/go1.25.1"' >> ~/.bashrc
echo 'export GOPATH="$HOME/go"' >> ~/.bashrc

# 重新加载环境变量
source ~/.bashrc

# 验证安装
go version
```

### 步骤 2: 安装 Wails CLI

```bash
# 安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 验证安装
wails version
```

### 步骤 3: 克隆并配置项目

```bash
# 进入项目目录
cd /path/to/inspdinfo

# 更新 Go 模块
go mod tidy

# 给脚本添加执行权限
chmod +x run.sh
chmod +x install.sh
chmod +x setup_playwright.sh
```

### 步骤 4: 安装 Playwright 浏览器

```bash
# 方法 1: 使用项目脚本
./setup_playwright.sh

# 方法 2: 手动安装
export PATH="$HOME/go/go1.25.1/bin:$HOME/go/bin:$PATH"
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install

# 验证安装
ls ~/.cache/ms-playwright/
```

### 步骤 5: 启动应用

```bash
# 使用启动脚本
./run.sh

# 或手动启动
export PATH="$HOME/go/go1.25.1/bin:$HOME/go/bin:$PATH"
wails dev -tags webkit2_41
```

### 步骤 6: 访问应用

打开浏览器访问：
- **主应用**: http://localhost:34115
- **前端开发**: http://localhost:5173 或 http://localhost:5174

## 常见问题解决

### 问题 1: Go 命令未找到

```bash
# 检查 Go 安装路径
ls ~/go/go1.25.1/bin/go

# 手动设置环境变量
export PATH="$HOME/go/go1.25.1/bin:$PATH"
```

### 问题 2: Wails 命令未找到

```bash
# 检查 GOPATH/bin
ls ~/go/bin/wails

# 重新安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 问题 3: Playwright 驱动错误

```bash
# 错误信息: "please install the driver (v1.47.2)"
# 解决方案: 安装正确版本的驱动
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install
```

### 问题 4: 浏览器不弹出

```bash
# 检查浏览器安装
ls ~/.cache/ms-playwright/chromium-*/

# 重新安装 Chromium
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install chromium
```

### 问题 5: 端口占用

```bash
# 检查端口占用
netstat -tulpn | grep :34115
netstat -tulpn | grep :5173

# 杀死占用进程
sudo kill -9 <PID>
```

### 问题 6: 前端编译错误

```bash
# 进入前端目录
cd frontend

# 重新安装依赖
npm install

# 清理缓存
npm run build
```

## 环境变量配置文件

创建 `~/.bashrc` 配置：

```bash
# Go 环境配置
export PATH="$HOME/go/go1.25.1/bin:$HOME/go/bin:$PATH"
export GOROOT="$HOME/go/go1.25.1"
export GOPATH="$HOME/go"

# 可选: 设置 Go 代理（中国用户）
export GOPROXY=https://goproxy.cn,direct
```

## 验证安装

运行以下命令验证所有组件：

```bash
# 检查 Go
go version

# 检查 Wails
wails version

# 检查 Playwright 浏览器
ls ~/.cache/ms-playwright/

# 检查项目依赖
go mod verify

# 测试编译
go build -o test-build
```

## 生产环境构建

```bash
# 构建生产版本
wails build

# 构建输出位置
ls build/bin/
```