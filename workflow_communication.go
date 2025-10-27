package main

import (
	"fmt"
	"log"
	"time"
)

// sendAndWait 发送消息并等待响应
func (we *WorkflowExecutor) sendAndWait(msg WSMessage, timeout time.Duration, expectedAction string) (ExecutionResult, error) {
	if !we.wsServer.IsRunning() {
		return ExecutionResult{Success: false}, fmt.Errorf("WebSocket 服务器未运行")
	}
	if !we.wsServer.HasClients() {
		return ExecutionResult{Success: false}, fmt.Errorf("没有浏览器扩展连接")
	}

	log.Printf("[Workflow] 发送: %s, 期望响应: %s", msg.Type, expectedAction)
	we.wsServer.Broadcast(msg)

	deadline := time.Now().Add(timeout)
	for {
		select {
		case response := <-we.responseCh:
			log.Printf("[Workflow] 收到响应: %+v", response.Data)
			if response.Type == "action_result" {
				if action, ok := response.Data["action"].(string); ok && action != expectedAction {
					log.Printf("[Workflow] 忽略不匹配响应: 期望 %s, 实际 %s", expectedAction, action)
					if time.Now().Before(deadline) {
						continue
					}
					break
				}
				if success, ok := response.Data["success"].(bool); ok && success {
					return ExecutionResult{Success: true, Message: "执行成功", Data: response.Data}, nil
				}
				errMsg := "执行失败"
				if err, ok := response.Data["error"].(string); ok {
					errMsg = err
				}
				return ExecutionResult{Success: false, Error: errMsg}, fmt.Errorf("%s", errMsg)
			}
		case <-time.After(time.Until(deadline)):
			return ExecutionResult{Success: false}, fmt.Errorf("执行超时")
		}
	}
}

// HandleResponse 处理来自插件的响应
func (we *WorkflowExecutor) HandleResponse(msg WSMessage) {
	select {
	case we.responseCh <- msg:
	default:
		log.Printf("[Workflow] 响应通道已满，丢弃消息")
	}
}
