# 工作流节点开发指南

## 概述

本文档提供添加新工作流节点的完整步骤，确保节点功能完整且连接点正常显示。

## 常见问题

❌ **节点连接点消失**  
❌ **无法连接到其他节点**  
❌ **节点拖拽后无法显示**

**原因**：遗漏了必要的配置文件修改

## 开发步骤（必须按顺序完成）

### 1. 添加节点配置

**文件**: `frontend/src/components/workflow/nodeConfigs.ts`

```typescript
{ 
  type: 'your_node_type',           // 节点类型（唯一标识）
  label: '节点名称',                 // 显示名称
  icon: '🔧',                        // emoji 图标
  color: 'rgba(255, 87, 34, 0.15)', // 背景色（半透明）
  description: '节点功能描述'        // 简短描述
}
```

**示例**：
```typescript
{ type: 'decrypt', label: '解密', icon: '🔓', color: 'rgba(255, 87, 34, 0.15)', description: '解密数据' }
```

---

### 2. 添加 TypeScript 类型定义 ⚠️ **关键步骤**

**文件**: `frontend/src/types/workflow.ts`

在 `NodeType` 类型中添加新节点类型：

```typescript
export type NodeType = 
  | 'start'
  | 'end'
  | 'your_node_type'  // ← 添加这里
  | ...
```

**⚠️ 如果遗漏此步骤，节点连接点将无法显示！**

---

### 3. 添加属性面板配置

**文件**: `frontend/src/components/workflow/PropertyPanel.vue`

#### 3.1 添加表单模板

在 `<template>` 中添加：

```vue
<!-- 你的节点名称 -->
<template v-if="nodeType === 'your_node_type'">
  <div class="form-item">
    <label>参数名称</label>
    <input v-model="formData.paramName" placeholder="提示文本" />
    <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">
      帮助说明
    </small>
  </div>
  
  <div class="form-item">
    <label>下拉选择</label>
    <select v-model="formData.selectParam">
      <option value="option1">选项1</option>
      <option value="option2">选项2</option>
    </select>
  </div>
</template>
```

#### 3.2 设置默认值

在 `watch(() => props.nodeData, ...)` 中添加：

```typescript
if (!formData.value.yourParam) {
  formData.value.yourParam = 'defaultValue'
}
```

在 `else` 分支中也添加默认值：

```typescript
} else {
  formData.value = { 
    ..., 
    yourParam: 'defaultValue' 
  }
}
```

---

### 4. 注册后端动作

**文件**: `workflow_actions.go`

在 `executeStep` 函数的 `switch` 中添加：

```go
case "your_node_type":
    return we.executeYourNodeType(step)
```

---

### 5. 实现后端逻辑

**文件**: `workflow_actions_your_feature.go` (新建文件)

```go
package main

import (
    "fmt"
    "log"
)

// executeYourNodeType 执行你的节点逻辑
func (we *WorkflowExecutor) executeYourNodeType(step ExecutionStep) (ExecutionResult, error) {
    // 1. 获取参数
    param, ok := step.Params["paramName"].(string)
    if !ok || param == "" {
        return ExecutionResult{Success: false}, fmt.Errorf("缺少参数")
    }

    // 2. 执行业务逻辑
    result := doSomething(param)

    // 3. 保存到变量（可选）
    if saveToVariable, ok := step.Params["saveToVariable"].(string); ok && saveToVariable != "" {
        we.variables[saveToVariable] = result
        log.Printf("[Workflow] ✓ 结果已保存到变量: %s", saveToVariable)
    }

    // 4. 返回结果
    return ExecutionResult{
        Success: true,
        Message: "执行成功",
        Data: map[string]interface{}{
            "result": result,
        },
    }, nil
}
```

---

## 完整示例：解密节点

### 1. nodeConfigs.ts
```typescript
{ type: 'decrypt', label: '解密', icon: '🔓', color: 'rgba(255, 87, 34, 0.15)', description: '解密数据' }
```

### 2. workflow.ts
```typescript
export type NodeType = 
  | ...
  | 'decrypt'
```

### 3. PropertyPanel.vue
```vue
<template v-if="nodeType === 'decrypt'">
  <div class="form-item">
    <label>数据来源</label>
    <input v-model="formData.dataVariable" placeholder="变量名" />
  </div>
  <div class="form-item">
    <label>解密算法</label>
    <select v-model="formData.algorithm">
      <option value="sm4-ecb">SM4-ECB</option>
      <option value="sm4-cbc">SM4-CBC</option>
    </select>
  </div>
</template>
```

