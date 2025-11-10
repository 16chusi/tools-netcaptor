#!/bin/bash

# 推送到所有远程仓库的脚本
# 使用方法: ./push-all.sh [branch_name]

set -e

BRANCH=${1:-main}

echo "🚀 开始推送到所有远程仓库..."

# 推送代码到所有仓库
echo "📤 推送分支 $BRANCH 到所有仓库..."
git push origin $BRANCH

# 推送所有标签到所有仓库
echo "🏷️  推送所有标签到所有仓库..."
git push origin --tags

echo "✅ 推送完成！"
echo ""
echo "📋 推送目标:"
echo "  • Gitee: git@gitee.com:duola-tools/tools-netcaptor.git"
echo "  • GitHub: git@github.com:16chusi/tools-netcaptor.git"
