package main

import (
	"fmt"
	"strconv"
	"strings"
)

// executeIf 执行条件判断
func (we *WorkflowExecutor) executeIf(step ExecutionStep, task WorkflowTask) (string, error) {
	leftValue, _ := step.Params["leftValue"].(string)
	operator, _ := step.Params["operator"].(string)
	rightValue, _ := step.Params["rightValue"].(string)

	if leftValue == "" || operator == "" {
		return "", fmt.Errorf("缺少条件参数")
	}

	truePort, _ := step.Params["truePort"].(string)
	if truePort == "" {
		truePort = "right"
	}

	falsePort, _ := step.Params["falsePort"].(string)
	if falsePort == "" {
		falsePort = "left"
	}

	AppLog.Info(fmt.Sprintf("[Workflow] 评估条件: %s %s %s", leftValue, operator, rightValue))

	// 评估条件
	result, err := we.compare(leftValue, operator, rightValue)
	if err != nil {
		return "", fmt.Errorf("条件评估失败: %w", err)
	}

	// 选择端口
	targetPort := falsePort
	if result {
		targetPort = truePort
		AppLog.Info(fmt.Sprintf("[Workflow] 条件为 true，选择端口: %s", targetPort))
	} else {
		AppLog.Info(fmt.Sprintf("[Workflow] 条件为 false，选择端口: %s", targetPort))
	}

	// 查找从指定端口出发的边
	for _, edge := range task.Edges {
		edgeSourcePort := ""
		if edge.SourcePort != nil {
			edgeSourcePort = *edge.SourcePort
		}

		if edge.Source == step.NodeID && edgeSourcePort == targetPort {
			AppLog.Info(fmt.Sprintf("[Workflow] 找到匹配的边: %s -> %s (端口: %s)", edge.Source, edge.Target, targetPort))
			return edge.Target, nil
		}
	}

	return "", fmt.Errorf("if 节点的 %s 端口未连接", targetPort)
}

// compare 比较两个值
func (we *WorkflowExecutor) compare(left, operator, right string) (bool, error) {
	// 字符串操作
	switch operator {
	case "contains":
		return strings.Contains(left, right), nil
	case "notContains":
		return !strings.Contains(left, right), nil
	case "startsWith":
		return strings.HasPrefix(left, right), nil
	case "endsWith":
		return strings.HasSuffix(left, right), nil
	}

	// 尝试数字比较
	leftNum, leftErr := strconv.ParseFloat(left, 64)
	rightNum, rightErr := strconv.ParseFloat(right, 64)

	if leftErr == nil && rightErr == nil {
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
