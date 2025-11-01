package main

import (
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
	if strings.Contains(path, ".") {
		parts := strings.SplitN(path, ".", 2)
		if val, ok := we.variables[parts[0]]; ok {
			if mapVal, isMap := val.(map[string]interface{}); isMap {
				return mapVal[parts[1]]
			}
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
		return we.variables[path]
	}
	return nil
}
