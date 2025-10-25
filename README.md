# 🌐 NetCaptor

<div align="center">

**强大的网络抓包与分析工具**

基于 Wails + Go + Vue3 构建的跨平台网络流量捕获工具

[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Wails](https://img.shields.io/badge/Wails-v2.10.2-DF0000?style=flat&logo=wails)](https://wails.io/)
[![Vue](https://img.shields.io/badge/Vue-3.2-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [使用指南](#-使用指南) • [技术栈](#-技术栈) • [截图](#-截图)

</div>

---

## ✨ 功能特性

### 🔍 网络抓包
- **HTTP/HTTPS 拦截**：基于 GoProxy 的透明代理，支持 HTTPS 流量解密
- **实时监控**：实时捕获和显示网络请求
- **智能过滤**：按类型（Fetch/XHR、JS、CSS、图片、文档）快速筛选
- **详细信息**：查看完整的请求头、响应头、载荷和响应内容

### 📊 数据分析
- **Chrome DevTools 风格**：熟悉的界面设计，零学习成本
- **性能指标**：显示请求耗时、响应大小等关键指标
- **内容预览**：
  - JSON 自动格式化
  - HTML 渲染预览
  - 图片直接显示
  - PDF 浏览器打开
  - 多种编码查看（文本、十六进制、Base64）

### 🛠️ 开发工具
- **请求重放**：一键生成 cURL、PowerShell、Fetch 代码
- **快速复制**：复制请求代码到剪贴板
- **数据导出**：导出筛选后的网络数据为 JSON
- **多浏览器支持**：Chrome、Edge、Firefox 一键启动

### 🔐 安全特性
- **自签名证书**：自动生成和管理 CA 证书
- **跨平台证书安装**：支持 Windows、macOS、Linux
- **端口自定义**：避免端口冲突，随机分配测试端口

### ⚙️ 灵活配置
- **代理端口设置**：自定义代理监听端口
- **下载路径管理**：可视化选择下载目录
- **默认浏览器**：配置常用浏览器
- **自定义 URL**：设置默认打开的测试页面

### 🔄 工作流编排
- **可视化任务流**: 通过拖拽节点（如“导航”、“点击”、“输入文本”）来创建自动化任务流。
- **流程自动化**: 无需编码即可实现复杂的浏览器操作序列，用于自动化测试或数据采集。

### 🧩 浏览器插件联动
- **WebSocket 通信**: 内置 WebSocket 服务，可与配套的 [浏览器插件](../chrome-extension/README.md) 联动，实现更深度的浏览器集成与交互。

---

## 🚀 快速开始

### 环境要求

- **Go**: 1.23 或更高版本 (与 go.mod 一致)
- **Node.js**: 14 或更高版本
- **Wails CLI**: v2.10.2

### 一键安装

```bash
# 克隆项目
git clone <repository-url>
cd netcaptor/tools-netcaptor

# 一键安装所有依赖
chmod +x quick-start.sh
./quick-start.sh

# 启动应用
./run.sh
```

### 手动安装

#### 1. 安装 Go 环境

```bash
# 下载 Go 1.23+
wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz
tar -xzf go1.25.1.linux-amd64.tar.gz -C ~/go/

# 设置环境变量
export PATH="$HOME/go/go1.25.1/bin:$HOME/go/bin:$PATH"
export GOROOT="$HOME/go/go1.25.1"
export GOPATH="$HOME/go"
```

#### 2. 安装 Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

#### 3. 安装依赖

```bash
cd tools-netcaptor
go mod tidy
cd frontend
npm install
cd ..
```

#### 4. 安装 Playwright 浏览器

```bash
# 安装 Playwright 浏览器驱动（必需）
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install

# 如果只需要 Chromium
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install chromium
```

> **注意**：Playwright 需要下载浏览器驱动，首次安装可能需要几分钟时间。

#### 5. 启动开发模式

```bash
wails dev -tags webkit2_41
```

---

## 📖 使用指南

### 基础使用

1. **启动代理**
   - 点击工具栏的 ● 按钮启动代理服务器
   - 默认端口：8888（可在设置中修改）

2. **配置浏览器**
   - 点击"打开浏览器"按钮自动配置代理
   - 或手动设置浏览器代理为 `127.0.0.1:8888`

3. **捕获流量**
   - 在浏览器中访问任何网站
   - 请求会实时显示在列表中

4. **查看详情**
   - 点击任意请求查看详细信息
   - 切换标签页查看标头、载荷、响应、请求代码

### HTTPS 抓包

1. **安装证书**
   - 点击工具栏的 🔒 按钮
   - 按照系统对应的安装步骤操作
   - 证书位置：`~/.netcaptor/certs/netcaptor-ca.crt`

2. **系统安装指南**

   **Windows:**
   ```
   1. 双击 netcaptor-ca.crt 文件
   2. 点击"安装证书" → 选择"当前用户"
   3. 选择"将所有证书放入下列存储"
   4. 浏览并选择"受信任的根证书颁发机构"
   5. 完成安装并重启浏览器
   ```

   **macOS:**
   ```
   1. 双击 netcaptor-ca.crt 文件
   2. 在钥匙串访问中找到 "NetCaptor CA"
   3. 双击证书，展开"信任"
   4. 选择"始终信任"
   5. 重启浏览器
   ```

   **Linux:**
   ```bash
   sudo cp ~/.netcaptor/certs/netcaptor-ca.crt /usr/local/share/ca-certificates/netcaptor.crt
   sudo update-ca-certificates
   # 重启浏览器
   ```

### 高级功能

#### 过滤请求
- 使用顶部过滤标签快速筛选：全部、Fetch/XHR、JS、CSS、图片、文档、其他
- 在搜索框输入关键词进行文本过滤

#### 导出数据
- 点击 ⬇️ 按钮导出当前筛选的请求
- 支持 JSON 格式，包含完整的请求和响应信息

#### 生成请求代码
1. 选择任意请求
2. 切换到"请求"标签页
3. 选择格式：cURL、PowerShell、Fetch
4. 点击"复制"按钮

#### 响应内容查看
- **自动识别**：根据 Content-Type 自动选择查看方式
- **手动切换**：在下拉菜单中选择查看格式
- **PDF/图片**：点击"在浏览器中打开"或"下载并打开"

---

## 🛠️ 技术栈

### 后端
- **Go 1.23**: 高性能后端语言
- **Wails v2**: 跨平台桌面应用框架
- **GoProxy**: HTTP/HTTPS 代理服务器
- **ChromeDP**: 浏览器自动化（用于独立的网页抓取功能，非代理抓包核心）
- **Playwright-Go**: 浏览器自动化（用于独立的网页抓取功能，非代理抓包核心）

### 前端
- **Vue 3**: 渐进式 JavaScript 框架
- **TypeScript**: 类型安全的 JavaScript
- **Vite**: 快速的前端构建工具

### 核心库
- `github.com/elazarl/goproxy` - HTTP 代理
- `github.com/wailsapp/wails/v2` - 桌面应用框架
- `github.com/chromedp/chromedp` - Chrome DevTools Protocol
- `github.com/playwright-community/playwright-go` - 浏览器自动化

---

## 📸 截图

### 主界面
![主界面](docs/img.png)

### 请求详情
![请求详情](docs/img_1.png)

### 证书安装
![证书安装](docs/img_2.png)

---

## 🔧 开发指南

### 项目结构

```
tools-netcaptor/
├── frontend/              # Vue3 前端
│   ├── src/
│   │   ├── components/   # Vue 组件
│   │   └── App.vue       # 主应用
│   └── wailsjs/          # Wails 生成的绑定
├── *.go                  # Go 后端代码
├── wails.json            # Wails 配置
└── README.md             # 项目文档
```

### 构建生产版本

```bash
# 构建所有平台
wails build

# 构建特定平台
wails build -platform linux/amd64
wails build -platform windows/amd64
wails build -platform darwin/amd64
```

### 调试

```bash
# 启动开发模式（热重载）
wails dev -tags webkit2_41

# 查看日志
tail -f app.log
```

### 常见问题

#### Playwright 驱动问题

如果出现 "please install the driver" 错误：

```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install
```

#### 浏览器不弹出

检查浏览器安装：

```bash
# 查看已安装的浏览器
ls ~/.cache/ms-playwright/

# 重新安装 Chromium
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4702.0 install chromium
```

#### 端口占用

检查并释放端口：

```bash
# 检查端口占用
lsof -i :8888
lsof -i :5173

# 杀死占用进程
kill -9 <PID>
```

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 开发流程
1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📝 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

---

## ⚠️ 免责声明

本工具仅供学习和合法用途使用。请遵守相关法律法规和网站使用条款。使用本工具进行的任何活动，其法律责任由使用者自行承担。

---

## 📮 联系方式

- **作者**: fzxs
- **邮箱**: fzxs88@yeah.net
- **问题反馈**: [GitHub Issues](../../issues)

---

<div align="center">

**如果这个项目对你有帮助，请给一个 ⭐️**

Made with ❤️ by fzxs

</div>
