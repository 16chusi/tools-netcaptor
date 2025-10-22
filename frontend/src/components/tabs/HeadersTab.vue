<template>
  <div class="headers-view">
    <div class="section">
      <div class="section-header">
        <h4>常规</h4>
        <button @click="$emit('copyGeneral')" class="copy-icon" title="复制">📋</button>
      </div>
      <div class="kv-list">
        <div class="kv-item"><span class="key">Request URL:</span><span class="value">{{ entry.url }}</span></div>
        <div class="kv-item"><span class="key">Request Method:</span><span class="value">{{ entry.method }}</span></div>
        <div class="kv-item" v-if="entry.status"><span class="key">Status Code:</span><span class="value">{{ entry.status }} {{ entry.statusText }}</span></div>
      </div>
    </div>
    <div class="section" v-if="entry.request.headers">
      <div class="section-header">
        <h4>请求标头</h4>
        <button @click="$emit('copyRequestHeaders')" class="copy-icon" title="复制">📋</button>
      </div>
      <div class="kv-list">
        <div class="kv-item" v-for="(value, key) in entry.request.headers" :key="key">
          <span class="key">{{ key }}:</span><span class="value">{{ value }}</span>
        </div>
      </div>
    </div>
    <div class="section" v-if="entry.response.headers">
      <div class="section-header">
        <h4>响应标头</h4>
        <button @click="$emit('copyResponseHeaders')" class="copy-icon" title="复制">📋</button>
      </div>
      <div class="kv-list">
        <div class="kv-item" v-for="(value, key) in entry.response.headers" :key="key">
          <span class="key">{{ key }}:</span><span class="value">{{ value }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  entry: any
}>()

defineEmits<{
  copyGeneral: []
  copyRequestHeaders: []
  copyResponseHeaders: []
}>()
</script>

<style scoped>
.section {
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.section h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #5f6368;
  text-align: left;
}

.copy-icon {
  width: 20px;
  height: 20px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 2px;
  transition: background 0.2s;
  padding: 0;
}

.copy-icon:hover {
  background: #e8eaed;
}

.kv-list {
  font-family: monospace;
  font-size: 11px;
}

.kv-item {
  display: flex;
  padding: 2px 0;
  line-height: 1.6;
  gap: 8px;
}

.kv-item .key {
  color: #881280;
  min-width: 180px;
  flex-shrink: 0;
  text-align: right;
  padding-right: 4px;
}

.kv-item .value {
  color: #1a1aa6;
  word-break: break-all;
  text-align: left;
  flex: 1;
}
</style>
