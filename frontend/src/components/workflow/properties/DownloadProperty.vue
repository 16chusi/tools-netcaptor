<template>
  <div class="data-section">
    <h5>💾 文件下载</h5>
    <div class="form-item">
      <label>URL来源</label>
      <select v-model="formData.urlSource">
        <option value="direct">直接输入</option>
        <option value="variable">从变量获取</option>
        <option value="template">模板拼接</option>
      </select>
    </div>
    <div class="form-item" v-if="formData.urlSource === 'direct'">
      <label>下载URL</label>
      <input v-model="formData.downloadUrl" placeholder="https://example.com/file.pdf" />
      <div class="variable-hint">💡 支持变量: {变量名}，如 {data.fileUrl}</div>
    </div>
    <div class="form-item" v-if="formData.urlSource === 'variable'">
      <label>变量名</label>
      <input v-model="formData.urlVariable" placeholder="downloadUrl" />
    </div>
    <div class="form-item" v-if="formData.urlSource === 'template'">
      <label>URL模板</label>
      <input v-model="formData.urlTemplate" placeholder="https://api.com/files/{fileId}" />
      <div class="variable-hint">💡 使用 {变量名} 进行替换</div>
    </div>
    <div class="form-item">
      <label>保存目录</label>
      <div class="form-row">
        <input v-model="formData.saveDirectory" placeholder="选择保存目录" readonly />
        <button @click="selectDirectory" type="button">选择</button>
      </div>
    </div>
    <div class="form-item">
      <label>文件名模板</label>
      <input v-model="formData.filenameTemplate" placeholder="{filename} 或 file_{index}.pdf" />
      <div class="variable-hint">💡 支持变量: {filename}, {index}, {timestamp}</div>
    </div>
  </div>
</template>

<script setup lang="ts">

import {SelectDownloadDirectory} from "../../../../wailsjs/go/network/NetworkApp";

const props = defineProps<{ formData: any }>()

const selectDirectory = async () => {
  try {
    const dir = await SelectDownloadDirectory()
    if (dir) {
      props.formData.saveDirectory = dir
    }
  } catch (error) {
    console.error('选择目录失败:', error)
  }
}
</script>

<style scoped>
@import './common.css';
</style>
