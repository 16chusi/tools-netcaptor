<template>
  <div v-if="visible" class="drawer-overlay" @click="$emit('close')">
    <div class="drawer" @click.stop>
      <div class="drawer-header">
        <h3>⚙️ 设置</h3>
        <button @click="$emit('close')" class="close-icon">✕</button>
      </div>
      <div class="drawer-content">
        <div class="setting-item">
          <label>代理端口</label>
          <input 
            :value="proxyPort" 
            @input="$emit('update:proxyPort', Number(($event.target as HTMLInputElement).value))"
            type="number" 
            class="setting-input" 
            :disabled="proxyRunning"
            min="1024"
            max="65535"
          />
          <small>修改后需要重启代理</small>
        </div>
        
        <div class="setting-item">
          <label>默认浏览器</label>
          <select :value="selectedBrowser" @change="$emit('update:selectedBrowser', ($event.target as HTMLSelectElement).value)" class="setting-select">
            <option value="chrome">Chrome</option>
            <option value="edge">Edge</option>
            <option value="firefox">Firefox</option>
          </select>
        </div>
        
        <div class="setting-item">
          <label>默认打开 URL</label>
          <input :value="targetUrl" @input="$emit('update:targetUrl', ($event.target as HTMLInputElement).value)" class="setting-input" placeholder="留空使用测试服务器" />
          <small>留空将打开内置测试服务器（随机端口）</small>
        </div>
        
        <div class="setting-item">
          <label>下载路径</label>
          <div class="path-input-group">
            <input :value="downloadPath" class="setting-input" placeholder="默认下载目录" readonly />
            <button @click="$emit('selectPath')" class="browse-btn">浏览...</button>
          </div>
          <small>留空使用系统默认下载目录</small>
        </div>
        
        <div class="setting-item">
          <label>历史记录数量</label>
          <input 
            v-model.number="maxHistoryEntries" 
            @change="updateMaxHistory"
            type="number" 
            class="setting-input" 
            min="10"
            max="1000"
            step="10"
          />
          <small>保存最近的 N 条记录（默认30，范围10-1000）</small>
        </div>
        
        <div class="setting-item">
          <label>Webhook 服务</label>
          <div style="display: flex; gap: 8px; align-items: center;">
            <div class="info-box" style="flex: 1;">
              <span class="port-value">{{ webhookRunning ? webhookPort : '未启动' }}</span>
            </div>
            <button v-if="webhookRunning" @click="copyWebhookUrl" class="ws-btn ws-btn-copy">复制</button>
            <button v-if="!webhookRunning" @click="$emit('startWebhook')" class="ws-btn ws-btn-start">启动</button>
            <button v-else @click="$emit('stopWebhook')" class="ws-btn ws-btn-stop">停止</button>
          </div>
          <small>HTTP接口 + 测试页面（随机端口）</small>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import {GetMaxHistoryEntries, SetMaxHistoryEntries} from "../../wailsjs/go/main/NetworkApp";

const props = defineProps<{
  visible: boolean
  proxyPort: number
  proxyRunning: boolean
  selectedBrowser: string
  targetUrl: string
  downloadPath: string
  wsPort?: number
  wsRunning?: boolean
  webhookPort?: number
  webhookRunning?: boolean
}>()

const maxHistoryEntries = ref(30)

// 加载当前设置
watch(() => props.visible, async (visible) => {
  if (visible) {
    try {
      maxHistoryEntries.value = await GetMaxHistoryEntries()
    } catch (e) {
      console.error('获取历史记录数量失败:', e)
    }
  }
})

async function updateMaxHistory() {
  try {
    await SetMaxHistoryEntries(maxHistoryEntries.value)
    console.log('历史记录数量已更新:', maxHistoryEntries.value)
  } catch (e) {
    console.error('设置历史记录数量失败:', e)
    alert('设置失败: ' + e)
  }
}


const emit = defineEmits<{
  close: []
  'update:proxyPort': [value: number]
  'update:selectedBrowser': [value: string]
  'update:targetUrl': [value: string]
  selectPath: []
  startWebSocket: []
  stopWebSocket: []
  startWebhook: []
  stopWebhook: []
}>()

function copyWebhookUrl() {
  if (props.webhookPort) {
    const url = `http://127.0.0.1:${props.webhookPort}/webhook`
    navigator.clipboard.writeText(url).then(() => {
      alert('已复制: ' + url)
    })
  }
}
</script>

<style scoped>
.drawer-overlay {
  position: fixed;
   top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
}

.drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 600px;
  background: white;
  box-shadow: -2px 0 8px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #e0e0e0;
  background: #f8f9fa;
}

.drawer-header h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.close-icon {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: #5f6368;
  cursor: pointer;
  font-size: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background 0.2s;
}

.close-icon:hover {
  background: #e8eaed;
}

.drawer-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  text-align: left;
}

.setting-item {
  margin-bottom: 24px;
}

.setting-item label {
  display: block;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.setting-item small {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: #666;
}

.setting-input,
.setting-select {
  width: 100%;
  height: 36px;
  padding: 0 12px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #333333;
  border-radius: 4px;
  font-size: 13px;
}

.setting-input:focus,
.setting-select:focus {
  outline: none;
  border-color: #1a73e8;
  box-shadow: 0 0 0 2px rgba(26, 115, 232, 0.1);
}

.setting-input:disabled {
  background: #f8f9fa;
  color: #80868b;
  cursor: not-allowed;
}

.path-input-group {
  display: flex;
  gap: 8px;
}

.path-input-group .setting-input {
  flex: 1;
}

.browse-btn {
  padding: 0 16px;
  height: 36px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #5f6368;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  white-space: nowrap;
}

.browse-btn:hover {
  background: #f8f9fa;
  border-color: #1a73e8;
}

.info-box {
  padding: 10px 12px;
  background: #f8f9fa;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
}

.port-value {
  font-size: 16px;
  font-weight: 600;
  color: #1a73e8;
  font-family: 'Courier New', monospace;
}

.ws-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
}

.ws-btn-start {
  background: #1a73e8;
  color: white;
}

.ws-btn-start:hover {
  background: #1557b0;
}

.ws-btn-stop {
  background: #d93025;
  color: white;
}

.ws-btn-stop:hover {
  background: #b71c1c;
}

.ws-btn-copy {
  background: #34a853;
  color: white;
}

.ws-btn-copy:hover {
  background: #2d8e47;
}
</style>
