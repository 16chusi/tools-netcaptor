package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

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
