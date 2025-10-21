<template>
  <div class="devtools-network">
    <div class="toolbar">
      <div class="toolbar-left">
        <button @click="toggleProxy" :class="['icon-btn', proxyRunning ? 'recording' : '']" :title="proxyRunning ? '停止代理' : '启动代理'">
          <span v-if="proxyRunning" class="record-dot"></span>
          <span v-else>●</span>
        </button>
        <button @click="clearAll" class="icon-btn" title="清空">🗑️</button>
        <input v-model="filterText" placeholder="Filter" class="filter-input" />
        <div class="divider"></div>
        <input v-model="targetUrl" placeholder="http://localhost:9999" class="url-input" />
        <select v-model="selectedBrowser" class="browser-select">
          <option value="chrome">Chrome</option>
          <option value="edge">Edge</option>
          <option value="firefox">Firefox</option>
        </select>
        <button @click="openBrowser" class="open-btn">打开</button>
      </div>
      <div class="toolbar-right">
        <span class="proxy-label">Proxy:</span>
        <input 
          v-model.number="proxyPort" 
          type="number" 
          class="proxy-port-input" 
          :disabled="proxyRunning"
          min="1024"
          max="65535"
        />
        <button @click="showCertDialog" class="icon-btn" title="证书">🔒</button>
        <button @click="exportData" class="icon-btn" title="导出">⬇️</button>
      </div>
    </div>

    <div v-if="certDialogVisible" class="cert-dialog-overlay" @click="certDialogVisible = false">
      <div class="cert-dialog" @click.stop>
        <h3>🔒 NetCaptor HTTPS 证书</h3>
        <p>要捕获 HTTPS 请求，需要安装并信任此 CA 证书。</p>
        <div class="cert-path">
          <strong>证书位置:</strong>
          <code>{{ certPath }}</code>
        </div>
        <div class="install-steps">
          <h4>Windows 安装步骤:</h4>
          <ol>
            <li>双击 goproxy-ca.crt 文件</li>
            <li>点击“安装证书”</li>
            <li>选择“当前用户”</li>
            <li>选择“将所有证书放入下列存储”</li>
            <li>浏览并选择“受信任的根证书颁发机构”</li>
            <li>完成安装并重启浏览器</li>
          </ol>
          <h4>macOS 安装步骤:</h4>
          <ol>
            <li>双击 goproxy-ca.crt 文件</li>
            <li>在钥匙串访问中找到 "NetCaptor CA"</li>
            <li>双击证书，展开“信任”</li>
            <li>选择“始终信任”</li>
            <li>重启浏览器</li>
          </ol>
          <h4>Linux/Ubuntu 安装步骤:</h4>
          <p><strong>方法1: 系统级安装（推荐）</strong></p>
          <pre>sudo cp {{ certPath }} /usr/local/share/ca-certificates/goproxy.crt
sudo update-ca-certificates
# 重启浏览器</pre>
          <p><strong>方法2: Chrome/Chromium 专用</strong></p>
          <pre>mkdir -p $HOME/.pki/nssdb
