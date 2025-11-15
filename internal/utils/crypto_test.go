package utils

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/tjfoc/gmsm/sm4"
)

// SM4DecryptECB 对应 sm4.decrypt(data, key, {mode: "ecb"})
func SM4DecryptECB(hexData, hexKey string) (string, error) {
	ciphertext, err := hex.DecodeString(hexData)
	if err != nil {
		return "", err
	}

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", err
	}

	decrypted, err := sm4.Sm4Ecb(key, ciphertext, false)
	if err != nil {
		return "", err
	}

	// 移除 PKCS7 填充
	padding := int(decrypted[len(decrypted)-1])
	if padding > 16 || padding == 0 {
		return string(decrypted), nil
	}
	return string(decrypted[:len(decrypted)-padding]), nil
}

func TestAESDecryptECB(t *testing.T) {
	data, err := os.ReadFile("encrypted.txt")
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	encrypted := strings.TrimSpace(string(data))
	key := "46696e32416e63304571753245727934"

	t.Logf("加密数据长度: %d", len(encrypted))
	t.Logf("密钥字符串: %s", key)
	t.Logf("前32字符: %s", encrypted[:32])

	// 测试前16字节
	resultFirst, _ := SM4DecryptECB(encrypted, key)
	t.Logf("前16字节解密: %q", resultFirst)
	t.Logf("前16字节hex: % x", []byte(resultFirst))

	result, err := SM4DecryptECB(encrypted, key)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	t.Logf("解密结果前100字符: %q", result[:min(100, len(result))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func mustHexDecode(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
