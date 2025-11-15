<template>
  <div class="data-section">
    <h5>📸 网页截图</h5>
    <div class="form-item">
      <label>保存格式</label>
      <select v-model="formData.format">
        <option value="png">PNG图片</option>
        <option value="jpeg">JPEG图片</option>
        <option value="pdf">PDF文档</option>
      </select>
    </div>
    <div class="form-item">
      <label>截图范围</label>
      <select v-model="formData.captureType">
        <option value="viewport">可视区域</option>
        <option value="fullpage">整个页面</option>
        <option value="element">指定元素</option>
      </select>
    </div>
    <div class="form-item" v-if="formData.captureType === 'element'">
      <label>元素选择器</label>
      <input v-model="formData.selector" placeholder="#element-id 或 .class-name" />
      <div class="variable-hint">💡 支持变量: {变量名}，如 #content-{index}</div>
    </div>
    <div class="form-item" v-if="formData.format !== 'pdf'">
      <label>图片质量</label>
      <select v-model="formData.quality">
        <option value="100">最高质量</option>
        <option value="80">高质量</option>
        <option value="60">中等质量</option>
        <option value="40">低质量</option>
      </select>
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
      <input v-model="formData.filenameTemplate" placeholder="screenshot_{timestamp} 或 page_{counter}" />
      <div class="variable-hint">💡 内置变量: {timestamp}, {date}, {time}, {uuid}, {uuid_short}, {random}, {random_6}, {counter}, {title}, {url}</div>
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