### 4. workflow_actions.go
```go
case "decrypt":
    return we.executeDecrypt(step)
```

### 5. workflow_actions_decrypt.go
```go
func (we *WorkflowExecutor) executeDecrypt(step ExecutionStep) (ExecutionResult, error) {
    // 实现逻辑
}
```

---

## 开发检查清单

使用此清单确保所有步骤完成：

- [ ] 1. 在 `nodeConfigs.ts` 中添加节点配置
- [ ] 2. 在 `workflow.ts` 中添加 TypeScript 类型 ⚠️
- [ ] 3. 在 `PropertyPanel.vue` 中添加表单和默认值
- [ ] 4. 在 `workflow_actions.go` 中注册动作
- [ ] 5. 创建 `workflow_actions_xxx.go` 实现逻辑
- [ ] 6. 测试节点拖拽和连接
- [ ] 7. 测试节点执行功能

---

## 常见错误排查

### 问题：节点连接点不显示

**原因**：未在 `workflow.ts` 中添加类型定义

**解决**：
```typescript
// frontend/src/types/workflow.ts
export type NodeType = 
  | ...
  | 'your_node_type'  // ← 添加这行
```

### 问题：节点拖拽后无法显示

**原因**：`nodeConfigs.ts` 中配置错误或缺失

**解决**：检查配置格式是否正确，特别是 `type` 字段

### 问题：属性面板不显示

**原因**：`PropertyPanel.vue` 中的 `nodeType` 判断错误

**解决**：确保 `v-if="nodeType === 'your_node_type'"` 中的类型名称一致

### 问题：后端执行报错"未知的操作类型"

**原因**：未在 `workflow_actions.go` 中注册

**解决**：在 `executeStep` 的 `switch` 中添加 case

---

## 节点类型命名规范

- 使用小写字母和下划线：`download_captured`
- 见名知意：`decrypt`, `extract`, `navigate`
- 避免缩写：使用 `intercept_request` 而非 `int_req`

---

## 变量系统

### 保存变量
```go
if saveToVariable, ok := step.Params["saveToVariable"].(string); ok && saveToVariable != "" {
    we.variables[saveToVariable] = yourData
    log.Printf("[Workflow] ✓ 变量已保存: %s", saveToVariable)
}
```

### 读取变量
```go
data, exists := we.variables[variableName]
if !exists {
    return ExecutionResult{Success: false}, fmt.Errorf("变量 %s 不存在", variableName)
}
```

### 前端引用变量
用户在配置中使用 `{variableName}` 引用变量，后端通过 `replaceVariables` 自动替换。

---

## 参考现有节点

- **简单节点**: `wait` (等待) - 最简单的实现
- **数据提取**: `extract` (提取数据) - 变量保存示例
- **复杂逻辑**: `download` (下载) - 多参数处理
- **条件分支**: `if` (条件判断) - 特殊执行流程
- **循环节点**: `jsonl_reader` (JSONL读取) - 循环实现

---

## 测试建议

1. **拖拽测试**：从组件面板拖拽节点到画布
2. **连接测试**：连接到其他节点，检查连接点是否正常
3. **配置测试**：点击节点，检查属性面板是否显示
4. **保存测试**：保存工作流，重新加载检查节点是否正常
5. **执行测试**：运行工作流，检查节点是否正确执行
6. **变量测试**：测试变量保存和读取功能

---

## 最佳实践

1. **错误处理**：所有参数都要验证，返回清晰的错误信息
2. **日志记录**：关键步骤使用 `log.Printf` 记录
3. **变量命名**：使用有意义的变量名，如 `decryptedData` 而非 `data1`
4. **用户提示**：在表单中添加 `<small>` 标签提供帮助信息
5. **默认值**：为所有参数提供合理的默认值
6. **类型安全**：使用类型断言时检查 `ok` 返回值

---

## 相关文档

- [工作流执行流程](./WORKFLOW_EXECUTION.md)
- [工作流变量系统](./WORKFLOW_VARIABLES.md)
- [解密功能使用指南](./WORKFLOW_DECRYPT_GUIDE.md)
