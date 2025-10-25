package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WorkflowExecutor 工作流执行器
type WorkflowExecutor struct {
	app        *NetworkApp
	wsServer   *WebSocketServer
	running    bool
	stopped    bool
	mu         sync.Mutex
	responseCh chan WSMessage
	variables  map[string]interface{}
}

// NewWorkflowExecutor 创建工作流执行器
func NewWorkflowExecutor(app *NetworkApp) *WorkflowExecutor {
	return &WorkflowExecutor{
		app:        app,
		wsServer:   app.wsServer,
		responseCh: make(chan WSMessage, 10),
		variables:  make(map[string]interface{}),
	}
}

// Execute 执行工作流
func (we *WorkflowExecutor) Execute(task WorkflowTask) error {
	we.mu.Lock()
	if we.running {
		we.mu.Unlock()
		return fmt.Errorf("工作流正在执行中")
	}
	we.running = true
	we.stopped = false
	we.mu.Unlock()

	defer func() {
		we.mu.Lock()
		we.running = false
		we.mu.Unlock()
	}()

	log.Printf("[Workflow] ========== 开始执行任务 ==========")
	log.Printf("[Workflow] 任务名称: %s", task.Name)
	log.Printf("[Workflow] 任务ID: %s", task.ID)
	log.Printf("[Workflow] 节点数量: %d", len(task.Nodes))
	log.Printf("[Workflow] 边数量: %d", len(task.Edges))

	// 打印所有节点的详细信息
	for i, node := range task.Nodes {
		log.Printf("[Workflow] 节点[%d] - ID: %s, Type: %s, Label: %s", i, node.ID, node.Type, node.Label)
		log.Printf("[Workflow] 节点[%d] - Data: %+v", i, node.Data)
		if node.Data != nil {
			for key, value := range node.Data {
				log.Printf("[Workflow] 节点[%d] - Data[%s] = %v (type: %T)", i, key, value, value)
			}
		}
	}

	// 构建执行计划
	steps, err := we.buildExecutionPlan(task)
	if err != nil {
		return fmt.Errorf("构建执行计划失败: %w", err)
	}

	log.Printf("[Workflow] 执行计划包含 %d 个步骤", len(steps))

	// 发送初始状态
	we.emitStatus(ExecutionStatus{
		TaskID:      task.ID,
		CurrentStep: 0,
		TotalSteps:  len(steps),
		Status:      "running",
	})

	// 逐步执行
	for i, step := range steps {
		if we.stopped {
			we.emitStatus(ExecutionStatus{
				TaskID:      task.ID,
				CurrentStep: i,
				TotalSteps:  len(steps),
				Status:      "stopped",
			})
			return fmt.Errorf("执行已停止")
		}

		log.Printf("[Workflow] 执行步骤 %d/%d: %s", i+1, len(steps), step.Action)

		we.emitStatus(ExecutionStatus{
			TaskID:      task.ID,
			CurrentStep: i + 1,
			TotalSteps:  len(steps),
			Status:      "running",
			CurrentNode: step.NodeID,
		})

		result, err := we.executeStep(step)
		if err != nil {
			errMsg := fmt.Sprintf("步骤执行失败: %v", err)
			log.Printf("[Workflow] %s", errMsg)
			we.emitStatus(ExecutionStatus{
				TaskID:       task.ID,
				CurrentStep:  i + 1,
				TotalSteps:   len(steps),
				Status:       "failed",
				CurrentNode:  step.NodeID,
				ErrorMessage: errMsg,
			})
			return err
		}

		log.Printf("[Workflow] 步骤完成: %v", result.Message)
	}

	we.emitStatus(ExecutionStatus{
		TaskID:      task.ID,
		CurrentStep: len(steps),
		TotalSteps:  len(steps),
		Status:      "success",
	})

	log.Printf("[Workflow] 任务执行完成")
	return nil
}

// Stop 停止执行
func (we *WorkflowExecutor) Stop() {
	we.mu.Lock()
	defer we.mu.Unlock()
	we.stopped = true
}

// IsRunning 是否正在运行
func (we *WorkflowExecutor) IsRunning() bool {
	we.mu.Lock()
	defer we.mu.Unlock()
	return we.running
}

