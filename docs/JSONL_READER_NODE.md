# JSONL 读取器节点实现文档

## 功能概述

JSONL 读取器节点用于读取 JSONL（JSON Lines）格式文件，并逐行循环执行后续工作流节点。每读取一行数据，就会触发一次后续流程的执行，实现数据驱动的批量自动化操作。

## 核心特性

1. **文件读取**: 通过系统文件选择器选择 JSONL 文件
2. **字段提取**: 支持提取指定字段或全部字段（使用 `*`）
3. **循环控制**: 
   - 可配置循环间隔时间（默认 100ms）
   - 可限制最大循环次数（默认为文件总行数）
4. **变量传递**: 提取的字段值自动注入到变量系统，供后续节点使用
5. **实时预览**: 加载文件后显示可用字段和总行数

## 使用方法

### 1. 添加节点

在工作流画布中，从节点面板拖拽 "JSONL读取器" 节点到画布。

### 2. 配置节点

点击节点打开属性面板，配置以下参数：

- **JSONL文件**: 点击"选择"按钮，从系统中选择 `.jsonl` 文件
- **加载可用字段**: 选择文件后，点击此按钮加载文件并显示所有可用字段
- **提取字段**: 
  - 输入要提取的字段名，多个字段用逗号分隔，如：`name,age,email`
  - 输入 `*` 表示提取全部字段
- **循环间隔(ms)**: 每次循环之间的等待时间，默认 100 毫秒
- **最大循环次数**: 限制循环次数，输入 0 表示处理全部数据

### 3. 连接后续节点

JSONL 读取器节点之后的所有节点都会被循环执行。例如：

```
[开始] → [JSONL读取器] → [导航] → [输入文本] → [点击] → [结束]
```

在这个流程中，导航、输入文本、点击这三个节点会针对 JSONL 文件的每一行数据执行一次。

### 4. 使用提取的变量

在后续节点中，可以通过 `{字段名}` 的方式引用提取的数据。例如：

JSONL 文件内容：
```jsonl
{"name": "张三", "url": "https://example.com/user/1"}
{"name": "李四", "url": "https://example.com/user/2"}
```

配置提取字段为 `name,url`，则在后续节点中可以使用：
- 导航节点的 URL: `{url}`
- 输入节点的文本: `{name}`

## 实现细节

### 后端实现

#### 1. JSONL 读取器 (`jsonl_reader.go`)

```go
type JSONLReader struct {
    filePath string
    lines    []map[string]interface{}
}
```

核心方法：
- `Load()`: 加载并解析 JSONL 文件
- `GetKeys()`: 获取所有可用的字段名
- `GetLineCount()`: 获取总行数
- `GetLine(index)`: 获取指定行的数据
- `ExtractValue(data, keys)`: 提取指定字段的值

#### 2. 工作流执行器扩展 (`workflow_executor.go`)

新增方法：
- `executeJSONLReaderLoop()`: 实现循环读取和执行逻辑

循环执行流程：
1. 加载 JSONL 文件
2. 解析提取字段配置
3. 遍历每一行数据
4. 提取字段值并注入到变量系统
5. 执行 JSONL 读取器之后的所有步骤
6. 等待配置的间隔时间
7. 继续下一行

#### 3. NetworkApp 绑定 (`network_app.go`)

新增 Wails 绑定方法：
- `SelectJSONLFile()`: 打开文件选择对话框
- `LoadJSONLFile(filePath)`: 加载文件并返回字段信息

### 前端实现

#### 1. 类型定义 (`workflow.ts`)

在 `NodeType` 中添加：
```typescript
| 'jsonl_reader'
```

#### 2. 节点配置 (`nodeConfigs.ts`)

```typescript
{ 
  type: 'jsonl_reader', 
  label: 'JSONL读取器', 
  icon: '📄', 
  color: '#b37feb', 
  description: '读取JSONL文件并循环' 
}
```

#### 3. 属性面板 (`PropertyPanel.vue`)

新增配置表单：
- 文件选择按钮
- 加载字段按钮
- 字段显示区域
- 提取字段输入框
- 循环间隔输入框
- 最大次数输入框
- 文件信息显示

## 使用示例

### 示例 1: 批量访问 URL

JSONL 文件 (`urls.jsonl`):
```jsonl
{"url": "https://example.com/page1", "title": "页面1"}
{"url": "https://example.com/page2", "title": "页面2"}
{"url": "https://example.com/page3", "title": "页面3"}
```

工作流配置：
1. JSONL读取器节点：
   - 文件: `urls.jsonl`
   - 提取字段: `url,title`
   - 循环间隔: 2000ms
2. 导航节点：
   - URL: `{url}`
3. 等待节点：
   - 时间: 3000ms

### 示例 2: 批量表单填写

JSONL 文件 (`users.jsonl`):
```jsonl
{"username": "user1", "email": "user1@example.com", "phone": "13800138001"}
{"username": "user2", "email": "user2@example.com", "phone": "13800138002"}
```

工作流配置：
1. 导航节点: 打开注册页面
2. JSONL读取器节点：
   - 文件: `users.jsonl`
   - 提取字段: `*`
   - 循环间隔: 500ms
3. 输入节点（用户名）: `{username}`
4. 输入节点（邮箱）: `{email}`
5. 输入节点（手机）: `{phone}`
6. 点击节点: 提交按钮

## 注意事项

1. **文件格式**: 确保文件是标准的 JSONL 格式，每行一个有效的 JSON 对象
2. **循环间隔**: 根据目标网站的响应速度合理设置间隔，避免请求过快
3. **错误处理**: 如果某一行数据处理失败，会记录日志但继续处理下一行
4. **变量覆盖**: 每次循环会覆盖之前的变量值
5. **性能考虑**: 大文件会一次性加载到内存，注意文件大小

## 后续优化方向

1. 支持流式读取大文件
2. 添加错误重试机制
3. 支持条件过滤（跳过某些行）
4. 支持嵌套字段提取（如 `user.name`）
5. 添加循环进度显示
6. 支持暂停/恢复功能