certutil -d sql:$HOME/.pki/nssdb -A -t "C,," -n "GoProxy CA" -i {{ certPath }}
# 如果没有certutil，先安装: sudo apt install libnss3-tools
# 重启浏览器</pre>
          <p><strong>方法3: 临时测试（不推荐）</strong></p>
          <p>启动Chrome时添加参数忽略证书错误：</p>
          <pre>google-chrome --proxy-server="127.0.0.1:8888" --ignore-certificate-errors</pre>
          <h4>Chrome 快捷方式:</h4>
          <ol>
            <li>打开 Chrome 设置</li>
            <li>搜索“证书”</li>
            <li>点击“管理证书”</li>
            <li>选择“受信任的根证书颁发机构”</li>
            <li>点击“导入”，选择 goproxy-ca.crt</li>
            <li>重启浏览器</li>
          </ol>
        </div>
        <button @click="certDialogVisible = false" class="close-btn">关闭</button>
      </div>
    </div>

    <div class="network-table">
      <div class="table-header">
        <div class="col-name">名称</div>
        <div class="col-path">路径</div>
        <div class="col-status">状态</div>
        <div class="col-type">类型</div>
        <div class="col-size">大小</div>
        <div class="col-time">时间</div>
      </div>
      <div class="table-body">
        <div 
          v-for="entry in filteredEntries" 
          :key="entry.id"
          :class="['table-row', {selected: selectedEntry?.id === entry.id}]"
          @click="selectEntry(entry)"
        >
          <div class="col-name">
            <span class="method" :class="'method-' + entry.method.toLowerCase()">{{ entry.method }}</span>
            <span class="url-name" :title="entry.url">{{ getDomain(entry.url) }}</span>
          </div>
          <div class="col-path">
            <span class="path-text" :title="entry.url">{{ getPath(entry.url) }}</span>
          </div>
          <div class="col-status">
            <span v-if="entry.status" :class="'status-' + getStatusClass(entry.status)">{{ entry.status }}</span>
            <span v-else class="status-pending">pending</span>
          </div>
          <div class="col-type">{{ getResourceType(entry) }}</div>
          <div class="col-size">{{ formatSize(entry.size) }}</div>
          <div class="col-time">{{ entry.duration || '-' }}ms</div>
        </div>
        <div v-if="filteredEntries.length === 0" class="empty-state">
          <div v-if="!proxyRunning">
            <h3>🚀 开始使用</h3>
            <ol>
              <li>点击 ▶️ 启动代理</li>
              <li>配置浏览器代理为 <code>127.0.0.1:8888</code></li>
              <li>访问 <strong>HTTP</strong> 网站，请求会在这里显示</li>
            </ol>
            <p style="color: #1a73e8; margin-top: 15px;">
              🔒 支持 HTTPS 抓包：点击工具栏的 🔒 按钮安装 CA 证书。
            </p>
          </div>
          <div v-else>
            <p>等待网络请求...</p>
            <small>在浏览器中访问 <strong>HTTP</strong> 网站后，请求会自动显示</small>
            <p style="color: #1a73e8; margin-top: 10px; font-size: 11px;">
              🔒 要抓取 HTTPS，请先安装 CA 证书
            </p>
          </div>
        </div>
      </div>
    </div>

    <div v-if="selectedEntry" class="details-panel">
      <div class="details-tabs">
        <button :class="['tab', {active: activeTab === 'headers'}]" @click="activeTab = 'headers'">Headers</button>
        <button :class="['tab', {active: activeTab === 'payload'}]" @click="activeTab = 'payload'">Payload</button>
        <button :class="['tab', {active: activeTab === 'response'}]" @click="activeTab = 'response'">Response</button>
      </div>
      <div class="details-content">
        <div v-if="activeTab === 'headers'" class="headers-view">
          <div class="section">
            <h4>General</h4>
            <div class="kv-list">
              <div class="kv-item"><span class="key">Request URL:</span><span class="value">{{ selectedEntry.url }}</span></div>
              <div class="kv-item"><span class="key">Request Method:</span><span class="value">{{ selectedEntry.method }}</span></div>
              <div class="kv-item" v-if="selectedEntry.status"><span class="key">Status Code:</span><span class="value">{{ selectedEntry.status }} {{ selectedEntry.statusText }}</span></div>
            </div>
          </div>
          <div class="section" v-if="selectedEntry.request.headers">
            <h4>Request Headers</h4>
            <div class="kv-list">
              <div class="kv-item" v-for="(value, key) in selectedEntry.request.headers" :key="key">
                <span class="key">{{ key }}:</span><span class="value">{{ value }}</span>
              </div>
            </div>
          </div>
          <div class="section" v-if="selectedEntry.response.headers">
            <h4>Response Headers</h4>
            <div class="kv-list">
              <div class="kv-item" v-for="(value, key) in selectedEntry.response.headers" :key="key">
                <span class="key">{{ key }}:</span><span class="value">{{ value }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-if="activeTab === 'payload'" class="payload-view">
          <pre v-if="selectedEntry.request.body">{{ selectedEntry.request.body }}</pre>
          <div v-else class="empty">无请求体</div>
        </div>
        <div v-if="activeTab === 'response'" class="response-view">
          <pre v-if="selectedEntry.response.body">{{ selectedEntry.response.body }}</pre>
          <div v-else class="empty">无响应体</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { GetAllEntries, ClearCapture, StartProxy, StartProxyWithPort, StopProxy, IsProxyRunning, GetProxyURL, GetCACertPath, OpenInChrome, OpenInEdge, OpenInFirefox } from '../../wailsjs/go/main/NetworkApp'

const entries = ref<any[]>([])
const selectedEntry = ref<any>(null)
const activeTab = ref('headers')
const filterText = ref('')
const proxyRunning = ref(false)
const proxyUrl = ref('')
const targetUrl = ref('http://localhost:9999')
const selectedBrowser = ref('chrome')
const proxyPort = ref(8888)
const certDialogVisible = ref(false)
const certPath = ref('')
let refreshInterval: any = null

const filteredEntries = computed(() => {
  if (!filterText.value) return entries.value
  const filter = filterText.value.toLowerCase()
  return entries.value.filter(e => 
    e.url.toLowerCase().includes(filter) ||
    e.method.toLowerCase().includes(filter)
  )
})

onMounted(async () => {
  proxyRunning.value = await IsProxyRunning()
  if (proxyRunning.value) {
    proxyUrl.value = await GetProxyURL()
  }
  certPath.value = await GetCACertPath()
  refreshInterval = setInterval(refreshData, 500)
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
})

async function refreshData() {
  entries.value = await GetAllEntries()
}

async function toggleProxy() {
  if (proxyRunning.value) {
    await StopProxy()
    proxyRunning.value = false
    proxyUrl.value = ''
  } else {
    try {
      await StartProxyWithPort(proxyPort.value)
      proxyRunning.value = true
      proxyUrl.value = await GetProxyURL()
    } catch (e: any) {
      alert(`启动代理失败: ${e}`)
    }
  }
}

async function clearAll() {
  await ClearCapture()
  entries.value = []
  selectedEntry.value = null
}

async function openBrowser() {
  let url = targetUrl.value
  if (!url.startsWith('http')) url = 'http://' + url
  
  try {
    if (selectedBrowser.value === 'chrome') {
      await OpenInChrome(url)
    } else if (selectedBrowser.value === 'edge') {
      await OpenInEdge(url)
    } else if (selectedBrowser.value === 'firefox') {
      await OpenInFirefox(url)
    }
  } catch (e: any) {
    alert(`打开浏览器失败: ${e}`)
  }
}

function selectEntry(entry: any) {
  selectedEntry.value = entry
  activeTab.value = 'headers'
}

function exportData() {
  const data = { entries: entries.value, exportTime: new Date().toISOString() }
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `network-${Date.now()}.json`
  a.click()
}

function getDomain(url: string) {
  try {
    const u = new URL(url)
    return u.hostname
  } catch {
    return url
  }
}

function getPath(url: string) {
  try {
    const u = new URL(url)
    let path = u.pathname
    // 如果有查询参数，也显示出来
    if (u.search) {
      path += u.search
    }
    return path || '/'
  } catch {
    return url
  }
}

function getResourceType(entry: any) {
  const ct = entry.response?.contentType || ''
  if (ct.includes('javascript')) return 'js'
  if (ct.includes('css')) return 'css'
  if (ct.includes('html')) return 'document'
  if (ct.includes('json')) return 'json'
  if (ct.includes('image')) return 'image'
  if (ct.includes('font')) return 'font'
  return 'other'
}

function getStatusClass(status: number) {
  if (status >= 200 && status < 300) return 'success'
  if (status >= 300 && status < 400) return 'redirect'
  if (status >= 400 && status < 500) return 'client-error'
  return 'server-error'
}

function formatSize(bytes: number) {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

function showCertDialog() {
  certDialogVisible.value = true
}
</script>

<style scoped>
.devtools-network {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #ffffff;
  color: #333333;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  font-size: 12px;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 8px;
  background: #f3f3f3;
  border-bottom: 1px solid #d0d0d0;
  height: 36px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-btn {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: #5f6368;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 2px;
  transition: background 0.1s;
}

.icon-btn:hover {
  background: #e8eaed;
}

.icon-btn.recording {
  color: #d93025;
}

.record-dot {
  width: 12px;
  height: 12px;
  background: #d93025;
  border-radius: 50%;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.open-btn {
  padding: 0 12px;
  height: 24px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #5f6368;
  border-radius: 2px;
  cursor: pointer;
  font-size: 11px;
  transition: all 0.1s;
}

.open-btn:hover {
  background: #f8f9fa;
  border-color: #c6c6c6;
}

.browser-select {
  height: 24px;
  padding: 0 6px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #333333;
  border-radius: 2px;
  font-size: 11px;
  cursor: pointer;
}

.browser-select:focus {
  outline: none;
  border-color: #1a73e8;
}

.proxy-label {
  font-size: 11px;
  color: #5f6368;
  white-space: nowrap;
}

.proxy-port-input {
  width: 60px;
  height: 24px;
  padding: 0 6px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #333333;
  border-radius: 2px;
  font-size: 11px;
  text-align: center;
}

.proxy-port-input:disabled {
  background: #f8f9fa;
  color: #80868b;
}

.proxy-port-input:focus {
  outline: none;
  border-color: #1a73e8;
  box-shadow: 0 0 0 1px #1a73e8;
}

.divider {
  width: 1px;
  height: 16px;
  background: #dadce0;
  margin: 0 6px;
}

.filter-input {
  width: 200px;
  height: 24px;
  padding: 0 8px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #333333;
  border-radius: 2px;
  font-size: 11px;
}

.filter-input:focus {
  outline: none;
  border-color: #1a73e8;
  box-shadow: 0 0 0 1px #1a73e8;
}

.url-input {
  flex: 1;
  height: 24px;
  padding: 0 8px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #333333;
  border-radius: 2px;
  font-size: 11px;
}

.url-input:focus {
  outline: none;
  border-color: #1a73e8;
  box-shadow: 0 0 0 1px #1a73e8;
}



.network-table {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.table-header {
  display: flex;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
  padding: 4px 8px;
  font-weight: 600;
  color: #5f6368;
}

.table-body {
  flex: 1;
  overflow-y: auto;
}

.table-row {
  display: flex;
  padding: 4px 8px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  align-items: center;
  color: #333333;
}

.table-row:hover {
  background: #f8f9fa;
}

.table-row.selected {
  background: #e8f0fe;
}

.col-name {
  width: 200px;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.col-path {
  flex: 1;
  min-width: 0;
}

.col-status { width: 60px; }
.col-type { width: 80px; }
.col-size { width: 80px; text-align: right; }
.col-time { width: 80px; text-align: right; }

.method {
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: 600;
  font-size: 10px;
  text-transform: uppercase;
}

.method-get { background: #e3f2fd; color: #1976d2; }
.method-post { background: #e8f5e9; color: #388e3c; }
.method-put { background: #fff3e0; color: #f57c00; }
.method-delete { background: #ffebee; color: #d32f2f; }

.url-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.path-text {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #5f6368;
  font-size: 11px;
}

.status-success { color: #0f9d58; }
.status-redirect { color: #f57c00; }
.status-client-error { color: #d32f2f; }
.status-server-error { color: #c62828; }
.status-pending { color: #999; }

.empty-state {
  padding: 60px 40px;
  text-align: center;
  color: #80868b;
}

.empty-state h3 {
  margin: 0 0 20px 0;
  font-size: 18px;
  color: #5f6368;
}

.empty-state ol {
  text-align: left;
  display: inline-block;
  margin: 0;
  padding-left: 20px;
}

.empty-state li {
  margin: 10px 0;
  line-height: 1.6;
}

.empty-state code {
  background: #f1f3f4;
  color: #d93025;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}

.details-panel {
  height: 300px;
  border-top: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
  background: #ffffff;
}

.details-tabs {
  display: flex;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
}

.tab {
  padding: 8px 16px;
  border: none;
  background: transparent;
  color: #5f6368;
  cursor: pointer;
  font-size: 12px;
  border-bottom: 2px solid transparent;
}

.tab.active {
  border-bottom-color: #1a73e8;
  color: #1a73e8;
}

.details-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.section {
  margin-bottom: 20px;
}

.section h4 {
  margin: 0 0 8px 0;
  font-size: 13px;
  font-weight: 600;
  color: #5f6368;
}

.kv-list {
  font-family: monospace;
  font-size: 11px;
}

.kv-item {
  display: flex;
  padding: 2px 0;
  line-height: 1.6;
}

.kv-item .key {
  color: #881280;
  min-width: 150px;
  flex-shrink: 0;
}

.kv-item .value {
  color: #1a1aa6;
  word-break: break-all;
}

.payload-view pre,
.response-view pre {
  margin: 0;
  padding: 12px;
  background: #f8f9fa;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  font-size: 11px;
  line-height: 1.5;
  overflow-x: auto;
  color: #333333;
}

.empty {
  padding: 40px;
  text-align: center;
  color: #80868b;
}

.cert-dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.cert-dialog {
  background: white;
  padding: 24px;
  border-radius: 8px;
  max-width: 600px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.cert-dialog h3 {
  margin: 0 0 16px 0;
  color: #333;
}

.cert-dialog p {
  color: #666;
  margin-bottom: 16px;
}

.cert-path {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 4px;
  margin-bottom: 16px;
}

.cert-path code {
  display: block;
  margin-top: 8px;
  padding: 8px;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  font-size: 12px;
  word-break: break-all;
}

.install-steps h4 {
  margin: 16px 0 8px 0;
  color: #333;
  font-size: 14px;
}

.install-steps ol {
  margin: 0 0 16px 0;
  padding-left: 20px;
  font-size: 13px;
  line-height: 1.6;
}

.install-steps pre {
  background: #2d2d30;
  color: #d4d4d4;
  padding: 12px;
  border-radius: 4px;
  font-size: 11px;
  overflow-x: auto;
}

.close-btn {
  width: 100%;
  padding: 10px;
  background: #1a73e8;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  margin-top: 16px;
}

.close-btn:hover {
  background: #1557b0;
}
</style>
