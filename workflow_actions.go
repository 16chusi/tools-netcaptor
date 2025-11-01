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
	case "jsonl_reader":
		return ExecutionResult{Success: false}, fmt.Errorf("jsonl_reader 节点不应该通过 executeStep 执行")
	case "if":
		return ExecutionResult{Success: false}, fmt.Errorf("if 节点不应该通过 executeStep 执行")
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
	selector, ok := step.Params["selector"].(string)
	if !ok || selector == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 selector 参数")
	}

	selectorType := "css"
	if st, ok := step.Params["selectorType"].(string); ok && st != "" {
		selectorType = st
	}

	msg := WSMessage{
		Type: "click_element",
		Data: map[string]interface{}{
			"selector":     selector,
			"selectorType": selectorType,
		},
	}

	return we.sendAndWait(msg, 10*time.Second, "click_element")
}

// executeInput 执行输入
func (we *WorkflowExecutor) executeInput(step ExecutionStep) (ExecutionResult, error) {
	selector, ok := step.Params["selector"].(string)
	if !ok || selector == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 selector 参数")
	}

	text, ok := step.Params["text"].(string)
	if !ok {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 text 参数")
	}

	selectorType := "css"
	if st, ok := step.Params["selectorType"].(string); ok && st != "" {
		selectorType = st
	}

	msg := WSMessage{
		Type: "input_text",
		Data: map[string]interface{}{
			"selector":     selector,
			"text":         text,
			"selectorType": selectorType,
		},
	}

	return we.sendAndWait(msg, 10*time.Second, "input_text")
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

// executeExtract 执行数据提取
func (we *WorkflowExecutor) executeExtract(step ExecutionStep) (ExecutionResult, error) {
	selector, _ := step.Params["selector"].(string)
	if selector == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 selector 参数")
	}

	selectorType := "css"
	if st, ok := step.Params["selectorType"].(string); ok && st != "" {
		selectorType = st
	}

	msg := WSMessage{
		Type: "extract_data",
		Data: map[string]interface{}{
			"selector":     selector,
			"attribute":    step.Params["attribute"],
			"selectorType": selectorType,
		},
	}

	result, err := we.sendAndWait(msg, 10*time.Second, "extract")

	if err == nil && result.Success {
		if saveToVariable, ok := step.Params["saveToVariable"].(string); ok && saveToVariable != "" {
			if data, ok := result.Data["value"]; ok {
				we.variables[saveToVariable] = data
				log.Printf("[Workflow] ✓ 变量已保存: %s = %v", saveToVariable, data)
			}
		}
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
