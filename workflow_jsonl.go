package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// executeJSONLReaderLoop 执行JSONL读取器的循环逻辑（已废弃，保留兼容）
func (we *WorkflowExecutor) executeJSONLReaderLoop(task WorkflowTask, steps []ExecutionStep, jsonlStepIndex int) error {
	return fmt.Errorf("JSONL读取器暂不支持新的执行模式")

}
