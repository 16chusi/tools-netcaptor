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
    <div class="form-item">
      <label>设备模式</label>
      <select v-model="formData.deviceMode">
        <option value="desktop">🖥️ 桌面模式</option>
        <option value="mobile">📱 移动模式 (iPhone)</option>
        <option value="tablet">📱 平板模式 (iPad)</option>
      </select>
      <div class="variable-hint">💡 选择网页显示的设备类型</div>
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
