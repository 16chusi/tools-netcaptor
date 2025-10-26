package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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

		// 处理JSONL读取器的循环逻辑
		if step.Action == "jsonl_reader" {
			if err := we.executeJSONLReaderLoop(task, steps, i); err != nil {
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
			break // JSONL读取器处理完成后结束
		}

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
	// 替换参数中的变量
	we.replaceVariables(&step)

	switch step.Action {
	case "navigate":
		return we.executeNavigate(step)
	case "click":
		return we.executeClick(step)
	case "input":
		return we.executeInput(step)
	case "wait":
		return we.executeWait(step)
	case "extract":
		return we.executeExtract(step)
	case "download":
		return we.executeDownload(step)
	case "scroll":
		return we.executeScroll(step)
	case "intercept_request":
		return we.executeInterceptRequest(step)
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

	openMode := "current"
	if mode, ok := step.Params["openMode"].(string); ok && mode != "" {
		openMode = mode
	}

	msg := WSMessage{
		Type: "navigate",
		Data: map[string]interface{}{
			"url":      url,
			"openMode": openMode,
		},
	}

	return we.sendAndWait(msg, 15*time.Second, "navigate")
}

// executeClick 执行点击
func (we *WorkflowExecutor) executeClick(step ExecutionStep) (ExecutionResult, error) {
	selector, ok := step.Params["selector"].(string)
	if !ok || selector == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 selector 参数")
	}

	selectorType := "css"
	if st, ok := step.Params["selectorType"].(string); ok && st != "" {
		selectorType = st
	}

	msg := WSMessage{
		Type: "click_element",
		Data: map[string]interface{}{
			"selector":     selector,
			"selectorType": selectorType,
		},
	}

	return we.sendAndWait(msg, 10*time.Second, "click_element")
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

	selectorType := "css"
	if st, ok := step.Params["selectorType"].(string); ok && st != "" {
		selectorType = st
	}

	msg := WSMessage{
		Type: "input_text",
		Data: map[string]interface{}{
			"selector":     selector,
			"text":         text,
			"selectorType": selectorType,
		},
	}

	return we.sendAndWait(msg, 10*time.Second, "input_text")
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

// executeExtract 执行数据提取
func (we *WorkflowExecutor) executeExtract(step ExecutionStep) (ExecutionResult, error) {
	log.Printf("[Workflow] ========== executeExtract 开始 ==========")

	selector, _ := step.Params["selector"].(string)
	attribute, _ := step.Params["attribute"].(string)
	saveToVariable, _ := step.Params["saveToVariable"].(string)

	log.Printf("[Workflow] selector: %s", selector)
	log.Printf("[Workflow] attribute: %s", attribute)
	log.Printf("[Workflow] saveToVariable: %s", saveToVariable)

	if selector == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 selector 参数")
	}

	selectorType := "css"
	if st, ok := step.Params["selectorType"].(string); ok && st != "" {
		selectorType = st
	}

	msg := WSMessage{
		Type: "extract_data",
		Data: map[string]interface{}{
			"selector":     selector,
			"attribute":    attribute,
			"selectorType": selectorType,
		},
	}

	result, err := we.sendAndWait(msg, 10*time.Second, "extract")

	log.Printf("[Workflow] 提取结果: success=%v, err=%v", result.Success, err)
	log.Printf("[Workflow] 结果数据: %+v", result.Data)

	if err == nil && result.Success && saveToVariable != "" {
		// 保存到变量
		if data, ok := result.Data["value"]; ok {
			we.variables[saveToVariable] = data
			log.Printf("[Workflow] ✓ 变量已保存: %s = %v (type: %T)", saveToVariable, data, data)
		} else {
			log.Printf("[Workflow] ❌ result.Data 中没有 'value' 字段")
		}
	}

	log.Printf("[Workflow] 当前所有变量: %+v", we.variables)

	return result, err
}

// executeScroll 执行滚动
func (we *WorkflowExecutor) executeScroll(step ExecutionStep) (ExecutionResult, error) {
	scrollType, _ := step.Params["scrollType"].(string)
	if scrollType == "" {
		scrollType = "bottom"
	}

	interval := 500
	if i, ok := step.Params["interval"].(float64); ok {
		interval = int(i)
	}

	msg := WSMessage{
		Type: "scroll_page",
		Data: map[string]interface{}{
			"scrollType": scrollType,
			"interval":   interval,
		},
	}

	if scrollType == "times" {
		if times, ok := step.Params["times"].(float64); ok {
			msg.Data["times"] = int(times)
		}
	} else if scrollType == "distance" {
		if distance, ok := step.Params["distance"].(float64); ok {
			msg.Data["distance"] = int(distance)
		}
	}

	return we.sendAndWait(msg, 30*time.Second, "scroll_page")
}