// buildExecutionPlan 构建执行计划
func (we *WorkflowExecutor) buildExecutionPlan(task WorkflowTask) ([]ExecutionStep, error) {
	steps := []ExecutionStep{}

	// 找到开始节点
	var startNode *WorkflowNode
	for i := range task.Nodes {
		if task.Nodes[i].Type == "start" {
			startNode = &task.Nodes[i]
			break
		}
	}

	if startNode == nil {
		return nil, fmt.Errorf("未找到开始节点")
	}

	// 构建节点映射
	nodeMap := make(map[string]*WorkflowNode)
	for i := range task.Nodes {
		nodeMap[task.Nodes[i].ID] = &task.Nodes[i]
	}

	// 构建边映射
	edgeMap := make(map[string][]string)
	for _, edge := range task.Edges {
		edgeMap[edge.Source] = append(edgeMap[edge.Source], edge.Target)
	}

	// 从开始节点遍历
	visited := make(map[string]bool)
	we.traverseNodes(startNode.ID, nodeMap, edgeMap, visited, &steps)

	return steps, nil
}

// traverseNodes 遍历节点
func (we *WorkflowExecutor) traverseNodes(nodeID string, nodeMap map[string]*WorkflowNode,
	edgeMap map[string][]string, visited map[string]bool, steps *[]ExecutionStep) {

	if visited[nodeID] {
		return
	}
	visited[nodeID] = true

	node := nodeMap[nodeID]
	if node == nil {
		return
	}

	// 跳过开始和结束节点
	if node.Type != "start" && node.Type != "end" {
		step := we.nodeToStep(node)
		*steps = append(*steps, step)
	}

	// 继续遍历子节点
	for _, targetID := range edgeMap[nodeID] {
		we.traverseNodes(targetID, nodeMap, edgeMap, visited, steps)
	}
}

// nodeToStep 将节点转换为执行步骤
func (we *WorkflowExecutor) nodeToStep(node *WorkflowNode) ExecutionStep {
	log.Printf("[Workflow] ========== nodeToStep 开始 ==========")
	log.Printf("[Workflow] 输入节点 - ID: %s, Type: %s", node.ID, node.Type)
	log.Printf("[Workflow] 输入节点 - Data: %+v", node.Data)
	log.Printf("[Workflow] 输入节点 - Data 类型: %T", node.Data)

	if node.Data != nil {
		log.Printf("[Workflow] node.Data 详细内容:")
		for key, value := range node.Data {
			log.Printf("[Workflow]   - %s = %v (类型: %T)", key, value, value)
		}
	} else {
		log.Printf("[Workflow] node.Data 为 nil")
	}

	step := ExecutionStep{
		NodeID: node.ID,
		Action: node.Type,
		Params: node.Data,
	}

	if step.Params == nil {
		log.Printf("[Workflow] step.Params 为 nil，创建空 map")
		step.Params = make(map[string]interface{})
	}

	log.Printf("[Workflow] 输出步骤 - NodeID: %s, Action: %s, Params: %+v", step.NodeID, step.Action, step.Params)

	return step
}

// executeStep 执行单个步骤
func (we *WorkflowExecutor) executeStep(step ExecutionStep) (ExecutionResult, error) {
	switch step.Action {
	case "navigate":
		return we.executeNavigate(step)
	case "click":
		return we.executeClick(step)
	case "input":
		return we.executeInput(step)
	case "wait":
		return we.executeWait(step)
	default:
		return ExecutionResult{Success: false}, fmt.Errorf("未知的操作类型: %s", step.Action)
	}
}

