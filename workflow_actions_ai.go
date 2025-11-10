package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// getModelIndex 安全地从参数中获取modelIndex
func getModelIndex(params map[string]interface{}) int {
	var index int

	LogDebug(fmt.Sprintf("[AI] 原始参数: %+v", params))

	if mi, ok := params["modelIndex"].(float64); ok {
		index = int(mi)
		LogDebug(fmt.Sprintf("[AI] 从 float64 获取模型索引: %d", index))
	} else if mis, ok := params["modelIndex"].(string); ok {
		if idx, err := strconv.Atoi(mis); err == nil {
			index = idx
			LogDebug(fmt.Sprintf("[AI] 从 string 获取模型索引: %d", index))
		} else {
			LogError(fmt.Sprintf("[AI] 字符串转换失败: %s, 错误: %v", mis, err))
		}
	} else {
		LogError(fmt.Sprintf("[AI] 无法获取 modelIndex 参数，类型: %T, 值: %v", params["modelIndex"], params["modelIndex"]))
	}

	LogInfo(fmt.Sprintf("[AI] 最终使用的模型索引: %d", index))
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
		AppLog.Info(fmt.Sprintf("[AI] 处理图片内容..."))

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

		AppLog.Info(fmt.Sprintf("[AI] 找到 %d 张图片", len(imageUrls)))

		// 使用自定义设置调用AI处理图片
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

		if err != nil {
			return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
		}
	}

	// 保存结果到变量
	if saveToVariable != "" {
		// 去掉花括号
		cleanSaveToVariable := saveToVariable
		we.setVariable(cleanSaveToVariable, result)
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

	result, err := we.aiService.CallAIWithCustomSettings(modelIndex, fullPrompt, systemPrompt, 3, 2, customSettings)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
	}

	if saveToVariable != "" {
		// 去掉花括号
		cleanSaveToVariable := saveToVariable
		we.setVariable(cleanSaveToVariable, result)
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("AI内容分析完成，结果保存到变量: %s", saveToVariable),
		Data:    map[string]interface{}{"result": result},
	}, nil
}

// executeAITransformData 执行AI数据转换
func (we *WorkflowExecutor) executeAITransformData(step ExecutionStep) (ExecutionResult, error) {
	modelIndex := getModelIndex(step.Params)
	prompt := step.Params["prompt"].(string)
	outputFormat := step.Params["outputFormat"].(string)
	saveToVariable := step.Params["saveToVariable"].(string)

	// dataSource 已经被 replaceVariables 替换成实际数据对象
	data := step.Params["dataSource"]
	if data == nil {
		return ExecutionResult{Success: false}, fmt.Errorf("数据源为空")
	}

	dataJSON, _ := json.Marshal(data)
	systemPrompt := fmt.Sprintf(`你是一个专业的数据转换专家。请将提供的数据按照要求进行转换。
输出格式要求: %s
请确保输出格式正确，不要包含任何解释文字。`, outputFormat)
	fullPrompt := fmt.Sprintf("原始数据:\n%s\n\n转换要求:\n%s", string(dataJSON), prompt)

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

	result, err := we.aiService.CallAIWithCustomSettings(modelIndex, fullPrompt, systemPrompt, 3, 2, customSettings)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("AI调用失败: %v", err)
	}

	if saveToVariable != "" {
		// 去掉花括号
		cleanSaveToVariable := saveToVariable
		we.setVariable(cleanSaveToVariable, result)
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("AI数据转换完成，结果保存到变量: %s", saveToVariable),
		Data:    map[string]interface{}{"result": result},
	}, nil
}

// getCurrentPageDOM 获取当前页面DOM
func (we *WorkflowExecutor) getCurrentPageDOM(timeout int, contentType string) (string, error) {
	AppLog.Info(fmt.Sprintf("[AI] 开始获取页面DOM，类型: %s", contentType))

	// 检查WebSocket连接状态
	if !we.wsServer.IsRunning() {
		AppLog.Info(fmt.Sprintf("[AI] ❌ WebSocket服务器未运行"))
		return "", fmt.Errorf("WebSocket服务器未运行")
	}
	if !we.wsServer.HasClients() {
		AppLog.Info(fmt.Sprintf("[AI] ❌ 没有浏览器扩展连接"))
		return "", fmt.Errorf("没有浏览器扩展连接")
	}

	AppLog.Info(fmt.Sprintf("[AI] ✅ WebSocket连接正常，客户端数量: %d", we.wsServer.GetClientCount()))

	// 通过WebSocket获取页面DOM
	message := WSMessage{
		Type: "get_page_dom",
		Data: map[string]interface{}{
			"contentType": contentType,
		},
	}

	LogDebug(fmt.Sprintf("[WebSocket] 发送WebSocket消息: %+v", message))
	result, err := we.sendAndWait(message, time.Duration(timeout)*time.Second, "get_page_dom")
	if err != nil {
		AppLog.Info(fmt.Sprintf("[AI] ❌ WebSocket通信失败: %v", err))
		return "", err
	}

	AppLog.Info(fmt.Sprintf("[AI] ✅ 收到WebSocket响应: Success=%v", result.Success))
	if !result.Success {
		AppLog.Info(fmt.Sprintf("[AI] ❌ DOM获取失败: %v", result.Error))
		return "", fmt.Errorf("获取DOM失败: %v", result.Error)
	}

	// Chrome扩展直接在响应中返回html字段
	if html, ok := result.Data["html"].(string); ok {
		contentSize := len(html)
		actualType := "unknown"
		if ct, ok := result.Data["contentType"].(string); ok {
			actualType = ct
		}
		AppLog.Info(fmt.Sprintf("[AI] ✅ 成功获取页面内容，类型: %s，长度: %d 字符", actualType, contentSize))
		return html, nil
	}

	AppLog.Info(fmt.Sprintf("[AI] ❌ DOM数据格式错误: %+v", result.Data))
	return "", fmt.Errorf("DOM数据格式错误")
}