// executeDownload 执行下载
func (we *WorkflowExecutor) executeDownload(step ExecutionStep) (ExecutionResult, error) {
	urlSource, _ := step.Params["urlSource"].(string)
	saveDirectory, _ := step.Params["saveDirectory"].(string)

	// 获取 URL
	var urls []string
	switch urlSource {
	case "variable":
		urlVariable, _ := step.Params["urlVariable"].(string)
		if val, ok := we.variables[urlVariable]; ok {
			if arr, isArray := val.([]interface{}); isArray {
				for _, item := range arr {
					urls = append(urls, fmt.Sprintf("%v", item))
				}
			} else {
				urls = []string{fmt.Sprintf("%v", val)}
			}
		} else {
			return ExecutionResult{Success: false}, fmt.Errorf("变量 %s 不存在", urlVariable)
		}
	case "template":
		urlTemplate, _ := step.Params["urlTemplate"].(string)
		urls = []string{we.replaceVariablesInString(urlTemplate)}
	default:
		downloadUrl, _ := step.Params["downloadUrl"].(string)
		urls = []string{downloadUrl}
	}

	if len(urls) == 0 {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少下载 URL")
	}

	if saveDirectory == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("请选择保存目录")
	}

	// 循环下载
	var downloadedFiles []string
	for i, url := range urls {
		if url == "" {
			continue
		}

		// 从 URL 提取文件名
		filename := filepath.Base(url)
		if idx := strings.Index(filename, "?"); idx > 0 {
			filename = filename[:idx]
		}
		if filename == "" || filename == "/" {
			filename = fmt.Sprintf("download_%d", i+1)
		}

		// 生成完整路径
		savePath := filepath.Join(saveDirectory, filename)

		// 检查文件是否存在，存在则添加 UUID
		if _, err := os.Stat(savePath); err == nil {
			ext := filepath.Ext(filename)
			base := strings.TrimSuffix(filename, ext)
			uuidStr := uuid.New().String()[:8]
			filename = fmt.Sprintf("%s_%s%s", base, uuidStr, ext)
			savePath = filepath.Join(saveDirectory, filename)
		}

		log.Printf("[Workflow] 下载文件 %d/%d: %s -> %s", i+1, len(urls), url, savePath)

		// 下载文件
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("[Workflow] 下载失败: %v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("[Workflow] 下载失败: HTTP %d", resp.StatusCode)
			continue
		}

		file, err := os.Create(savePath)
		if err != nil {
			log.Printf("[Workflow] 创建文件失败: %v", err)
			continue
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Printf("[Workflow] 写入文件失败: %v", err)
			os.Remove(savePath)
			continue
		}

		log.Printf("[Workflow] ✓ 下载成功: %s", filename)
		downloadedFiles = append(downloadedFiles, savePath)
	}

	if len(downloadedFiles) == 0 {
		return ExecutionResult{Success: false}, fmt.Errorf("没有文件被下载")
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("下载完成: %d 个文件", len(downloadedFiles)),
	}, nil
}

// replaceVariables 替换步骤参数中的变量
func (we *WorkflowExecutor) replaceVariables(step *ExecutionStep) {
	for key, value := range step.Params {
		if strVal, ok := value.(string); ok {
			step.Params[key] = we.replaceVariablesInString(strVal)
		}
	}
}

// replaceVariablesInString 替换字符串中的变量
func (we *WorkflowExecutor) replaceVariablesInString(str string) string {
	// 匹配 {varName}, {varName.field}, {varName[key]}
	for {
		start := strings.Index(str, "{")
		if start == -1 {
			break
		}
		end := strings.Index(str[start:], "}")
		if end == -1 {
			break
		}
		end += start

		placeholder := str[start : end+1]
		varPath := str[start+1 : end]

		// 解析变量路径
		value := we.resolveVariablePath(varPath)
		if value != nil {
			str = strings.Replace(str, placeholder, fmt.Sprintf("%v", value), 1)
		} else {
			str = str[:start] + str[end+1:]
		}
	}
	return str
}

// resolveVariablePath 解析变量路径 支持 data.url 和 data[url]
func (we *WorkflowExecutor) resolveVariablePath(path string) interface{} {
	// 处理 data.field 或 data[field]
	if strings.Contains(path, ".") {
		parts := strings.SplitN(path, ".", 2)
		if val, ok := we.variables[parts[0]]; ok {
			if mapVal, isMap := val.(map[string]interface{}); isMap {
				return mapVal[parts[1]]
			}
		}
	} else if strings.Contains(path, "[") && strings.Contains(path, "]") {
		start := strings.Index(path, "[")
		end := strings.Index(path, "]")
		varName := path[:start]
		key := path[start+1 : end]
		if val, ok := we.variables[varName]; ok {
			if mapVal, isMap := val.(map[string]interface{}); isMap {
				return mapVal[key]
			}
		}
	} else {
		// 简单变量
		return we.variables[path]
	}
	return nil
}

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

