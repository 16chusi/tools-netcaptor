package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RouteRule 路由规则
type RouteRule struct {
	ID        string    `json:"id"`
	Pattern   string    `json:"pattern"`    // 域名模式: google.com, *.openai.com
	RouteType string    `json:"route_type"` // "direct", "proxy", "auto"
	Source    string    `json:"source"`     // "manual", "auto_learned"
	Enabled   bool      `json:"enabled"`    // 是否启用此规则
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

// SmartProxyManager 智能代理管理器
type SmartProxyManager struct {
	configPath     string
	rules          []*RouteRule
	proxyConfigMgr *ProxyConfigManager
}

// NewSmartProxyManager 创建智能代理管理器
func NewSmartProxyManager(proxyConfigMgr *ProxyConfigManager) *SmartProxyManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".netcaptor")
	os.MkdirAll(configDir, 0755)

	configPath := filepath.Join(configDir, "smart_proxy_rules.json")

	manager := &SmartProxyManager{
		configPath:     configPath,
		rules:          make([]*RouteRule, 0),
		proxyConfigMgr: proxyConfigMgr,
	}

	// 加载现有规则
	manager.LoadRules()

	return manager
}

// LoadRules 加载规则
func (sm *SmartProxyManager) LoadRules() error {
	if _, err := os.Stat(sm.configPath); os.IsNotExist(err) {
		// 添加一些默认规则
		sm.addDefaultRules()
		return sm.SaveRules()
	}

	data, err := ioutil.ReadFile(sm.configPath)
	if err != nil {
		return fmt.Errorf("读取规则文件失败: %w", err)
	}

	err = json.Unmarshal(data, &sm.rules)
	if err != nil {
		return fmt.Errorf("解析规则文件失败: %w", err)
	}

	log.Printf("[SmartProxy] 加载了 %d 条路由规则", len(sm.rules))
	return nil
}

// SaveRules 保存规则
func (sm *SmartProxyManager) SaveRules() error {
	data, err := json.MarshalIndent(sm.rules, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化规则失败: %w", err)
	}

	err = ioutil.WriteFile(sm.configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("保存规则文件失败: %w", err)
	}

	return nil
}

// addDefaultRules 添加默认规则
func (sm *SmartProxyManager) addDefaultRules() {
	defaultRules := []*RouteRule{
		{
			ID:        "default-localhost",
			Pattern:   "localhost",
			RouteType: "direct",
			Source:    "manual",
			Enabled:   true,
			CreatedAt: time.Now(),
		},
		{
			ID:        "default-127",
			Pattern:   "127.0.0.1",
			RouteType: "direct",
			Source:    "manual",
			Enabled:   true,
			CreatedAt: time.Now(),
		},
		{
			ID:        "default-local",
			Pattern:   "*.local",
			RouteType: "direct",
			Source:    "manual",
			Enabled:   true,
			CreatedAt: time.Now(),
		},
	}

	sm.rules = append(sm.rules, defaultRules...)
}

// DecideRoute 决定路由方式
func (sm *SmartProxyManager) DecideRoute(host string) string {
	// 查找匹配的规则
	rule := sm.findMatchingRule(host)

	if rule != nil && rule.Enabled {
		// 更新最后使用时间
		rule.LastUsed = time.Now()
		sm.SaveRules()

		log.Printf("[SmartProxy] %s 匹配规则: %s -> %s (%s)", host, rule.Pattern, rule.RouteType, rule.Source)
		return rule.RouteType
	}

	// 没有匹配规则，默认直连
	log.Printf("[SmartProxy] %s 无匹配规则，使用直连", host)
	return "direct"
}

// findMatchingRule 查找匹配的规则
func (sm *SmartProxyManager) findMatchingRule(host string) *RouteRule {
	var bestMatch *RouteRule

	for _, rule := range sm.rules {
		if !rule.Enabled {
			continue
		}

		if sm.matchPattern(rule.Pattern, host) {
			// 优先级：手动 > 自动，精确匹配 > 通配符
			if bestMatch == nil {
				bestMatch = rule
			} else {
				// 手动规则优先
				if rule.Source == "manual" && bestMatch.Source == "auto_learned" {
					bestMatch = rule
				} else if rule.Source == bestMatch.Source {
					// 同类型规则，精确匹配优先
					if !strings.Contains(rule.Pattern, "*") && strings.Contains(bestMatch.Pattern, "*") {
						bestMatch = rule
					}
				}
			}
		}
	}

	return bestMatch
}

// matchPattern 匹配模式
func (sm *SmartProxyManager) matchPattern(pattern, host string) bool {
	if pattern == host {
		return true // 精确匹配
	}

	if strings.HasPrefix(pattern, "*.") {
		// 通配符匹配
		suffix := pattern[2:] // 去掉 "*."
		return strings.HasSuffix(host, "."+suffix) || host == suffix
	}

	return false
}

// LearnFromFailure 从失败中学习
func (sm *SmartProxyManager) LearnFromFailure(host string, reason string) {
	// 检查是否已有规则
	if sm.findMatchingRule(host) != nil {
		return // 已有规则，不重复添加
	}

	// 添加自动学习的代理规则
	rule := &RouteRule{
		ID:        fmt.Sprintf("auto-%s-%d", host, time.Now().Unix()),
		Pattern:   host,
		RouteType: "proxy", // 默认启用代理
		Source:    "auto_learned",
		Enabled:   true,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	sm.rules = append(sm.rules, rule)
	sm.SaveRules()

	log.Printf("[SmartProxy] 自动学习: %s 直连失败(%s)，添加代理规则", host, reason)
}

// AddManualRule 添加手动规则
func (sm *SmartProxyManager) AddManualRule(pattern, routeType string) error {
	// 检查是否已存在相同模式的手动规则
	for _, rule := range sm.rules {
		if rule.Pattern == pattern && rule.Source == "manual" {
			// 更新现有规则
			rule.RouteType = routeType
			rule.Enabled = true
			rule.LastUsed = time.Now()
			return sm.SaveRules()
		}
	}

	// 添加新规则
	rule := &RouteRule{
		ID:        fmt.Sprintf("manual-%s-%d", pattern, time.Now().Unix()),
		Pattern:   pattern,
		RouteType: routeType,
		Source:    "manual",
		Enabled:   true,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	sm.rules = append(sm.rules, rule)
	log.Printf("[SmartProxy] 添加手动规则: %s -> %s", pattern, routeType)
	return sm.SaveRules()
}

// RemoveRule 删除规则
func (sm *SmartProxyManager) RemoveRule(id string) error {
	for i, rule := range sm.rules {
		if rule.ID == id {
			sm.rules = append(sm.rules[:i], sm.rules[i+1:]...)
			log.Printf("[SmartProxy] 删除规则: %s", rule.Pattern)
			return sm.SaveRules()
		}
	}
	return fmt.Errorf("规则不存在: %s", id)
}

// GetRules 获取所有规则
func (sm *SmartProxyManager) GetRules() []*RouteRule {
	return sm.rules
}

// ClearAutoLearnedRules 清空自动学习的规则
func (sm *SmartProxyManager) ClearAutoLearnedRules() error {
	newRules := make([]*RouteRule, 0)
	for _, rule := range sm.rules {
		if rule.Source != "auto_learned" {
			newRules = append(newRules, rule)
		}
	}

	removed := len(sm.rules) - len(newRules)
	sm.rules = newRules

	log.Printf("[SmartProxy] 清空了 %d 条自动学习规则", removed)
	return sm.SaveRules()
}
