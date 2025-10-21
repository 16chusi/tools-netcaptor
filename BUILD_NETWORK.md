# 构建和运行网络抓包工具

## 前置要求

1. Go 1.18+
2. Node.js 16+
3. Wails CLI v2

## 安装 Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## 构建步骤

### 1. 安装前端依赖

```bash
cd inspdinfo/frontend
npm install
```

### 2. 开发模式运行

```bash
cd inspdinfo
wails dev
```

### 3. 生产构建

```bash
cd inspdinfo
wails build
```

构建完成后,可执行文件位于 `build/bin/` 目录

## 项目结构

```
inspdinfo/
├── frontend/                    # 前端代码
│   ├── src/
│   │   ├── components/
│   │   │   └── NetworkCapture.vue   # 网络抓包组件
│   │   ├── App.vue              # 主应用
│   │   └── main.ts
│   └── package.json
├── app.go                       # 原有应用逻辑
├── browser.go                   # Playwright 浏览器管理
├── chromedp_browser.go          # ChromeDP 浏览器管理
├── network_capture.go           # 网络捕获核心逻辑
├── proxy_server.go              # 本地代理服务器
├── network_app.go               # 网络应用控制器
├── downloader.go                # 下载管理器
├── main.go                      # 主入口
└── wails.json                   # Wails 配置
```

## 功能说明

### 1. 自动下载工具 (原有功能)
- 使用 Playwright 或 ChromeDP 自动化浏览器
- 抓取网页下载链接
- 批量下载文件
- 支持翻页

### 2. 网络抓包工具 (新功能)
- **前端拦截模式**: 在 iframe 中加载网页,拦截 fetch/XHR 请求
- **代理模式**: 启动本地代理服务器,捕获所有网络流量
- 实时显示请求和响应
- 支持导出 JSON 格式数据

## 使用方法

### 启动应用

```bash
# 开发模式
wails dev

# 或运行构建后的可执行文件
./build/bin/inspdinfo
```

### 测试网络抓包

1. 切换到"网络抓包工具"标签页
2. 输入测试页面地址: `file:///path/to/test_network.html`
3. 点击"加载"按钮
4. 在测试页面中点击按钮发送请求
5. 在右侧面板查看捕获的请求和响应

### 使用代理模式

1. 点击"启动代理"按钮
2. 配置浏览器使用代理 `http://127.0.0.1:8888`
3. 访问任意网站,所有请求将被捕获

## 技术栈

### 后端
- **Go**: 主要编程语言
- **Wails v2**: 桌面应用框架
- **net/http**: HTTP 代理服务器
- **sync**: 并发安全

### 前端
- **Vue 3**: UI 框架
- **TypeScript**: 类型安全
- **Vite**: 构建工具

## 已知限制

1. **跨域限制**: 前端拦截模式无法注入脚本到跨域页面
2. **HTTPS 代理**: 当前版本的代理模式对 HTTPS 支持有限
3. **性能**: 大量请求可能影响性能
4. **数据大小**: 响应体默认只保存前 5KB

## 故障排除

### 问题: 看不到任何请求

**原因**: 跨域限制或脚本注入失败

**解决方案**:
- 使用代理模式
- 测试同源页面
- 使用提供的 `test_network.html` 测试

### 问题: 代理模式无法启动

**原因**: 端口被占用

**解决方案**:
```bash
# 检查端口占用
lsof -i :8888

# 修改 proxy_server.go 中的端口号
```

### 问题: Wails 构建失败

**原因**: 依赖未安装或版本不兼容

**解决方案**:
```bash
# 重新安装依赖
cd frontend
rm -rf node_modules package-lock.json
npm install

# 更新 Go 依赖
go mod tidy
```

## 开发建议

### 添加新功能

1. 后端: 在 `network_app.go` 中添加新方法
2. 前端: 在 `NetworkCapture.vue` 中调用
3. 重新生成绑定: `wails dev` 会自动生成

### 调试

```bash
# 启用详细日志
wails dev -v

# 查看前端控制台
# 在应用中按 F12 打开开发者工具
```

## 贡献

欢迎提交 Issue 和 Pull Request!

## 许可证

MIT
