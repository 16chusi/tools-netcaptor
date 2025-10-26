package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// JSONLReader JSONL文件读取器
type JSONLReader struct {
	filePath string
	lines    []map[string]interface{}
}

// NewJSONLReader 创建JSONL读取器
func NewJSONLReader(filePath string) *JSONLReader {
	return &JSONLReader{
		filePath: filePath,
		lines:    make([]map[string]interface{}, 0),
	}
}

// Load 加载JSONL文件
func (r *JSONLReader) Load() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	r.lines = make([]map[string]interface{}, 0)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	var buffer strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		buffer.WriteString(line)

		// 尝试解析
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(buffer.String()), &data); err == nil {
			r.lines = append(r.lines, data)
			buffer.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	return nil
}

// GetKeys 获取所有可用的key
func (r *JSONLReader) GetKeys() []string {
	if len(r.lines) == 0 {
		return []string{}
	}

	keysMap := make(map[string]bool)
	for _, line := range r.lines {
		for key := range line {
			keysMap[key] = true
		}
	}

	keys := make([]string, 0, len(keysMap))
	for key := range keysMap {
		keys = append(keys, key)
	}
	return keys
}

// GetLineCount 获取总行数
func (r *JSONLReader) GetLineCount() int {
	return len(r.lines)
}

// GetLine 获取指定行的数据
func (r *JSONLReader) GetLine(index int) (map[string]interface{}, error) {
	if index < 0 || index >= len(r.lines) {
		return nil, fmt.Errorf("索引越界: %d", index)
	}
	return r.lines[index], nil
}

// ExtractValue 提取指定key的值
func (r *JSONLReader) ExtractValue(data map[string]interface{}, keys []string) map[string]interface{} {
	result := make(map[string]interface{})

	// 如果keys包含"*"，返回全部数据
	for _, key := range keys {
		if key == "*" {
			return data
		}
	}

	// 提取指定的keys
	for _, key := range keys {
		if val, ok := data[key]; ok {
			result[key] = val
		}
	}

	return result
}
