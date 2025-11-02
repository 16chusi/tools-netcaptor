package main

import (
	"fmt"
	"log"
	"sync"

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

	return we.executeFromNode(task, we.findStartNode(task), 0)
}

// executeFromNode 从指定节点开始执行
func (we *WorkflowExecutor) executeFromNode(task WorkflowTask, nodeID string, stepCount int) error {
	if nodeID == "" {
		return fmt.Errorf("节点ID为空")
	}

	node := we.findNode(task, nodeID)
	if node == nil {
		return fmt.Errorf("未找到节点: %s", nodeID)
	}

	// 结束节点
	if node.Type == "end" {
		log.Printf("[Workflow] 到达结束节点")
		we.emitStatus(ExecutionStatus{
			TaskID:      task.ID,
			CurrentStep: stepCount,
			TotalSteps:  stepCount,
			Status:      "success",
			CurrentNode: nodeID,
		})
		return nil
	}

	// 开始节点，直接跳过
	if node.Type == "start" {
		nextNodeID := we.findNextNode(task, nodeID, "")
		if nextNodeID == "" {
			return fmt.Errorf("开始节点没有后续节点")
		}
		return we.executeFromNode(task, nextNodeID, stepCount)
	}

	// 检查是否停止
	if we.stopped {
		we.emitStatus(ExecutionStatus{
			TaskID:      task.ID,
			CurrentStep: stepCount,
			TotalSteps:  stepCount,
			Status:      "stopped",
		})
		return fmt.Errorf("执行已停止")
	}

	stepCount++
	log.Printf("[Workflow] 执行步骤 %d: %s", stepCount, node.Type)

	we.emitStatus(ExecutionStatus{
		TaskID:      task.ID,
		CurrentStep: stepCount,
		TotalSteps:  stepCount,
		Status:      "running",
		CurrentNode: nodeID,
	})

	// jsonl_reader 节点特殊处理
	if node.Type == "jsonl_reader" {
		step := ExecutionStep{
			NodeID: nodeID,
			Action: node.Type,
			Params: node.Data,
		}
		we.replaceVariables(&step)

		err := we.executeJSONLReader(step, task, stepCount)
		if err != nil {
			errMsg := fmt.Sprintf("步骤执行失败: %v", err)
			log.Printf("[Workflow] %s", errMsg)
			we.emitStatus(ExecutionStatus{
				TaskID:       task.ID,
				CurrentStep:  stepCount,
				TotalSteps:   stepCount,
				Status:       "failed",
				CurrentNode:  nodeID,
				ErrorMessage: errMsg,
			})
			return err
		}
		// JSONL读取器执行完成，循环体已经包含了所有后续节点，直接返回
		log.Printf("[Workflow] JSONL读取器循环执行完成")
		endNodeID := we.findEndNode(task)
		we.emitStatus(ExecutionStatus{
			TaskID:      task.ID,
			CurrentStep: stepCount,
			TotalSteps:  stepCount,
			Status:      "success",
			CurrentNode: endNodeID,
		})
		return nil
	}

	// if 节点特殊处理
	if node.Type == "if" {
		step := ExecutionStep{
			NodeID: nodeID,
			Action: node.Type,
			Params: node.Data,
		}
		we.replaceVariables(&step)

		nextNodeID, err := we.executeIf(step, task)
		if err != nil {
			errMsg := fmt.Sprintf("步骤执行失败: %v", err)
			log.Printf("[Workflow] %s", errMsg)
			we.emitStatus(ExecutionStatus{
				TaskID:       task.ID,
				CurrentStep:  stepCount,
				TotalSteps:   stepCount,
				Status:       "failed",
				CurrentNode:  nodeID,
				ErrorMessage: errMsg,
			})
			return err
		}
		return we.executeFromNode(task, nextNodeID, stepCount)
	}

	// for 节点特殊处理
	if node.Type == "for" {
		step := ExecutionStep{
			NodeID: nodeID,
			Action: node.Type,
			Params: node.Data,
		}
		we.replaceVariables(&step)

		err := we.executeFor(step, task, stepCount)
		if err != nil {
			errMsg := fmt.Sprintf("步骤执行失败: %v", err)
			log.Printf("[Workflow] %s", errMsg)
			we.emitStatus(ExecutionStatus{
				TaskID:       task.ID,
				CurrentStep:  stepCount,
				TotalSteps:   stepCount,
				Status:       "failed",
				CurrentNode:  nodeID,
				ErrorMessage: errMsg,
			})
			return err
		}
		endNodeID := we.findEndNode(task)
		we.emitStatus(ExecutionStatus{
			TaskID:      task.ID,
			CurrentStep: stepCount,
			TotalSteps:  stepCount,
			Status:      "success",
			CurrentNode: endNodeID,
		})
		return nil
	}

	// 普通节点
	step := ExecutionStep{
		NodeID: nodeID,
		Action: node.Type,
		Params: node.Data,
	}

	log.Printf("[Workflow] ========== 执行节点前的变量 ==========")
	for key, value := range we.variables {
		log.Printf("[Workflow]   %s (%T) = %v", key, value, value)
	}
	log.Printf("[Workflow] =====================================")

	result, err := we.executeStep(step)
	if err != nil {
		errMsg := fmt.Sprintf("步骤执行失败: %v", err)
		log.Printf("[Workflow] %s", errMsg)
		we.emitStatus(ExecutionStatus{
			TaskID:       task.ID,
			CurrentStep:  stepCount,
			TotalSteps:   stepCount,
			Status:       "failed",
			CurrentNode:  nodeID,
			ErrorMessage: errMsg,
		})
		return err
	}

	log.Printf("[Workflow] 步骤完成: %v", result.Message)
	log.Printf("[Workflow] 节点返回的 result.Data: %+v", result.Data)
	log.Printf("[Workflow] ========== 执行节点后的变量 ==========")
	for key, value := range we.variables {
		log.Printf("[Workflow]   %s (%T) = %v", key, value, value)
	}
	log.Printf("[Workflow] =====================================")

	// 查找下一个节点
	nextNodeID := we.findNextNode(task, nodeID, "")
	if nextNodeID == "" {
		return fmt.Errorf("节点 %s 没有后续节点", nodeID)
	}

	return we.executeFromNode(task, nextNodeID, stepCount)
}

// findStartNode 查找开始节点
func (we *WorkflowExecutor) findStartNode(task WorkflowTask) string {
	for _, node := range task.Nodes {
		if node.Type == "start" {
			return node.ID
		}
	}
	return ""
}

// findNode 查找节点
func (we *WorkflowExecutor) findNode(task WorkflowTask, nodeID string) *WorkflowNode {
	for i := range task.Nodes {
		if task.Nodes[i].ID == nodeID {
			return &task.Nodes[i]
		}
	}
	return nil
}

// findNextNode 查找下一个节点
func (we *WorkflowExecutor) findNextNode(task WorkflowTask, nodeID string, port string) string {
	for _, edge := range task.Edges {
		if edge.Source == nodeID {
			if port == "" || edge.SourcePort == nil || *edge.SourcePort == port {
				return edge.Target
			}
		}
	}
	return ""
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

// findEndNode 查找结束节点
func (we *WorkflowExecutor) findEndNode(task WorkflowTask) string {
	for i := range task.Nodes {
		if task.Nodes[i].Type == "end" {
			return task.Nodes[i].ID
		}
	}
	return ""
}

// emitStatus 发送状态到前端
func (we *WorkflowExecutor) emitStatus(status ExecutionStatus) {
	if we.app.ctx != nil {
		runtime.EventsEmit(we.app.ctx, "workflow_status", status)
	}
}
