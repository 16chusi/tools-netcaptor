<template>
  <div class="control-section">
    <h5>🔄 循环控制</h5>
    <div class="form-item">
      <label>循环次数</label>
      <input v-model.number="formData.count" type="number" placeholder="10" />
      <div class="variable-hint">💡 支持变量: {变量名}，如 {totalCount}</div>
    </div>
    <div class="form-item">
      <label>循环变量名</label>
      <input v-model="formData.variable" placeholder="index" />
      <div class="variable-hint">当前循环索引将保存到: {{ '{' + (formData.variable || 'index') + '}' }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ formData: any }>()
</script>

<style scoped>
@import './common.css';
</style>
