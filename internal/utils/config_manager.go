package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type ConfigManager struct {
	db *sql.DB
	mu sync.RWMutex
}

var (
	configManager     *ConfigManager
	configManagerOnce sync.Once
)

// GetConfigManager 获取配置管理器单例
func GetConfigManager() *ConfigManager {
	configManagerOnce.Do(func() {
		homeDir, _ := os.UserHomeDir()
		if homeDir == "" {
			homeDir = "."
		}
		configDir := filepath.Join(homeDir, ".netcaptor")
		os.MkdirAll(configDir, 0755)

		dbPath := filepath.Join(configDir, "workflow.db")
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			panic(fmt.Sprintf("无法打开数据库: %v", err))
		}

		configManager = &ConfigManager{db: db}
		configManager.initTables()
		configManager.migrateOldConfigs()
	})
	return configManager
}

// initTables 初始化配置表
func (cm *ConfigManager) initTables() {
	// 创建通用配置表
	_, err := cm.db.Exec(`
		CREATE TABLE IF NOT EXISTS app_config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		AppLog.Error(fmt.Sprintf("创建配置表失败: %v", err))
	}
}

// migrateOldConfigs 迁移旧配置文件到SQLite
func (cm *ConfigManager) migrateOldConfigs() {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		return
	}
	configDir := filepath.Join(homeDir, ".netcaptor")

	// 迁移 proxy_config.json
	proxyConfigPath := filepath.Join(configDir, "proxy_config.json")
	if data, err := os.ReadFile(proxyConfigPath); err == nil {
		cm.SetConfig("proxy_config", string(data))
		AppLog.Info("已迁移 proxy_config.json 到数据库")
	}

	// 迁移 smart_proxy_rules.json
	smartProxyPath := filepath.Join(configDir, "smart_proxy_rules.json")
	if data, err := os.ReadFile(smartProxyPath); err == nil {
		cm.SetConfig("smart_proxy_rules", string(data))
		AppLog.Info("已迁移 smart_proxy_rules.json 到数据库")
	}

	// 迁移 ai_models 表到 KV 格式
	cm.migrateAIModelsTable()
}

// SetConfig 保存配置
func (cm *ConfigManager) SetConfig(key string, value string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	_, err := cm.db.Exec(`
		INSERT OR REPLACE INTO app_config (key, value, updated_at)
		VALUES (?, ?, datetime('now'))
	`, key, value)

	return err
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig(key string) (string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var value string
	err := cm.db.QueryRow("SELECT value FROM app_config WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetConfigJSON 保存JSON配置
func (cm *ConfigManager) SetConfigJSON(key string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return cm.SetConfig(key, string(jsonData))
}

// GetConfigJSON 获取JSON配置
func (cm *ConfigManager) GetConfigJSON(key string, target interface{}) error {
	value, err := cm.GetConfig(key)
	if err != nil || value == "" {
		return err
	}
	return json.Unmarshal([]byte(value), target)
}

// migrateAIModelsTable 迁移 ai_models 表到 KV 格式
func (cm *ConfigManager) migrateAIModelsTable() {
	// 检查是否已经迁移过
	if value, _ := cm.GetConfig("ai_models"); value != "" {
		return // 已经有KV格式的数据，跳过迁移
	}

	// 从旧表读取数据
	rows, err := cm.db.Query("SELECT provider, name, api_key, base_url FROM ai_models ORDER BY id")
	if err != nil {
		return // 表可能不存在或为空
	}
	defer rows.Close()

	type AIModel struct {
		Provider string `json:"provider"`
		Name     string `json:"name"`
		APIKey   string `json:"apiKey"`
		BaseURL  string `json:"baseUrl"`
	}

	var models []AIModel
	for rows.Next() {
		var model AIModel
		if err := rows.Scan(&model.Provider, &model.Name, &model.APIKey, &model.BaseURL); err == nil {
			models = append(models, model)
		}
	}

	if len(models) > 0 {
		cm.SetConfigJSON("ai_models", models)
		AppLog.Info(fmt.Sprintf("已迁移 %d 个AI模型配置到KV格式", len(models)))
	}
}
