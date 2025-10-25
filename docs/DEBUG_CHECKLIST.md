# 工作流调试检查清单

## 问题：执行时提示"缺少 URL 参数"

## 调试步骤

### 1. 清除旧数据（必须）
```javascript
// 在浏览器控制台执行
localStorage.clear()
// 然后刷新页面
```

### 2. 创建新任务并设置 URL

1. 打开"任务流编排"
2. 点击"+ 新建"
3. 拖拽"导航"节点到画布
4. **点击"导航"节点**（重要！）
5. 在右侧属性面板输入 URL
6. **点击属性面板底部的"保存"按钮**（重要！）
7. **点击画布工具栏的"💾 保存"按钮**（重要！）

### 3. 查看浏览器控制台日志

执行任务时，应该看到以下日志序列：

#### 前端日志（浏览器控制台）

```
[PropertyPanel] 保存节点数据: navigate {url: "http://www.qq.com"}
[WorkflowDrawer] ========== 保存节点数据 ==========
[WorkflowDrawer] 节点ID: xxx
[WorkflowDrawer] 节点类型: navigate
[WorkflowDrawer] 保存的数据: {url: "http://www.qq.com"}
[WorkflowDrawer] 当前节点数据: {type: "navigate"}
[WorkflowDrawer] 合并后的数据: {type: "navigate", url: "http://www.qq.com"}
[WorkflowDrawer] 验证保存后的数据: {type: "navigate", url: "http://www.qq.com"}
[WorkflowDrawer] ✓ 节点数据已更新

[FlowCanvas] ========== emitChange 开始 ==========
[FlowCanvas] 序列化节点[1] - ID: xxx
[FlowCanvas] 序列化节点[1] - getData(): {type: "navigate", url: "http://www.qq.com"}
[FlowCanvas] 序列化节点[1] - 结果: {id: "xxx", type: "navigate", data: {type: "navigate", url: "http://www.qq.com"}}

[WorkflowDrawer] ========== 保存任务 ==========
[WorkflowDrawer] 保存节点[1] data: {type: "navigate", url: "http://www.qq.com"}
[WorkflowDrawer] ✓ 保存到 localStorage 成功

[FlowCanvas] ========== 开始执行任务 ==========
[FlowCanvas] 节点[1] - Data.type = navigate (type: string)
[FlowCanvas] 节点[1] - Data.url = http://www.qq.com (type: string)
```

#### 后端日志（应用日志）

```
[Workflow] ========== 开始执行任务 ==========
[Workflow] 节点[1] - ID: xxx, Type: navigate, Label: 导航
[Workflow] 节点[1] - Data: map[type:navigate url:http://www.qq.com]
[Workflow] 节点[1] - Data[type] = navigate (type: string)
[Workflow] 节点[1] - Data[url] = http://www.qq.com (type: string)

[Workflow] ========== nodeToStep 开始 ==========
[Workflow] 输入节点 - ID: xxx, Type: navigate
[Workflow] 输入节点 - Data: map[type:navigate url:http://www.qq.com]
[Workflow] node.Data 详细内容:
[Workflow]   - type = navigate (类型: string)
[Workflow]   - url = http://www.qq.com (类型: string)
[Workflow] 输出步骤 - Params: map[type:navigate url:http://www.qq.com]

[Workflow] ========== executeNavigate 开始 ==========
[Workflow] step.Params 键值对:
[Workflow]   - type = navigate (类型: string)
[Workflow]   - url = http://www.qq.com (类型: string)
[Workflow] 尝试获取 URL: ok=true, url=http://www.qq.com
[Workflow] ✓ 执行导航: http://www.qq.com
```

### 4. 问题诊断

#### 情况 A：前端日志显示 data 为空或没有 url
**原因**：节点数据未正确保存
**解决**：
- 确保点击了属性面板的"保存"按钮
- 确保点击了画布的"💾 保存"按钮
- 清除 localStorage 重新创建

#### 情况 B：前端日志正常，但后端日志显示 Data 为空
**原因**：数据传输到后端时丢失
**解决**：
- 检查 Wails 绑定是否正常
- 查看完整的前端日志中的"转换后的任务"
- 可能是 JSON 序列化问题

#### 情况 C：后端日志显示 Data 有 type 但没有 url
**原因**：url 字段未被保存或被覆盖
**解决**：
- 检查 `onSaveNodeData` 函数是否正确合并数据
- 检查 `emitChange` 函数是否正确序列化数据
- 查看"验证保存后的数据"日志

### 5. 手动验证数据

在浏览器控制台执行：

```javascript
// 查看保存的任务
const tasks = JSON.parse(localStorage.getItem('workflow-tasks'))
console.log('所有任务:', tasks)

// 查看第一个任务的节点
if (tasks && tasks[0]) {
  console.log('第一个任务的节点:', tasks[0].nodes)
  
  // 查找导航节点
  const navNode = tasks[0].nodes.find(n => n.type === 'navigate')
  console.log('导航节点:', navNode)
  console.log('导航节点的 data:', navNode?.data)
  console.log('导航节点的 url:', navNode?.data?.url)
}
```

### 6. 预期的正确数据结构

```json
{
  "id": "1234567890",
  "name": "任务 1",
  "nodes": [
    {
      "id": "node-xxx",
      "type": "navigate",
      "x": 300,
      "y": 150,
      "label": "导航",
      "data": {
        "type": "navigate",
        "url": "http://www.qq.com"
      }
    }
  ]
}
```

**关键点**：`data` 对象必须同时包含 `type` 和 `url` 字段。

### 7. 如果问题仍然存在

请提供以下信息：
1. 完整的浏览器控制台日志（从保存到执行）
2. 完整的后端日志（从接收任务到执行失败）
3. localStorage 中的任务数据（执行上面的验证脚本）

这样可以精确定位问题发生在哪个环节。
