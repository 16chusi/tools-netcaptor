# 工作流节点和连线删除功能

## 功能概述

为 NetCaptor 工作流编辑器添加了完整的节点和连线删除功能，支持多种删除方式。

## 实现的功能

### 1. 🖱️ 工具栏删除按钮
- **位置**：工具栏中的"🗑️ 删除"按钮
- **功能**：删除当前选中的所有节点和连线
- **状态**：
  - 未选中任何元素时按钮禁用（灰色）
  - 选中元素后按钮激活（红色边框）
- **使用方法**：
  1. 点击选择一个或多个节点/连线
  2. 点击工具栏的"🗑️ 删除"按钮

### 2. ⌨️ 键盘快捷键删除
- **快捷键**：`Delete` 或 `Backspace`
- **功能**：快速删除选中的节点和连线
- **使用方法**：
  1. 点击选择节点或连线
  2. 按下 `Delete` 或 `Backspace` 键

### 3. 🖱️ 右键菜单删除
- **触发方式**：右键点击节点或连线
- **功能**：显示上下文菜单，提供删除选项
- **使用方法**：
  1. 右键点击要删除的节点或连线
  2. 在弹出菜单中点击"🗑️ 删除"

### 4. 🔲 多选删除
- **功能**：支持同时选择和删除多个元素
- **使用方法**：
  - **框选**：按住鼠标左键拖动，框选多个节点
  - **Ctrl/Cmd + 点击**：按住 Ctrl（Windows/Linux）或 Cmd（Mac）点击多个节点
  - 选中后使用上述任意删除方式

## 技术实现

### 依赖安装
```bash
npm install @antv/x6-plugin-selection --save
```

### 核心代码

#### 1. 导入 Selection 插件
```typescript
import { Selection } from '@antv/x6-plugin-selection'
```

#### 2. 启用选择功能
```typescript
graph.use(
  new Selection({
    enabled: true,        // 启用选择
    multiple: true,       // 支持多选
    rubberband: true,     // 启用框选
    movable: true,        // 选中后可移动
    showNodeSelectionBox: true,  // 显示选择框
  })
)
```

#### 3. 监听选择状态
```typescript
graph.on('selection:changed', ({ selected }) => {
  hasSelection.value = selected.length > 0
})
```

#### 4. 删除功能实现
```typescript
function deleteSelected() {
  if (!graph) return
  
  const selectedCells = graph.getSelectedCells()
  if (selectedCells.length === 0) return
  
  selectedCells.forEach(cell => {
    graph!.removeCell(cell)
  })
  
  hasSelection.value = false
}
```

#### 5. 键盘绑定
```typescript
graph.bindKey(['backspace', 'delete'], () => {
  deleteSelected()
})
```

#### 6. 右键菜单
```typescript
graph.on('node:contextmenu', ({ node, e }) => {
  e.preventDefault()
  showContextMenu(node, e)
})

graph.on('edge:contextmenu', ({ edge, e }) => {
  e.preventDefault()
  showContextMenu(edge, e)
})
```

## 用户体验优化

### 视觉反馈
- ✅ 选中的节点显示蓝色边框
- ✅ 删除按钮根据选择状态动态启用/禁用
- ✅ 右键菜单样式美观，带阴影效果
- ✅ 按钮悬停时有颜色变化

### 交互优化
- ✅ 支持框选多个节点
- ✅ 支持 Ctrl/Cmd + 点击多选
- ✅ 删除后自动清除选择状态
- ✅ 右键菜单点击外部自动关闭

### 样式定义
```css
/* 删除按钮样式 */
.toolbar-btn.danger {
  color: #ff4d4f;
  border-color: #ff4d4f;
}

.toolbar-btn.danger:hover:not(:disabled) {
  background: #ff4d4f;
  color: white;
}

.toolbar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 右键菜单样式 */
.context-menu {
  position: fixed;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  z-index: 9999;
  min-width: 120px;
}

.menu-item {
  padding: 8px 16px;
  cursor: pointer;
  font-size: 12px;
  transition: background 0.2s;
}

.menu-item:hover {
  background: #f5f5f5;
}
```

## 使用场景

### 场景 1：删除单个节点
1. 点击选中节点
2. 按 `Delete` 键或点击工具栏删除按钮

### 场景 2：删除多个节点
1. 按住鼠标左键拖动框选多个节点
2. 按 `Delete` 键删除所有选中节点

### 场景 3：删除连线
1. 点击选中连线（连线会高亮显示）
2. 右键点击选择"删除"或按 `Delete` 键

### 场景 4：快速清理
1. 使用框选选中不需要的节点和连线
2. 按 `Backspace` 键快速删除

## 注意事项

⚠️ **重要提示**
- 删除操作不可撤销，请谨慎操作
- 删除节点会自动删除与其相连的所有连线
- 建议在删除前先保存工作流

💡 **最佳实践**
- 使用框选功能批量删除多个节点
- 使用右键菜单精确删除单个元素
- 定期保存工作流以防误删

## 测试建议

### 功能测试
- [ ] 测试单个节点删除
- [ ] 测试多个节点删除
- [ ] 测试连线删除
- [ ] 测试框选删除
- [ ] 测试键盘快捷键
- [ ] 测试右键菜单
- [ ] 测试删除按钮状态切换

### 边界测试
- [ ] 删除开始节点
- [ ] 删除结束节点
- [ ] 删除所有节点
- [ ] 空画布时点击删除按钮

## 后续优化建议

1. **撤销/重做功能**
   - 添加 Ctrl+Z 撤销删除
   - 添加 Ctrl+Y 重做删除

2. **删除确认**
   - 删除重要节点时弹出确认对话框
   - 批量删除时显示数量提示

3. **快捷键提示**
   - 在工具栏按钮上显示快捷键提示
   - 添加键盘快捷键帮助面板

4. **删除动画**
   - 添加节点删除的淡出动画
   - 提升用户体验

## 相关文件

- `frontend/src/components/workflow/FlowCanvas.vue` - 主要实现文件
- `frontend/package.json` - 依赖配置

## 版本信息

- **功能版本**: v1.0.0
- **实现日期**: 2024
- **依赖版本**: 
  - @antv/x6: ^2.18.1
  - @antv/x6-plugin-selection: ^2.1.7
