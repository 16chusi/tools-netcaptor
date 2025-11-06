package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// 全局计数器
var globalCounter int = 0

// replaceVariables 替换步骤参数中的变量
func (we *WorkflowExecutor) replaceVariables(step *ExecutionStep) {
	for key, value := range step.Params {
		if strVal, ok := value.(string); ok {
			step.Params[key] = we.replaceVariablesInString(strVal)
		}
		// 数字类型的变量替换在各个具体函数中处理
	}
}

// replaceVariablesInString 替换字符串中的变量
func (we *WorkflowExecutor) replaceVariablesInString(str string) string {
	original := str
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

		// 处理内置变量
		if builtinValue := we.resolveBuiltinVariable(varPath); builtinValue != "" {
			str = strings.Replace(str, placeholder, builtinValue, 1)
			continue
		}

		// 处理用户定义变量
		value := we.resolveVariablePath(varPath)
		if value != nil {
			str = strings.Replace(str, placeholder, fmt.Sprintf("%v", value), 1)
		} else {
			// 如果变量不存在，保留原样
			break
		}
	}
	if str != original {
		log.Printf("[Workflow] 变量替换: %s -> %s", original, str)
	}
	return str
}

// resolveBuiltinVariable 解析内置变量
func (we *WorkflowExecutor) resolveBuiltinVariable(varName string) string {
	now := time.Now()

	switch varName {
	case "timestamp":
		return now.Format("20060102_150405")
	case "date":
		return now.Format("20060102")
	case "time":
		return now.Format("150405")
	case "uuid":
		return uuid.New().String()
	case "uuid_short":
		return uuid.New().String()[:8]
	case "random":
		return strconv.Itoa(rand.Intn(10000))
	case "random_6":
		return fmt.Sprintf("%06d", rand.Intn(1000000))
	case "counter":
		globalCounter++
		return strconv.Itoa(globalCounter)
	case "title":
		// 页面标题需要通过WebSocket从浏览器获取
		return we.getPageTitle()
	case "url":
		// 当前页面URL需要通过WebSocket从浏览器获取
		return we.getCurrentURL()
	}

	return ""
}

// getPageTitle 获取页面标题
func (we *WorkflowExecutor) getPageTitle() string {
	if !we.wsServer.IsRunning() || !we.wsServer.HasClients() {
		return "unknown_title"
	}

	msg := WSMessage{
		Type: "get_page_info",
		Data: map[string]interface{}{
			"info": "title",
		},
	}

	result, err := we.sendAndWait(msg, 15*time.Second, "get_page_info")
	if err != nil || !result.Success {
		return "unknown_title"
	}

	if title, ok := result.Data["title"].(string); ok {
		// 清理标题中的特殊字符
		title = strings.ReplaceAll(title, "/", "_")
		title = strings.ReplaceAll(title, "\\", "_")
		title = strings.ReplaceAll(title, ":", "_")
		title = strings.ReplaceAll(title, "*", "_")
		title = strings.ReplaceAll(title, "?", "_")
		title = strings.ReplaceAll(title, "\"", "_")
		title = strings.ReplaceAll(title, "<", "_")
		title = strings.ReplaceAll(title, ">", "_")
		title = strings.ReplaceAll(title, "|", "_")
		return title
	}

	return "unknown_title"
}

// getCurrentURL 获取当前页面URL
func (we *WorkflowExecutor) getCurrentURL() string {
	if !we.wsServer.IsRunning() || !we.wsServer.HasClients() {
		return "unknown_url"
	}

	msg := WSMessage{
		Type: "get_page_info",
		Data: map[string]interface{}{
			"info": "url",
		},
	}

	result, err := we.sendAndWait(msg, 15*time.Second, "get_page_info")
	if err != nil || !result.Success {
		return "unknown_url"
	}

	if url, ok := result.Data["url"].(string); ok {
		return url
	}

	return "unknown_url"
}

// resolveVariablePath 解析变量路径 支持 data.url 和 data[url]
func (we *WorkflowExecutor) resolveVariablePath(path string) interface{} {
	log.Printf("[Workflow] resolveVariablePath: %s", path)

	if strings.Contains(path, ".") {
		parts := strings.SplitN(path, ".", 2)
		log.Printf("[Workflow] 解析嵌套路径: 变量=%s, 字段=%s", parts[0], parts[1])

		if val, ok := we.variables[parts[0]]; ok {
			log.Printf("[Workflow] 变量 %s 存在，类型: %T, 值: %v", parts[0], val, val)

			// 如果是 map，直接访问
			if mapVal, isMap := val.(map[string]interface{}); isMap {
				log.Printf("[Workflow] 变量是 map，访问字段: %s", parts[1])
				if result, exists := mapVal[parts[1]]; exists {
					log.Printf("[Workflow] ✓ 找到字段 %s: %v", parts[1], result)
					return result
				}
				log.Printf("[Workflow] ✗ map 中不存在字段: %s", parts[1])
			}

			// 如果是 JSON 字符串，尝试解析
			if strVal, isStr := val.(string); isStr {
				log.Printf("[Workflow] 变量是字符串，尝试解析 JSON")
				var jsonData map[string]interface{}
				if err := json.Unmarshal([]byte(strVal), &jsonData); err == nil {
					log.Printf("[Workflow] JSON 解析成功，访问字段: %s", parts[1])
					if result, exists := jsonData[parts[1]]; exists {
						log.Printf("[Workflow] ✓ 找到字段 %s: %v", parts[1], result)
						return result
					}
					log.Printf("[Workflow] ✗ JSON 中不存在字段: %s", parts[1])
				} else {
					log.Printf("[Workflow] JSON 解析失败: %v", err)
				}
			}
		} else {
			log.Printf("[Workflow] ✗ 变量 %s 不存在", parts[0])
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
		log.Printf("[Workflow] 直接访问变量: %s", path)
		if val, ok := we.variables[path]; ok {
			log.Printf("[Workflow] ✓ 变量存在，类型: %T", val)
			return val
		}
		log.Printf("[Workflow] ✗ 变量不存在: %s", path)
	}
	return nil
}

// setVariable 设置变量
func (we *WorkflowExecutor) setVariable(name string, value interface{}) {
	we.variables[name] = value
	log.Printf("[Workflow] 设置变量 %s = %v", name, value)
}

// getVariable 获取变量
func (we *WorkflowExecutor) getVariable(name string) interface{} {
	if value, exists := we.variables[name]; exists {
		return value
	}
	return nil
}
