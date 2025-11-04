# 循环组件开发快速检查清单

## 🚀 开发新循环组件时的必备检查项

### 1. 组件分类确认
```
□ 确认组件类型：
  □ 纯后端逻辑组件（JSONL、For、If、Wait、Download）
  □ 需要浏览器插件组件（Click、Input、Extract、Navigate、Scroll）

□ 如果是浏览器插件组件：
  □ 必须添加循环间等待机制
  □ 必须添加详细日志标识
```

### 2. 循环间等待机制
```go
// 必须添加的代码模式
for i := 0; i < count; i++ {
    if i > 0 { // 第一次不需要等待
        log.Printf("[组件名] 等待页面稳定...")
        time.Sleep(3 * time.Second) // 关键：3秒等待
    }
    // 执行具体操作
}
```

### 3. 日志标识规范
```go
// 必须使用的日志格式
log.Printf("[组件名] 开始执行")
log.Printf("[组件名] 执行参数: param1=%s, param2=%s", param1, param2)
log.Printf("[组件名] ✅ 执行成功")
log.Printf("[组件名] ❌ 执行失败: %v", err)
```

### 4. 错误处理模式
```go
// 标准错误处理
result, err := executeOperation()
if err != nil {
    log.Printf("[组件名] 执行失败: %v", err)
    return fmt.Errorf("操作失败: %w", err)
} else {
    log.Printf("[组件名] 执行成功")
}
```

### 5. 测试验证步骤
```
□ 单独测试组件功能
□ 测试单层循环（组件 → 简单节点 → 结束）
□ 测试多层嵌套（JSONL → 组件 → 其他组件 → 结束）
□ 验证数据一致性
□ 检查日志输出清晰度
```

## ⚠️ 常见错误及解决方案

### 错误1: "Could not establish connection"
```
原因: 循环执行太快，页面状态不稳定
解决: 增加 time.Sleep(3 * time.Second)
```

### 错误2: "元素未找到"
```
原因: CSS选择器配置错误或页面未加载完成
解决: 检查选择器 + 确保页面稳定
```

### 错误3: 日志难以调试
```
原因: 缺乏详细的节点标识
解决: 使用 [组件名] 前缀 + 详细参数日志
```

## 🎯 成功标准

### 功能正常的标志
```
✅ 单层循环正常工作
✅ 多层嵌套循环正常工作
✅ 日志清晰易读，能快速定位问题
✅ 数据采集准确，无错乱
✅ 执行稳定，重复测试无异常
```

### 代码质量标准
```go
// ✅ 好的循环组件代码示例
func (we *WorkflowExecutor) executeMyLoop() error {
    log.Printf("[我的循环组件] 开始执行，共 %d 次", count)
    
    for i := 0; i < count; i++ {
        if i > 0 {
            log.Printf("[我的循环组件] 等待页面稳定...")
            time.Sleep(3 * time.Second)
        }
        
        log.Printf("[我的循环组件] 执行第 %d/%d 次", i+1, count)
        
        if err := executeStep(); err != nil {
            log.Printf("[我的循环组件] ❌ 第 %d 次执行失败: %v", i+1, err)
            return fmt.Errorf("执行第 %d 次失败: %w", i+1, err)
        }
        
        log.Printf("[我的循环组件] ✅ 第 %d 次执行成功", i+1)
    }
    
    log.Printf("[我的循环组件] 所有循环执行完成")
    return nil
}
```

---

**记住：稳定性 > 速度，清晰的日志 > 简洁的代码**
