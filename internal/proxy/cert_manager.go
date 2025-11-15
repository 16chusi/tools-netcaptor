package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"netcaptor/internal/types"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CertManager struct {
	caCert    *x509.Certificate
	caKey     *rsa.PrivateKey
	certCache map[string]*tls.Certificate
	mu        sync.RWMutex
	certDir   string
}

func NewCertManager() (*CertManager, error) {
	homeDir, _ := os.UserHomeDir()
	certDir := filepath.Join(homeDir, ".netcaptor", "certs")
	os.MkdirAll(certDir, 0755)

	cm := &CertManager{
		certCache: make(map[string]*tls.Certificate),
		certDir:   certDir,
	}

	// 加载或生成CA证书
	if err := cm.loadOrGenerateCA(); err != nil {
		return nil, err
	}

	return cm, nil
}

func (cm *CertManager) loadOrGenerateCA() error {
	caFile := filepath.Join(cm.certDir, "netcaptor-ca.crt")
	keyFile := filepath.Join(cm.certDir, "netcaptor-ca.key")

	// 尝试加载现有证书
	if _, err := os.Stat(caFile); err == nil {
		certPEM, _ := os.ReadFile(caFile)
		keyPEM, _ := os.ReadFile(keyFile)

		certBlock, _ := pem.Decode(certPEM)
		keyBlock, _ := pem.Decode(keyPEM)

		if certBlock != nil && keyBlock != nil {
			cert, err := x509.ParseCertificate(certBlock.Bytes)
			if err == nil {
				key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
				if err == nil {
					cm.caCert = cert
					cm.caKey = key
					return nil
				}
			}
		}
	}

	// 生成新的CA证书
	return cm.generateCA()
}

func (cm *CertManager) generateCA() error {
	// 生成私钥
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"NetCaptor Proxy"},
			CommonName:   "NetCaptor Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// 自签名
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return err
	}

	cm.caCert = cert
	cm.caKey = key

	// 保存证书
	caFile := filepath.Join(cm.certDir, "netcaptor-ca.crt")
	keyFile := filepath.Join(cm.certDir, "netcaptor-ca.key")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	os.WriteFile(caFile, certPEM, 0644)
	os.WriteFile(keyFile, keyPEM, 0600)

	return nil
}

func (cm *CertManager) GetCACert() *tls.Certificate {
	// 返回用于GoProxy的CA证书
	cert := &tls.Certificate{
		Certificate: [][]byte{cm.caCert.Raw},
		PrivateKey:  cm.caKey,
	}
	return cert
}

func (cm *CertManager) GetCertForHost(host string) (*tls.Certificate, error) {
	cm.mu.RLock()
	if cert, ok := cm.certCache[host]; ok {
		cm.mu.RUnlock()
		return cert, nil
	}
	cm.mu.RUnlock()

	// 生成新证书
	cert, err := cm.generateCertForHost(host)
	if err != nil {
		return nil, err
	}

	cm.mu.Lock()
	cm.certCache[host] = cert
	cm.mu.Unlock()

	return cert, nil
}

func (cm *CertManager) generateCertForHost(host string) (*tls.Certificate, error) {
	// 生成私钥
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// 创建证书模板
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"NetCaptor Proxy"},
			CommonName:   host,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{host},
	}

	// 使用CA签名
	certDER, err := x509.CreateCertificate(rand.Reader, &template, cm.caCert, &key.PublicKey, cm.caKey)
	if err != nil {
		return nil, err
	}

	cert := &tls.Certificate{
		Certificate: [][]byte{certDER, cm.caCert.Raw},
		PrivateKey:  key,
	}

	return cert, nil
}

func (cm *CertManager) GetCACertPath() string {
	return filepath.Join(cm.certDir, "netcaptor-ca.crt")
}

func (cm *CertManager) GetCACertPEM() ([]byte, error) {
	return os.ReadFile(cm.GetCACertPath())
}

func (cm *CertManager) GetCACertInfo() *types.CertInfo {
	certPath := cm.GetCACertPath()
	info := &types.CertInfo{
		Path: certPath,
	}

	// 检查文件是否存在
	if stat, err := os.Stat(certPath); err == nil {
		info.Exists = true
		info.CreatedAt = stat.ModTime().Format("2006-01-02 15:04:05")

		// 读取证书详细信息
		if cm.caCert != nil {
			info.NotBefore = cm.caCert.NotBefore.Format("2006-01-02 15:04:05")
			info.NotAfter = cm.caCert.NotAfter.Format("2006-01-02 15:04:05")
			info.Subject = cm.caCert.Subject.String()
			info.Issuer = cm.caCert.Issuer.String()
		}
	}

	return info
}
