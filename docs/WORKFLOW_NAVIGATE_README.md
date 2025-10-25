# 🔄 工作流导航功能

## 快速开始

### 1️⃣ 启动服务
```bash
./run.sh
```

### 2️⃣ 启动 WebSocket
点击工具栏 WebSocket 按钮

### 3️⃣ 安装插件
1. 打开 `chrome://extensions/`
2. 加载 `chrome-extension` 目录
3. 连接到 WebSocket

### 4️⃣ 创建工作流
1. 点击"🔄 任务流"按钮
2. 点击"+ 新建任务"
3. 拖拽"🌐 导航"节点到画布
4. 连接: 开始 → 导航 → 结束

### 5️⃣ 配置节点
1. 点击"导航"节点
2. 输入 URL: `https://www.baidu.com`
3. 点击"保存"

### 6️⃣ 执行
1. 点击"💾 保存"
2. 点击"▶️ 运行"
3. 观察浏览器自动打开页面

## 📚 详细文档

- [快速开始指南](./docs/WORKFLOW_NAVIGATE_QUICKSTART.md) - 5分钟快速测试
- [测试文档](./docs/WORKFLOW_NAVIGATE_TEST.md) - 完整测试步骤
- [验证清单](./docs/WORKFLOW_NAVIGATE_CHECKLIST.md) - 功能验证
- [实现详解](./docs/WORKFLOW_NAVIGATE_IMPLEMENTATION.md) - 技术细节
- [开发总结](./WORKFLOW_NAVIGATE_SUMMARY.md) - 完整总结

## 🐛 常见问题

### 缺少 URL 参数？
点击两次保存：属性面板的"保存" + 工具栏的"💾 保存"

### 浏览器未打开？
检查 WebSocket 服务器和插件连接状态

### 执行超时？
检查网络连接和浏览器插件日志

## ✅ 功能状态

- ✅ 可视化工作流编辑
- ✅ 节点拖拽和连接
- ✅ 节点属性配置
- ✅ 工作流执行
- ✅ 浏览器自动导航
- ✅ 实时状态更新
- ✅ 错误处理
- ✅ 数据持久化

## 🎯 测试 URL

- https://www.baidu.com
- https://github.com
- https://www.google.com

## 📊 代码修复

已修复 5 处关键问题：
1. ✅ 节点数据保存
2. ✅ 节点数据加载
3. ✅ 节点数据序列化
4. ✅ 导航超时优化
5. ✅ 导航完成检测

## 🚀 下一步

- 添加更多节点类型（点击、输入、等待）
- 实现条件判断和循环
- 添加变量系统
- 实现数据提取

---

**版本**: 1.0.0 | **状态**: ✅ 可用
