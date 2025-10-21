#!/bin/bash

echo "🚀 启动网络抓包工具..."

# 检查 Wails 是否安装
if ! command -v wails &> /dev/null; then
    echo "❌ Wails 未安装,正在安装..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
fi

# 检查前端依赖
if [ ! -d "frontend/node_modules" ]; then
    echo "📦 安装前端依赖..."
    cd frontend
    npm install
    cd ..
fi

# 检查 Go 依赖
echo "📦 检查 Go 依赖..."
go mod tidy

# 启动开发模式
echo "🎯 启动开发模式..."
wails dev
