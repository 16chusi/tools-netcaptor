package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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
	overwriteMode, _ := step.Params["overwriteMode"].(string)
	if overwriteMode == "" {
		overwriteMode = "skip" // 默认跳过
	}

	// 等待时间（默认 3 秒）
	waitTime := 3000
	if wt, ok := step.Params["waitTime"].(float64); ok && wt > 0 {
		waitTime = int(wt)
	}

	log.Printf("[Workflow] 从已捕获请求下载: urlPattern=%s, saveDirectory=%s, fileExtension=%s, waitTime=%dms", urlPattern, saveDirectory, fileExtension, waitTime)

	// 等待页面加载
	if waitTime > 0 {
		log.Printf("[Workflow] 等待 %d ms 让页面加载...", waitTime)
		time.Sleep(time.Duration(waitTime) * time.Millisecond)
	}

	// 获取所有已捕获的条目
	entries := we.app.capture.GetEntries()
	if len(entries) == 0 {
		return ExecutionResult{Success: false}, fmt.Errorf("没有已捕获的请求")
	}

	log.Printf("[Workflow] 共有 %d 个已捕获的请求", len(entries))
	for i, entry := range entries {
		log.Printf("[Workflow]   [%d] Path=%s", i+1, entry.Path)
	}

	// 转换通配符模式为正则表达式
	regexPattern := regexp.QuoteMeta(urlPattern)
	regexPattern = strings.ReplaceAll(regexPattern, "\\*", ".*")
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("URL 模式无效: %w", err)
	}

	// 查找匹配的请求
	var matchedEntries []NetworkEntry
	for _, entry := range entries {
		// 匹配完整 URL 或路径部分
		urlMatch := regex.MatchString(entry.URL)
		pathMatch := regex.MatchString(entry.Path)
		if urlMatch || pathMatch {
			matchedEntries = append(matchedEntries, entry)
			log.Printf("[Workflow] ✓ 匹配到: %s", entry.Path)
		}
	}

	if len(matchedEntries) == 0 {
		log.Printf("[Workflow] ⚠ 没有找到匹配 %s 的请求，跳过", urlPattern)
		return ExecutionResult{
			Success: true,
			Message: "没有匹配的请求，已跳过",
		}, nil
	}

	log.Printf("[Workflow] 找到 %d 个匹配的请求", len(matchedEntries))

	// 下载每个匹配的响应
	downloadedCount := 0
	skippedCount := 0
	for _, entry := range matchedEntries {
		if entry.Response.Body == "" {
			log.Printf("[Workflow] 跳过无响应体的请求: %s", entry.URL)
			continue
		}

		// 生成文件名（使用 MD5）
		filename := generateFilenameByMD5(entry.Response.Body, fileExtension)
		savePath := filepath.Join(saveDirectory, filename)

		// 检查文件是否存在
		if _, err := os.Stat(savePath); err == nil {
			if overwriteMode == "skip" {
				log.Printf("[Workflow] ⊘ 文件已存在，跳过: %s", filename)
				skippedCount++
				continue
			}
			log.Printf("[Workflow] ⚠ 文件已存在，覆盖: %s", filename)
		}

		// 写入文件
		err := os.WriteFile(savePath, []byte(entry.Response.Body), 0644)
		if err != nil {
			log.Printf("[Workflow] 保存文件失败: %v", err)
			continue
		}

		log.Printf("[Workflow] ✓ 已保存: %s", filename)
		downloadedCount++
	}

	if downloadedCount == 0 && skippedCount == 0 {
		return ExecutionResult{Success: false}, fmt.Errorf("没有文件被保存")
	}

	msg := fmt.Sprintf("已保存 %d 个响应", downloadedCount)
	if skippedCount > 0 {
		msg += fmt.Sprintf("，跳过 %d 个重复文件", skippedCount)
	}

	return ExecutionResult{
		Success: true,
		Message: msg,
	}, nil
}

// generateFilenameByMD5 使用 MD5 生成文件名
func generateFilenameByMD5(content string, extension string) string {
	// 计算 MD5
	hash := md5.Sum([]byte(content))
	md5Str := hex.EncodeToString(hash[:])

	// 如果没有指定扩展名，默认为 txt
	if extension == "" {
		extension = "txt"
	}

	// 确保扩展名不带点
	extension = strings.TrimPrefix(extension, ".")

	return fmt.Sprintf("%s.%s", md5Str, extension)
}
