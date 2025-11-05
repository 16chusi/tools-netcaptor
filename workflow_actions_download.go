package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
