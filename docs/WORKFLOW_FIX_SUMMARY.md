# 工作流问题修复总结

## 已修复的问题

### 1. URL 参数未保存问题

**问题原因**：
- 节点数据保存后，没有触发 `change` 事件
- 导致数据未被序列化到任务对象中

**解决方案**：
- 在 `onSaveNodeData` 函数中，保存数据后通过微调节点位置来触发 `change` 事件
- 这会自动调用 `emitChange` 函数，将数据序列化到任务中

**代码位置**：`frontend/src/components/WorkflowDrawer.vue`

### 2. Alert 弹窗问题

**问题**：
- alert 弹窗样式不美观
- 可能弹出多次

**解决方案**：
- 创建了 `Toast` 组件 (`frontend/src/components/Toast.vue`)
- 创建了 `toast` 工具函数 (`frontend/src/utils/toast.ts`)
- 替换所有 `alert` 为 `toast`

**使用方法**：
```typescript
import { toast } from '../utils/toast'

toast.success('保存成功')
toast.error('执行失败')
toast.warning('警告信息')
toast.info('提示信息')
```

## 使用步骤

### 正确的工作流创建步骤

1. **清除旧数据**（如果之前有问题）
   ```javascript
   // 在浏览器控制台执行
   localStorage.clear()
   ```

2. **创建新任务**
   - 打开"任务流编排"
   - 点击"+ 新建"

3. **添加节点**
   - 从左侧拖拽"导航"节点到画布
   - 连接：开始 → 导航 → 结束

4. **设置属性**
   - **点击"导航"节点**（重要！）
   - 在右侧属性面板输入 URL：`http://www.qq.com`
   - **点击属性面板底部的"保存"按钮**（重要！）
   - 此时会自动触发数据保存

5. **保存任务**
   - 点击画布工具栏的"💾 保存"按钮
   - 看到绿色提示"保存成功"

6. **执行任务**
   - 点击"▶️ 运行"按钮

## 验证数据

在浏览器控制台执行：

```javascript
// 查看保存的任务
const tasks = JSON.parse(localStorage.getItem('workflow-tasks'))
console.log('任务列表:', tasks)

// 查看导航节点的数据
const navNode = tasks[0]?.nodes.find(n => n.type === 'navigate')
console.log('导航节点:', navNode)
console.log('导航节点 data:', navNode?.data)

// 应该看到：
// {
//   type: "navigate",
//   url: "http://www.qq.com"
// }
```

## 调试日志

### 保存节点数据时的日志

```
[PropertyPanel] 保存节点数据: navigate {url: "http://www.qq.com"}
[WorkflowDrawer] ========== 保存节点数据 ==========
[WorkflowDrawer] 保存的数据: {url: "http://www.qq.com"}
[WorkflowDrawer] 合并后的数据: {type: "navigate", url: "http://www.qq.com"}
[WorkflowDrawer] ✓ 节点数据已更新
[FlowCanvas] ========== emitChange 开始 ==========
[FlowCanvas] 序列化节点[1] - getData(): {type: "navigate", url: "http://www.qq.com"}
```

### 执行任务时的日志

```
[FlowCanvas] ========== 开始执行任务 ==========
[FlowCanvas] 节点[1] - Data.url = http://www.qq.com (type: string)
```

后端日志：
```
[Workflow] 节点[1] - Data[url] = http://www.qq.com (type: string)
[Workflow] ✓ 执行导航: http://www.qq.com
```

## 注意事项

1. **必须点击"保存"按钮**：属性面板的"保存"按钮会触发数据更新
2. **数据自动序列化**：保存后会自动触发 `change` 事件，无需手动保存任务
3. **Toast 提示**：所有操作都会有友好的 Toast 提示，不再使用 alert

## 如果仍有问题

请提供：
1. 浏览器控制台的完整日志
2. localStorage 中的数据（执行上面的验证脚本）
3. 后端日志

这样可以快速定位问题。
