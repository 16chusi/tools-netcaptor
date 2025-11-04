package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// executeJSONLReader 执行JSONL读取器
func (we *WorkflowExecutor) executeJSONLReader(step ExecutionStep, task WorkflowTask, stepCount int) error {
	return we.executeJSONLReaderWithDepth(step, task, stepCount, 0)
}

// executeJSONLReaderWithDepth 执行JSONL读取器（支持嵌套深度控制）
func (we *WorkflowExecutor) executeJSONLReaderWithDepth(step ExecutionStep, task WorkflowTask, stepCount int, depth int) error {
	log.Printf("[JSONL读取器] ========== 开始执行 (深度:%d) ==========", depth)
	log.Printf("[JSONL读取器] 参数: %+v", step.Params)
	
	// 获取文件路径
	filePath, ok := step.Params["filePath"].(string)
	if !ok || filePath == "" {
		log.Printf("[JSONL读取器] 错误: 缺少 filePath 参数")
		return fmt.Errorf("缺少 filePath 参数")
	}
	
	log.Printf("[JSONL读取器] 文件路径: %s", filePath)

	// 获取提取字段
	extractKeysStr, _ := step.Params["extractKeys"].(string)
	if extractKeysStr == "" {
		extractKeysStr = "*"
	}
	extractKeys := strings.Split(extractKeysStr, ",")
	for i := range extractKeys {
		extractKeys[i] = strings.TrimSpace(extractKeys[i])
	}

	// 获取保存变量名
	saveToVariable, _ := step.Params["saveToVariable"].(string)
	if saveToVariable == "" {
		saveToVariable = "data"
	}

	// 获取循环间隔时间（默认500ms）
	loopInterval := 500
	if interval, ok := step.Params["interval"].(float64); ok {
		if interval >= 0 {
			loopInterval = int(interval)
		}
	}

	// 获取最大循环次数（默认为全部行数，最小为1）
	maxCount := 0
	if max, ok := step.Params["maxCount"].(float64); ok {
		if max >= 1 {
			maxCount = int(max)
		}
	}

	log.Printf("[Workflow] JSONL读取器: 文件=%s, 字段=%v, 变量=%s, 间隔=%dms, 最大次数=%d", filePath, extractKeys, saveToVariable, loopInterval, maxCount)

	// 创建JSONL读取器
	reader := NewJSONLReader(filePath)
	if err := reader.Load(); err != nil {
		return fmt.Errorf("加载JSONL文件失败: %w", err)
	}

	lineCount := reader.GetLineCount()
	log.Printf("[Workflow] JSONL文件加载成功，共 %d 行", lineCount)

	if lineCount == 0 {
		return fmt.Errorf("JSONL文件为空")
	}

	// 确定实际循环次数
	actualLoops := lineCount
	if maxCount > 0 && maxCount < lineCount {
		actualLoops = maxCount
	}
	log.Printf("[Workflow] 将处理 %d 行数据", actualLoops)

	// 查找JSONL节点的下一个节点（循环体的起始节点）
	nextNodeID := we.findNextNode(task, step.NodeID, "")
	if nextNodeID == "" {
		return fmt.Errorf("JSONL读取器节点没有后续节点")
	}

	// 遍历每一行数据
	for i := 0; i < actualLoops; i++ {
		// 检查是否停止
		if we.stopped {
			return fmt.Errorf("执行已停止")
		}

		log.Printf("[Workflow] ========== 处理第 %d/%d 行 ==========", i+1, lineCount)

		// 每次循环前等待，确保页面和扩展状态稳定
		if i > 0 { // 第一次不需要等待
			log.Printf("[JSONL读取器] 等待页面稳定...")
			time.Sleep(3 * time.Second)
		}

		// 获取当前行数据
		lineData, err := reader.GetLine(i)
		if err != nil {
			return fmt.Errorf("读取第 %d 行失败: %w", i+1, err)
		}

		// 提取指定字段
		extractedData := reader.ExtractValue(lineData, extractKeys)

		// 保存到变量
		we.variables[saveToVariable] = extractedData
		log.Printf("[JSONL读取器] ✓ 变量已保存: %s = %v", saveToVariable, extractedData)
		
		log.Printf("[JSONL读取器] ========== 保存后的所有变量 ==========")
		for key, value := range we.variables {
			log.Printf("[JSONL读取器] 变量: %s (%T) = %v", key, value, value)
		}
		log.Printf("[JSONL读取器] =====================================")

		// 执行循环体（从下一个节点开始，直到结束节点）
		log.Printf("[JSONL读取器] 准备执行循环体...")
		
		// 临时使用原始的executeLoopBody，不使用多层循环
		if err := we.executeLoopBody(task, nextNodeID, stepCount); err != nil {
			return fmt.Errorf("执行第 %d 行时失败: %w", i+1, err)
		}

		// 循环间隔（最后一次不需要等待）
		if i < actualLoops-1 && loopInterval > 0 {
			time.Sleep(time.Duration(loopInterval) * time.Millisecond)
		}
	}

	log.Printf("[Workflow] JSONL读取器执行完成，共处理 %d 行", actualLoops)
	return nil
}

// executeLoopBody 执行循环体（从当前节点到结束节点）
func (we *WorkflowExecutor) executeLoopBody(task WorkflowTask, startNodeID string, baseStepCount int) error {
	return we.executeLoopBodyWithDepth(task, startNodeID, baseStepCount, 0)
}

