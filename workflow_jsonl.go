package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// executeJSONLReader 执行JSONL读取器
func (we *WorkflowExecutor) executeJSONLReader(step ExecutionStep, task WorkflowTask, stepCount int) error {
	// 获取文件路径
	filePath, ok := step.Params["filePath"].(string)
	if !ok || filePath == "" {
		return fmt.Errorf("缺少 filePath 参数")
	}

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

	// 查找循环体的结束节点（回到JSONL节点的边）
	loopEndNodeID := we.findLoopEndNode(task, step.NodeID)

	// 遍历每一行数据
	for i := 0; i < actualLoops; i++ {
		// 检查是否停止
		if we.stopped {
			return fmt.Errorf("执行已停止")
		}

		log.Printf("[Workflow] 处理第 %d/%d 行", i+1, lineCount)

		// 获取当前行数据
		lineData, err := reader.GetLine(i)
		if err != nil {
			return fmt.Errorf("读取第 %d 行失败: %w", i+1, err)
		}

		// 提取指定字段
		extractedData := reader.ExtractValue(lineData, extractKeys)

		// 保存到变量
		we.variables[saveToVariable] = extractedData
		log.Printf("[Workflow] ✓ 变量已保存: %s = %v", saveToVariable, extractedData)

		// 执行循环体（从下一个节点开始，直到循环结束节点）
		if err := we.executeLoopBody(task, nextNodeID, loopEndNodeID, stepCount); err != nil {
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

// findLoopEndNode 查找循环体的结束节点
func (we *WorkflowExecutor) findLoopEndNode(task WorkflowTask, loopNodeID string) string {
	// 查找所有指向循环节点的边，这些边的source就是循环体的结束节点
	for _, edge := range task.Edges {
		if edge.Target == loopNodeID {
			return edge.Source
		}
	}
	return ""
}

// executeLoopBody 执行循环体
func (we *WorkflowExecutor) executeLoopBody(task WorkflowTask, startNodeID string, endNodeID string, baseStepCount int) error {
	currentNodeID := startNodeID
	stepCount := baseStepCount

	for {
		// 如果到达结束节点或循环结束节点，退出
		if currentNodeID == "" || currentNodeID == endNodeID {
			break
		}

		node := we.findNode(task, currentNodeID)
		if node == nil {
			return fmt.Errorf("未找到节点: %s", currentNodeID)
		}

		// 如果是结束节点，退出
		if node.Type == "end" {
			break
		}

		// 检查是否停止
		if we.stopped {
			return fmt.Errorf("执行已停止")
		}

		stepCount++
		log.Printf("[Workflow] 循环体步骤 %d: %s", stepCount, node.Type)

		we.emitStatus(ExecutionStatus{
			TaskID:      task.ID,
			CurrentStep: stepCount,
			TotalSteps:  stepCount,
			Status:      "running",
			CurrentNode: currentNodeID,
		})

		// if 节点特殊处理
		if node.Type == "if" {
			step := ExecutionStep{
				NodeID: currentNodeID,
				Action: node.Type,
				Params: node.Data,
			}
			we.replaceVariables(&step)

			nextNodeID, err := we.executeIf(step, task)
			if err != nil {
				return fmt.Errorf("if节点执行失败: %w", err)
			}
			currentNodeID = nextNodeID
			continue
		}

		// 普通节点
		step := ExecutionStep{
			NodeID: currentNodeID,
			Action: node.Type,
			Params: node.Data,
		}

		result, err := we.executeStep(step)
		if err != nil {
			return fmt.Errorf("步骤执行失败: %w", err)
		}

		log.Printf("[Workflow] 步骤完成: %v", result.Message)

		// 查找下一个节点
		currentNodeID = we.findNextNode(task, currentNodeID, "")
	}

	return nil
}

// executeJSONLReaderLoop 执行JSONL读取器的循环逻辑（已废弃，保留兼容）
func (we *WorkflowExecutor) executeJSONLReaderLoop(task WorkflowTask, steps []ExecutionStep, jsonlStepIndex int) error {
	return fmt.Errorf("JSONL读取器暂不支持新的执行模式")
}
