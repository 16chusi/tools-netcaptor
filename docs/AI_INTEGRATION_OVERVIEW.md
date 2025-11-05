# NetCaptor AI集成概览

## 文档索引

本文档提供NetCaptor AI功能集成的完整技术概览和实现细节。

### 相关文档
- [AI功能用户指南](./AI_FEATURES.md) - 用户使用手册
- [AI组件开发指南](./AI_COMPONENT_DEVELOPMENT.md) - 开发者指南
- [AI架构设计](./AI_ARCHITECTURE.md) - 技术架构文档

## 实现概览

### 1. 核心组件

#### 前端组件
```
frontend/src/components/
├── tabs/AIModelTab.vue          # AI模型管理界面
├── workflow/nodeConfigs.ts      # AI节点配置
└── workflow/PropertyPanel.vue   # AI节点属性面板
```

#### 后端服务
```
tools-netcaptor/
├── ai_models.go                 # AI模型管理服务
├── workflow_actions_ai.go       # AI工作流动作实现
└── workflow_variables.go        # 变量管理（新增AI支持）
```

#### 浏览器扩展
```
chrome-extension/
└── content.js                   # 新增get_page_dom消息处理
```

### 2. AI组件类型

#### 数据分析类
- `ai_extract_data` - 智能数据提取
- `ai_analyze_content` - 内容理解分析
- `ai_validate_data` - 数据验证
- `ai_transform_data` - 数据转换

#### 浏览器控制类
- `ai_smart_click` - 智能点击
- `ai_form_fill` - 智能表单填写
- `ai_navigation` - 智能导航

### 3. 技术架构

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Vue3 前端     │    │    Go 后端       │    │   AI 服务商     │
│                 │    │                  │    │                 │
│ ┌─────────────┐ │    │ ┌──────────────┐ │    │ ┌─────────────┐ │
│ │AI模型管理   │◄┼────┼►│AIService     │◄┼────┼►│OpenAI/Claude│ │
│ └─────────────┘ │    │ └──────────────┘ │    │ └─────────────┘ │
│                 │    │                  │    │                 │
│ ┌─────────────┐ │    │ ┌──────────────┐ │    │                 │
│ │AI组件面板   │◄┼────┼►│WorkflowExecutor│    │                 │
│ └─────────────┘ │    │ └──────────────┘ │    │                 │
│                 │    │        │         │    │                 │
│ ┌─────────────┐ │    │        ▼         │    │                 │
│ │属性配置     │ │    │ ┌──────────────┐ │    │                 │
│ └─────────────┘ │    │ │AI Actions    │ │    │                 │
└─────────────────┘    │ └──────────────┘ │    └─────────────────┘
                       └──────────────────┘
                                │
                       ┌──────────────────┐
                       │  Chrome扩展      │
                       │                  │
                       │ ┌──────────────┐ │
                       │ │DOM获取       │ │
                       │ └──────────────┘ │
                       └──────────────────┘
```

### 4. 数据流

#### AI数据提取流程
```
1. 用户配置AI节点 → 2. 获取页面DOM → 3. 调用AI分析 → 4. 保存结果到变量
   PropertyPanel      Chrome Extension     AIService        VariableManager
```

#### AI浏览器控制流程
```
1. 用户描述操作 → 2. AI分析DOM → 3. 生成选择器 → 4. 执行浏览器操作
   PropertyPanel     AIService        ActionExecutor     Chrome Extension
```

## 关键实现细节

### 1. AI模型管理

**配置存储**: localStorage
```typescript
interface AIModel {
  provider: string    // openai, anthropic, azure, custom
  name: string       // 模型名称
  apiKey: string     // API密钥
  baseUrl: string    // 自定义端点
}
```

**支持的供应商**:
- OpenAI: GPT-3.5/GPT-4系列
- Anthropic: Claude系列
- Azure OpenAI: 企业版
- 自定义: 兼容OpenAI API格式

### 2. 工作流集成

**节点类型扩展**:
```typescript
// frontend/src/types/workflow.ts
export type NodeType = 
  | 'ai_extract_data'
  | 'ai_analyze_content'
  | 'ai_validate_data'
  | 'ai_transform_data'
  | 'ai_smart_click'
  | 'ai_form_fill'
  | 'ai_navigation'
