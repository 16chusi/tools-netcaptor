package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// executeAIExtractData 执行AI数据提取
func (we *WorkflowExecutor) executeAIExtractData(step ExecutionStep) (ExecutionResult, error) {
	modelIndex, _ := strconv.Atoi(step.Params["modelIndex"].(string))
	prompt := step.Params["prompt"].(string)
	outputFormat := step.Params["outputFormat"].(string)
	saveToVariable := step.Params["saveToVariable"].(string)
	
	// 获取重试配置
	retryCount := 3
	retryDelay := 2
	if rc, ok := step.Params["retryCount"].(float64); ok {
		retryCount = int(rc)
	}
	if rd, ok := step.Params["retryDelay"].(float64); ok {
		retryDelay = int(rd)
	}

	// 获取当前页面DOM
	domData, err := we.getCurrentPageDOM()
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("获取页面DOM失败: %v", err)
	}

	// 构建系统提示词
	systemPrompt := fmt.Sprintf(`你是一个专业的网页数据提取助手。请分析提供的HTML内容，根据用户要求提取数据。
输出格式要求: %s
请确保输出格式正确，不要包含任何解释文字，只返回提取的数据。`, outputFormat)

	// 构建用户提示词
	fullPrompt := fmt.Sprintf("页面HTML内容:\n%s\n\n提取要求:\n%s", domData, prompt)

	// 调用AI（带重试）
	result, err := we.aiService.CallAIWithRetry(modelIndex, fullPrompt, systemPrompt, retryCount, retryDelay)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
	}

	// 保存结果到变量
	if saveToVariable != "" {
		we.setVariable(saveToVariable, result)
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("AI数据提取完成，结果保存到变量: %s", saveToVariable),
		Data:    map[string]interface{}{"result": result},
	}, nil
}

// executeAIAnalyzeContent 执行AI内容分析
func (we *WorkflowExecutor) executeAIAnalyzeContent(step ExecutionStep) (ExecutionResult, error) {
	modelIndex, _ := strconv.Atoi(step.Params["modelIndex"].(string))
	prompt := step.Params["prompt"].(string)
	saveToVariable := step.Params["saveToVariable"].(string)

	// 获取当前页面DOM
	domData, err := we.getCurrentPageDOM()
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("获取页面DOM失败: %v", err)
	}

	systemPrompt := "你是一个专业的网页内容分析师。请分析提供的HTML内容，理解页面结构和内容，提供详细的分析报告。"
	fullPrompt := fmt.Sprintf("页面HTML内容:\n%s\n\n分析要求:\n%s", domData, prompt)

	result, err := we.aiService.CallAI(modelIndex, fullPrompt, systemPrompt)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
	}

	if saveToVariable != "" {
		we.setVariable(saveToVariable, result)
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("AI内容分析完成，结果保存到变量: %s", saveToVariable),
		Data:    map[string]interface{}{"result": result},
	}, nil
}

// executeAIValidateData 执行AI数据验证
func (we *WorkflowExecutor) executeAIValidateData(step ExecutionStep) (ExecutionResult, error) {
	modelIndex, _ := strconv.Atoi(step.Params["modelIndex"].(string))
	dataSource := step.Params["dataSource"].(string)
	prompt := step.Params["prompt"].(string)
	saveToVariable := step.Params["saveToVariable"].(string)

	// 获取要验证的数据
	data := we.getVariable(dataSource)
	if data == nil {
		return ExecutionResult{Success: false}, fmt.Errorf("未找到数据源: %s", dataSource)
	}

	dataJSON, _ := json.Marshal(data)
	systemPrompt := "你是一个专业的数据验证专家。请验证提供的数据是否符合要求，返回验证结果和建议。"
	fullPrompt := fmt.Sprintf("待验证数据:\n%s\n\n验证要求:\n%s", string(dataJSON), prompt)

	result, err := we.aiService.CallAI(modelIndex, fullPrompt, systemPrompt)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
	}

	if saveToVariable != "" {
		we.setVariable(saveToVariable, result)
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("AI数据验证完成，结果保存到变量: %s", saveToVariable),
		Data:    map[string]interface{}{"result": result},
	}, nil
}

// executeAITransformData 执行AI数据转换
func (we *WorkflowExecutor) executeAITransformData(step ExecutionStep) (ExecutionResult, error) {
	modelIndex, _ := strconv.Atoi(step.Params["modelIndex"].(string))
	dataSource := step.Params["dataSource"].(string)
	prompt := step.Params["prompt"].(string)
	outputFormat := step.Params["outputFormat"].(string)
	saveToVariable := step.Params["saveToVariable"].(string)

	// 获取要转换的数据
	data := we.getVariable(dataSource)
	if data == nil {
		return ExecutionResult{Success: false}, fmt.Errorf("未找到数据源: %s", dataSource)
	}

	dataJSON, _ := json.Marshal(data)
	systemPrompt := fmt.Sprintf(`你是一个专业的数据转换专家。请将提供的数据按照要求进行转换。
输出格式要求: %s
请确保输出格式正确，不要包含任何解释文字。`, outputFormat)
	fullPrompt := fmt.Sprintf("原始数据:\n%s\n\n转换要求:\n%s", string(dataJSON), prompt)

	result, err := we.aiService.CallAI(modelIndex, fullPrompt, systemPrompt)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
	}

	if saveToVariable != "" {
		we.setVariable(saveToVariable, result)
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("AI数据转换完成，结果保存到变量: %s", saveToVariable),
		Data:    map[string]interface{}{"result": result},
	}, nil
}

