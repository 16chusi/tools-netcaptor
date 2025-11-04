# 等待模式功能技术文档

## 功能概述

为解决 NetCaptor 工作流中数据提取时机问题，特别是异步请求导致的陈旧数据问题，在 Chrome 扩展中实现了全局等待模式设置功能。

## 问题背景

### 原始问题
- 在多层嵌套循环执行中，数据提取组件返回上一次循环的陈旧数据
- 页面异步请求未完成时就开始数据提取
- 固定延时无法适应不同页面的加载特性

### 根本原因
- 页面 DOM 更新与异步请求完成存在时间差
- 浏览器插件在页面未完全稳定时就执行数据提取
- 缺乏灵活的等待策略来适应不同场景

## 解决方案

### 架构设计
```
Popup 面板 → Chrome Storage → Content Script → 等待策略 → 数据提取
```

### 三种等待模式

#### 1. 标准模式 (DOM稳定)
**适用场景**: 大多数静态页面或轻量级动态内容

**工作原理**:
```javascript
// 监控页面内容变化
let lastContent = document.body.innerHTML;
let stableCount = 0;

const checkStability = () => {
  const currentContent = document.body.innerHTML;
  if (currentContent === lastContent) {
    stableCount++;
    if (stableCount >= 3) { // 连续3次相同
      resolve(); // 开始执行
    }
  } else {
    stableCount = 0;
    lastContent = currentContent;
  }
  setTimeout(checkStability, 500);
};
```

**特点**:
- 检查间隔: 500ms
- 稳定条件: 连续 3 次内容相同
- 初始等待: 1 秒

#### 2. 完整等待 (异步请求)
**适用场景**: 有大量异步加载内容的页面

**工作原理**:
```javascript
// 拦截 fetch 请求
const originalFetch = window.fetch;
window.fetch = function(...args) {
  pendingRequests++;
  return originalFetch.apply(this, args).finally(() => {
    pendingRequests--;
    lastActivity = Date.now();
  });
};

// 拦截 XMLHttpRequest
XMLHttpRequest.prototype.send = function(...args) {
  if (this._netcaptor_tracked) {
    pendingRequests++;
    const onComplete = () => {
      pendingRequests--;
      lastActivity = Date.now();
    };
    this.addEventListener('load', onComplete);
    this.addEventListener('error', onComplete);
  }
  return originalSend.apply(this, args);
};
```

**特点**:
- 监控所有网络请求
- 等待条件: 无待完成请求 + 距离最后活动 > 1秒
- 自动恢复原始方法避免内存泄漏

#### 3. 无需等待 (立即执行)
**适用场景**: 简单静态页面或需要快速执行的场景

**工作原理**:
```javascript
if (waitMode === 'none') {
  resolve(); // 立即执行
}
```

## 实现细节

### 1. 前端设置界面
**文件**: `chrome-extension/popup.html`

```html
<select id="waitMode">
  <option value="standard">🔄 标准模式 (DOM稳定)</option>
  <option value="complete">⏳ 完整等待 (异步请求)</option>
  <option value="none">⚡ 无需等待 (立即执行)</option>
</select>
```

### 2. 设置持久化
**文件**: `chrome-extension/popup.js`

```javascript
// 保存设置
document.getElementById('waitMode').addEventListener('change', async (e) => {
  const mode = e.target.value;
  await chrome.storage.local.set({ waitMode: mode });
});

// 加载设置
const stored = await chrome.storage.local.get(['waitMode']);
if (stored.waitMode) {
  document.getElementById('waitMode').value = stored.waitMode;
}
```

### 3. 运行时应用
**文件**: `chrome-extension/content.js`

```javascript
case 'extract_data':
  // 获取全局设置
  chrome.storage.local.get(['waitMode'], (result) => {
    const waitMode = result.waitMode || 'standard';
    
    // 根据模式执行不同等待策略
    const waitForPageComplete = () => {
      return new Promise((resolve) => {
        if (waitMode === 'none') {
          resolve();
        } else if (waitMode === 'complete') {
          // 完整等待逻辑
        } else {
          // 标准等待逻辑
        }
      });
    };
    
    waitForPageComplete().then(() => {
      // 执行数据提取
    });
  });
```

### 4. 后端简化
**文件**: `tools-netcaptor/workflow_actions.go`

移除了组件级别的 `waitForComplete` 参数，统一使用全局设置：

```go
// 移除前
msg := WSMessage{
    Type: "extract_data",
    Data: map[string]interface{}{
        "selector":        selector,
        "waitForComplete": waitForComplete, // 已移除
        "selectorType":    selectorType,
    },
}

// 移除后
msg := WSMessage{
    Type: "extract_data",
    Data: map[string]interface{}{
        "selector":     selector,
        "selectorType": selectorType,
    },
}
```

## 性能考虑

### 内存管理
- 完整等待模式会临时替换原生方法
- 执行完成后自动恢复原始方法
- 避免长期占用内存

### 执行效率
- 标准模式: 轻量级，适合大多数场景
- 完整等待: 重量级，仅在必要时使用
- 无需等待: 最高效，适合简单场景

### 兼容性
- 支持所有现代浏览器
- 兼容 Manifest V3 规范
- 不影响页面原有功能

## 使用建议

### 场景选择
1. **电商网站**: 完整等待模式（商品信息异步加载）
2. **新闻网站**: 标准模式（内容相对静态）
3. **简单表单**: 无需等待模式（快速填写）

### 调试技巧
1. 开启浏览器控制台查看 `[NetCaptor]` 日志
2. 观察网络面板的请求完成情况
3. 根据实际效果调整等待模式

### 故障排除
- 数据仍然陈旧 → 尝试完整等待模式
- 执行过慢 → 降级到标准模式或无需等待
- 请求监控异常 → 刷新页面重新加载扩展

## 版本记录

### v1.0 (2024-11-04)
- 实现三种等待模式
- 添加全局设置面板
- 移除组件级别配置
- 优化用户体验（移除 alert 提示）

## 相关文件

- `chrome-extension/popup.html` - 设置界面
- `chrome-extension/popup.js` - 设置逻辑
- `chrome-extension/content.js` - 等待策略实现
- `tools-netcaptor/workflow_actions.go` - 后端简化
