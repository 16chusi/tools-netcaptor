package main

import (
	"fmt"
	"log"
	"time"
)

// executeFor 执行for循环
func (we *WorkflowExecutor) executeFor(step ExecutionStep, task WorkflowTask, stepCount int) error {
	count, ok := step.Params["count"].(float64)
	if !ok || count <= 0 {
		return fmt.Errorf("缺少有效的循环次数")
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
	log.Printf("[Workflow] For循环: 次数=%d, 变量=%s, 间隔=%dms", loopCount, variable, interval)

	nextNodeID := we.findNextNode(task, step.NodeID, "")
	if nextNodeID == "" {
		return fmt.Errorf("for循环节点没有后续节点")
	}

	for i := 1; i <= loopCount; i++ {
		if we.stopped {
			return fmt.Errorf("执行已停止")
		}

		log.Printf("[Workflow] ========== 循环第 %d/%d 次 ==========", i, loopCount)

		we.variables[variable] = i
		log.Printf("[Workflow] ✓ 变量已设置: %s = %d", variable, i)
		we.printAllVariables()

		if err := we.executeLoopBody(task, nextNodeID, stepCount); err != nil {
			return fmt.Errorf("循环第 %d 次失败: %w", i, err)
		}

		if i < loopCount && interval > 0 {
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}

	log.Printf("[Workflow] For循环执行完成，共循环 %d 次", loopCount)
	return nil
}

func (we *WorkflowExecutor) printAllVariables() {
	log.Printf("[Workflow] ========== 当前所有变量 ==========")
	for key, value := range we.variables {
		log.Printf("[Workflow]   %s (%T) = %v", key, value, value)
	}
	log.Printf("[Workflow] =====================================")
}
