#!/bin/bash

# 该脚本用于交叉编译应用程序，生成适用于 Linux, Windows, 和 macOS 的包。

echo "🚀 开始 NetCaptor 跨平台构建..."

# 设置 Go 环境以确保一致性，与 run.sh 保持一致
export PATH="/home/fzxs/go/go1.25.1/bin:/home/fzxs/go/bin:$PATH"
export GOROOT="/home/fzxs/go/go1.25.1"
export GOPATH="/home/fzxs/go"

# 显式定义 Wails 可执行文件的路径
WAILS_CMD="$GOPATH/bin/wails"

# 检查 wails 命令是否存在于指定路径
if ! [ -x "$WAILS_CMD" ]; then
    echo "❌ 在 $WAILS_CMD 未找到 Wails CLI。"
    echo "   请先运行 go install github.com/wailsapp/wails/v2/cmd/wails@latest 进行安装。"
    exit 1
fi

# 打印版本信息
echo "Go Version: $(go version)"
echo "Wails Version: $($WAILS_CMD version)"
echo "----------------------------------"

# 清理旧的构建产物
echo "🧹 清理旧的构建..."
$WAILS_CMD build -clean

# --- 构建 Linux (amd64) ---
echo "🐧 正在构建 Linux (amd64) 安装包..."

# -appimage 标志会创建一个 AppImage 安装包
wails build -platform linux/amd64 -tags webkit2_41 -appimage
if [ $? -ne 0 ]; then
    echo "❌ Linux AppImage 打包失败。"
    exit 1
fi

# -deb 标志会创建一个 Debian 安装包
$WAILS_CMD build -platform linux/amd64 -tags webkit2_41 -deb
if [ $? -ne 0 ]; then
    echo "❌ Linux .deb 打包失败。"
    exit 1
fi

# --- 构建 Windows (amd64) ---
echo "🪟 正在构建 Windows (amd64)..."
$WAILS_CMD build -platform windows/amd64
if [ $? -ne 0 ]; then
    echo "❌ Windows (amd64) 构建失败。"
    exit 1
fi

# --- 构建 macOS (Intel amd64) ---
echo "🍎 正在构建 macOS (amd64)..."
$WAILS_CMD build -platform darwin/amd64
if [ $? -ne 0 ]; then
    echo "❌ macOS (amd64) 构建失败。"
    exit 1
fi

# --- 构建 macOS (Apple Silicon arm64) ---
echo "🍎 正在构建 macOS (arm64)..."
$WAILS_CMD build -platform darwin/arm64
if [ $? -ne 0 ]; then
    echo "❌ macOS (arm64) 构建失败。"
    exit 1
fi

echo "----------------------------------"
echo "✅ 所有平台构建完成！"
echo "📦 可执行文件位于 'build/bin' 目录中。"

# 列出产物以供确认
ls -l build/bin
