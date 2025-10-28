package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// executeDownloadCaptured 从已捕获的请求中下载响应
func (we *WorkflowExecutor) executeDownloadCaptured(step ExecutionStep) (ExecutionResult, error) {
	urlPattern, _ := step.Params["urlPattern"].(string)
	if urlPattern == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少 URL 匹配模式")
	}

	saveDirectory, _ := step.Params["saveDirectory"].(string)
	if saveDirectory == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("请选择保存目录")
	}

	fileExtension, _ := step.Params["fileExtension"].(string)

	log.Printf("[Workflow] 从已捕获请求下载: urlPattern=%s, saveDirectory=%s, fileExtension=%s", urlPattern, saveDirectory, fileExtension)

	// 获取所有已捕获的条目
	entries := we.app.capture.GetEntries()
	if len(entries) == 0 {
		return ExecutionResult{Success: false}, fmt.Errorf("没有已捕获的请求")
	}

	log.Printf("[Workflow] 共有 %d 个已捕获的请求", len(entries))

	// 转换通配符模式为正则表达式
	regexPattern := "^" + regexp.QuoteMeta(urlPattern) + "$"
	regexPattern = strings.ReplaceAll(regexPattern, "\\*", ".*")
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("URL 模式无效: %w", err)
	}

	// 查找匹配的请求
	var matchedEntries []NetworkEntry
	for _, entry := range entries {
		if regex.MatchString(entry.URL) {
			matchedEntries = append(matchedEntries, entry)
		}
	}

	if len(matchedEntries) == 0 {
		return ExecutionResult{Success: false}, fmt.Errorf("没有找到匹配 %s 的请求", urlPattern)
	}

	log.Printf("[Workflow] 找到 %d 个匹配的请求", len(matchedEntries))

	// 下载每个匹配的响应
	downloadedCount := 0
	for _, entry := range matchedEntries {
		if entry.Response.Body == "" {
			log.Printf("[Workflow] 跳过无响应体的请求: %s", entry.URL)
			continue
		}

		// 生成文件名
		filename := generateFilename(entry.URL, fileExtension)
		savePath := filepath.Join(saveDirectory, filename)

		// 写入文件
		err := os.WriteFile(savePath, []byte(entry.Response.Body), 0644)
		if err != nil {
			log.Printf("[Workflow] 保存文件失败: %v", err)
			continue
		}

		log.Printf("[Workflow] ✓ 已保存: %s", filename)
		downloadedCount++
	}

	if downloadedCount == 0 {
		return ExecutionResult{Success: false}, fmt.Errorf("没有文件被保存")
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("已保存 %d 个响应", downloadedCount),
	}, nil
}

// generateFilename 生成文件名
func generateFilename(url string, extension string) string {
	// 使用 UUID 生成唯一文件名
	uniqueID := uuid.New().String()[:8]

	// 如果没有指定扩展名，尝试从 URL 推断
	if extension == "" {
		if strings.Contains(url, ".json") || strings.Contains(url, "/api/") {
			extension = "json"
		} else if strings.Contains(url, ".html") {
			extension = "html"
		} else if strings.Contains(url, ".xml") {
			extension = "xml"
		} else {
			extension = "txt"
		}
	}

	// 确保扩展名不带点
	extension = strings.TrimPrefix(extension, ".")

	return fmt.Sprintf("response_%s.%s", uniqueID, extension)
}
