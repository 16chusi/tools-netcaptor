package main

import (
	"fmt"
	"strconv"
	"time"
)

// executeFor 执行for循环
func (we *WorkflowExecutor) executeFor(step ExecutionStep, task WorkflowTask, stepCount int) error {
	return we.executeForWithDepth(step, task, stepCount, 0)
}

// executeForWithDepth 执行for循环（支持嵌套深度控制）
func (we *WorkflowExecutor) executeForWithDepth(step ExecutionStep, task WorkflowTask, stepCount int, depth int) error {
	var count float64
	var ok bool

	// 尝试从 float64 获取循环次数
	if count, ok = step.Params["count"].(float64); ok {
		// 数字类型，直接使用
	} else if countStr, isStr := step.Params["count"].(string); isStr {
		// 字符串类型，可能包含变量，先替换变量再转换
		replacedStr := we.replaceVariablesInString(countStr)
		if parsedCount, err := strconv.ParseFloat(replacedStr, 64); err == nil {
			count = parsedCount
		} else {
			return fmt.Errorf("循环次数无法解析为数字: %s", replacedStr)
		}
	} else {
		return fmt.Errorf("缺少有效的循环次数")
	}

	if count <= 0 {
		return fmt.Errorf("循环次数必须大于0，当前值: %v", count)
	}

	variable, _ := step.Params["variable"].(string)
	if variable == "" {
		variable = "index"
	}

	interval := 500
	if i, ok := step.Params["interval"].(float64); ok && i >= 0 {
		interval = int(i)
	}

	loopCount := int(count)
	AppLog.Info(fmt.Sprintf("[Workflow] For循环 (深度:%d): 次数=%d, 变量=%s, 间隔=%dms", depth, loopCount, variable, interval))

	nextNodeID := we.findNextNode(task, step.NodeID, "")
	if nextNodeID == "" {
		return fmt.Errorf("for循环节点没有后续节点")
	}

	for i := 1; i <= loopCount; i++ {
		if we.stopped {
			return fmt.Errorf("执行已停止")
		}

		LogDebug(fmt.Sprintf("[For] ========== 循环第 %d/%d 次 (深度:%d) ==========", i, loopCount, depth))

		we.variables[variable] = i
		AppLog.Info(fmt.Sprintf("[Workflow] ✓ 变量已设置: %s = %d", variable, i))
		we.printAllVariables()

		if err := we.executeLoopBodyWithDepth(task, nextNodeID, stepCount, depth); err != nil {
			return fmt.Errorf("循环第 %d 次失败: %w", i, err)
		}

		if i < loopCount && interval > 0 {
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}

	AppLog.Info(fmt.Sprintf("[Workflow] For循环执行完成 (深度:%d)，共循环 %d 次", depth, loopCount))
	return nil
}

func (we *WorkflowExecutor) printAllVariables() {
	LogDebug(fmt.Sprintf("[Workflow] ========== 当前所有变量 =========="))
	for key, value := range we.variables {
		AppLog.Info(fmt.Sprintf("[Workflow]   %s (%T) = %v", key, value, value))
	}
	AppLog.Info(fmt.Sprintf("[Workflow] ====================================="))
}
