<template>
  <div class="data-section">
    <h5>🔍 请求拦截</h5>
    <div class="form-item">
      <label>URL匹配</label>
      <input v-model="formData.urlPattern" placeholder="https://api.example.com/*" />
      <div class="variable-hint">💡 支持通配符 * 和变量 {变量名}</div>
    </div>
    <div class="form-item">
      <label>匹配方式</label>
      <select v-model="formData.matchType">
        <option value="contains">包含</option>
        <option value="exact">完全匹配</option>
        <option value="regex">正则表达式</option>
        <option value="starts_with">开头匹配</option>
      </select>
    </div>
    <div class="form-item">
      <label>提取字段</label>
      <input v-model="formData.extractKeys" placeholder="* 或 name,age,email" />
      <div class="variable-hint">💡 * 表示全部，或指定字段名用逗号分隔</div>
    </div>
    <div class="form-item">
      <label>保存到变量</label>
      <input v-model="formData.saveToVariable" placeholder="responseData" />
      <div class="variable-hint">结果将保存到: {{ '{' + (formData.saveToVariable || 'responseData') + '}' }}</div>
    </div>
    <div class="form-item">
      <label>超时时间(ms)</label>
      <input v-model.number="formData.timeout" type="number" placeholder="30000" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ formData: any }>()
</script>

<style scoped>
@import './common.css';
</style>
