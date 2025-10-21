#!/bin/bash

# 设置 Go 环境
export PATH="/home/fzxs/go/go1.25.1/bin:/home/fzxs/go/bin:$PATH"
export GOROOT="/home/fzxs/go/go1.25.1"
export GOPATH="/home/fzxs/go"

echo "🚀 启动自动下载工具..."
echo "Go 版本: $(go version)"
echo "Wails 版本: $(wails version)"

# 使用系统 Chrome浏览器
echo "🌐 使用系统 Chrome浏览器..."

# 启动开发模式
wails dev -tags webkit2_41