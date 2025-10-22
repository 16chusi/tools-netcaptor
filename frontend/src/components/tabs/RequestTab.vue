<template>
  <div class="request-view">
    <div class="request-toolbar">
      <label>格式:</label>
      <select :value="format" @change="$emit('update:format', ($event.target as HTMLSelectElement).value)" class="format-select">
        <option value="curl">cURL</option>
        <option value="powershell">PowerShell</option>
        <option value="fetch">Fetch</option>
      </select>
      <button @click="$emit('copy')" class="copy-request-btn">复制</button>
    </div>
    <pre class="request-code">{{ code }}</pre>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  format: string
  code: string
}>()

defineEmits(['update:format', 'copy'])
</script>

<style scoped>
.request-view {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.request-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
  font-size: 11px;
}

.format-select {
  height: 24px;
  padding: 0 6px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #333333;
  border-radius: 2px;
  font-size: 11px;
  cursor: pointer;
}

.format-select:focus {
  outline: none;
  border-color: #1a73e8;
}

.copy-request-btn {
  padding: 4px 12px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #5f6368;
  border-radius: 2px;
  cursor: pointer;
  font-size: 11px;
}

.copy-request-btn:hover {
  background: #f8f9fa;
}

.request-code {
  flex: 1;
  margin: 0;
  padding: 12px;
  background: #f8f9fa;
  border: none;
  font-size: 12px;
  line-height: 1.6;
  overflow: auto;
  color: #333333;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  white-space: pre;
  text-align: left;
}
</style>
