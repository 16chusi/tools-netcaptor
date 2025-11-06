<template>
  <div class="ai-section">
    <h5>🤖 AI 数据验证</h5>
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
      <input v-model="formData.dataSource" placeholder="{extractedData}" />
      <div class="variable-hint">💡 当前引用: {{ formData.dataSource || '{extractedData}' }}</div>
    </div>
    <div class="form-item">
      <label>验证规则</label>
      <textarea v-model="formData.prompt" rows="4" placeholder="验证数据格式是否正确，检查必填字段..."></textarea>
      <div class="variable-hint">💡 支持变量: {变量名}</div>
    </div>
    <div class="form-item">
      <label>保存到变量</label>
      <input v-model="formData.saveToVariable" placeholder="validationResult" />
      <div class="variable-hint">结果将保存到: {{ '{' + (formData.saveToVariable || 'validationResult') + '}' }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { GetAIModels } from '../../../../wailsjs/go/main/NetworkApp'

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
