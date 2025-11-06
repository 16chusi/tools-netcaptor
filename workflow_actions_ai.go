package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// getModelIndex 安全地从参数中获取modelIndex
func getModelIndex(params map[string]interface{}) int {
	var index int
	if mi, ok := params["modelIndex"].(float64); ok {
		index = int(mi)
	} else if mis, ok := params["modelIndex"].(string); ok {
		if idx, err := strconv.Atoi(mis); err == nil {
			index = idx
		}
	}
	log.Printf("[AI] 用户选择的模型索引: %d", index)
	return index
}

// executeAIExtractData 执行AI数据提取
func (we *WorkflowExecutor) executeAIExtractData(step ExecutionStep) (ExecutionResult, error) {
	modelIndex := getModelIndex(step.Params)
	prompt := step.Params["prompt"].(string)
	outputFormat := step.Params["outputFormat"].(string)
	saveToVariable := step.Params["saveToVariable"].(string)

	// 获取用户配置的超时时间，默认60秒
	timeout := 60
	if t, ok := step.Params["timeout"].(float64); ok && t > 0 {
		timeout = int(t)
	}

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
	contentType := "text" // 默认值
	if ct, ok := step.Params["contentType"].(string); ok {
		contentType = ct
	}
	domData, err := we.getCurrentPageDOM(timeout, contentType)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("获取页面DOM失败: %v", err)
	}

	// 构建系统提示词和用户提示词
	var result string

	if contentType == "image" {
		// 图片内容的特殊处理
		log.Printf("[AI] 处理图片内容...")

		// 解析图片数据
		var imageUrls []string
		if err := json.Unmarshal([]byte(domData), &imageUrls); err != nil {
			// 如果解析失败，尝试解析为图片对象数组
			var imageData []map[string]interface{}
			if err := json.Unmarshal([]byte(domData), &imageData); err != nil {
				return ExecutionResult{Success: false}, fmt.Errorf("图片数据格式错误: %v", err)
			}
			// 提取图片URL
			for _, img := range imageData {
				if src, ok := img["src"].(string); ok && src != "" {
					imageUrls = append(imageUrls, src)
				}
			}
		}

		if len(imageUrls) == 0 {
			return ExecutionResult{Success: false}, fmt.Errorf("未找到有效的图片URL")
		}

		log.Printf("[AI] 找到 %d 张图片", len(imageUrls))

		// 使用自定义设置调用AI处理图片
		var customSettings *AICustomSettings
		if useCustom, ok := step.Params["useCustomSettings"].(bool); ok && useCustom {
			customSettings = &AICustomSettings{}
			if tm, ok := step.Params["thinkingMode"].(string); ok {
				customSettings.ThinkingMode = tm
			}
			if tp, ok := step.Params["topP"].(float64); ok {
				customSettings.TopP = tp
			}
			if temp, ok := step.Params["temperature"].(float64); ok {
				customSettings.Temperature = temp
			}
			if mt, ok := step.Params["maxTokens"].(float64); ok {
				customSettings.MaxTokens = int(mt)
			}
		}

		result, err = we.aiService.CallAIWithImages(modelIndex, prompt, imageUrls, customSettings)
		if err != nil {
			return ExecutionResult{Success: false}, fmt.Errorf("AI图片处理失败: %v", err)
		}

	} else {
		// 文本内容的常规处理
		systemPrompt := fmt.Sprintf(`你是一个专业的网页数据提取助手。请分析提供的内容，根据用户要求提取数据。
输出格式要求: %s
请确保输出格式正确，不要包含任何解释文字，只返回提取的数据。`, outputFormat)

		fullPrompt := fmt.Sprintf("页面内容:\n%s\n\n提取要求:\n%s", domData, prompt)

		if useCustom, ok := step.Params["useCustomSettings"].(bool); ok && useCustom {
			// 使用自定义设置
			customSettings := &AICustomSettings{}
			if tm, ok := step.Params["thinkingMode"].(string); ok {
				customSettings.ThinkingMode = tm
			}
			if tp, ok := step.Params["topP"].(float64); ok {
				customSettings.TopP = tp
			}
			if temp, ok := step.Params["temperature"].(float64); ok {
				customSettings.Temperature = temp
			}
			if mt, ok := step.Params["maxTokens"].(float64); ok {
				customSettings.MaxTokens = int(mt)
			}
			result, err = we.aiService.CallAIWithCustomSettings(modelIndex, fullPrompt, systemPrompt, retryCount, retryDelay, customSettings)
		} else {
			// 使用全局设置
			result, err = we.aiService.CallAIWithRetry(modelIndex, fullPrompt, systemPrompt, retryCount, retryDelay)
		}

		if err != nil {
			return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
		}
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
	modelIndex := getModelIndex(step.Params)
	prompt := step.Params["prompt"].(string)
	saveToVariable := step.Params["saveToVariable"].(string)

	// 获取当前页面DOM
	domData, err := we.getCurrentPageDOM(60, "text") // 默认60秒超时，文本内容
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
	modelIndex := getModelIndex(step.Params)
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
	modelIndex := getModelIndex(step.Params)
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
	modelIndex := getModelIndex(step.Params)
	prompt := step.Params["prompt"].(string)
	waitTime := int(step.Params["waitTime"].(float64))

	// 获取当前页面DOM
	domData, err := we.getCurrentPageDOM(60, "text") // 默认60秒超时，文本内容
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
	modelIndex := getModelIndex(step.Params)
	prompt := step.Params["prompt"].(string)
	dataSource := step.Params["dataSource"].(string)

	// 获取当前页面DOM
	domData, err := we.getCurrentPageDOM(60, "text") // 默认60秒超时，文本内容
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
	modelIndex := getModelIndex(step.Params)
	prompt := step.Params["prompt"].(string)
	waitTime := int(step.Params["waitTime"].(float64))

	// 获取当前页面DOM
	domData, err := we.getCurrentPageDOM(60, "text") // 默认60秒超时，文本内容
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
func (we *WorkflowExecutor) getCurrentPageDOM(timeout int, contentType string) (string, error) {
	log.Printf("[AI] 开始获取页面DOM，类型: %s", contentType)

	// 检查WebSocket连接状态
	if !we.wsServer.IsRunning() {
		log.Printf("[AI] ❌ WebSocket服务器未运行")
		return "", fmt.Errorf("WebSocket服务器未运行")
	}
	if !we.wsServer.HasClients() {
		log.Printf("[AI] ❌ 没有浏览器扩展连接")
		return "", fmt.Errorf("没有浏览器扩展连接")
	}

	log.Printf("[AI] ✅ WebSocket连接正常，客户端数量: %d", we.wsServer.GetClientCount())

	// 通过WebSocket获取页面DOM
	message := WSMessage{
		Type: "get_page_dom",
		Data: map[string]interface{}{
			"contentType": contentType,
		},
	}

	log.Printf("[AI] 发送WebSocket消息: %+v", message)
	result, err := we.sendAndWait(message, time.Duration(timeout)*time.Second, "get_page_dom")
	if err != nil {
		log.Printf("[AI] ❌ WebSocket通信失败: %v", err)
		return "", err
	}

	log.Printf("[AI] ✅ 收到WebSocket响应: Success=%v", result.Success)
	if !result.Success {
		log.Printf("[AI] ❌ DOM获取失败: %v", result.Error)
		return "", fmt.Errorf("获取DOM失败: %v", result.Error)
	}

	// Chrome扩展直接在响应中返回html字段
	if html, ok := result.Data["html"].(string); ok {
		contentSize := len(html)
		actualType := "unknown"
		if ct, ok := result.Data["contentType"].(string); ok {
			actualType = ct
		}
		log.Printf("[AI] ✅ 成功获取页面内容，类型: %s，长度: %d 字符", actualType, contentSize)
		return html, nil
	}

	log.Printf("[AI] ❌ DOM数据格式错误: %+v", result.Data)
	return "", fmt.Errorf("DOM数据格式错误")
}
