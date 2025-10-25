# 工作流执行功能说明

## 📋 功能概述

通过可视化流程编排，生成指令序列，通过 WebSocket 控制浏览器插件执行自动化任务。

## 🏗️ 架构设计

```
前端 (Vue) → Go 后端 → WebSocket → Chrome 插件 → 网页操作
```

### 核心文件

**后端 (Go)**
- `workflow_types.go` - 工作流类型定义
- `workflow_executor.go` - 执行引擎核心逻辑
- `network_app.go` - Wails 绑定方法
- `websocket_server.go` - WebSocket 通信

**前端 (Vue)**
- `FlowCanvas.vue` - 画布组件，包含执行按钮和状态显示
- `WorkflowDrawer.vue` - 工作流抽屉容器

**浏览器插件**
- `background.js` - 接收指令并执行

## 🎯 支持的节点类型

| 节点类型 | 说明 | 参数 |
|---------|------|------|
| `start` | 开始节点 | 无 |
| `navigate` | 导航到 URL | `url`: 目标网址 |
| `click` | 点击元素 | `selector`: CSS 选择器 |
| `input` | 输入文本 | `selector`: CSS 选择器, `text`: 输入内容 |
| `wait` | 等待延时 | `duration`: 毫秒数 (默认 1000) |
| `end` | 结束节点 | 无 |

## 🚀 使用流程

### 1. 启动 WebSocket 服务
```typescript
// 前端会自动启动，或手动调用
await StartWebSocketServer()
const port = await GetWebSocketPort()
```

### 2. 连接浏览器插件
- 打开 Chrome 插件 popup
- 输入 WebSocket 端口号
- 点击"连接"

### 3. 创建工作流
- 点击"任务流编排"按钮
- 点击"+ 新建"创建任务
- 从左侧面板拖拽节点到画布
- 连接节点形成流程

### 4. 配置节点
- 点击节点打开属性面板
- 填写必要参数（如 URL、选择器等）
- 保存配置

### 5. 执行工作流
- 点击"▶️ 运行"按钮
- 观察执行状态和进度
- 当前执行的节点会高亮显示

### 6. 停止执行
- 点击"⏹️ 停止"按钮

## 📊 执行状态

执行过程中会实时显示：
- 当前步骤 / 总步骤数
- 执行状态：运行中 / 成功 / 失败 / 已停止
- 当前执行的节点（红色边框高亮）

## 🔧 测试示例

### 示例 1: 简单导航
```
开始 → 导航(https://example.com) → 结束
```

### 示例 2: 表单填写
```
开始 
  → 导航(https://example.com/form)
  → 输入(#username, "testuser")
  → 输入(#password, "123456")
  → 点击(#submit)
  → 等待(2000)
  → 结束
```

## ⚠️ 注意事项

1. **WebSocket 连接**: 确保插件已连接到 WebSocket 服务器
2. **选择器准确性**: CSS 选择器必须能准确定位到目标元素
3. **等待时间**: 页面加载或动画需要适当的等待时间
4. **错误处理**: 如果某步失败，整个流程会停止

## 🔮 后续扩展

### Phase 2 (计划中)
- `extract` - 提取页面数据
- `download` - 下载文件
- `intercept` - 设置拦截规则
- 变量系统 - 节点间数据传递

### Phase 3 (计划中)
- `if` - 条件判断
- `for` - 循环执行
- 错误重试机制
- 子流程调用

## 🐛 调试

查看日志：
- **Go 后端**: 控制台输出 `[Workflow]` 前缀
- **WebSocket**: 控制台输出 `[WebSocket]` 前缀
- **浏览器插件**: Chrome DevTools Console，`[NetCaptor]` 前缀
- **前端**: 浏览器 Console，`[FlowCanvas]` 前缀

## 📝 API 参考

### Go 绑定方法

```go
// 执行工作流（异步）
ExecuteWorkflow(task WorkflowTask) error

// 停止工作流
StopWorkflow()

// 获取运行状态
IsWorkflowRunning() bool
```

### 前端事件

```typescript
// 监听执行状态
EventsOn('workflow_status', (status) => {
  // status.currentStep, status.totalSteps, status.status
})

// 监听错误
EventsOn('workflow_error', (data) => {
  // data.error
})
```

### WebSocket 消息格式

```json
// 发送到插件
{
  "type": "click_element",
  "data": {
    "selector": "#button"
  }
}

// 插件响应
{
  "type": "action_result",
  "data": {
    "action": "click_element",
    "success": true,
    "error": "错误信息（如果失败）"
  }
}
```
