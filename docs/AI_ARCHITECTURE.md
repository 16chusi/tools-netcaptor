# NetCaptor AI架构设计

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        NetCaptor AI 系统                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌──────────────────┐    ┌─────────────┐ │
│  │   前端 Vue3     │    │    后端 Go       │    │  AI 服务商  │ │
│  │                 │    │                  │    │             │ │
│  │ ┌─────────────┐ │    │ ┌──────────────┐ │    │ ┌─────────┐ │ │
│  │ │AI模型管理   │◄┼────┼►│AIService     │◄┼────┼►│OpenAI   │ │ │
│  │ └─────────────┘ │    │ └──────────────┘ │    │ └─────────┘ │ │
│  │                 │    │                  │    │             │ │
│  │ ┌─────────────┐ │    │ ┌──────────────┐ │    │ ┌─────────┐ │ │
│  │ │AI组件面板   │◄┼────┼►│WorkflowExecutor│    │ │Claude   │ │ │
│  │ └─────────────┘ │    │ └──────────────┘ │    │ └─────────┘ │ │
│  │                 │    │        │         │    │             │ │
│  │ ┌─────────────┐ │    │        ▼         │    │ ┌─────────┐ │ │
│  │ │属性配置     │ │    │ ┌──────────────┐ │    │ │Azure    │ │ │
│  │ └─────────────┘ │    │ │AI Actions    │ │    │ └─────────┘ │ │
│  └─────────────────┘    │ └──────────────┘ │    │             │ │
│                         │        │         │    │ ┌─────────┐ │ │
│  ┌─────────────────┐    │        ▼         │    │ │Custom   │ │ │
│  │  Chrome扩展     │    │ ┌──────────────┐ │    │ └─────────┘ │ │
│  │                 │    │ │Variable Mgr  │ │    └─────────────┘ │
│  │ ┌─────────────┐ │    │ └──────────────┘ │                    │
│  │ │DOM获取      │◄┼────┼─────────────────┐│                    │
│  │ └─────────────┘ │    │                 ││                    │
│  │                 │    │ ┌──────────────┐││                    │
│  │ ┌─────────────┐ │    │ │WebSocket     │││                    │
│  │ │页面操作     │◄┼────┼►│Server        │││                    │
│  │ └─────────────┘ │    │ └──────────────┘││                    │
│  └─────────────────┘    └──────────────────┘│                    │
│                                             │                    │
└─────────────────────────────────────────────┼────────────────────┘
                                              │
                                              ▼
                                    ┌──────────────────┐
                                    │   浏览器页面     │
                                    │                  │
                                    │ ┌──────────────┐ │
                                    │ │DOM Tree      │ │
                                    │ └──────────────┘ │
                                    │                  │
                                    │ ┌──────────────┐ │
                                    │ │User Interface│ │
                                    │ └──────────────┘ │
                                    └──────────────────┘
```

## 核心组件设计

### 1. AIService (AI服务层)

**职责**: 统一管理AI模型调用和配置

```go
type AIService struct {
    models []AIModel
}

type AIModel struct {
    Provider string `json:"provider"` // openai, anthropic, azure, custom
    Name     string `json:"name"`     // gpt-4, claude-3-sonnet
    APIKey   string `json:"apiKey"`   // API密钥
    BaseURL  string `json:"baseUrl"`  // 自定义端点
}
```

**核心方法**:
- `CallAI(modelIndex, prompt, systemPrompt) -> (result, error)`
- `TestModel(model) -> error`
- `UpdateModels(models) -> void`

**设计特点**:
- 支持多供应商统一接口
- 自动重试和错误处理
- 请求/响应日志记录
- 超时控制和并发限制

### 2. WorkflowExecutor (工作流执行器)

**职责**: 集成AI功能到工作流执行引擎

```go
type WorkflowExecutor struct {
    app        *NetworkApp
    wsServer   *WebSocketServer
    aiService  *AIService        // 新增AI服务
    variables  map[string]interface{}
    // ...
}
```

**AI集成点**:
- 执行AI组件节点
- 管理AI调用上下文
- 处理AI结果和变量
- 协调DOM获取和操作

### 3. AI Actions (AI动作执行器)

**职责**: 实现具体的AI组件逻辑

**组件分类**:

#### 数据分析类
```go
// AI数据提取
func (we *WorkflowExecutor) executeAIExtractData(step ExecutionStep) (ExecutionResult, error)

// AI内容分析  
func (we *WorkflowExecutor) executeAIAnalyzeContent(step ExecutionStep) (ExecutionResult, error)

// AI数据验证
func (we *WorkflowExecutor) executeAIValidateData(step ExecutionStep) (ExecutionResult, error)