// executeLoopBodyWithDepth 执行循环体（支持嵌套深度控制）
func (we *WorkflowExecutor) executeLoopBodyWithDepth(task WorkflowTask, startNodeID string, baseStepCount int, depth int) error {
	const MAX_NEST_DEPTH = 5 // 最大嵌套深度

	if depth > MAX_NEST_DEPTH {
		return fmt.Errorf("嵌套循环深度超过限制: %d", MAX_NEST_DEPTH)
	}

	currentNodeID := startNodeID
	stepCount := baseStepCount

	for {
		// 如果节点ID为空，退出
		if currentNodeID == "" {
			break
		}

		node := we.findNode(task, currentNodeID)
		if node == nil {
			return fmt.Errorf("未找到节点: %s", currentNodeID)
		}

		// 如果是结束节点，退出
		if node.Type == "end" {
			log.Printf("[Workflow] 循环体到达结束节点")
			break
		}

		// 检查是否停止
		if we.stopped {
			return fmt.Errorf("执行已停止")
		}

		stepCount++
		log.Printf("[Workflow] 循环体步骤 %d (深度:%d): %s", stepCount, depth, node.Type)

		we.emitStatus(ExecutionStatus{
			TaskID:      task.ID,
			CurrentStep: stepCount,
			TotalSteps:  stepCount,
			Status:      "running",
			CurrentNode: currentNodeID,
		})

		// for 节点特殊处理 - 支持嵌套
		if node.Type == "for" {
			paramsCopy := make(map[string]interface{})
			for k, v := range node.Data {
				paramsCopy[k] = v
			}
			step := ExecutionStep{
				NodeID: currentNodeID,
				Action: node.Type,
				Params: paramsCopy,
			}
			we.replaceVariables(&step)

			log.Printf("[Workflow] 执行嵌套for循环 (深度:%d)", depth+1)
			err := we.executeForWithDepth(step, task, stepCount, depth+1)
			if err != nil {
				return fmt.Errorf("嵌套for循环执行失败: %w", err)
			}

			// 查找下一个节点
			currentNodeID = we.findNextNode(task, currentNodeID, "")
			continue
		}

		// jsonl_reader 节点特殊处理 - 支持嵌套
		if node.Type == "jsonl_reader" {
			paramsCopy := make(map[string]interface{})
			for k, v := range node.Data {
				paramsCopy[k] = v
			}
			step := ExecutionStep{
				NodeID: currentNodeID,
				Action: node.Type,
				Params: paramsCopy,
			}
			we.replaceVariables(&step)

			log.Printf("[Workflow] 执行嵌套JSONL读取器 (深度:%d)", depth+1)
			err := we.executeJSONLReaderWithDepth(step, task, stepCount, depth+1)
			if err != nil {
				return fmt.Errorf("嵌套JSONL读取器执行失败: %w", err)
			}

			// 查找下一个节点
			currentNodeID = we.findNextNode(task, currentNodeID, "")
			continue
		}

		// if 节点特殊处理 - 深拷贝 Params
		if node.Type == "if" {
			paramsCopy := make(map[string]interface{})
			for k, v := range node.Data {
				paramsCopy[k] = v
			}
			step := ExecutionStep{
				NodeID: currentNodeID,
				Action: node.Type,
				Params: paramsCopy,
			}
			we.replaceVariables(&step)

			nextNodeID, err := we.executeIf(step, task)
			if err != nil {
				return fmt.Errorf("if节点执行失败: %w", err)
			}
			currentNodeID = nextNodeID
			continue
		}

		// 普通节点 - 深拷贝 Params 避免修改原始数据
		paramsCopy := make(map[string]interface{})
		for k, v := range node.Data {
			paramsCopy[k] = v
		}
		step := ExecutionStep{
			NodeID: currentNodeID,
			Action: node.Type,
			Params: paramsCopy,
		}
		we.replaceVariables(&step)

		log.Printf("[循环体执行器] ========== 当前变量状态 ==========")
		for key, value := range we.variables {
			log.Printf("[循环体执行器] 变量: %s (%T) = %v", key, value, value)
		}
		log.Printf("[循环体执行器] =====================================")

		// 获取节点的自定义标签或类型作为显示名称
		nodeDisplayName := node.Type
		if customLabel, ok := node.Data["customLabel"].(string); ok && customLabel != "" {
			nodeDisplayName = fmt.Sprintf("%s(%s)", customLabel, node.Type)
		}

		log.Printf("[循环体执行器] 准备执行节点: [%s] ID=%s, 参数=%+v", nodeDisplayName, currentNodeID, step.Params)
		
		result, err := we.executeStep(step)
		if err != nil {
			log.Printf("[循环体执行器] ❌ 节点执行失败: [%s] ID=%s, 错误=%v", nodeDisplayName, currentNodeID, err)
			return fmt.Errorf("步骤执行失败: %w", err)
		}

		// 安全访问result.Message
		message := "步骤完成"
		if result.Message != "" {
			message = result.Message
		}
		log.Printf("[循环体执行器] ✅ 节点执行成功: [%s] %v", nodeDisplayName, message)

		// 查找下一个节点
		currentNodeID = we.findNextNode(task, currentNodeID, "")
	}

	return nil
}

// executeJSONLReaderLoop 执行JSONL读取器的循环逻辑（已废弃，保留兼容）
func (we *WorkflowExecutor) executeJSONLReaderLoop(task WorkflowTask, steps []ExecutionStep, jsonlStepIndex int) error {
	return fmt.Errorf("JSONL读取器暂不支持新的执行模式")
}
