package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// executeIf 执行条件判断
func (we *WorkflowExecutor) executeIf(step ExecutionStep, task WorkflowTask) (string, error) {
	condition, ok := step.Params["condition"].(string)
	if !ok || condition == "" {
		return "", fmt.Errorf("缺少 condition 参数")
	}

	truePort, _ := step.Params["truePort"].(string)
	if truePort == "" {
		truePort = "right"
	}

	falsePort, _ := step.Params["falsePort"].(string)
	if falsePort == "" {
		falsePort = "left"
	}

	// 评估条件
	result, err := we.evaluateCondition(condition)
	if err != nil {
		return "", fmt.Errorf("条件评估失败: %w", err)
	}

	// 选择端口
	targetPort := falsePort
	if result {
		targetPort = truePort
		log.Printf("[Workflow] 条件为 true，选择端口: %s", targetPort)
	} else {
		log.Printf("[Workflow] 条件为 false，选择端口: %s", targetPort)
	}

	// 查找从指定端口出发的边
	for _, edge := range task.Edges {
		edgeSourcePort := ""
		if edge.SourcePort != nil {
			edgeSourcePort = *edge.SourcePort
		}

		if edge.Source == step.NodeID && edgeSourcePort == targetPort {
			log.Printf("[Workflow] 找到匹配的边: %s -> %s (端口: %s)", edge.Source, edge.Target, targetPort)
			return edge.Target, nil
		}
	}

	return "", fmt.Errorf("if 节点的 %s 端口未连接", targetPort)
}

// evaluateCondition 评估条件表达式
func (we *WorkflowExecutor) evaluateCondition(condition string) (bool, error) {
	// 支持的运算符
	operators := []string{"==", "!=", ">=", "<=", ">", "<"}

	var operator string
	var parts []string

	// 查找运算符
	for _, op := range operators {
		if strings.Contains(condition, op) {
			operator = op
			parts = strings.SplitN(condition, op, 2)
			break
		}
	}

	if operator == "" || len(parts) != 2 {
		return false, fmt.Errorf("无效的条件表达式: %s", condition)
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	log.Printf("[Workflow] 评估条件: %s %s %s", left, operator, right)

	// 比较
	return we.compare(left, operator, right)
}

// compare 比较两个值
func (we *WorkflowExecutor) compare(left, operator, right string) (bool, error) {
	// 尝试数字比较
	leftNum, leftErr := strconv.ParseFloat(left, 64)
	rightNum, rightErr := strconv.ParseFloat(right, 64)

	if leftErr == nil && rightErr == nil {
		// 数字比较
		switch operator {
		case "==":
			return leftNum == rightNum, nil
		case "!=":
			return leftNum != rightNum, nil
		case ">":
			return leftNum > rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		}
	}

	// 字符串比较
	switch operator {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case ">":
		return left > right, nil
	case "<":
		return left < right, nil
	case ">=":
		return left >= right, nil
	case "<=":
		return left <= right, nil
	}

	return false, fmt.Errorf("不支持的运算符: %s", operator)
}
