<template>
  <div class="data-section">
    <h5>📊 数据收集器</h5>
    <div class="form-item">
      <label>数据来源</label>
      <input v-model="formData.dataVariable" placeholder="变量名，如 data" />
      <div class="variable-hint">输入变量名，不需要 {}</div>
    </div>
    <div class="form-item">
      <label>保存文件</label>
      <div class="form-row">
        <input v-model="formData.filePath" placeholder="选择保存位置" readonly />
        <button @click="selectFile" type="button">保存到</button>
      </div>
      <div class="variable-hint">数据将追加到此文件</div>
    </div>
    <div class="form-item">
      <label>数据格式</label>
      <select v-model="formData.format">
        <option value="jsonl">JSONL (每行一个JSON)</option>
        <option value="text">文本 (每行追加)</option>
      </select>
    </div>
  </div>
</template>

<script setup lang="ts">

import {SelectSaveFilePath} from "../../../../wailsjs/go/network/NetworkApp";

const props = defineProps<{ formData: any }>()

const selectFile = async () => {
  const file = await SelectSaveFilePath('data.jsonl')
  if (file) props.formData.filePath = file
}
</script>

<style scoped>
@import './common.css';
</style>
