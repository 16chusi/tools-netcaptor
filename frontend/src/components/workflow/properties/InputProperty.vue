<template>
  <div class="basic-section">
    <h5>⌨️ 输入文本</h5>
    <div class="form-item">
      <label>选择器类型</label>
      <select v-model="formData.selectorType">
        <option value="css">CSS 选择器</option>
        <option value="xpath">XPath</option>
      </select>
    </div>
    <div class="form-item">
      <label>元素选择器</label>
      <input v-model="formData.selector" :placeholder="formData.selectorType === 'xpath' ? '//*[@id=&quot;username&quot;]' : '#username'" />
      <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 #user-{index} 或 [data-name='{data.name}']</div>
    </div>
    <div class="form-item">
      <label>输入内容</label>
      <input v-model="formData.text" placeholder="要输入的文本" />
      <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {data.name} 或 用户: {data.username}</div>
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
