<template>
  <div class="ai-model-tab">
    <div class="section">
      <h3>🤖 AI模型配置</h3>
      <div class="model-list">
        <div v-for="(model, index) in models" :key="index" class="model-item">
          <div class="model-info">
            <div class="form-row">
              <label>供应商:</label>
              <select v-model="model.provider" class="provider-select">
                <option value="openai">OpenAI</option>
                <option value="anthropic">Anthropic</option>
                <option value="azure">Azure OpenAI</option>
                <option value="custom">自定义</option>
              </select>
            </div>
            <div class="form-row">
              <label>模型名称:</label>
              <input v-model="model.name" placeholder="gpt-4, claude-3-sonnet" class="model-input" />
            </div>
            <div class="form-row">
              <label>API Key:</label>
              <input v-model="model.apiKey" type="password" placeholder="sk-..." class="model-input" />
            </div>
            <div class="form-row">
              <label>Base URL:</label>
              <input v-model="model.baseUrl" placeholder="https://api.openai.com/v1 (包含版本号)" class="model-input" />
            </div>
          </div>
          <div class="model-actions">
            <button @click="testAndSaveModel(index)" class="test-btn" :disabled="!model.apiKey || testingStates[index]">
              {{ testingStates[index] ? '测试中...' : '测试并保存' }}
            </button>
            <button @click="removeModel(index)" class="remove-btn">删除</button>
          </div>
          <!-- 测试结果区域 -->
          <div v-if="testResults[index]" class="test-result" :class="testResults[index].type">
            <div class="result-header">
              <span class="result-icon">{{ testResults[index].type === 'success' ? '✅' : '❌' }}</span>
              <span class="result-title">{{ testResults[index].title }}</span>
              <span class="result-time">{{ testResults[index].time }}</span>
            </div>
            <div v-if="testResults[index].message" class="result-message">
              {{ testResults[index].message }}
            </div>
          </div>
        </div>
      </div>
      <button @click="addModel" class="add-btn">+ 添加模型</button>
    </div>

    <div class="section">
      <h3>⚙️ 默认设置</h3>
      <div class="default-settings">
        <label>
          默认模型:
          <select v-model="defaultModel" class="default-select">
            <option v-for="(model, index) in models" :key="index" :value="index">
              {{ model.name || `${model.provider} 模型 ${index + 1}` }}
            </option>
          </select>
        </label>
        <label>
          默认温度:
          <input v-model.number="defaultTemperature" type="range" min="0" max="1" step="0.1" />
          <span>{{ defaultTemperature }}</span>
        </label>
        <label>
          最大Token数:
          <input v-model.number="defaultMaxTokens" type="number" min="100" max="8000" />
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, reactive } from 'vue'
import { toast } from '../../utils/toast'
import { TestAIModel, UpdateAIModels, GetAIModels, CallAI, ShowQuestionDialog } from '../../../wailsjs/go/main/NetworkApp'

interface AIModel {
  provider: string
  name: string
  apiKey: string
  baseUrl: string
}

const models = ref<AIModel[]>([])
const defaultModel = ref(0)
const defaultTemperature = ref(0.7)
const defaultMaxTokens = ref(2000)

// 测试状态管理
const testingStates = reactive<Record<number, boolean>>({})
const testResults = reactive<Record<number, any>>({})

// 防抖定时器
const debounceTimers = reactive<Record<string, number>>({})

function addModel() {
  models.value.push({
    provider: 'openai',
    name: '',
    apiKey: '',
    baseUrl: ''
  })
}

async function removeModel(index: number) {
  const model = models.value[index]
  const modelName = model.name || `${model.provider} 模型`
  
  try {
    const result = await ShowQuestionDialog(
      '确认删除',
      `确定要删除模型配置 "${modelName}" 吗？\n\n此操作不可撤销。`
    )
    
    if (result === 'Yes') {
      models.value.splice(index, 1)
      delete testingStates[index]
      delete testResults[index]
      
      if (defaultModel.value >= models.value.length) {
        defaultModel.value = Math.max(0, models.value.length - 1)
      }
      
      toast.success('模型配置已删除')
    }
  } catch (error) {
    console.log('删除操作已取消或出错:', error)
  }
}

function debounce(key: string, fn: Function, delay: number = 1000) {
  if (debounceTimers[key]) {
    clearTimeout(debounceTimers[key])
  }
  debounceTimers[key] = window.setTimeout(() => {
    fn()
    delete debounceTimers[key]
  }, delay)
}

function setTestResult(index: number, type: 'success' | 'error', title: string, message?: string) {
  testResults[index] = {
    type,
    title,
    message,
    time: new Date().toLocaleTimeString()
  }
}