// executeJSONLReaderLoop 执行JSONL读取器的循环逻辑
func (we *WorkflowExecutor) executeJSONLReaderLoop(task WorkflowTask, steps []ExecutionStep, jsonlStepIndex int) error {
	step := steps[jsonlStepIndex]
	filePath, _ := step.Params["filePath"].(string)
	if filePath == "" {
		return fmt.Errorf("缺少文件路径")
	}

	extractKeysStr, _ := step.Params["extractKeys"].(string)
	if extractKeysStr == "" {
		extractKeysStr = "*"
	}

	interval := 100
	if i, ok := step.Params["interval"].(float64); ok {
		interval = int(i)
	}

	// 加载文件
	reader := NewJSONLReader(filePath)
	if err := reader.Load(); err != nil {
		return fmt.Errorf("加载文件失败: %w", err)
	}

	// 解析keys
	var extractKeys []string
	if extractKeysStr == "*" {
		extractKeys = []string{"*"}
	} else {
		extractKeys = strings.Split(extractKeysStr, ",")
		for i := range extractKeys {
			extractKeys[i] = strings.TrimSpace(extractKeys[i])
		}
	}

	totalLines := reader.GetLineCount()
	if m, ok := step.Params["maxCount"].(float64); ok && int(m) > 0 && int(m) < totalLines {
		totalLines = int(m)
	}

	log.Printf("[Workflow] JSONL读取器: 文件=%s, 总行数=%d, 提取字段=%v, 间隔=%dms",
		filePath, totalLines, extractKeys, interval)

	// 获取JSONL读取器之后的所有步骤
	loopSteps := steps[jsonlStepIndex+1:]

	// 循环执行每一行
	for lineIndex := 0; lineIndex < totalLines; lineIndex++ {
		if we.stopped {
			return fmt.Errorf("执行已停止")
		}

		log.Printf("[Workflow] JSONL循环: 第 %d/%d 行", lineIndex+1, totalLines)

		// 读取当前行
		lineData, err := reader.GetLine(lineIndex)
		if err != nil {
			log.Printf("[Workflow] 读取第 %d 行失败: %v", lineIndex+1, err)
			continue
		}

		// 提取数据
		extractedData := reader.ExtractValue(lineData, extractKeys)

		// 将提取的数据保存到变量中
		saveToVariable, _ := step.Params["saveToVariable"].(string)
		if saveToVariable != "" {
			// 保存为嵌套对象
			we.variables[saveToVariable] = extractedData
			log.Printf("[Workflow] 变量设置: %s = %v", saveToVariable, extractedData)
		} else {
			// 直接保存各个字段
			for key, value := range extractedData {
				we.variables[key] = value
				log.Printf("[Workflow] 变量设置: %s = %v", key, value)
			}
		}

		// 执行后续步骤
		for _, loopStep := range loopSteps {
			if we.stopped {
				return fmt.Errorf("执行已停止")
			}

			log.Printf("[Workflow] 执行循环内步骤: %s", loopStep.Action)

			// 创建步骤副本避免变量污染
			stepCopy := ExecutionStep{
				NodeID: loopStep.NodeID,
				Action: loopStep.Action,
				Params: make(map[string]interface{}),
			}
			for k, v := range loopStep.Params {
				stepCopy.Params[k] = v
			}

			_, err := we.executeStep(stepCopy)
			if err != nil {
				log.Printf("[Workflow] 循环内步骤执行失败: %v", err)
				return err
			}
		}

		// 等待间隔
		if interval > 0 && lineIndex < totalLines-1 {
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}

	log.Printf("[Workflow] JSONL循环完成: 共处理 %d 行", totalLines)
	return nil
}

// executeInterceptRequest 执行请求拦截
func (we *WorkflowExecutor) executeInterceptRequest(step ExecutionStep) (ExecutionResult, error) {
	urlPattern, _ := step.Params["urlPattern"].(string)
	if urlPattern == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 URL 匹配模式")
	}

	action, _ := step.Params["action"].(string)
	if action == "" {
		action = "block"
	}

	mockResponse, _ := step.Params["mockResponse"].(string)
	redirectUrl, _ := step.Params["redirectUrl"].(string)
	saveDirectory, _ := step.Params["saveDirectory"].(string)
	statusCode := 403
	if sc, ok := step.Params["statusCode"].(float64); ok {
		statusCode = int(sc)
	}

	log.Printf("[Workflow] 设置请求拦截: urlPattern=%s, action=%s", urlPattern, action)

	msg := WSMessage{
		Type: "setup_intercept",
		Data: map[string]interface{}{
			"urlPattern":    urlPattern,
			"action":        action,
			"mockResponse":  mockResponse,
			"redirectUrl":   redirectUrl,
			"saveDirectory": saveDirectory,
			"statusCode":    statusCode,
		},
	}

	return we.sendAndWait(msg, 10*time.Second, "setup_intercept")
}

// emitStatus 发送状态到前端
func (we *WorkflowExecutor) emitStatus(status ExecutionStatus) {
	if we.app.ctx != nil {
		runtime.EventsEmit(we.app.ctx, "workflow_status", status)
	}
}
