package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// executeDownload 执行下载
func (we *WorkflowExecutor) executeDownload(step ExecutionStep) (ExecutionResult, error) {
	urlSource, _ := step.Params["urlSource"].(string)
	saveDirectory, _ := step.Params["saveDirectory"].(string)

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

	var downloadedFiles []string
	for i, url := range urls {
		if url == "" {
			continue
		}

		filename := filepath.Base(url)
		if idx := strings.Index(filename, "?"); idx > 0 {
			filename = filename[:idx]
		}
		if filename == "" || filename == "/" {
			filename = fmt.Sprintf("download_%d", i+1)
		}

		savePath := filepath.Join(saveDirectory, filename)

		if _, err := os.Stat(savePath); err == nil {
			ext := filepath.Ext(filename)
			base := strings.TrimSuffix(filename, ext)
			uuidStr := uuid.New().String()[:8]
			filename = fmt.Sprintf("%s_%s%s", base, uuidStr, ext)
			savePath = filepath.Join(saveDirectory, filename)
		}

		log.Printf("[Workflow] 下载文件 %d/%d: %s -> %s", i+1, len(urls), url, savePath)

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

// executeInterceptRequest 执行请求拦截
func (we *WorkflowExecutor) executeInterceptRequest(step ExecutionStep) (ExecutionResult, error) {
	log.Printf("[Workflow] WebSocket 运行状态: %v, 客户端连接: %v", we.wsServer.IsRunning(), we.wsServer.HasClients())

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
	fileExtension, _ := step.Params["fileExtension"].(string)
	dataFormat, _ := step.Params["dataFormat"].(string)
	saveToVariable, _ := step.Params["saveToVariable"].(string)
	statusCode := 403
	if sc, ok := step.Params["statusCode"].(float64); ok {
		statusCode = int(sc)
	}

	log.Printf("[Workflow] 设置请求拦截: urlPattern=%s, action=%s, dataFormat=%s, saveToVariable=%s", urlPattern, action, dataFormat, saveToVariable)

	msg := WSMessage{
		Type: "setup_intercept",
		Data: map[string]interface{}{
			"urlPattern":     urlPattern,
			"action":         action,
			"mockResponse":   mockResponse,
			"redirectUrl":    redirectUrl,
			"saveDirectory":  saveDirectory,
			"fileExtension":  fileExtension,
			"statusCode":     statusCode,
			"dataFormat":     dataFormat,
			"saveToVariable": saveToVariable,
		},
	}
	log.Printf("[Workflow] 发送 WebSocket 消息: %+v", msg.Data)

	result, err := we.sendAndWait(msg, 10*time.Second, "setup_intercept")
	if err != nil {
		return result, err
	}

	// 如果是捕获数据模式，保存到变量
	if action == "capture" && saveToVariable != "" && result.Success {
		if capturedData, ok := result.Data["capturedData"]; ok {
			// 转换数据格式
			if dataFormat == "" {
				dataFormat = "text"
			}
			converted, err := convertDataFormat(capturedData, dataFormat)
			if err != nil {
				log.Printf("[Workflow] 数据格式转换失败: %v", err)
				we.variables[saveToVariable] = capturedData
			} else {
				we.variables[saveToVariable] = converted
				log.Printf("[Workflow] ✓ 拦截数据已转换为 %s 格式并保存到变量: %s", dataFormat, saveToVariable)
			}
		}
	}

	return result, nil
}

// convertDataFormat 转换数据格式
func convertDataFormat(data interface{}, format string) (string, error) {
	var dataStr string
	switch v := data.(type) {
	case string:
		dataStr = v
	case []byte:
		dataStr = string(v)
	default:
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("数据序列化失败: %w", err)
		}
		dataStr = string(jsonBytes)
	}

	switch format {
	case "hex":
		return hex.EncodeToString([]byte(dataStr)), nil
	case "json":
		var jsonData interface{}
		if err := json.Unmarshal([]byte(dataStr), &jsonData); err != nil {
			return "", fmt.Errorf("JSON解析失败: %w", err)
		}
		formatted, err := json.Marshal(jsonData)
		if err != nil {
			return "", fmt.Errorf("JSON格式化失败: %w", err)
		}
		return string(formatted), nil
	case "text":
		fallthrough
	default:
		return dataStr, nil
	}
}
