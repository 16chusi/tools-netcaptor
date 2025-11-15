<template>
  <div class="ai-section">
    <h5>🤖 AI 数据转换</h5>
    <div class="form-item">
      <label>AI模型</label>
      <select v-model="formData.modelIndex">
        <option v-for="(model, index) in aiModels" :key="index" :value="index">
          {{ model.name || `${model.provider} 模型 ${index + 1}` }}
        </option>
      </select>
    </div>
    <div class="form-item">
      <label>数据来源</label>
      <input v-model="formData.dataSource" placeholder="{rawData}" />
      <div class="variable-hint">💡 当前引用: {{ formData.dataSource || '{rawData}' }}</div>
    </div>
    <div class="form-item">
      <label>转换指令</label>
      <textarea v-model="formData.prompt" rows="4" placeholder="将数据转换为标准格式，清洗无效数据..."></textarea>
      <div class="variable-hint">💡 支持变量: {变量名}</div>
    </div>
    <div class="form-item">
      <label>输出格式</label>
      <select v-model="formData.outputFormat">
        <option value="json">JSON</option>
        <option value="csv">CSV</option>
        <option value="xml">XML</option>
      </select>
    </div>
    <div class="form-item">
      <label>保存到变量</label>
      <input v-model="formData.saveToVariable" placeholder="transformedData" />
      <div class="variable-hint">结果将保存到: {{ '{' + (formData.saveToVariable || 'transformedData') + '}' }}</div>
    </div>
    
    <details class="advanced-options">
      <summary>🔧 高级选项</summary>
      <div class="form-item">
        <label>超时时间(秒)</label>
        <input v-model.number="formData.timeout" type="number" min="10" max="300" placeholder="100" />
        <div class="variable-hint">AI处理超时时间，默认100秒</div>
      </div>
      <div class="form-item">
        <label>思考模式</label>
        <select v-model="formData.thinkingMode">
          <option value="enabled">开启</option>
          <option value="disabled">关闭</option>
        </select>
      </div>
      <div class="form-item">
        <label>Top-p</label>
        <input v-model.number="formData.topP" type="number" min="0" max="1" step="0.1" placeholder="0.9" />
        <div class="variable-hint">控制输出随机性，0.1-1.0</div>
      </div>
      <div class="form-item">
        <label>Temperature</label>
        <input v-model.number="formData.temperature" type="number" min="0" max="2" step="0.1" placeholder="0.7" />
        <div class="variable-hint">控制创造性，0.1-2.0</div>
      </div>
      <div class="form-item">
        <label>Max Tokens</label>
        <input v-model.number="formData.maxTokens" type="number" min="100" max="8000" placeholder="2000" />
        <div class="variable-hint">最大输出长度</div>
      </div>
    </details>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {GetAIModels} from "../../../../wailsjs/go/network/NetworkApp";

defineProps<{ formData: any }>()

const aiModels = ref<any[]>([])

onMounted(async () => {
  try {
    // 从后端API获取AI模型配置
    const models = await GetAIModels()
    aiModels.value = models || []
  } catch (error) {
    console.error('获取AI模型配置失败:', error)
    aiModels.value = []
  }
})
</script>

<style scoped>
@import './common.css';
</style>
