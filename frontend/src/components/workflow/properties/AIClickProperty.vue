<template>
  <div class="ai-section">
    <h5>🤖 AI 智能点击</h5>
    <div class="form-item">
      <label>AI模型</label>
      <select v-model="formData.modelIndex">
        <option v-for="(model, index) in aiModels" :key="index" :value="index">
          {{ model.name || `${model.provider} 模型 ${index + 1}` }}
        </option>
      </select>
    </div>
    <div class="form-item">
      <label>点击描述</label>
      <textarea v-model="formData.clickDescription" rows="2" placeholder="点击登录按钮"></textarea>
      <div class="variable-hint">💡 描述要点击的元素，AI会自动识别</div>
    </div>
    <div class="form-item">
      <label>等待元素出现(ms)</label>
      <input v-model.number="formData.waitTime" type="number" placeholder="3000" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

defineProps<{ formData: any }>()

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
