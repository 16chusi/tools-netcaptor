<template>
  <div class="ai-section">
    <h5>🤖 AI 数据提取</h5>
    <div class="form-item">
      <label>AI模型</label>
      <select v-model="formData.modelIndex">
        <option v-for="(model, index) in aiModels" :key="index" :value="index">
          {{ model.name || `${model.provider} 模型 ${index + 1}` }}
        </option>
      </select>
    </div>
    <div class="form-item">
      <label>提取指令</label>
      <textarea v-model="formData.extractPrompt" rows="3" placeholder="请描述需要提取的数据，如：提取页面中的商品名称、价格和库存信息"></textarea>
      <div class="variable-hint">💡 支持变量: {变量名}，如 从 {pageContent} 中提取商品信息</div>
    </div>
    <div class="form-item">
      <label>输出格式</label>
      <select v-model="formData.outputFormat">
        <option value="json">JSON</option>
        <option value="text">纯文本</option>
        <option value="csv">CSV</option>
      </select>
    </div>
    <div class="form-item">
      <label>保存到变量</label>
      <input v-model="formData.saveToVariable" placeholder="extractedData" />
      <div class="variable-hint">结果将保存到: {{ '{' + (formData.saveToVariable || 'extractedData') + '}' }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

defineProps<{
  formData: any
}>()

const aiModels = ref<any[]>([])

onMounted(() => {
  const saved = localStorage.getItem('ai-model-settings')
  if (saved) {
    const settings = JSON.parse(saved)
    aiModels.value = settings.models || []
  }
})
</script>

<style scoped>
@import './common.css';
</style>
