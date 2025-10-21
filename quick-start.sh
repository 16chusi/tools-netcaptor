#!/bin/bash

echo "🚀 自动下载工具 - 一键安装脚本"
echo "=================================="

# 检查系统
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo "❌ 此脚本仅支持 Linux 系统"
    exit 1
fi

# 设置变量
GO_VERSION="1.25.1"
GO_DIR="$HOME/go"
GO_ROOT="$GO_DIR/go$GO_VERSION"

echo "📦 开始安装..."

# 1. 安装 Go
if [ ! -d "$GO_ROOT" ]; then
    echo "📥 下载 Go $GO_VERSION..."
    mkdir -p "$GO_DIR"
    cd "$GO_DIR"
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    tar -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    mv go "go${GO_VERSION}"
    rm "go${GO_VERSION}.linux-amd64.tar.gz"
fi

# 2. 设置环境变量
export PATH="$GO_ROOT/bin:$GO_DIR/bin:$PATH"
export GOROOT="$GO_ROOT"
export GOPATH="$GO_DIR"

echo "✅ Go 环境已配置"

# 3. 安装 Wails
if ! command -v wails &> /dev/null; then
    echo "📦 安装 Wails..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
fi

echo "✅ Wails 已安装"

# 4. 更新项目依赖
echo "📦 更新项目依赖..."
go mod tidy

# 5. 安装 Playwright 浏览器
echo "📦 安装 Playwright 浏览器..."
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install

echo "🎉 安装完成！"
echo ""
echo "环境变量配置（添加到 ~/.bashrc）："
echo "export PATH=\"$GO_ROOT/bin:$GO_DIR/bin:\$PATH\""
echo "export GOROOT=\"$GO_ROOT\""
echo "export GOPATH=\"$GO_DIR\""
echo ""
echo "启动应用："
echo "./run.sh"
echo ""
echo "访问地址："
echo "http://localhost:34115"