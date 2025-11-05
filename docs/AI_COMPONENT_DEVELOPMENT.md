# AI组件开发指南

## 概述

本文档详细说明如何在NetCaptor中开发新的AI组件，包括前端配置界面、后端执行逻辑和类型定义。

## 开发流程

### 1. 定义组件类型

在 `frontend/src/types/workflow.ts` 中添加新的节点类型：

```typescript
export type NodeType = 
  | 'existing_types...'
  | 'ai_new_component'  // 新增AI组件类型
```

### 2. 配置组件信息

在 `frontend/src/components/workflow/nodeConfigs.ts` 中添加组件定义：

```typescript
{
  name: 'ai_analysis',  // 或 ai_control
  title: '🤖 AI数据分析',
  nodes: [
    { 
      type: 'ai_new_component', 
      label: 'AI新功能', 
      icon: '🎯', 
      color: 'rgba(114, 46, 209, 0.15)', 
      description: '组件功能描述' 
    }
  ]
}
```

### 3. 创建属性面板

在 `frontend/src/components/workflow/PropertyPanel.vue` 中添加配置界面：

```vue
<!-- AI新组件 -->
<template v-if="nodeType === 'ai_new_component'">
  <div class="form-item">
    <label>AI模型</label>
    <select v-model="formData.modelIndex">
      <option v-for="(model, index) in aiModels" :key="index" :value="index">
        {{ model.name || `${model.provider} 模型 ${index + 1}` }}
      </option>
    </select>
  </div>
  <div class="form-item">
    <label>功能配置</label>
    <textarea v-model="formData.prompt" placeholder="描述AI任务..." rows="3"></textarea>
  </div>
  <div class="form-item">
    <label>保存变量</label>
    <input v-model="formData.saveToVariable" placeholder="resultData" />
  </div>
</template>
```

### 4. 实现后端逻辑

在 `workflow_actions_ai.go` 中添加执行函数：

```go
// executeAINewComponent 执行AI新组件
func (we *WorkflowExecutor) executeAINewComponent(step ExecutionStep) (ExecutionResult, error) {
	modelIndex, _ := strconv.Atoi(step.Params["modelIndex"].(string))
	prompt := step.Params["prompt"].(string)
	saveToVariable := step.Params["saveToVariable"].(string)

	// 获取页面DOM（如果需要）
	domData, err := we.getCurrentPageDOM()
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("获取页面DOM失败: %v", err)
	}

	// 构建AI提示词
	systemPrompt := "你是一个专业的助手，请根据用户要求处理任务。"
	fullPrompt := fmt.Sprintf("页面内容:\n%s\n\n任务要求:\n%s", domData, prompt)

	// 调用AI
	result, err := we.aiService.CallAI(modelIndex, fullPrompt, systemPrompt)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
	}

	// 保存结果
	if saveToVariable != "" {
		we.setVariable(saveToVariable, result)
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("AI新组件执行完成，结果保存到变量: %s", saveToVariable),
		Data:    map[string]interface{}{"result": result},
	}, nil
}
```

### 5. 注册执行器

在 `workflow_actions.go` 的switch语句中添加case：

```go
switch step.Action {
	// ... 其他case
	case "ai_new_component":
		return we.executeAINewComponent(step)
	default:
		return ExecutionResult{Success: false}, fmt.Errorf("未知的操作类型: %s", step.Action)
}
```

### 6. 添加默认值

在 `PropertyPanel.vue` 的AI组件默认值设置中添加：

```typescript
// AI组件默认值
if (props.nodeType?.startsWith('ai_')) {
  // ... 现有默认值
  if (props.nodeType === 'ai_new_component') {
    if (!formData.value.customParam) {
      formData.value.customParam = 'defaultValue'
    }
  }
}
```

## 组件类型指南

### 数据分析类组件

**特点**:
- 主要处理页面内容和数据
- 通常需要获取DOM
- 结果保存到变量供后续使用

**模板**:
```go
func (we *WorkflowExecutor) executeAIAnalysisComponent(step ExecutionStep) (ExecutionResult, error) {
	// 1. 获取参数
	modelIndex, _ := strconv.Atoi(step.Params["modelIndex"].(string))
	prompt := step.Params["prompt"].(string)
	
	// 2. 获取页面数据
	domData, err := we.getCurrentPageDOM()
	if err != nil {
		return ExecutionResult{Success: false}, err
	}
	
	// 3. 调用AI分析
	result, err := we.aiService.CallAI(modelIndex, prompt, systemPrompt)
	if err != nil {
		return ExecutionResult{Success: false}, err
	}
	
	// 4. 保存结果
	we.setVariable(saveToVariable, result)
	
	return ExecutionResult{Success: true, Message: "分析完成"}, nil
}
```

### 浏览器控制类组件

**特点**:
- 主要执行浏览器操作
- AI生成操作指令
- 通过WebSocket与浏览器通信

