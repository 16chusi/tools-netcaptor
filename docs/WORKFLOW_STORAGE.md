# 工作流持久化方案

## 技术选型
使用 SQLite 作为本地持久化存储，数据库文件位于 `~/.netcaptor/workflow.db`

## 数据表结构

### workflow_tasks
| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT | 主键，任务ID |
| name | TEXT | 任务名称 |
| description | TEXT | 任务描述 |
| created_at | TEXT | 创建时间 |
| updated_at | TEXT | 更新时间 |
| nodes_json | TEXT | 节点数据（JSON） |
| edges_json | TEXT | 边数据（JSON） |

## API 接口

### 后端 (Go)
- `SaveWorkflowTask(task)` - 保存任务
- `GetWorkflowTask(id)` - 获取单个任务
- `GetAllWorkflowTasks()` - 获取所有任务
- `DeleteWorkflowTask(id)` - 删除任务

### 前端 (Vue)
通过 Wails 自动生成的绑定调用后端接口

## 使用方式

1. **创建任务**: 自动保存到数据库
2. **编辑任务**: 点击"保存"按钮持久化
3. **加载任务**: 组件挂载时自动从数据库加载

## 数据流
```
前端 → Wails Bridge → Go Backend → SQLite
```

## 优势
- 轻量级，无需额外服务
- 支持并发访问
- 数据持久化可靠
- 跨平台兼容
