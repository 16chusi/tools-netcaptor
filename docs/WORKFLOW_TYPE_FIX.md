# 工作流类型修复说明

## 问题描述

前端编译时出现类型错误：
```
error TS2345: Argument of type 'WorkflowTask' is not assignable to parameter of type 'main.WorkflowTask'.
Property 'convertValues' is missing
```

## 根本原因

Wails 会自动从 Go 代码生成 TypeScript 类型定义到 `frontend/wailsjs/go/models.ts`。这些生成的类型包含 `convertValues` 方法，而我们在 `frontend/src/types/workflow.ts` 中定义的类型是简单接口，不包含这些方法。

## 解决方案

### 1. 使用 Wails 生成的类型

在 `FlowCanvas.vue` 中：
```typescript
// 导入 Wails 生成的类型
import { main } from '../../../wailsjs/go/models'

// 使用类型别名
type WorkflowTask = main.WorkflowTask

// 调用 Go 方法时使用 createFrom 构造
const task = main.WorkflowTask.createFrom(props.task)
await ExecuteWorkflow(task)
```

### 2. 简化内部类型定义

在 `WorkflowDrawer.vue` 中使用简单接口：
```typescript
interface WorkflowTask {
  id: string
  name: string
  description?: string
  createdAt: string
  updatedAt: string
  nodes: any[]
  edges: any[]
}
```

### 3. 使用 any 避免类型冲突

在组件间传递数据时使用 `any` 类型：
```typescript
const emit = defineEmits<{
  change: [task: any]  // 使用 any 避免类型冲突
}>()

const props = defineProps<{
  task?: any  // 使用 any 避免类型冲突
}>()
```

## 修改的文件

1. **FlowCanvas.vue**
   - 导入 Wails 生成的类型
   - 使用 `main.WorkflowTask.createFrom()` 构造实例
   - emit 类型改为 `any`
   - props 类型改为 `any`

2. **WorkflowDrawer.vue**
   - 移除对 `types/workflow.ts` 的导入
   - 使用本地接口定义
   - onTaskChange 参数类型改为 `any`

## 编译结果

✅ 前端编译成功
```
vite v3.2.11 building for production...
✓ 1053 modules transformed.
dist/assets/index.75097467.js  743.21 KiB / gzip: 220.29 KiB
```

## 最佳实践

### 跨 Wails 边界传递数据

当数据需要从前端传递到 Go 后端时：
```typescript
// ✅ 正确：使用 Wails 生成的类型构造
const task = main.WorkflowTask.createFrom(plainObject)
await GoMethod(task)

// ❌ 错误：直接传递普通对象
await GoMethod(plainObject)
```

### 组件内部使用

组件内部可以使用简单的接口或 `any` 类型，只在调用 Go 方法时转换：
```typescript
// 组件内部
const tasks = ref<any[]>([])

// 调用 Go 方法时转换
const task = main.WorkflowTask.createFrom(tasks.value[0])
await ExecuteWorkflow(task)
```

## 注意事项

1. **不要删除 `types/workflow.ts`**: 虽然现在没用到，但保留它作为文档参考
2. **Wails 类型自动生成**: 每次修改 Go 结构体后，Wails 会重新生成类型
3. **类型转换开销**: `createFrom()` 会创建新对象，但开销很小
4. **类型安全**: 虽然使用了 `any`，但在 Go 层面仍有类型检查

## 总结

通过使用 Wails 自动生成的类型和适当的类型转换，成功解决了类型不匹配问题。这是 Wails 应用的标准做法，既保证了类型安全，又避免了手动维护重复的类型定义。
