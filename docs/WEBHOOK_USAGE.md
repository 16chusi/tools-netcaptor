# Webhook 服务使用说明

## 功能概述

Webhook 服务提供 HTTP 接口，允许页面通过 POST 请求发送数据到应用程序。同时提供测试页面用于验证网络抓包功能。

## 启动服务

1. 打开设置面板（点击工具栏的 ⚙️ 按钮）
2. 找到 "Webhook 服务" 选项
3. 点击"启动"按钮
4. 服务将在随机端口启动，端口号会显示在设置面板中

## API 接口

### 端点
```
GET/POST http://127.0.0.1:{PORT}/webhook
```

注：`{PORT}` 为服务启动后分配的随机端口，可在设置面板中查看。

支持两种请求方式：
- **GET**: 通过 URL 参数传递数据
- **POST**: 通过 JSON Body 传递数据

### 其他端点

- `GET http://127.0.0.1:{PORT}/` - 测试页面（包含 Webhook 测试功能）
- `GET http://127.0.0.1:{PORT}/api/test` - API 测试接口
- `GET http://127.0.0.1:{PORT}/api/data` - API 数据接口

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 否 | 操作类型：`save` 或 `print`，默认 `print` |
| file | string | 否 | 文件名（action=save 时必填） |
| data | string | 是 | 要传递的数据 |
| type | string | 否 | 数据类型：`txt`、`json`、`base64`、`hex`，默认 `txt` |

### 操作说明

#### 1. save - 保存到文件
将数据保存到指定文件，文件保存在程序根目录。

**POST 请求示例：**
```javascript
// 保存 JSON 数据
fetch('http://127.0.0.1:PORT/webhook', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    action: 'save',
    file: 'output.json',
    data: JSON.stringify({ message: 'Hello World' }),
    type: 'json'
  })
})

// 保存 Base64 图片
fetch('http://127.0.0.1:PORT/webhook', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    action: 'save',
    file: 'image.png',
    data: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
    type: 'base64'
  })
})
```

**GET 请求示例：**
```javascript
// 使用 fetch 保存文本文件
fetch('http://127.0.0.1:PORT/webhook?action=save&file=test.txt&data=Hello+World')

// 使用 fetch 保存 JSON 文件
const params = new URLSearchParams({
  action: 'save',
  file: 'data.json',
  data: JSON.stringify({ key: 'value' }),
  type: 'json'
})
fetch(`http://127.0.0.1:PORT/webhook?${params}`)
```

```bash
# 使用 curl
curl "http://127.0.0.1:PORT/webhook?action=save&file=test.txt&data=Hello+World"
```

#### 2. print - 打印到控制台
将数据打印到应用程序的控制台日志。

**POST 请求示例：**
```javascript
// 打印调试信息（使用默认值）
fetch('http://127.0.0.1:PORT/webhook', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    data: 'Debug message: ' + new Date().toISOString()
  })
})

// 明确指定 action 和 type
fetch('http://127.0.0.1:PORT/webhook', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    action: 'print',
    data: 'Debug message',
    type: 'txt'
  })
})
```

**GET 请求示例：**
```javascript
// 使用 fetch 打印文本（使用默认值）
fetch('http://127.0.0.1:PORT/webhook?data=Hello+World')

// 使用 URLSearchParams 构建参数
const params = new URLSearchParams({
  action: 'print',
  data: 'Debug info',
  type: 'txt'
})
fetch(`http://127.0.0.1:PORT/webhook?${params}`)

// 在浏览器中直接访问
window.location.href = 'http://127.0.0.1:PORT/webhook?data=Test+message'
```

```bash
# 使用 curl
curl "http://127.0.0.1:PORT/webhook?data=Hello+World"
```

### 数据类型

- **txt**: 纯文本（默认）
- **json**: JSON 字符串
- **base64**: Base64 编码的数据
- **hex**: 十六进制编码的数据

### 响应格式

**操作成功响应：**
```json
{
  "success": true,
  "message": "文件已保存: output.json"
}
```

**服务状态响应（GET 无参数）：**
```json
{
  "status": "ok",
  "message": "Webhook服务运行中",
  "port": 12345
}
```

**错误响应：**
```json
{
  "error": "错误信息"
}
```

### 快速测试

```bash
# 1. 检查服务状态
curl "http://127.0.0.1:PORT/webhook"

# 2. 打印消息到控制台
curl "http://127.0.0.1:PORT/webhook?data=Hello"

# 3. 保存文件
curl "http://127.0.0.1:PORT/webhook?action=save&file=test.txt&data=Hello+World"
```

## 使用场景

1. **数据采集**: 从网页中提取数据并保存到本地文件
2. **调试日志**: 将页面运行时信息发送到应用程序控制台
3. **自动化测试**: 在工作流中保存测试结果
4. **数据导出**: 批量导出处理后的数据
5. **网络测试**: 使用内置测试页面验证抓包功能

## 安全提示

- Webhook 服务仅监听本地地址 (127.0.0.1)
- 不对外网开放，仅本机访问
- 建议仅在需要时启动服务
- 使用随机端口提高安全性
