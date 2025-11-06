<template>
  <div class="ai-section">
    <h5>🤖 AI 智能导航</h5>
    <div class="form-item">
      <label>AI模型</label>
      <select v-model="formData.modelIndex">
        <option v-for="(model, index) in aiModels" :key="index" :value="index">
          {{ model.name || `${model.provider} 模型 ${index + 1}` }}
        </option>
      </select>
    </div>
    <div class="form-item">
      <label>导航描述</label>
      <textarea v-model="formData.prompt" rows="4" placeholder="找到并点击'产品中心'菜单"></textarea>
      <div class="variable-hint">💡 支持变量: {变量名}</div>
    </div>
    <div class="form-item">
      <label>等待时间(ms)</label>
      <input v-model.number="formData.waitTime" type="number" placeholder="3000" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {GetAIModels} from "../../../../wailsjs/go/main/NetworkApp";

defineProps<{ formData: any }>()

const aiModels = ref<any[]>([])

onMounted(async () => {
  try {
    // 从后端API获取AI模型配置
    const models = await  GetAIModels()
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
