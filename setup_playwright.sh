#!/bin/bash

export PATH="/home/fzxs/go/go1.25.1/bin:/home/fzxs/go/bin:$PATH"
export GOROOT="/home/fzxs/go/go1.25.1"
export GOPATH="/home/fzxs/go"

echo "📦 安装 Playwright..."

# 安装驱动
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install

echo "✅ Playwright 安装完成"