**模板**:
```go
func (we *WorkflowExecutor) executeAIControlComponent(step ExecutionStep) (ExecutionResult, error) {
	// 1. 获取参数
	modelIndex, _ := strconv.Atoi(step.Params["modelIndex"].(string))
	prompt := step.Params["prompt"].(string)
	
	// 2. 获取页面DOM
	domData, err := we.getCurrentPageDOM()
	if err != nil {
		return ExecutionResult{Success: false}, err
	}
	
	// 3. AI生成操作指令
	systemPrompt := "分析HTML，返回CSS选择器或操作指令"
	instruction, err := we.aiService.CallAI(modelIndex, prompt, systemPrompt)
	if err != nil {
		return ExecutionResult{Success: false}, err
	}
	
	// 4. 执行浏览器操作
	actionStep := ExecutionStep{
		Action: "click", // 或其他基础操作
		Params: map[string]interface{}{
			"selector": instruction,
			"selectorType": "css",
		},
	}
	
	return we.executeClick(actionStep)
}
```

## 最佳实践

### 1. 错误处理

```go
// 参数验证
if modelIndex >= len(we.aiService.models) {
	return ExecutionResult{Success: false}, fmt.Errorf("无效的模型索引: %d", modelIndex)
}

// AI调用错误处理
result, err := we.aiService.CallAI(modelIndex, prompt, systemPrompt)
if err != nil {
	return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
}

// 结果验证
if strings.TrimSpace(result) == "" {
	return ExecutionResult{Success: false}, fmt.Errorf("AI返回空结果")
}
```

### 2. 提示词设计

```go
// 明确的系统提示词
systemPrompt := `你是一个专业的网页分析专家。
请分析提供的HTML内容，根据用户要求执行任务。
输出格式: JSON
要求: 只返回结果，不要包含解释文字。`

// 结构化的用户提示词
fullPrompt := fmt.Sprintf(`页面HTML内容:
%s

任务要求:
%s

输出示例:
{"data": [...], "count": 10}`, domData, prompt)
```

### 3. 性能优化

```go
// DOM内容压缩
func compressDOM(html string) string {
	// 移除注释、多余空白等
	// 只保留必要的内容
	return cleanedHTML
}

// 结果缓存
var domCache = make(map[string]string)

func (we *WorkflowExecutor) getCachedDOM() (string, error) {
	url := we.getCurrentURL()
	if cached, exists := domCache[url]; exists {
		return cached, nil
	}
	
	dom, err := we.getCurrentPageDOM()
	if err == nil {
		domCache[url] = dom
	}
	return dom, err
}
```

### 4. 调试支持

```go
import "log"

func (we *WorkflowExecutor) executeAIComponent(step ExecutionStep) (ExecutionResult, error) {
	log.Printf("[AI组件] 开始执行: %s", step.Action)
	log.Printf("[AI组件] 参数: %+v", step.Params)
	
	// 执行逻辑...
	
	log.Printf("[AI组件] AI提示词: %s", fullPrompt)
	log.Printf("[AI组件] AI响应: %s", result)
	log.Printf("[AI组件] 执行完成")
	
	return ExecutionResult{Success: true}, nil
}
```

## 测试指南

### 1. 单元测试

```go
func TestAINewComponent(t *testing.T) {
	// 创建测试执行器
	executor := &WorkflowExecutor{
		aiService: &MockAIService{},
		variables: make(map[string]interface{}),
	}
	
	// 准备测试步骤
	step := ExecutionStep{
		Action: "ai_new_component",
		Params: map[string]interface{}{
			"modelIndex": "0",
			"prompt": "测试提示词",
			"saveToVariable": "testResult",
		},
	}
	
	// 执行测试
	result, err := executor.executeAINewComponent(step)
	
	// 验证结果
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, executor.variables["testResult"])
}
```

### 2. 集成测试

```go
func TestAIComponentIntegration(t *testing.T) {
	// 启动完整的测试环境
	app := setupTestApp()
	defer app.cleanup()
	
	// 创建测试工作流
	task := WorkflowTask{
		Nodes: []WorkflowNode{
			{Type: "ai_new_component", Data: testParams},
		},
	}
	
	// 执行工作流
	err := app.workflowExecutor.Execute(task)
	assert.NoError(t, err)
}
```

## 常见问题

### Q: 如何处理AI返回的非结构化数据？

A: 使用正则表达式或字符串处理提取有用信息：

```go
// 提取JSON部分
jsonStart := strings.Index(result, "{")
jsonEnd := strings.LastIndex(result, "}") + 1
if jsonStart >= 0 && jsonEnd > jsonStart {
	jsonStr := result[jsonStart:jsonEnd]
	// 解析JSON
}
```

### Q: 如何优化AI调用的性能？

A: 
1. 压缩DOM内容
2. 使用更快的模型
3. 实现结果缓存
4. 批量处理相似任务

### Q: 如何处理AI调用超时？

A: 在AI服务中设置合理的超时时间：

```go
client := &http.Client{Timeout: 30 * time.Second}
```

### Q: 如何调试AI组件？

A: 
1. 查看工作流执行日志
2. 输出AI的提示词和响应
3. 验证变量状态
4. 使用浏览器开发者工具检查DOM

## 扩展示例

### 智能表格提取组件

```typescript
// 前端配置
{
  type: 'ai_table_extract',
  label: 'AI表格提取',
  icon: '📊',
  color: 'rgba(114, 46, 209, 0.15)'
}
```

```go
// 后端实现
func (we *WorkflowExecutor) executeAITableExtract(step ExecutionStep) (ExecutionResult, error) {
	systemPrompt := `分析HTML中的表格，提取所有行和列的数据。
返回JSON格式: {"headers": [...], "rows": [[...], [...]]}`
	
	// 实现逻辑...
}
```

这样就完成了一个新AI组件的开发！
