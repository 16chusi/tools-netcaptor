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
