<template>
  <div class="ai-model-tab">
    <div class="section">
      <h3>🤖 AI模型配置</h3>
      <p class="description">支持所有OpenAI兼容接口，包括OpenAI、智谱AI、DeepSeek、Ollama等</p>
      
      <div class="model-list">
        <div v-for="(model, index) in models" :key="index" class="model-item">
          <div class="model-info">
            <div class="form-row">
              <label>模型名称:</label>
              <input v-model="model.name" placeholder="gpt-4, glm-4, deepseek-chat, llama3" class="model-input" />
            </div>
            <div class="form-row">
              <label>API Key:</label>
              <input v-model="model.apiKey" type="password" placeholder="sk-..." class="model-input" />
            </div>
            <div class="form-row">
              <label>Base URL:</label>
              <input v-model="model.baseUrl" placeholder="https://api.openai.com/v1" class="model-input" />
              <div class="url-examples">
                <span>常用地址:</span>
                <button @click="setBaseUrl(index, 'https://api.openai.com/v1')" class="url-btn">OpenAI</button>
                <button @click="setBaseUrl(index, 'https://open.bigmodel.cn/api/paas/v4')" class="url-btn">智谱AI</button>
                <button @click="setBaseUrl(index, 'https://api.deepseek.com/v1')" class="url-btn">DeepSeek</button>
                <button @click="setBaseUrl(index, 'http://localhost:11434/v1')" class="url-btn">Ollama</button>
              </div>
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
              {{ model.name || `模型 ${index + 1}` }}
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
import {
  CallAI,
  GetAIModels,
  ShowQuestionDialog,
  TestAIModel,
  UpdateAIModels
} from "../../../wailsjs/go/network/NetworkApp";

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
    provider: 'openai', // 统一使用openai兼容
    name: '',
    apiKey: '',
    baseUrl: ''
  })
}

function setBaseUrl(index: number, url: string) {
  models.value[index].baseUrl = url
}

async function removeModel(index: number) {
  const model = models.value[index]
  const modelName = model.name || `模型 ${index + 1}`
  
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
      console.log('[前端] 测试模型连接...')
      await TestAIModel(model)
      
      console.log('[前端] 保存配置到数据库...')
      await saveModels()
      
      setTestResult(index, 'success', '测试成功并已保存', '连接正常')
      
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
    toast.success('AI模型设置已保存')
  } catch (error) {
    console.error('保存模型配置失败:', error)
    throw error
  }
}

// 加载模型配置
onMounted(async () => {
  try {
    const loadedModels = await GetAIModels()
    if (loadedModels && loadedModels.length > 0) {
      models.value = loadedModels
      console.log('[AIModelTab] 已加载', loadedModels.length, '个模型配置')
    }
  } catch (error) {
    console.error('[AIModelTab] 加载模型配置失败:', error)
  }
})

// 监听变化自动保存
watch([models, defaultModel, defaultTemperature, defaultMaxTokens], () => {
  debounce('auto-save', async () => {
    try {
      await UpdateAIModels(models.value)
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

.description {
  margin: 0 0 16px 0;
  padding: 12px;
  background: #e6f7ff;
  border: 1px solid #91d5ff;
  border-radius: 4px;
  font-size: 13px;
  color: #0050b3;
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

.model-input {
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 13px;
}

.url-examples {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
  font-size: 12px;
}

.url-examples span {
  color: #666;
}

.url-btn {
  padding: 2px 8px;
  background: #f0f0f0;
  border: 1px solid #d9d9d9;
  border-radius: 3px;
  cursor: pointer;
  font-size: 11px;
  transition: all 0.2s;
}

.url-btn:hover {
  background: #e6f7ff;
  border-color: #91d5ff;
}

.model-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.test-btn, .remove-btn, .add-btn {
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

.test-btn:disabled {
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