// AI数据转换
func (we *WorkflowExecutor) executeAITransformData(step ExecutionStep) (ExecutionResult, error)
```

#### 浏览器控制类
```go
// AI智能点击
func (we *WorkflowExecutor) executeAISmartClick(step ExecutionStep) (ExecutionResult, error)

// AI表单填写
func (we *WorkflowExecutor) executeAIFormFill(step ExecutionStep) (ExecutionResult, error)

// AI智能导航
func (we *WorkflowExecutor) executeAINavigation(step ExecutionStep) (ExecutionResult, error)
```

### 4. DOM获取机制

**Chrome扩展端**:
```javascript
// content.js
case 'get_page_dom':
  const html = document.documentElement.outerHTML;
  const url = window.location.href;
  const title = document.title;
  
  sendResponse({ 
    success: true, 
    html: html,
    url: url,
    title: title
  });
```

**后端调用**:
```go
func (we *WorkflowExecutor) getCurrentPageDOM() (string, error) {
    message := WSMessage{
        Type: "get_page_dom",
        Data: map[string]interface{}{},
    }
    
    result, err := we.sendAndWait(message, 10*time.Second, "get_page_dom")
    if err != nil {
        return "", err
    }
    
    return result.Data["html"].(string), nil
}
```

### 5. 变量管理系统

**扩展支持AI结果**:
```go
// 设置AI结果变量
func (we *WorkflowExecutor) setVariable(name string, value interface{}) {
    we.variables[name] = value
    log.Printf("[Workflow] 设置变量 %s = %v", name, value)
}

// 获取变量用于AI输入
func (we *WorkflowExecutor) getVariable(name string) interface{} {
    if value, exists := we.variables[name]; exists {
        return value
    }
    return nil
}
```

**变量类型支持**:
- 字符串: AI文本结果
- JSON对象: 结构化数据
- 数组: 列表数据
- 数值: 统计结果

## 数据流设计

### 1. AI数据提取流程

```
用户配置 → DOM获取 → AI分析 → 结果保存 → 后续使用
    │         │        │        │         │
    ▼         ▼        ▼        ▼         ▼
PropertyPanel → Chrome → AIService → Variables → NextNode
              Extension
```

**详细步骤**:
1. 用户在PropertyPanel配置AI提取参数
2. WorkflowExecutor调用getCurrentPageDOM()
3. Chrome扩展返回完整DOM内容
4. AIService构建提示词并调用AI API
5. AI返回结构化数据结果
6. 结果保存到变量系统
7. 后续节点可引用提取的数据

### 2. AI浏览器控制流程

```
用户描述 → DOM分析 → AI理解 → 指令生成 → 浏览器执行
    │         │        │        │          │
    ▼         ▼        ▼        ▼          ▼
PropertyPanel → Chrome → AIService → ActionStep → Chrome
              Extension                        Extension
```

**详细步骤**:
1. 用户用自然语言描述操作需求
2. 获取当前页面DOM结构
3. AI分析DOM并理解用户意图
4. 生成具体的操作指令(CSS选择器等)
5. 转换为标准的浏览器操作步骤
6. 通过WebSocket执行浏览器操作

### 3. 配置数据流

```
前端配置 → 本地存储 → 后端同步 → AI服务配置
    │         │         │          │
    ▼         ▼         ▼          ▼
AIModelTab → localStorage → NetworkApp → AIService
```

## 安全架构

### 1. API密钥管理

**前端安全**:
- 使用password类型输入框
- 不在控制台输出敏感信息
- 本地存储加密(可选)

**后端安全**:
- 内存中临时存储
- 不写入日志文件
- 进程结束时清理内存

### 2. 数据传输安全

**DOM数据**:
- 用户确认后才发送给AI
- 支持数据脱敏选项
- 可配置本地AI模型

**AI响应**:
- 验证返回数据格式
- 过滤恶意代码
- 限制执行权限

### 3. 操作权限控制

**浏览器操作**:
- 白名单机制
- 操作确认
- 危险操作拦截

## 性能优化

### 1. AI调用优化

**请求优化**:
```go
// 请求池管理
type AIRequestPool struct {
    semaphore chan struct{}
    timeout   time.Duration
}

// 并发控制
func (pool *AIRequestPool) Execute(fn func() error) error {
    select {
    case pool.semaphore <- struct{}{}:
        defer func() { <-pool.semaphore }()
        return fn()
    case <-time.After(pool.timeout):
        return errors.New("请求超时")
    }
}
```

**缓存策略**:
```go
// DOM分析结果缓存
type DOMAnalysisCache struct {
    cache map[string]CacheEntry
    mutex sync.RWMutex
}