```

**执行器集成**:
```go
// workflow_actions.go
switch step.Action {
case "ai_extract_data":
    return we.executeAIExtractData(step)
case "ai_smart_click":
    return we.executeAISmartClick(step)
// ...
}
```

### 3. DOM获取机制

**Chrome扩展支持**:
```javascript
// content.js
case 'get_page_dom':
  const html = document.documentElement.outerHTML;
  sendResponse({ success: true, html: html });
```

**后端调用**:
```go
// workflow_actions_ai.go
func (we *WorkflowExecutor) getCurrentPageDOM() (string, error) {
    message := WSMessage{Type: "get_page_dom"}
    result, err := we.sendAndWait(message, 10*time.Second, "get_page_dom")
    return result.Data["html"].(string), nil
}
```

### 4. 变量系统扩展

**AI结果存储**:
```go
// workflow_variables.go
func (we *WorkflowExecutor) setVariable(name string, value interface{}) {
    we.variables[name] = value
}

func (we *WorkflowExecutor) getVariable(name string) interface{} {
    return we.variables[name]
}
```

## 配置示例

### AI模型配置
```json
{
  "models": [
    {
      "provider": "openai",
      "name": "gpt-4",
      "apiKey": "sk-...",
      "baseUrl": ""
    }
  ],
  "defaultModel": 0,
  "defaultTemperature": 0.7,
  "defaultMaxTokens": 2000
}
```

### AI节点配置
```json
{
  "type": "ai_extract_data",
  "data": {
    "modelIndex": 0,
    "prompt": "提取所有商品信息，包括名称、价格、链接",
    "outputFormat": "json",
    "saveToVariable": "products"
  }
}
```

## 开发指南

### 添加新的AI组件

1. **前端**: 在`nodeConfigs.ts`中添加节点定义
2. **类型**: 在`workflow.ts`中添加NodeType
3. **属性面板**: 在`PropertyPanel.vue`中添加配置界面
4. **后端**: 在`workflow_actions_ai.go`中实现执行逻辑
5. **路由**: 在`workflow_actions.go`中添加case分支

### 扩展AI供应商

1. **前端**: 在`AIModelTab.vue`中添加选项
2. **后端**: 在`ai_models.go`中实现API调用逻辑
3. **测试**: 添加连接测试和快速测试支持

## 性能考虑

### 1. API调用优化
- 合理设置超时时间
- 实现请求重试机制
- 缓存相似页面的分析结果

### 2. DOM数据优化
- 压缩传输的HTML内容
- 只提取必要的DOM部分
- 实现增量更新机制

### 3. 内存管理
- 及时清理大型变量
- 限制AI响应内容大小
- 实现变量生命周期管理

## 安全考虑

### 1. API密钥保护
- 前端使用password类型输入框
- 后端内存中临时存储
- 不在日志中输出敏感信息

### 2. 数据隐私
- 用户确认页面内容发送给AI
- 支持本地AI模型部署
- 提供数据脱敏选项

### 3. 输入验证
- 验证AI返回的选择器安全性
- 限制执行的操作类型
- 实现操作权限控制

## 故障排除

### 常见问题
1. **AI调用失败**: 检查API密钥和网络连接
2. **DOM获取失败**: 确认Chrome扩展已连接
3. **变量未找到**: 检查变量名和作用域
4. **选择器无效**: 验证AI返回的选择器格式

### 调试工具
- 工作流执行日志
- AI调用请求/响应记录
- 变量状态监控
- 性能指标统计

## 版本历史

### v1.0.0 (2025-11-05)
- ✅ 基础AI模型管理
- ✅ 7个核心AI组件
- ✅ DOM获取机制
- ✅ 变量系统集成
- ✅ 多供应商支持

### 计划功能
- 🔄 AI模型本地部署支持
- 🔄 批量操作优化
- 🔄 可视化调试工具
- 🔄 AI训练数据收集
