<template>
  <div class="data-section">
    <h5>📄 JSONL读取器</h5>
    <div class="form-item">
      <label>JSONL文件</label>
      <div class="form-row">
        <input v-model="formData.filePath" placeholder="选择JSONL文件" readonly />
        <button @click="selectFile" type="button">选择</button>
      </div>
    </div>
    <div class="form-item" v-if="availableKeys.length > 0">
      <label>可用字段</label>
      <div class="variable-hint">{{ availableKeys.join(', ') }}</div>
    </div>
    <div class="form-item" v-if="totalLines > 0">
      <label>总行数</label>
      <div class="variable-hint">{{ totalLines }} 条记录</div>
    </div>
    <div class="form-item">
      <label>保存到变量</label>
      <input v-model="formData.saveToVariable" placeholder="data" />
      <div class="variable-hint">当前行数据将保存到: {{ '{' + (formData.saveToVariable || 'data') + '}' }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { SelectJSONLFile, LoadJSONLFile } from '../../../../wailsjs/go/main/NetworkApp'

const props = defineProps<{ formData: any }>()

const availableKeys = ref<string[]>([])
const totalLines = ref(0)

const selectFile = async () => {
  try {
    const file = await SelectJSONLFile()
    if (file) {
      props.formData.filePath = file
      const result = await LoadJSONLFile(file)
      if (result.success) {
        availableKeys.value = result.keys || []
        totalLines.value = result.totalLines || 0
      }
    }
  } catch (error) {
    console.error('选择文件失败:', error)
  }
}
</script>

<style scoped>
@import './common.css';
</style>
