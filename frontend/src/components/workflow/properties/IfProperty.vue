<template>
  <div class="control-section">
    <h5>🔀 条件判断</h5>
    <div class="form-item">
      <label>左值</label>
      <input v-model="formData.leftValue" placeholder="{variable} 或 固定值" />
      <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}</div>
    </div>
    <div class="form-item">
      <label>比较操作符</label>
      <select v-model="formData.operator">
        <option value="==">等于 (==)</option>
        <option value="!=">不等于 (!=)</option>
        <option value=">">大于 (>)</option>
        <option value="<">小于 (<)</option>
        <option value=">=">大于等于 (>=)</option>
        <option value="<=">小于等于 (<=)</option>
        <option value="contains">包含</option>
        <option value="not_contains">不包含</option>
        <option value="starts_with">开头是</option>
        <option value="ends_with">结尾是</option>
        <option value="regex">正则匹配</option>
        <option value="exists">存在</option>
        <option value="not_exists">不存在</option>
      </select>
    </div>
    <div class="form-item" v-if="!['exists', 'not_exists'].includes(formData.operator)">
      <label>右值</label>
      <input v-model="formData.rightValue" placeholder="{variable} 或 固定值" />
      <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ formData: any }>()
</script>

<style scoped>
@import './common.css';
</style>