// executeNavigate 执行导航
func (we *WorkflowExecutor) executeNavigate(step ExecutionStep) (ExecutionResult, error) {
	log.Printf("[Workflow] ========== executeNavigate 开始 ==========")
	log.Printf("[Workflow] step.NodeID: %s", step.NodeID)
	log.Printf("[Workflow] step.Action: %s", step.Action)
	log.Printf("[Workflow] step.Params: %+v", step.Params)
	log.Printf("[Workflow] step.Params 类型: %T", step.Params)

	if step.Params != nil {
		log.Printf("[Workflow] step.Params 键值对:")
		for key, value := range step.Params {
			log.Printf("[Workflow]   - %s = %v (类型: %T)", key, value, value)
		}
	} else {
		log.Printf("[Workflow] step.Params 为 nil")
	}

	url, ok := step.Params["url"].(string)
	log.Printf("[Workflow] 尝试获取 URL: ok=%v, url=%s", ok, url)

	if !ok || url == "" {
		log.Printf("[Workflow] ❌ URL 参数无效或为空")
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 URL 参数")
	}

	log.Printf("[Workflow] ✓ 执行导航: %s", url)

	msg := WSMessage{
		Type: "navigate",
		Data: map[string]interface{}{"url": url},
	}

	return we.sendAndWait(msg, 15*time.Second)
}

// executeClick 执行点击
func (we *WorkflowExecutor) executeClick(step ExecutionStep) (ExecutionResult, error) {
	selector, ok := step.Params["selector"].(string)
	if !ok || selector == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 selector 参数")
	}

	msg := WSMessage{
		Type: "click_element",
		Data: map[string]interface{}{"selector": selector},
	}

	return we.sendAndWait(msg, 10*time.Second)
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

	msg := WSMessage{
		Type: "input_text",
		Data: map[string]interface{}{
			"selector": selector,
			"text":     text,
		},
	}

	return we.sendAndWait(msg, 10*time.Second)
}

// executeWait 执行等待
func (we *WorkflowExecutor) executeWait(step ExecutionStep) (ExecutionResult, error) {
	duration := 1000 // 默认 1 秒
	if d, ok := step.Params["duration"].(float64); ok {
		duration = int(d)
	}

	time.Sleep(time.Duration(duration) * time.Millisecond)

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("等待 %d 毫秒", duration),
	}, nil
}

// sendAndWait 发送消息并等待响应
func (we *WorkflowExecutor) sendAndWait(msg WSMessage, timeout time.Duration) (ExecutionResult, error) {
	// 检查 WebSocket 连接
	if !we.wsServer.IsRunning() {
		log.Printf("[Workflow] ❌ WebSocket 服务器未运行")
		return ExecutionResult{Success: false}, fmt.Errorf("WebSocket 服务器未运行，请确保浏览器扩展已连接")
	}

	if !we.wsServer.HasClients() {
		log.Printf("[Workflow] ❌ 没有 WebSocket 客户端连接")
		return ExecutionResult{Success: false}, fmt.Errorf("没有浏览器扩展连接，请先安装并启用扩展")
	}

	log.Printf("[Workflow] 发送消息: %s, 数据: %+v", msg.Type, msg.Data)

	// 发送消息
	we.wsServer.Broadcast(msg)
	log.Printf("[Workflow] 消息已广播，等待响应 (超时: %v)...", timeout)

	// 等待响应
	select {
	case response := <-we.responseCh:
		log.Printf("[Workflow] 收到响应: %s, 数据: %+v", response.Type, response.Data)
		if response.Type == "action_result" {
			if success, ok := response.Data["success"].(bool); ok && success {
				return ExecutionResult{Success: true, Message: "执行成功"}, nil
			}
			errMsg := "执行失败"
			if err, ok := response.Data["error"].(string); ok {
				errMsg = err
			}
			return ExecutionResult{Success: false, Error: errMsg}, fmt.Errorf(errMsg)
		}
	case <-time.After(timeout):
		log.Printf("[Workflow] ❌ 执行超时 (等待了 %v)，浏览器扩展可能未响应", timeout)
		return ExecutionResult{Success: false}, fmt.Errorf("执行超时：浏览器扩展未在 %v 内响应，请检查扩展是否正常工作", timeout)
	}

	return ExecutionResult{Success: false}, fmt.Errorf("未收到响应")
}

// HandleResponse 处理来自插件的响应
func (we *WorkflowExecutor) HandleResponse(msg WSMessage) {
	select {
	case we.responseCh <- msg:
	default:
		log.Printf("[Workflow] 响应通道已满，丢弃消息")
	}
}

// emitStatus 发送状态到前端
func (we *WorkflowExecutor) emitStatus(status ExecutionStatus) {
	if we.app.ctx != nil {
		runtime.EventsEmit(we.app.ctx, "workflow_status", status)
	}
}
