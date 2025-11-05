<template>
  <div class="data-section">
    <h5>💾 下载已捕获响应</h5>
    <div class="form-item">
      <label>URL匹配</label>
      <input v-model="formData.urlPattern" placeholder="*/api/data*" />
      <div class="variable-hint">支持通配符 *，匹配已捕获的请求</div>
    </div>
    <div class="form-item">
      <label>保存目录</label>
      <div class="form-row">
        <input v-model="formData.saveDirectory" placeholder="选择保存目录" readonly />
        <button @click="selectDirectory" type="button">选择</button>
      </div>
    </div>
    <div class="form-item">
      <label>文件扩展名</label>
      <input v-model="formData.fileExtension" placeholder="json" />
      <div class="variable-hint">如 json、txt、html 等，留空则自动检测</div>
    </div>
    <div class="form-item">
      <label>重复文件</label>
      <select v-model="formData.overwriteMode">
        <option value="skip">跳过</option>
        <option value="overwrite">覆盖</option>
      </select>
    </div>
  </div>
</template>

<script setup lang="ts">
import { SelectDownloadDirectory } from '../../../../wailsjs/go/main/NetworkApp'

const props = defineProps<{ formData: any }>()

const selectDirectory = async () => {
  const dir = await SelectDownloadDirectory()
  if (dir) props.formData.saveDirectory = dir
}
</script>

<style scoped>
@import './common.css';
</style>
