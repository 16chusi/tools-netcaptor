# NetCaptor 日志系统使用指南

## 概述
NetCaptor 使用 Wails 自带的 FileLogger，将日志记录到 `netcaptor.log` 文件中。

## 日志级别
- **TRACE**: 最详细的跟踪信息
- **DEBUG**: 调试信息  
- **INFO**: 一般信息
- **WARNING**: 警告信息
- **ERROR**: 错误信息
- **FATAL**: 致命错误，会导致程序退出

## 使用方法

### 基本用法
```go
// 使用全局日志实例
AppLog.Info("应用启动成功")
AppLog.Debug("调试信息")
AppLog.Warning("这是一个警告")
AppLog.Error("发生错误")
```

### 字符串拼接
由于 Wails logger 不支持格式化，需要手动拼接：
```go
AppLog.Info("用户登录: " + username)
AppLog.Debug("处理请求: " + method + " " + url)
AppLog.Error("连接失败: " + err.Error())
```

### 替换现有日志
```go
// 旧的方式
fmt.Printf("[Module] 信息: %s\n", message)
log.Printf("错误: %v", err)

// 新的方式
AppLog.Info("Module 信息: " + message)
AppLog.Error("错误: " + err.Error())
```

## 日志输出
- **文件**: 保存到 `netcaptor.log` 文件中，格式如：
  ```
  INFO  | 应用启动成功
  DEBUG | 收到请求: GET https://example.com
  ERROR | 连接失败: connection refused
  ```

## 优势
- **简单**: 直接使用 Wails 官方实现，无需维护自定义代码
- **可靠**: 经过 Wails 团队测试和维护
- **标准**: 符合 Wails logger 接口规范

## 迁移指南
逐步替换现有的日志调用：

1. **fmt.Printf** → **AppLog.Info**
2. **log.Printf** → **AppLog.Info** 或 **AppLog.Error**  
3. **fmt.Println** → **AppLog.Info**

注意：需要手动进行字符串拼接，不支持 `Printf` 风格的格式化。
