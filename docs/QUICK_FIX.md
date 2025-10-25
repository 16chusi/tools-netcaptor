# 快速修复：缺少 URL 参数

## 问题确认

后端日志显示：
```
step.Params: map[type:navigate]  ❌ 只有 type，没有 url
```

这说明前端保存的节点数据中没有包含 `url` 字段。

## 立即解决方案

### 方案 1: 使用测试工具创建正确的任务

1. 在浏览器中打开：
   ```
   file:///home/fzxs/workspaces/duola-tools/netcaptor/tools-netcaptor/test-node-data.html
   ```

2. 点击"创建测试任务"按钮

3. 点击"验证数据"确认 URL 已保存

4. 回到应用，刷新页面

5. 打开"任务流编排"，应该能看到测试任务

6. 点击"运行"

### 方案 2: 手动修复 localStorage 数据

在浏览器控制台执行：

```javascript
// 1. 读取当前任务
const tasks = JSON.parse(localStorage.getItem('workflow-tasks') || '[]');

// 2. 找到第一个任务
const task = tasks[0];

// 3. 找到导航节点并添加 URL
if (task && task.nodes) {
  const navNode = task.nodes.find(n => n.type === 'navigate');
  if (navNode) {
    // 确保 data 对象存在
    if (!navNode.data) {
      navNode.data = {};
    }
    // 添加 URL
    navNode.data.url = 'http://www.qq.com';
    navNode.data.type = 'navigate';
    
    console.log('已修复导航节点:', navNode);
  }
}

// 4. 保存回 localStorage
localStorage.setItem('workflow-tasks', JSON.stringify(tasks));

console.log('✓ 数据已修复，请刷新页面');
```

### 方案 3: 重新创建任务（正确步骤）

1. **清除旧数据**
   ```javascript
   localStorage.clear()
   ```
   然后刷新页面

2. **创建新任务**
   - 打开"任务流编排"
   - 点击"+ 新建"

3. **添加导航节点**
   - 从左侧拖拽"导航"节点到画布
   - 连接：开始 → 导航 → 结束

4. **设置 URL（关键步骤）**
   - **点击"导航"节点**（节点会被选中）
   - 右侧会弹出属性面板
   - 在"目标URL"输入框输入：`http://www.qq.com`
   - **点击属性面板底部的"保存"按钮**（重要！）
   - 等待看到控制台日志：`[WorkflowDrawer] ✓ 节点数据已更新`

5. **保存任务**
   - **点击画布工具栏的"💾 保存"按钮**（重要！）
   - 等待看到"保存成功"提示

6. **验证数据**
   在控制台执行：
   ```javascript
   const tasks = JSON.parse(localStorage.getItem('workflow-tasks'));
   const navNode = tasks[0].nodes.find(n => n.type === 'navigate');
   console.log('导航节点 data:', navNode.data);
   // 应该看到: {type: "navigate", url: "http://www.qq.com"}
   ```

7. **执行任务**
   - 点击"▶️ 运行"按钮

## 检查清单

执行前请确认：

- [ ] 已清除 localStorage
- [ ] 已创建新任务
- [ ] 已添加导航节点
- [ ] 已点击导航节点（节点被选中）
- [ ] 已在属性面板输入 URL
- [ ] 已点击属性面板的"保存"按钮
- [ ] 已点击画布的"💾 保存"按钮
- [ ] 已验证 localStorage 中的数据包含 url 字段

## 调试日志检查

### 保存时应该看到的日志：

```
[PropertyPanel] 保存节点数据: navigate {url: "http://www.qq.com"}
[WorkflowDrawer] ========== 保存节点数据 ==========
[WorkflowDrawer] 保存的数据: {url: "http://www.qq.com"}
[WorkflowDrawer] 合并后的数据: {type: "navigate", url: "http://www.qq.com"}
[WorkflowDrawer] ✓ 节点数据已更新
```

### 执行时应该看到的日志：

```
[FlowCanvas] 节点[1] - Data.url = http://www.qq.com (type: string)
```

后端日志：
```
[Workflow] 节点[1] - Data[url] = http://www.qq.com (type: string)
[Workflow] step.Params 键值对:
[Workflow]   - url = http://www.qq.com (类型: string)
[Workflow] ✓ 执行导航: http://www.qq.com
```

## 如果问题仍然存在

请提供以下信息：

1. 浏览器控制台的完整日志（从点击"保存"到点击"运行"）
2. 执行以下命令的输出：
   ```javascript
   JSON.parse(localStorage.getItem('workflow-tasks'))
   ```
3. 后端的完整日志

这样我可以精确定位问题所在。