// executeAISmartClick 执行AI智能点击
func (we *WorkflowExecutor) executeAISmartClick(step ExecutionStep) (ExecutionResult, error) {
	modelIndex, _ := strconv.Atoi(step.Params["modelIndex"].(string))
	prompt := step.Params["prompt"].(string)
	waitTime := int(step.Params["waitTime"].(float64))

	// 获取当前页面DOM
	domData, err := we.getCurrentPageDOM()
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("获取页面DOM失败: %v", err)
	}

	systemPrompt := `你是一个专业的网页自动化专家。请分析HTML内容，找到用户描述的元素，返回最佳的CSS选择器。
只返回CSS选择器，不要包含任何解释文字。如果找不到合适的元素，返回"NOT_FOUND"。`
	fullPrompt := fmt.Sprintf("页面HTML内容:\n%s\n\n要点击的元素描述:\n%s", domData, prompt)

	selector, err := we.aiService.CallAI(modelIndex, fullPrompt, systemPrompt)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
	}

	selector = strings.TrimSpace(selector)
	if selector == "NOT_FOUND" {
		return ExecutionResult{Success: false}, fmt.Errorf("AI未找到匹配的元素: %s", prompt)
	}

	// 执行点击操作
	clickStep := ExecutionStep{
		Action: "click",
		Params: map[string]interface{}{
			"selector":     selector,
			"selectorType": "css",
			"waitTime":     float64(waitTime),
		},
	}

	return we.executeClick(clickStep)
}

// executeAIFormFill 执行AI表单填写
func (we *WorkflowExecutor) executeAIFormFill(step ExecutionStep) (ExecutionResult, error) {
	modelIndex, _ := strconv.Atoi(step.Params["modelIndex"].(string))
	prompt := step.Params["prompt"].(string)
	dataSource := step.Params["dataSource"].(string)

	// 获取当前页面DOM
	domData, err := we.getCurrentPageDOM()
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("获取页面DOM失败: %v", err)
	}

	// 获取填写数据（如果有）
	var fillData string
	if dataSource != "" {
		data := we.getVariable(dataSource)
		if data != nil {
			dataJSON, _ := json.Marshal(data)
			fillData = string(dataJSON)
		}
	}

	systemPrompt := `你是一个专业的表单填写专家。请分析HTML内容，根据用户要求生成表单填写指令。
返回JSON格式的填写指令，格式如下:
[{"selector": "CSS选择器", "value": "填写值", "action": "input"}]
只返回JSON数组，不要包含任何解释文字。`

	fullPrompt := fmt.Sprintf("页面HTML内容:\n%s\n\n填写要求:\n%s", domData, prompt)
	if fillData != "" {
		fullPrompt += fmt.Sprintf("\n\n可用数据:\n%s", fillData)
	}

	result, err := we.aiService.CallAI(modelIndex, fullPrompt, systemPrompt)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
	}

	// 解析AI返回的填写指令
	var instructions []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &instructions); err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("解析AI指令失败: %v", err)
	}

	// 执行填写操作
	for _, instruction := range instructions {
		inputStep := ExecutionStep{
			Action: "input",
			Params: map[string]interface{}{
				"selector":     instruction["selector"],
				"selectorType": "css",
				"text":         instruction["value"],
				"waitTime":     3000.0,
			},
		}

		result, err := we.executeInput(inputStep)
		if err != nil || !result.Success {
			return ExecutionResult{Success: false}, fmt.Errorf("填写失败: %v", err)
		}
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("AI表单填写完成，共填写 %d 个字段", len(instructions)),
	}, nil
}

// executeAINavigation 执行AI智能导航
func (we *WorkflowExecutor) executeAINavigation(step ExecutionStep) (ExecutionResult, error) {
	modelIndex, _ := strconv.Atoi(step.Params["modelIndex"].(string))
	prompt := step.Params["prompt"].(string)
	waitTime := int(step.Params["waitTime"].(float64))

	// 获取当前页面DOM
	domData, err := we.getCurrentPageDOM()
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("获取页面DOM失败: %v", err)
	}

	systemPrompt := `你是一个专业的网页导航专家。请分析HTML内容，找到用户描述的导航元素（链接、按钮、菜单等），返回最佳的CSS选择器。
只返回CSS选择器，不要包含任何解释文字。如果找不到合适的元素，返回"NOT_FOUND"。`
	fullPrompt := fmt.Sprintf("页面HTML内容:\n%s\n\n导航目标描述:\n%s", domData, prompt)

	selector, err := we.aiService.CallAI(modelIndex, fullPrompt, systemPrompt)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
	}

	selector = strings.TrimSpace(selector)
	if selector == "NOT_FOUND" {
		return ExecutionResult{Success: false}, fmt.Errorf("AI未找到匹配的导航元素: %s", prompt)
	}

	// 执行点击导航
	clickStep := ExecutionStep{
		Action: "click",
		Params: map[string]interface{}{
			"selector":     selector,
			"selectorType": "css",
			"waitTime":     float64(waitTime),
		},
	}

	return we.executeClick(clickStep)
}

// getCurrentPageDOM 获取当前页面DOM
func (we *WorkflowExecutor) getCurrentPageDOM() (string, error) {
	// 检查WebSocket连接状态
	if !we.wsServer.IsRunning() {
		return "", fmt.Errorf("WebSocket服务器未运行")
	}
	if !we.wsServer.HasClients() {
		return "", fmt.Errorf("没有浏览器扩展连接")
	}

	// 通过WebSocket获取页面DOM
	message := WSMessage{
		Type: "get_page_dom",
		Data: map[string]interface{}{},
	}

	result, err := we.sendAndWait(message, 10*time.Second, "get_page_dom")
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("获取DOM失败: %v", result.Error)
	}

	if html, ok := result.Data["html"].(string); ok {
		return html, nil
	}

	return "", fmt.Errorf("DOM数据格式错误")
}
