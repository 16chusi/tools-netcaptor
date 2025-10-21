package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DownloadTask struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Error    string `json:"error,omitempty"`
}

type Downloader struct {
	tasks      []DownloadTask
	savePath   string
	maxRetries int
}

func NewDownloader(savePath string) *Downloader {
	return &Downloader{
		tasks:      make([]DownloadTask, 0),
		savePath:   savePath,
		maxRetries: 3,
	}
}

func (d *Downloader) AddTask(url string) {
	filename := d.extractFilename(url)
	task := DownloadTask{
		URL:      url,
		Filename: filename,
		Status:   "pending",
		Progress: 0,
	}
	d.tasks = append(d.tasks, task)
}

func (d *Downloader) extractFilename(url string) string {
	// 从URL中提取文件名
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	// 如果没有扩展名，添加默认扩展名
	if !strings.Contains(filename, ".") {
		filename += ".html"
	}

	// 清理文件名中的非法字符
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "&", "_")
	filename = strings.ReplaceAll(filename, "=", "_")

	return filename
}

func (d *Downloader) DownloadAll() error {
	// 确保保存目录存在
	err := os.MkdirAll(d.savePath, 0755)
	if err != nil {
		return fmt.Errorf("创建保存目录失败: %v", err)
	}

	for i := range d.tasks {
		task := &d.tasks[i]
		task.Status = "downloading"

		err := d.downloadFile(task)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			continue
		}

		task.Status = "completed"
		task.Progress = 100
	}

	return nil
}

func (d *Downloader) downloadFile(task *DownloadTask) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}

	var lastErr error
	for retry := 0; retry < d.maxRetries; retry++ {
		if retry > 0 {
			time.Sleep(time.Duration(retry) * time.Second)
		}

		req, err := http.NewRequest("GET", task.URL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		// 设置请求头模拟真实浏览器
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Header.Set("Connection", "keep-alive")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP错误: %d", resp.StatusCode)
			continue
		}

		// 创建文件
		filePath := filepath.Join(d.savePath, task.Filename)
		file, err := os.Create(filePath)
		if err != nil {
			resp.Body.Close()
			lastErr = err
			continue
		}

		// 下载文件
		_, err = io.Copy(file, resp.Body)
		file.Close()
		resp.Body.Close()

		if err != nil {
			os.Remove(filePath) // 删除不完整的文件
			lastErr = err
			continue
		}

		return nil // 下载成功
	}

	return fmt.Errorf("下载失败，重试%d次后仍然失败: %v", d.maxRetries, lastErr)
}

func (d *Downloader) GetTasks() []DownloadTask {
	return d.tasks
}

func (d *Downloader) ClearTasks() {
	d.tasks = make([]DownloadTask, 0)
}
