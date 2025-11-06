# 代理连接测试功能更新

## 更新内容

将代理连接测试功能从固定URL测试改为可自定义URL测试，提供更灵活的测试体验。

## 新功能特性

### 1. 自定义测试URL
- 添加了URL输入框，用户可以输入任意要测试的网址
- 默认测试URL设置为 `https://www.google.com`
- 支持HTTP和HTTPS协议的URL测试

### 2. 改进的用户界面
- URL输入框位于测试按钮上方
- 清晰的布局，便于用户操作
- 实时的测试结果反馈

### 3. 更详细的测试结果
- 显示具体访问的URL和HTTP状态码
- 更准确的错误信息反馈

## 使用方法

1. **打开代理设置**
   - 进入设置面板 → 网络代理标签页

2. **配置代理服务器**
   - 启用代理并填写服务器信息

3. **自定义测试URL**
   - 在"连接测试"部分的输入框中输入要测试的URL
   - 默认为 `https://www.google.com`，可以修改为任意网址

4. **执行测试**
   - 点击"测试连接"按钮
   - 查看测试结果

## 测试示例

### 常用测试URL
```
https://www.google.com          # 测试Google访问
https://www.baidu.com           # 测试百度访问
http://httpbin.org/ip           # 测试IP显示服务
https://api.openai.com          # 测试OpenAI API访问
https://www.github.com          # 测试GitHub访问
```

### 测试结果示例
- **成功**: "代理连接成功，访问 https://www.google.com 返回 HTTP 200"
- **失败**: "访问 https://www.google.com 失败: HTTP 403"
- **错误**: "连接失败: dial tcp: lookup proxy.example.com: no such host"

## 技术实现

### 后端更新
- 新增 `TestConnectionWithURL(testURL string)` 方法
- 保留原有 `TestConnection()` 方法作为默认测试
- 支持任意HTTP/HTTPS URL的代理测试

### 前端更新
- 添加 `testUrl` 响应式变量，默认值为 `https://www.google.com`
- 更新UI布局，URL输入框和测试按钮垂直排列
- 调用新的 `TestProxyConnectionWithURL` API方法

## 优势

1. **灵活性**: 用户可以测试任意网站的代理访问
2. **实用性**: 可以直接测试目标服务（如AI API）的代理连通性
3. **调试友好**: 详细的错误信息帮助排查问题
4. **用户体验**: 直观的界面和清晰的反馈

## 向后兼容

- 保留了原有的 `TestProxyConnection()` 方法
- 前端默认URL设置确保无需额外配置即可使用
- 所有现有功能保持不变
