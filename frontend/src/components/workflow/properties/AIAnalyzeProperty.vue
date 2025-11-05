<template>
  <div class="ai-section">
    <h5>🤖 AI 内容分析</h5>
    <div class="form-item">
      <label>AI模型</label>
      <select v-model="formData.modelIndex">
        <option v-for="(model, index) in aiModels" :key="index" :value="index">
          {{ model.name || `${model.provider} 模型 ${index + 1}` }}
        </option>
      </select>
    </div>
    <div class="form-item">
      <label>分析任务</label>
      <textarea v-model="formData.prompt" rows="4" placeholder="分析页面结构，识别表格、列表、表单等元素..."></textarea>
      <div class="variable-hint">💡 支持变量: {变量名}，如 {pageContent}</div>
    </div>
    <div class="form-item">
      <label>保存到变量</label>
      <input v-model="formData.saveToVariable" placeholder="analysisResult" />
      <div class="variable-hint">结果将保存到: {{ '{' + (formData.saveToVariable || 'analysisResult') + '}' }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

defineProps<{ formData: any }>()

const aiModels = ref<any[]>([])

onMounted(() => {
  const saved = localStorage.getItem('ai-models')
  if (saved) {
    const settings = JSON.parse(saved)
    aiModels.value = settings.models || []
  }
})
</script>

<style scoped>
@import './common.css';
</style>
