package workflow

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"netcaptor/internal/types"
	"netcaptor/internal/utils"
)

// executeScreenshot 执行网页截图
func (we *WorkflowExecutor) executeScreenshot(step ExecutionStep) (ExecutionResult, error) {
	utils.AppLog.Info(fmt.Sprintf("[Workflow] WebSocket 运行状态: %v, 客户端连接: %v", we.wsServer.IsRunning(), we.wsServer.HasClients()))

	format, _ := step.Params["format"].(string)
	if format == "" {
		format = "png"
	}

	captureType, _ := step.Params["captureType"].(string)
	if captureType == "" {
		captureType = "viewport"
	}

	saveDirectory, _ := step.Params["saveDirectory"].(string)
	if saveDirectory == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少保存目录")
	}

	filenameTemplate, _ := step.Params["filenameTemplate"].(string)
	if filenameTemplate == "" {
		filenameTemplate = "screenshot_{timestamp}"
	}

	selector, _ := step.Params["selector"].(string)
	quality := 80
	if q, ok := step.Params["quality"].(string); ok {
		switch q {
		case "100":
			quality = 100
		case "80":
			quality = 80
		case "60":
			quality = 60
		case "40":
			quality = 40
		}
	}

	// 处理变量替换
	selector = we.replaceVariablesInString(selector)
	filenameTemplate = we.replaceVariablesInString(filenameTemplate)

	// 生成完整文件路径
	var filename string
	if format == "pdf" {
		filename = filenameTemplate + ".pdf"
	} else {
		filename = filenameTemplate + "." + format
	}
	fullPath := filepath.Join(saveDirectory, filename)

	utils.AppLog.Info(fmt.Sprintf("[Workflow] 执行截图: format=%s, captureType=%s, quality=%d, path=%s", format, captureType, quality, fullPath))

	msg := types.WSMessage{
		Type: "screenshot",
		Data: map[string]interface{}{
			"format":      format,
			"captureType": captureType,
			"selector":    selector,
			"quality":     quality,
			"savePath":    fullPath,
		},
	}

	result, err := we.sendAndWait(msg, 30*time.Second, "screenshot")
	if err != nil {
		return result, err
	}

	// 处理返回的截图数据
	if result.Success {
		if imageData, ok := result.Data["imageData"].(string); ok && imageData != "" {
			// 解码base64数据
			data, err := base64.StdEncoding.DecodeString(imageData)
			if err != nil {
				return ExecutionResult{Success: false}, fmt.Errorf("解码截图数据失败: %w", err)
			}

			// 确保目录存在
			if err := os.MkdirAll(saveDirectory, 0755); err != nil {
				return ExecutionResult{Success: false}, fmt.Errorf("创建目录失败: %w", err)
			}

			// 保存文件
			if err := os.WriteFile(fullPath, data, 0644); err != nil {
				return ExecutionResult{Success: false}, fmt.Errorf("保存截图失败: %w", err)
			}

			utils.AppLog.Info(fmt.Sprintf("[Workflow] ✓ 截图已保存: %s", fullPath))
			result.Message = fmt.Sprintf("截图已保存: %s", filename)
		} else {
			return ExecutionResult{Success: false}, fmt.Errorf("未收到截图数据")
		}
	}

	return result, nil
}