async function testAndSaveModel(index: number) {
  const key = `test-save-${index}`
  debounce(key, async () => {
    const model = models.value[index]
    console.log('[前端] 测试并保存模型:', model)
    
    if (!model.apiKey) {
      setTestResult(index, 'error', '测试失败', '请先配置API Key')
      return
    }
    
    testingStates[index] = true
    
    try {
      // 第一步：测试连接
      console.log('[前端] 步骤1: 测试模型连接...')
      await TestAIModel(model)
      
      // 第二步：AI响应测试
      console.log('[前端] 步骤2: 测试AI响应...')
      await UpdateAIModels(models.value) // 先更新配置
      const result = await CallAI(index, "Hello", "You are a helpful assistant. Reply with just 'OK'.")
      
      // 第三步：保存到数据库
      console.log('[前端] 步骤3: 保存配置到数据库...')
      await saveModels()
      
      setTestResult(index, 'success', '测试成功并已保存', 
        `连接正常，AI响应: ${result.substring(0, 50)}${result.length > 50 ? '...' : ''}`)
      
    } catch (error: any) {
      console.error('[前端] 测试或保存失败:', error)
      setTestResult(index, 'error', '测试失败', error.message)
    } finally {
      testingStates[index] = false
    }
  })
}

async function saveModels() {
  try {
    // 保存到后端数据库
    await UpdateAIModels(models.value)
    // 同时保存到localStorage作为备份
    saveSettings()
  } catch (error) {
    console.error('保存模型配置失败:', error)
    throw error
  }
}

function saveSettings() {
  const settings = {
    models: models.value,
    defaultModel: defaultModel.value,
    defaultTemperature: defaultTemperature.value,
    defaultMaxTokens: defaultMaxTokens.value
  }
  
  localStorage.setItem('ai-model-settings', JSON.stringify(settings))
  toast.success('AI模型设置已保存')
}

function loadSettings() {
  const saved = localStorage.getItem('ai-model-settings')
  if (saved) {
    const settings = JSON.parse(saved)
    models.value = settings.models || []
    defaultModel.value = settings.defaultModel || 0
    defaultTemperature.value = settings.defaultTemperature || 0.7
    defaultMaxTokens.value = settings.defaultMaxTokens || 2000
  }
  
  if (models.value.length === 0) {
    addModel()
  }
}

onMounted(async () => {
  try {
    // 从后端数据库加载
    const savedModels = await GetAIModels()
    if (savedModels && savedModels.length > 0) {
      models.value = savedModels
      console.log('[AIModelTab] 从数据库加载了', savedModels.length, '个模型')
    } else {
      // 如果数据库为空，尝试从 localStorage 加载
      loadSettings()
    }
  } catch (error) {
    console.error('[AIModelTab] 加载模型失败，使用 localStorage:', error)
    loadSettings()
  }
})

// 监听变化自动保存
watch([models, defaultModel, defaultTemperature, defaultMaxTokens], () => {
  debounce('auto-save', async () => {
    try {
      await UpdateAIModels(models.value)
      saveSettings() // 同时保存到 localStorage 作为备份
    } catch (error) {
      console.error('自动保存失败:', error)
    }
  }, 2000)
}, { deep: true })
</script>

<style scoped>
.ai-model-tab {
  padding: 20px;
}

.section {
  margin-bottom: 30px;
  padding: 20px;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  background: #fafafa;
}

.section h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
  color: #333;
}

.model-item {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  margin-bottom: 12px;
  background: white;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
}

.model-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-row label {
  font-size: 12px;
  font-weight: 500;
  color: #333;
}

.provider-select, .model-input {
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 13px;
}

.model-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.test-btn, .quick-test-btn, .remove-btn, .add-btn {
  padding: 6px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.test-btn {
  background: #52c41a;
  color: white;
  border-color: #52c41a;
}

.quick-test-btn {
  background: #1890ff;
  color: white;
  border-color: #1890ff;
}

.test-btn:disabled, .quick-test-btn:disabled {
  background: #f5f5f5;
  color: #999;
  cursor: not-allowed;
}

.test-result {
  margin-top: 12px;
  padding: 12px;
  border-radius: 6px;
  border-left: 4px solid;
  font-size: 12px;
}

.test-result.success {
  background: #f6ffed;
  border-left-color: #52c41a;
  color: #389e0d;
}

.test-result.error {
  background: #fff2f0;
  border-left-color: #ff4d4f;
  color: #cf1322;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.result-icon {
  font-size: 14px;
}

.result-title {
  font-weight: 500;
  flex: 1;
}

.result-time {
  font-size: 11px;
  opacity: 0.7;
}

.result-message {
  font-size: 11px;
  opacity: 0.8;
  line-height: 1.4;
  word-break: break-all;
}

.remove-btn {
  background: #ff4d4f;
  color: white;
  border-color: #ff4d4f;
}

.add-btn {
  background: #1890ff;
  color: white;
  border-color: #1890ff;
}

.default-settings {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.default-settings label {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
}

.default-select {
  padding: 6px 8px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  min-width: 200px;
}
</style>
