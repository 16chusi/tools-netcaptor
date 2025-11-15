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
      <label>内容抓取类型</label>
      <select v-model="formData.contentType">
        <option value="text">纯文本内容</option>
        <option value="html">HTML结构</option>
        <option value="table">表格数据</option>
        <option value="image">图片信息</option>
        <option value="mixed">混合内容</option>
      </select>
      <div class="variable-hint">
        纯文本：只获取文字内容，速度快成本低<br>
        HTML结构：保留标签结构，适合复杂布局<br>
        表格数据：专门提取表格，保持数据结构<br>
        图片信息：获取图片链接和alt文本<br>
        混合内容：包含文本、表格、图片的完整信息
      </div>
    </div>
    <div class="form-item">
      <label>保存到变量</label>
      <input v-model="formData.saveToVariable" placeholder="extractedData" />
      <div class="variable-hint">结果将保存到: {{ '{' + (formData.saveToVariable || 'extractedData') + '}' }}</div>
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

defineProps<{
  formData: any
}>()

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
