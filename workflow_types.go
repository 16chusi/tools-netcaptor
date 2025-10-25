package main

// WorkflowTask 工作流任务
type WorkflowTask struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	CreatedAt   string         `json:"createdAt"`
	UpdatedAt   string         `json:"updatedAt"`
	Nodes       []WorkflowNode `json:"nodes"`
	Edges       []WorkflowEdge `json:"edges"`
}

// WorkflowNode 工作流节点
type WorkflowNode struct {
	ID    string                 `json:"id"`
	Type  string                 `json:"type"`
	X     float64                `json:"x"`
	Y     float64                `json:"y"`
	Label string                 `json:"label"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

// WorkflowEdge 工作流边
type WorkflowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label,omitempty"`
}

// ExecutionStep 执行步骤
type ExecutionStep struct {
	NodeID string                 `json:"nodeId"`
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// ExecutionStatus 执行状态
type ExecutionStatus struct {
	TaskID       string `json:"taskId"`
	CurrentStep  int    `json:"currentStep"`
	TotalSteps   int    `json:"totalSteps"`
	Status       string `json:"status"` // running, success, failed, stopped
	CurrentNode  string `json:"currentNode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}