type CacheEntry struct {
    Result    string
    Timestamp time.Time
    TTL       time.Duration
}
```

### 2. DOM数据优化

**内容压缩**:
```go
func compressDOM(html string) string {
    // 移除注释
    html = regexp.MustCompile(`<!--.*?-->`).ReplaceAllString(html, "")
    
    // 移除多余空白
    html = regexp.MustCompile(`\s+`).ReplaceAllString(html, " ")
    
    // 移除不必要的属性
    html = regexp.MustCompile(`\s(style|class)="[^"]*"`).ReplaceAllString(html, "")
    
    return html
}
```

**增量更新**:
```go
func (we *WorkflowExecutor) getDOMDiff() (string, error) {
    current := we.getCurrentPageDOM()
    if we.lastDOM == "" {
        we.lastDOM = current
        return current, nil
    }
    
    // 计算差异
    diff := calculateDOMDiff(we.lastDOM, current)
    we.lastDOM = current
    
    return diff, nil
}
```

### 3. 内存管理

**变量生命周期**:
```go
type VariableManager struct {
    variables map[string]VariableEntry
    mutex     sync.RWMutex
}

type VariableEntry struct {
    Value     interface{}
    CreatedAt time.Time
    TTL       time.Duration
    Size      int64
}

func (vm *VariableManager) Cleanup() {
    vm.mutex.Lock()
    defer vm.mutex.Unlock()
    
    now := time.Now()
    for name, entry := range vm.variables {
        if now.Sub(entry.CreatedAt) > entry.TTL {
            delete(vm.variables, name)
        }
    }
}
```

## 扩展性设计

### 1. 插件化AI供应商

```go
type AIProvider interface {
    Name() string
    Call(request AIRequest) (AIResponse, error)
    Test(config AIModel) error
}

type OpenAIProvider struct{}
type AnthropicProvider struct{}
type CustomProvider struct{}

// 注册机制
var providers = map[string]AIProvider{
    "openai":    &OpenAIProvider{},
    "anthropic": &AnthropicProvider{},
    "custom":    &CustomProvider{},
}
```

### 2. 可配置AI组件

```go
type AIComponentConfig struct {
    Type        string                 `json:"type"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  []ParameterDefinition  `json:"parameters"`
    Template    string                 `json:"template"`
}

type ParameterDefinition struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"`
    Required    bool        `json:"required"`
    Default     interface{} `json:"default"`
    Description string      `json:"description"`
}
```

### 3. 自定义AI模型

```go
type CustomAIModel struct {
    Endpoint    string            `json:"endpoint"`
    Headers     map[string]string `json:"headers"`
    RequestFormat string          `json:"requestFormat"`
    ResponseFormat string         `json:"responseFormat"`
    Transformer func(string) string `json:"-"`
}
```

## 监控和调试

### 1. 日志系统

```go
type AILogger struct {
    logger *log.Logger
    level  LogLevel
}

func (l *AILogger) LogAICall(modelIndex int, prompt string, response string, duration time.Duration) {
    l.logger.Printf("[AI调用] 模型:%d 耗时:%v 提示词长度:%d 响应长度:%d", 
        modelIndex, duration, len(prompt), len(response))
}
```

### 2. 性能指标

```go
type AIMetrics struct {
    TotalCalls    int64         `json:"totalCalls"`
    SuccessCalls  int64         `json:"successCalls"`
    FailedCalls   int64         `json:"failedCalls"`
    AverageTime   time.Duration `json:"averageTime"`
    TotalTokens   int64         `json:"totalTokens"`
    TotalCost     float64       `json:"totalCost"`
}
```

### 3. 调试工具

```go
type AIDebugger struct {
    enabled bool
    traces  []AITrace
}

type AITrace struct {
    Timestamp   time.Time `json:"timestamp"`
    ModelIndex  int       `json:"modelIndex"`
    Prompt      string    `json:"prompt"`
    Response    string    `json:"response"`
    Duration    time.Duration `json:"duration"`
    Error       string    `json:"error,omitempty"`
}
```

## 部署架构

### 1. 单机部署

```
┌─────────────────────────────────┐
│         NetCaptor 应用          │
├─────────────────────────────────┤
│  ┌─────────────┐ ┌────────────┐ │
│  │  前端UI     │ │  Go后端    │ │
│  └─────────────┘ └────────────┘ │
│  ┌─────────────┐ ┌────────────┐ │
│  │ Chrome扩展  │ │ AI服务     │ │
│  └─────────────┘ └────────────┘ │
└─────────────────────────────────┘
              │
              ▼
    ┌─────────────────┐
    │   外部AI API    │
    └─────────────────┘
```

### 2. 分布式部署

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  前端应用   │    │  API网关    │    │  AI服务集群 │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │
              ┌─────────────────┐
              │   配置中心      │
              └─────────────────┘
```

这个架构设计为NetCaptor的AI功能提供了完整的技术框架，支持灵活扩展和高性能运行。
