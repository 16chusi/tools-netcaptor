<template>
  <div class="basic-section">
    <h5>🌐 页面导航</h5>
    <div class="form-item">
      <label>目标URL</label>
      <input v-model="formData.url" placeholder="https://example.com 或 {url}" />
      <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {url} 或 {data.website}</div>
    </div>
    <div class="form-item">
      <label>打开方式</label>
      <select v-model="formData.openMode">
        <option value="current">当前窗口</option>
        <option value="new">新窗口</option>
      </select>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  formData: any
}>()
</script>

<style scoped>
@import './common.css';
</style>
