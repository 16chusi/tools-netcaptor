<template>
  <div class="basic-section">
    <h5>🖱️ 点击元素</h5>
    <div class="form-item">
      <label>选择器类型</label>
      <select v-model="formData.selectorType">
        <option value="css">CSS 选择器</option>
        <option value="xpath">XPath</option>
      </select>
    </div>
    <div class="form-item">
      <label>元素选择器</label>
      <input v-model="formData.selector" :placeholder="formData.selectorType === 'xpath' ? '//*[@id=&quot;btn&quot;]' : '#id 或 .class'" />
      <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 #btn-{index} 或 [data-id='{data.id}']</div>
    </div>
    <div class="form-item">
      <label>等待元素出现(ms)</label>
      <input v-model.number="formData.waitTime" type="number" placeholder="3000" />
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
