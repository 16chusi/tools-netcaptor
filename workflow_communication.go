package main

import (
	"fmt"
	"time"
)

// sendAndWait 发送消息并等待响应
func (we *WorkflowExecutor) sendAndWait(msg WSMessage, timeout time.Duration, expectedAction string) (ExecutionResult, error) {
	if !we.wsServer.IsRunning() {
		AppLog.Info(fmt.Sprintf("[WebSocket] ❌ 服务器未运行"))
		return ExecutionResult{Success: false}, fmt.Errorf("WebSocket 服务器未运行")
	}
	if !we.wsServer.HasClients() {
		AppLog.Info(fmt.Sprintf("[WebSocket] ❌ 没有客户端连接"))
		return ExecutionResult{Success: false}, fmt.Errorf("没有浏览器扩展连接")
	}

	AppLog.Info(fmt.Sprintf("[WebSocket] 发送消息: %s, 期望响应: %s, 超时: %.1f秒", msg.Type, expectedAction, timeout.Seconds()))
	we.wsServer.Broadcast(msg)

	deadline := time.Now().Add(timeout)
	AppLog.Info(fmt.Sprintf("[WebSocket] 开始等待响应，截止时间: %s", deadline.Format("15:04:05")))

	for {
		select {
		case response := <-we.responseCh:
			AppLog.Info(fmt.Sprintf("[WebSocket] 收到响应: Type=%s, Success=%v", response.Type, response.Data["success"]))

			// 处理action_result类型响应
			if response.Type == "action_result" {
				if action, ok := response.Data["action"].(string); ok && action != expectedAction {
					AppLog.Info(fmt.Sprintf("[WebSocket] ⚠️ 忽略不匹配响应: 期望 %s, 实际 %s", expectedAction, action))
					if time.Now().Before(deadline) {
						continue
					}
					break
				}
				if success, ok := response.Data["success"].(bool); ok && success {
					AppLog.Info(fmt.Sprintf("[WebSocket] ✅ 执行成功"))
					return ExecutionResult{Success: true, Message: "执行成功", Data: response.Data}, nil
				}
				errMsg := "执行失败"
				if err, ok := response.Data["error"].(string); ok {
					errMsg = err
				}
				AppLog.Info(fmt.Sprintf("[WebSocket] ❌ 执行失败: %s", errMsg))
				return ExecutionResult{Success: false, Error: errMsg}, fmt.Errorf("%s", errMsg)
			}

			// 处理直接类型响应 (如screenshot, get_page_info)
			if response.Type == expectedAction {
				if success, ok := response.Data["success"].(bool); ok && success {
					AppLog.Info(fmt.Sprintf("[WebSocket] ✅ 直接响应成功"))
					return ExecutionResult{Success: true, Message: "执行成功", Data: response.Data}, nil
				}
				errMsg := "执行失败"
				if err, ok := response.Data["error"].(string); ok {
					errMsg = err
				}
				AppLog.Info(fmt.Sprintf("[WebSocket] ❌ 直接响应失败: %s", errMsg))
				return ExecutionResult{Success: false, Error: errMsg}, fmt.Errorf("%s", errMsg)
			}

			AppLog.Info(fmt.Sprintf("[WebSocket] ⚠️ 收到未匹配的响应类型: %s", response.Type))

		case <-time.After(time.Until(deadline)):
			AppLog.Info(fmt.Sprintf("[WebSocket] ❌ 执行超时，等待时间: %.1f秒", timeout.Seconds()))
			return ExecutionResult{Success: false}, fmt.Errorf("执行超时")
		}
	}
}

// HandleResponse 处理来自插件的响应
func (we *WorkflowExecutor) HandleResponse(msg WSMessage) {
	AppLog.Info(fmt.Sprintf("[响应处理器] 收到响应消息: Type=%s, Data=%+v", msg.Type, msg.Data))
	select {
	case we.responseCh <- msg:
		AppLog.Info(fmt.Sprintf("[响应处理器] 消息已发送到响应通道"))
	default:
		AppLog.Info(fmt.Sprintf("[响应处理器] 响应通道已满，丢弃消息"))
	}
}
