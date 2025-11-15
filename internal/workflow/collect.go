package workflow

import (
	"encoding/json"
	"fmt"
	"os"

	"netcaptor/internal/utils"
)

// executeCollect 执行数据收集
func (we *WorkflowExecutor) executeCollect(step ExecutionStep) (ExecutionResult, error) {
	dataVariable, ok := step.Params["dataVariable"].(string)
	if !ok || dataVariable == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少数据来源变量")
	}

	filePath, ok := step.Params["filePath"].(string)
	if !ok || filePath == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少保存文件路径")
	}

	format, ok := step.Params["format"].(string)
	if !ok || format == "" {
		format = "jsonl"
	}

	data := we.resolveVariablePath(dataVariable)
	if data == nil {
		return ExecutionResult{Success: false}, fmt.Errorf("变量 %s 不存在", dataVariable)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var content string
	switch format {
	case "jsonl":
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return ExecutionResult{Success: false}, fmt.Errorf("JSON序列化失败: %w", err)
		}
		content = string(jsonBytes) + "\n"
	case "text":
		content = fmt.Sprintf("%v\n", data)
	default:
		return ExecutionResult{Success: false}, fmt.Errorf("不支持的格式: %s", format)
	}

	if _, err := file.WriteString(content); err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("写入文件失败: %w", err)
	}

	utils.AppLog.Info(fmt.Sprintf("[Workflow] ✓ 数据已追加到文件: %s", filePath))

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("数据已追加到 %s", filePath),
	}, nil
}
