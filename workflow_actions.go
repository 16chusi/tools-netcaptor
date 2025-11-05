package main

import (
	"fmt"
	"log"
	"time"
)

// executeStep 执行单个步骤
func (we *WorkflowExecutor) executeStep(step ExecutionStep) (ExecutionResult, error) {
	we.replaceVariables(&step)

	switch step.Action {
	case "navigate":
		return we.executeNavigate(step)
	case "click":
		return we.executeClick(step)
	case "input":
		return we.executeInput(step)
	case "wait":
		return we.executeWait(step)
	case "extract":
		return we.executeExtract(step)
	case "download":
		return we.executeDownload(step)
	case "scroll":
		return we.executeScroll(step)
	case "intercept_request":
		return we.executeInterceptRequest(step)
	case "download_captured":
		return we.executeDownloadCaptured(step)
	case "collect":
		return we.executeCollect(step)
	case "decrypt":
		return we.executeDecrypt(step)
	case "jsonl_reader":
		return ExecutionResult{Success: false}, fmt.Errorf("jsonl_reader 节点不应该通过 executeStep 执行")
	case "if":
		return ExecutionResult{Success: false}, fmt.Errorf("if 节点不应该通过 executeStep 执行")
	case "for":
		return ExecutionResult{Success: false}, fmt.Errorf("for 节点不应该通过 executeStep 执行")
	case "ai_extract_data":
		return we.executeAIExtractData(step)
	case "ai_analyze_content":
		return we.executeAIAnalyzeContent(step)
	case "ai_validate_data":
		return we.executeAIValidateData(step)
	case "ai_transform_data":
		return we.executeAITransformData(step)
	case "ai_smart_click":
		return we.executeAISmartClick(step)
	case "ai_form_fill":
		return we.executeAIFormFill(step)
	case "ai_navigation":
		return we.executeAINavigation(step)
	default:
		return ExecutionResult{Success: false}, fmt.Errorf("未知的操作类型: %s", step.Action)
	}
}

// executeNavigate 执行导航
func (we *WorkflowExecutor) executeNavigate(step ExecutionStep) (ExecutionResult, error) {
	url, ok := step.Params["url"].(string)
	if !ok || url == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 URL 参数")
	}

	log.Printf("[Workflow] 导航到: %s", url)
	log.Printf("[Workflow] 当前变量: %+v", we.variables)

	openMode := "current"
	if mode, ok := step.Params["openMode"].(string); ok && mode != "" {
		openMode = mode
	}

	msg := WSMessage{
		Type: "navigate",
		Data: map[string]interface{}{
			"url":      url,
			"openMode": openMode,
		},
	}

	return we.sendAndWait(msg, 15*time.Second, "navigate")
}

// executeClick 执行点击
func (we *WorkflowExecutor) executeClick(step ExecutionStep) (ExecutionResult, error) {
	log.Printf("[点击元素节点] 开始执行")
	selector, ok := step.Params["selector"].(string)
	if !ok || selector == "" {
		log.Printf("[点击元素节点] 错误: 缺少 selector 参数")
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 selector 参数")
	}

	selectorType := "css"
	if st, ok := step.Params["selectorType"].(string); ok && st != "" {
		selectorType = st
	}

	log.Printf("[点击元素节点] 执行参数: selector=%s, selectorType=%s", selector, selectorType)

	msg := WSMessage{
		Type: "click_element",
		Data: map[string]interface{}{
			"selector":     selector,
			"selectorType": selectorType,
		},
	}

	result, err := we.sendAndWait(msg, 10*time.Second, "click_element")
	if err != nil {
		log.Printf("[点击元素节点] 执行失败: %v", err)
	} else {
		log.Printf("[点击元素节点] 执行成功")
	}
	return result, err
}

// executeInput 执行输入
func (we *WorkflowExecutor) executeInput(step ExecutionStep) (ExecutionResult, error) {
	log.Printf("[输入文本节点] 开始执行")
	log.Printf("[输入文本节点] 参数: %+v", step.Params)
	
	// 检查WebSocket连接状态
	if !we.wsServer.IsRunning() {
		log.Printf("[输入文本节点] WebSocket服务器未运行")
		return ExecutionResult{Success: false}, fmt.Errorf("WebSocket服务器未运行")
	}
	if !we.wsServer.HasClients() {
		log.Printf("[输入文本节点] 没有浏览器扩展连接")
		return ExecutionResult{Success: false}, fmt.Errorf("没有浏览器扩展连接")
	}
	log.Printf("[输入文本节点] WebSocket连接正常")
	
	selector, ok := step.Params["selector"].(string)
	if !ok || selector == "" {
		log.Printf("[输入文本节点] 错误: 缺少 selector 参数")
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 selector 参数")
	}

	text, ok := step.Params["text"].(string)
	if !ok {
		log.Printf("[输入文本节点] 错误: 缺少 text 参数")
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 text 参数")
	}

	selectorType := "css"
	if st, ok := step.Params["selectorType"].(string); ok && st != "" {
		selectorType = st
	}

	log.Printf("[输入文本节点] 执行参数: selector=%s, text=%s, selectorType=%s", selector, text, selectorType)

	msg := WSMessage{
		Type: "input_text",
		Data: map[string]interface{}{
			"selector":     selector,
			"text":         text,
			"selectorType": selectorType,
		},
	}

	log.Printf("[输入文本节点] 发送WebSocket消息")
	result, err := we.sendAndWait(msg, 10*time.Second, "input_text")
	if err != nil {
		log.Printf("[输入文本节点] 执行失败: %v", err)
	} else {
		log.Printf("[输入文本节点] 执行成功")
	}
	
	return result, err
}

