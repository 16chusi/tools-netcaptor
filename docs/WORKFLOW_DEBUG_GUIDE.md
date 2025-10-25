# 工作流执行调试指南

## 问题描述
执行工作流时提示"步骤执行失败: 缺少 URL 参数"

## 数据流分析

### 前端数据流
1. **PropertyPanel.vue**: 用户输入 URL → `formData.url = "http://www.qq.com"`
2. **WorkflowDrawer.vue**: 保存数据 → `node.setData({type: 'navigate', url: 'http://www.qq.com'})`
3. **FlowCanvas.vue**: 序列化任务 → `{nodes: [{data: {type: 'navigate', url: 'http://www.qq.com'}}]}`
4. **发送到后端**: `ExecuteWorkflow(task)`

### 后端数据流
1. **workflow_executor.go**: 接收任务 → `WorkflowTask`
2. **nodeToStep**: 转换节点 → `step.Params = node.Data`
3. **executeNavigate**: 读取参数 → `step.Params["url"]`

## 调试步骤

### 1. 检查前端数据保存
打开浏览器控制台，执行工作流时查看日志：
```
[PropertyPanel] 保存节点数据: navigate {url: "http://www.qq.com"}
[WorkflowDrawer] 节点数据已更新: xxx {type: "navigate", url: "http://www.qq.com"}
[FlowCanvas] 执行任务: {...}
```

### 2. 检查后端接收数据
查看应用日志：
```
[Workflow] 节点转步骤 - ID: xxx, Type: navigate, Params: map[type:navigate url:http://www.qq.com]
```

### 3. 常见问题

#### 问题1: data 字段为空
**症状**: 后端日志显示 `Params: map[]`
**原因**: 节点数据未正确保存
**解决**: 
- 确保在 PropertyPanel 中点击"保存"按钮
- 确保在 FlowCanvas 工具栏点击"保存"按钮保存任务

#### 问题2: url 字段不存在
**症状**: 后端日志显示 `Params: map[type:navigate]`
**原因**: PropertyPanel 保存的数据未包含 url
**解决**:
- 检查 PropertyPanel 中 `v-model="formData.url"` 是否正确绑定
- 检查 `handleSave` 是否正确发送 `formData.value`

#### 问题3: 旧数据缓存
**症状**: 修改后仍然报错
**原因**: localStorage 中保存了旧的任务数据
**解决**:
```javascript
// 在浏览器控制台执行
localStorage.removeItem('workflow-tasks')
// 然后刷新页面，重新创建任务
```

## 快速修复

### 方案1: 清除缓存重新创建
1. 打开浏览器控制台 (F12)
2. 执行: `localStorage.clear()`
3. 刷新页面
4. 创建新任务
5. 添加"导航"节点
6. 点击节点，在右侧属性面板输入 URL
7. 点击"保存"按钮（属性面板底部）
8. 点击"保存"按钮（画布工具栏）
9. 点击"运行"按钮

### 方案2: 手动检查数据
1. 打开浏览器控制台
2. 执行: `JSON.parse(localStorage.getItem('workflow-tasks'))`
3. 检查输出中的 `nodes[].data` 字段是否包含 `url`
4. 如果不包含，说明数据未正确保存

## 代码修改说明

已修改的文件：
1. **WorkflowDrawer.vue**: 优化节点数据保存逻辑
2. **FlowCanvas.vue**: 添加调试日志，确保数据正确序列化
3. **PropertyPanel.vue**: 添加调试日志
4. **workflow_executor.go**: 添加调试日志

## 测试用例

### 正确的任务数据结构
```json
{
  "id": "1234567890",
  "name": "任务 1",
  "nodes": [
    {
      "id": "node1",
      "type": "start",
      "x": 300,
      "y": 50,
      "label": "开始",
      "data": {
        "type": "start"
      }
    },
    {
      "id": "node2",
      "type": "navigate",
      "x": 300,
      "y": 150,
      "label": "导航",
      "data": {
        "type": "navigate",
        "url": "http://www.qq.com"
      }
    },
    {
      "id": "node3",
      "type": "end",
      "x": 300,
      "y": 250,
      "label": "结束",
      "data": {
        "type": "end"
      }
    }
  ],
  "edges": [
    {
      "id": "edge1",
      "source": "node1",
      "target": "node2"
    },
    {
      "id": "edge2",
      "source": "node2",
      "target": "node3"
    }
  ]
}
```

注意 `node2` 的 `data` 字段必须包含 `url` 属性。
