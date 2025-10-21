#!/bin/bash

echo "🚀 自动下载工具安装脚本"
echo "=========================="

# 设置 Go 路径
export PATH="/home/fzxs/go/go1.25.1/bin:/home/fzxs/go/bin:$PATH"
export GOROOT="/home/fzxs/go/go1.25.1"
export GOPATH="/home/fzxs/go"

# 检查 Go 是否可用
if ! command -v go &> /dev/null; then
    echo "❌ Go 未找到，请检查路径: /home/fzxs/go/go1.25.1/bin"
    exit 1
fi

echo "✅ Go 版本: $(go version)"

# 检查 Wails 是否已安装
if ! command -v wails &> /dev/null; then
    echo "📦 安装 Wails..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
fi

echo "✅ Wails 版本: $(wails version)"

# 安装 Playwright 浏览器
echo "📦 安装 Playwright 浏览器..."
go mod tidy
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install chromium

echo "🔧 生成 Wails 绑定..."
wails generate module

echo "🎉 安装完成！"
echo ""
echo "使用方法："
echo "  开发模式: wails dev -tags webkit2_41"
echo "  构建应用: wails build"
echo ""
echo "功能特性："
echo "  ✨ 反检测浏览器配置"
echo "  🔍 智能链接提取"
echo "  📄 自动翻页支持"
echo "  ⬇️  批量文件下载"
echo "  🛡️  绕过 F12 禁用"