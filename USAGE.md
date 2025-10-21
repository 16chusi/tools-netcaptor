# 🚀 自动下载工具使用指南

## 功能特性

### 🛡️ 反检测技术
- **移除 webdriver 标识**：清除自动化工具特征
- **伪造浏览器指纹**：模拟真实用户环境
- **禁用开发者工具检测**：绕过 F12 禁用限制
- **随机延迟**：模拟人类操作行为

### 🔍 智能链接提取
- **CSS 选择器支持**：精确定位下载链接
- **默认规则**：自动识别常见下载链接格式
- **相对链接处理**：自动转换为绝对路径

### 📄 自动翻页
- **智能翻页检测**：自动识别下一页按钮
- **页数限制**：防止无限循环
- **状态监控**：实时显示抓取进度

## 使用步骤

### 1. 基本配置
```
网站URL: https://example.com/downloads
下载链接选择器: a[href*="download"] (可选)
下一页选择器: .next (可选)
最大页数: 10
下载路径: ./downloads
```

### 2. 高级选择器示例

#### 下载链接选择器
```css
/* PDF 文件 */
a[href$=".pdf"]

/* 包含 download 的链接 */
a[href*="download"]

/* 特定类名的链接 */
.download-link

/* 组合选择器 */
a[href$=".pdf"], a[href$=".doc"], a[href$=".zip"]
```

#### 下一页选择器
```css
/* 文本包含"下一页" */
a:has-text("下一页")

/* 类名为 next */
.next

/* 分页器中的下一页 */
.pagination .next

/* 箭头符号 */
a:has-text("›")
```

### 3. 操作流程
1. **输入网站URL**：要抓取的目标网站
2. **配置选择器**：根据网站结构调整（可选）
3. **开始抓取**：点击"开始抓取"按钮
4. **查看结果**：在右侧面板查看找到的链接
5. **开始下载**：点击"下载"按钮批量下载文件

## 常见网站配置

### 文档下载站点
```
链接选择器: a[href*="download"], .download-btn
下一页选择器: .pagination-next, a:has-text("下一页")
```

### 资源分享站点
```
链接选择器: a[href$=".zip"], a[href$=".rar"]
下一页选择器: .next-page, a:has-text("Next")
```

### 学术论文站点
```
链接选择器: a[href$=".pdf"], .pdf-download
下一页选择器: .page-next, a[title*="next"]
```

## 故障排除

### 抓取失败
- 检查网站URL是否正确
- 确认网站可以正常访问
- 调整选择器匹配网站结构

### 找不到链接
- 使用浏览器开发者工具检查页面结构
- 调整下载链接选择器
- 确认链接是否为动态加载

### 翻页失败
- 检查下一页按钮的选择器
- 确认是否存在下一页
- 调整最大页数限制

### 下载失败
- 检查下载路径权限
- 确认网络连接正常
- 查看错误日志信息

## 技术原理

### 反检测机制
```javascript
// 移除 webdriver 标识
Object.defineProperty(navigator, 'webdriver', {
    get: () => undefined,
});

// 伪造插件信息
Object.defineProperty(navigator, 'plugins', {
    get: () => [1, 2, 3, 4, 5],
});

// 阻止开发者工具检测
document.addEventListener('keydown', e => {
    if (e.key === 'F12') e.stopPropagation();
}, true);
```

### 浏览器启动参数
```go
Args: []string{
    "--no-sandbox",
    "--disable-blink-features=AutomationControlled",
    "--disable-web-security",
    "--disable-dev-shm-usage",
}
```

## 注意事项

⚠️ **使用须知**
- 请遵守网站的 robots.txt 和使用条款
- 避免对服务器造成过大压力
- 尊重版权和知识产权
- 仅用于合法用途

🔒 **安全提醒**
- 不要在不信任的网站上使用
- 定期更新工具版本
- 注意保护个人隐私信息

📈 **性能优化**
- 合理设置最大页数
- 选择合适的下载路径
- 定期清理下载文件