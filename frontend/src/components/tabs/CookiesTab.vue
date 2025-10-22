<template>
  <div class="cookies-view">
    <div class="section">
      <div class="section-header">
        <h4>请求 Cookies</h4>
        <button @click="$emit('copyRequestCookies')" class="copy-icon" title="复制">📋</button>
      </div>
      <div v-if="requestCookies.length > 0" class="cookie-list">
        <div class="cookie-item" v-for="(cookie, index) in requestCookies" :key="index">
          <div class="cookie-name">{{ cookie.name }}</div>
          <div class="cookie-value">{{ cookie.value }}</div>
        </div>
      </div>
      <div v-else class="empty">无 Cookie</div>
    </div>
    <div class="section">
      <div class="section-header">
        <h4>响应 Cookies</h4>
        <button @click="$emit('copyResponseCookies')" class="copy-icon" title="复制">📋</button>
      </div>
      <div v-if="responseCookies.length > 0" class="cookie-list">
        <div class="cookie-item" v-for="(cookie, index) in responseCookies" :key="index">
          <div class="cookie-name">{{ cookie.name }}</div>
          <div class="cookie-value">{{ cookie.value }}</div>
        </div>
      </div>
      <div v-else class="empty">无 Set-Cookie</div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  requestCookies: Array<{ name: string; value: string }>
  responseCookies: Array<{ name: string; value: string }>
}>()

defineEmits<{
  copyRequestCookies: []
  copyResponseCookies: []
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

.cookie-list {
  font-family: monospace;
  font-size: 11px;
}

.cookie-item {
  display: flex;
  padding: 2px 0;
  line-height: 1.6;
  gap: 8px;
}

.cookie-name {
  color: #881280;
  min-width: 180px;
  flex-shrink: 0;
  text-align: right;
  padding-right: 4px;
}

.cookie-value {
  flex: 1;
  color: #1a1aa6;
  word-break: break-all;
  text-align: left;
}

.empty {
  padding: 20px;
  text-align: center;
  color: #80868b;
}
</style>
