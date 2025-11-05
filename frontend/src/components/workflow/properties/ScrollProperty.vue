<template>
  <div class="basic-section">
    <h5>📜 页面滚动</h5>
    <div class="form-item">
      <label>滚动类型</label>
      <select v-model="formData.scrollType">
        <option value="bottom">滚动到底部</option>
        <option value="top">滚动到顶部</option>
        <option value="element">滚动到元素</option>
        <option value="pixels">滚动指定像素</option>
      </select>
    </div>
    <div class="form-item" v-if="formData.scrollType === 'element'">
      <label>目标元素选择器</label>
      <input v-model="formData.selector" placeholder="#element-id 或 .class-name" />
    </div>
    <div class="form-item" v-if="formData.scrollType === 'pixels'">
      <label>滚动距离(px)</label>
      <input v-model.number="formData.pixels" type="number" placeholder="500" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ formData: any }>()
</script>

<style scoped>
@import './common.css';
</style>
