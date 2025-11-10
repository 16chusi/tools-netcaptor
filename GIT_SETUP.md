# 🔄 Git 多仓库配置指南

## 📋 当前配置

项目配置了双仓库推送：
- **Gitee**: `git@gitee.com:duola-tools/tools-netcaptor.git`
- **GitHub**: `git@github.com:16chusi/tools-netcaptor.git`

## 🔧 配置说明

### .git/config 配置
```ini
[remote "origin"]
    url = git@gitee.com:duola-tools/tools-netcaptor.git
    pushurl = git@gitee.com:duola-tools/tools-netcaptor.git
    pushurl = git@github.com:16chusi/tools-netcaptor.git
    fetch = +refs/heads/*:refs/remotes/origin/*
```

这个配置实现：
- **拉取(fetch)**: 从Gitee仓库拉取
- **推送(push)**: 同时推送到Gitee和GitHub

## 🚀 使用方法

### 日常推送
```bash
# 推送代码到两个仓库
git push origin main

# 推送标签到两个仓库  
git push origin --tags

# 一键推送（使用脚本）
./push-all.sh
```

### 发布版本
```bash
# 使用发布脚本（会自动推送到两个仓库）
../scripts/release.sh v1.0.0
```

### 仓库同步
如果两个仓库不同步，使用同步脚本：
```bash
./sync-repos.sh
```

## 🛠️ 脚本说明

### push-all.sh
- 推送当前分支到所有仓库
- 推送所有标签到所有仓库

### sync-repos.sh  
- 从Gitee拉取最新代码
- 强制推送到GitHub进行同步
- 同步所有标签

### test-push.sh
- 测试推送配置（dry-run模式）
- 检查远程仓库状态

## 🔍 故障排除

### 推送被拒绝
```bash
# 如果GitHub推送被拒绝，先同步
./sync-repos.sh

# 或者手动解决
git fetch github
git merge github/main
git push origin main
```

### SSH密钥问题
```bash
# 测试SSH连接
ssh -T git@github.com
ssh -T git@gitee.com

# 如果失败，检查SSH密钥配置
cat ~/.ssh/config
```

### 查看推送状态
```bash
# 查看最后一次推送结果
git log --oneline -5

# 查看远程分支状态
git branch -r
```

## 📊 验证配置

运行测试脚本验证配置：
```bash
./test-push.sh
```

正常输出应该显示两个pushurl配置。

## 🎯 最佳实践

1. **开发流程**:
   ```bash
   git add .
   git commit -m "feat: 新功能"
   ./push-all.sh
   ```

2. **发布流程**:
   ```bash
   ../scripts/release.sh v1.0.0
   ```

3. **定期同步**:
   ```bash
   ./sync-repos.sh  # 每周运行一次
   ```

这样配置后，每次推送都会自动同步到Gitee和GitHub两个仓库！🎉
