<template>
  <div class="smart-proxy-tab">
    <!-- 添加规则 -->
    <div class="setting-item">
      <label>添加路由规则</label>
      <div class="add-rule-section">
        <input 
          v-model="newRule.pattern" 
          type="text" 
          class="setting-input" 
          placeholder="域名模式，如: google.com 或 *.openai.com"
        />
        <select v-model="newRule.routeType" class="setting-select">
          <option value="direct">直连</option>
          <option value="proxy">代理</option>
          <option value="auto">自动</option>
        </select>
        <button @click="addRule" :disabled="!newRule.pattern.trim()" class="add-btn">
          添加规则
        </button>
      </div>
    </div>

    <!-- 规则列表 -->
    <div class="setting-item">
      <label>路由规则列表 ({{ rules.length }} 条)</label>
      <div class="rules-list">
        <div v-if="rules.length === 0" class="empty-state">
          暂无规则，系统将默认使用直连
        </div>
        <div v-for="rule in sortedRules" :key="rule.id" class="rule-item">
          <div class="rule-info">
            <div class="rule-pattern">
              <span :class="['rule-icon', rule.source]">
                {{ rule.source === 'manual' ? '🔧' : '🤖' }}
              </span>
              <strong>{{ rule.pattern }}</strong>
            </div>
            <div class="rule-details">
              <span :class="['route-type', rule.route_type]">
                {{ getRouteTypeText(rule.route_type) }}
              </span>
              <span class="rule-source">{{ rule.source === 'manual' ? '手动' : '自动学习' }}</span>
              <span class="rule-time">{{ formatTime(rule.last_used) }}</span>
            </div>
          </div>
          <div class="rule-actions">
            <button @click="removeRule(rule.id)" class="remove-btn">删除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 批量操作 -->
    <div class="setting-item">
      <label>批量操作</label>
      <div class="batch-actions">
        <button @click="clearAutoRules" class="clear-btn">
          清空自动学习规则
        </button>
        <button @click="refreshRules" class="refresh-btn">
          刷新规则列表
        </button>
      </div>
    </div>

    <!-- 路由测试 -->
    <div class="setting-item">
      <label>路由测试</label>
      <div class="test-section">
        <input 
          v-model="testURL" 
          type="text" 
          class="setting-input" 
          placeholder="输入测试URL，如: https://www.google.com"
        />
        <button @click="testRouting" :disabled="testing" class="test-btn">
          {{ testing ? '测试中...' : '测试路由' }}
        </button>
      </div>
      <div v-if="testResult" class="test-result">
        <pre>{{ testResult }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {proxy} from "../../../wailsjs/go/models";
import RouteRule = proxy.RouteRule;
import {
  AddSmartProxyRule,
  ClearAutoLearnedRules,
  GetSmartProxyRules,
  RemoveSmartProxyRule, TestSmartProxyRouting
} from "../../../wailsjs/go/network/NetworkApp";

const rules = ref< RouteRule[]>([])
const newRule = ref({
  pattern: '',
  routeType: 'proxy'
})

// 测试相关变量
const testURL = ref('https://www.google.com')
const testing = ref(false)
const testResult = ref('')

const sortedRules = computed(() => {
  return [...rules.value].sort((a, b) => {
    // 手动规则优先
    if (a.source !== b.source) {
      return a.source === 'manual' ? -1 : 1
    }
    // 按最后使用时间排序
    return new Date(b.last_used).getTime() - new Date(a.last_used).getTime()
  })
})

onMounted(() => {
  loadRules()
})

async function loadRules() {
  try {
    const result = await GetSmartProxyRules()
    rules.value = result || []
  } catch (error) {
    console.error('加载规则失败:', error)
    alert('加载规则失败: ' + error)
  }
}

async function addRule() {
  if (!newRule.value.pattern.trim()) {
    alert('请输入域名模式')
    return
  }

  try {
    await AddSmartProxyRule(newRule.value.pattern.trim(), newRule.value.routeType)
    newRule.value.pattern = ''
    newRule.value.routeType = 'proxy'
    await loadRules()
  } catch (error) {
    console.error('添加规则失败:', error)
    alert('添加规则失败: ' + error)
  }
}

async function removeRule(id: string) {
  if (!confirm('确定要删除这条规则吗？')) {
    return
  }

  try {
    await RemoveSmartProxyRule(id)
    await loadRules()
  } catch (error) {
    console.error('删除规则失败:', error)
    alert('删除规则失败: ' + error)
  }
}

async function clearAutoRules() {
  if (!confirm('确定要清空所有自动学习的规则吗？')) {
    return
  }

  try {
    await ClearAutoLearnedRules()
    await loadRules()
  } catch (error) {
    console.error('清空规则失败:', error)
    alert('清空规则失败: ' + error)
  }
}

async function refreshRules() {
  await loadRules()
}

async function testRouting() {
  if (!testURL.value.trim()) {
    alert('请输入测试URL')
    return
  }

  testing.value = true
  testResult.value = ''

  try {
    const result = await TestSmartProxyRouting(testURL.value.trim())
    testResult.value = result
  } catch (error) {
    console.error('测试路由失败:', error)
    testResult.value = `测试失败: ${error}`
  } finally {
    testing.value = false
  }
}

function getRouteTypeText(routeType: string): string {
  switch (routeType) {
    case 'direct': return '直连'
    case 'proxy': return '代理'
    case 'auto': return '自动'
    default: return routeType
  }
}

function formatTime(timeStr: string): string {
  if (!timeStr) return '未使用'
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return `${Math.floor(diff / 86400000)}天前`
}
</script>

<style scoped>
.smart-proxy-tab {
  padding: 0;
}

.setting-item {
  margin-bottom: 24px;
}

.setting-item label {
  display: block;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.add-rule-section {
  display: flex;
  gap: 8px;
  align-items: center;
}

.setting-input,
.setting-select {
  height: 36px;
  padding: 0 12px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #333333;
  border-radius: 4px;
  font-size: 13px;
  box-sizing: border-box;
}

.setting-input {
  flex: 1;
}

.setting-select {
  width: 80px;
}

.add-btn {
  padding: 8px 16px;
  height: 36px;
  border: 1px solid #1a73e8;
  background: #1a73e8;
  color: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  white-space: nowrap;
}

.add-btn:hover:not(:disabled) {
  background: #1557b0;
}

.add-btn:disabled {
  background: #dadce0;
  border-color: #dadce0;
  cursor: not-allowed;
}

.rules-list {
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  max-height: 400px;
  overflow-y: auto;
}

.empty-state {
  padding: 20px;
  text-align: center;
  color: #666;
  font-style: italic;
}

.rule-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.rule-item:last-child {
  border-bottom: none;
}

.rule-info {
  flex: 1;
}

.rule-pattern {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.rule-icon {
  font-size: 16px;
}

.rule-details {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #666;
}

.route-type {
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: 500;
}

.route-type.direct {
  background: #e8f5e8;
  color: #2e7d32;
}

.route-type.proxy {
  background: #fff3e0;
  color: #f57c00;
}

.route-type.auto {
  background: #e3f2fd;
  color: #1976d2;
}

.remove-btn {
  padding: 4px 8px;
  border: 1px solid #d93025;
  background: transparent;
  color: #d93025;
  border-radius: 3px;
  cursor: pointer;
  font-size: 12px;
}

.remove-btn:hover {
  background: #d93025;
  color: white;
}

.batch-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.clear-btn,
.refresh-btn {
  padding: 8px 16px;
  border: 1px solid #dadce0;
  background: white;
  color: #5f6368;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.clear-btn:hover {
  border-color: #d93025;
  color: #d93025;
}

.refresh-btn:hover {
  border-color: #1a73e8;
  color: #1a73e8;
}

.test-section {
  display: flex;
  gap: 8px;
  align-items: center;
}

.test-btn {
  padding: 8px 16px;
  border: 1px solid #1a73e8;
  background: #1a73e8;
  color: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  white-space: nowrap;
}

.test-btn:hover:not(:disabled) {
  background: #1557b0;
}

.test-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.test-result {
  margin-top: 12px;
  padding: 12px;
  background: #f8f9fa;
  border: 1px solid #dadce0;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  white-space: pre-wrap;
  max-height: 300px;
  overflow-y: auto;
}
</style>
