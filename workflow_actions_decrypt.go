package main

import (
	"encoding/hex"
	"fmt"

	"github.com/tjfoc/gmsm/sm4"
)

// executeDecrypt 执行解密
func (we *WorkflowExecutor) executeDecrypt(step ExecutionStep) (ExecutionResult, error) {
	dataVariable, ok := step.Params["dataVariable"].(string)
	if !ok || dataVariable == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少数据来源变量")
	}

	algorithm, ok := step.Params["algorithm"].(string)
	if !ok || algorithm == "" {
		algorithm = "sm4-ecb"
	}

	key, ok := step.Params["key"].(string)
	if !ok || key == "" {
		return ExecutionResult{Success: false}, fmt.Errorf("缺少密钥")
	}

	data := we.resolveVariablePath(dataVariable)
	if data == nil {
		return ExecutionResult{Success: false}, fmt.Errorf("变量 %s 不存在", dataVariable)
	}

	dataStr, ok := data.(string)
	if !ok {
		return ExecutionResult{Success: false}, fmt.Errorf("变量 %s 不是字符串类型，实际类型: %T", dataVariable, data)
	}

	var decrypted string
	var err error

	switch algorithm {
	case "sm4-ecb":
		decrypted, err = sm4DecryptECB(dataStr, key)
	case "sm4-cbc":
		iv := ""
		if ivParam, ok := step.Params["iv"].(string); ok {
			iv = ivParam
		}
		decrypted, err = sm4DecryptCBC(dataStr, key, iv)
	default:
		return ExecutionResult{Success: false}, fmt.Errorf("不支持的算法: %s", algorithm)
	}

	if err != nil {
		return ExecutionResult{Success: false}, fmt.Errorf("解密失败: %w", err)
	}

	if saveToVariable, ok := step.Params["saveToVariable"].(string); ok && saveToVariable != "" {
		we.variables[saveToVariable] = decrypted
		AppLog.Info(fmt.Sprintf("[Workflow] ✓ 解密完成，保存到变量: %s", saveToVariable))
	}

	return ExecutionResult{
		Success: true,
		Message: fmt.Sprintf("解密成功，长度: %d", len(decrypted)),
		Data: map[string]interface{}{
			"decrypted": decrypted,
		},
	}, nil
}

// sm4DecryptECB SM4-ECB 解密
func sm4DecryptECB(hexData, hexKey string) (string, error) {
	ciphertext, err := hex.DecodeString(hexData)
	if err != nil {
		return "", fmt.Errorf("解码数据失败: %w", err)
	}

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", fmt.Errorf("解码密钥失败: %w", err)
	}

	decrypted, err := sm4.Sm4Ecb(key, ciphertext, false)
	if err != nil {
		return "", fmt.Errorf("SM4解密失败: %w", err)
	}

	// 移除 PKCS7 填充
	padding := int(decrypted[len(decrypted)-1])
	if padding > 16 || padding == 0 {
		return string(decrypted), nil
	}
	return string(decrypted[:len(decrypted)-padding]), nil
}

// sm4DecryptCBC SM4-CBC 解密
func sm4DecryptCBC(hexData, hexKey, hexIV string) (string, error) {
	ciphertext, err := hex.DecodeString(hexData)
	if err != nil {
		return "", fmt.Errorf("解码数据失败: %w", err)
	}

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", fmt.Errorf("解码密钥失败: %w", err)
	}

	var iv []byte
	if hexIV != "" {
		iv, err = hex.DecodeString(hexIV)
		if err != nil {
			return "", fmt.Errorf("解码IV失败: %w", err)
		}
	} else {
		iv = make([]byte, 16)
	}

	// 合并 IV 和密文
	data := append(iv, ciphertext...)
	decrypted, err := sm4.Sm4Cbc(key, data, false)
	if err != nil {
		return "", fmt.Errorf("SM4解密失败: %w", err)
	}

	// 移除 PKCS7 填充
	padding := int(decrypted[len(decrypted)-1])
	if padding > 16 || padding == 0 {
		return string(decrypted), nil
	}
	return string(decrypted[:len(decrypted)-padding]), nil
}
