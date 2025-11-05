<template>
  <div class="data-section">
    <h5>📊 提取数据</h5>
    <div class="form-item">
      <label>选择器类型</label>
      <select v-model="formData.selectorType">
        <option value="css">CSS 选择器</option>
        <option value="xpath">XPath</option>
      </select>
    </div>
    <div class="form-item">
      <label>元素选择器</label>
      <input v-model="formData.selector" :placeholder="formData.selectorType === 'xpath' ? '//*[@class=&quot;title&quot;]' : '.title'" />
      <div class="variable-hint">💡 支持变量: {变量名}，如 .item-{index} 或 [data-id='{data.id}']</div>
    </div>
    <div class="form-item">
      <label>提取属性</label>
      <select v-model="formData.attribute">
        <option value="text">文本内容</option>
        <option value="href">链接地址</option>
        <option value="src">图片地址</option>
        <option value="value">表单值</option>
        <option value="custom">自定义属性</option>
      </select>
    </div>
    <div class="form-item" v-if="formData.attribute === 'custom'">
      <label>自定义属性名</label>
      <input v-model="formData.customAttribute" placeholder="data-id" />
    </div>
    <div class="form-item">
      <label>保存到变量</label>
      <input v-model="formData.saveToVariable" placeholder="myData" />
      <div class="variable-hint">结果将保存到: {{ '{' + (formData.saveToVariable || 'myData') + '}' }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ formData: any }>()
</script>

<style scoped>
@import './common.css';
</style>
