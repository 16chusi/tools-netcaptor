#!/bin/bash

# 同步多个远程仓库
# 使用方法: ./sync-repos.sh

set -e

echo "🔄 开始同步远程仓库..."

# 添加单独的远程仓库用于同步
echo "📡 配置远程仓库..."
git remote remove gitee 2>/dev/null || true
git remote remove github 2>/dev/null || true

git remote add gitee git@gitee.com:duola-tools/tools-netcaptor.git
git remote add github git@github.com:16chusi/tools-netcaptor.git

echo "📥 从Gitee拉取最新代码..."
git fetch gitee

echo "📤 推送到GitHub..."
git push github main --force-with-lease

echo "🏷️  同步所有标签..."
git push github --tags

echo "✅ 仓库同步完成！"
echo ""
echo "📋 同步结果:"
echo "  • Gitee ➜ GitHub: ✅"
echo "  • 标签同步: ✅"
echo ""
echo "💡 现在可以使用 ./push-all.sh 进行双仓库推送"