// executeWait 执行等待
func (we *WorkflowExecutor) executeWait(step ExecutionStep) (ExecutionResult, error) {
	duration := 1000
	if d, ok := step.Params["duration"].(float64); ok {
		duration = int(d)
	}

	time.Sleep(time.Duration(duration) * time.Millisecond)

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("等待 %d 毫秒", duration),
	}, nil
}

// executeExtract 执行网页内容获取
func (we *WorkflowExecutor) executeExtract(step ExecutionStep) (ExecutionResult, error) {
	log.Printf("[获取网页内容节点] 开始执行")
	selector, _ := step.Params["selector"].(string)
	if selector == "" {
		log.Printf("[获取网页内容节点] 错误: 缺少 selector 参数")
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 selector 参数")
	}

	selectorType := "css"
	if st, ok := step.Params["selectorType"].(string); ok && st != "" {
		selectorType = st
	}

	attribute := "text"
	if attr, ok := step.Params["attribute"].(string); ok && attr != "" {
		attribute = attr
	}

	log.Printf("[获取网页内容节点] 执行参数: selector=%s, selectorType=%s, attribute=%s", selector, selectorType, attribute)

	msg := WSMessage{
		Type: "extract_data",
		Data: map[string]interface{}{
			"selector":     selector,
			"attribute":    step.Params["attribute"],
			"selectorType": selectorType,
		},
	}

	log.Printf("[获取网页内容节点] 发送WebSocket消息到浏览器扩展: %+v", msg)
	log.Printf("[获取网页内容节点] 等待浏览器扩展响应...")
	
	result, err := we.sendAndWait(msg, 10*time.Second, "extract")

	if err == nil && result.Success {
		log.Printf("[获取网页内容节点] 收到成功响应: %+v", result.Data)
		if saveToVariable, ok := step.Params["saveToVariable"].(string); ok && saveToVariable != "" {
			if data, ok := result.Data["value"]; ok {
				log.Printf("[获取网页内容节点] 准备保存数据到变量 %s: %v", saveToVariable, data)
				
				// 检查是否覆盖了现有变量
				if oldValue, exists := we.variables[saveToVariable]; exists {
					log.Printf("[获取网页内容节点] ⚠️ 覆盖现有变量 %s: 旧值=%v, 新值=%v", saveToVariable, oldValue, data)
				}
				
				we.variables[saveToVariable] = data
				log.Printf("[获取网页内容节点] ✓ 数据已保存到变量: %s = %v", saveToVariable, data)
				
				// 验证保存是否成功
				if savedValue, exists := we.variables[saveToVariable]; exists {
					log.Printf("[获取网页内容节点] ✓ 验证保存成功: %s = %v", saveToVariable, savedValue)
				} else {
					log.Printf("[获取网页内容节点] ❌ 验证保存失败: 变量 %s 不存在", saveToVariable)
				}
			} else {
				log.Printf("[获取网页内容节点] ❌ 响应中没有 value 字段: %+v", result.Data)
			}
		} else {
			log.Printf("[获取网页内容节点] 没有配置 saveToVariable 参数，不保存数据")
		}
		log.Printf("[获取网页内容节点] 执行成功")
	} else {
		log.Printf("[获取网页内容节点] 执行失败: %v", err)
	}

	return result, err
}

// executeScroll 执行滚动
func (we *WorkflowExecutor) executeScroll(step ExecutionStep) (ExecutionResult, error) {
	scrollType, _ := step.Params["scrollType"].(string)
	if scrollType == "" {
		scrollType = "bottom"
	}

	interval := 500
	if i, ok := step.Params["interval"].(float64); ok {
		interval = int(i)
	}

	msg := WSMessage{
		Type: "scroll_page",
		Data: map[string]interface{}{
			"scrollType": scrollType,
			"interval":   interval,
		},
	}

	if scrollType == "times" {
		if times, ok := step.Params["times"].(float64); ok {
			msg.Data["times"] = int(times)
		}
	} else if scrollType == "distance" {
		if distance, ok := step.Params["distance"].(float64); ok {
			msg.Data["distance"] = int(distance)
		}
	}

	return we.sendAndWait(msg, 30*time.Second, "scroll_page")
}
