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
        <div class="filter-tabs">
          <button :class="['filter-tab', {active: filterType === 'all'}]" @click="filterType = 'all'">全部</button>
          <button :class="['filter-tab', {active: filterType === 'fetch'}]" @click="filterType = 'fetch'">Fetch/XHR</button>
          <button :class="['filter-tab', {active: filterType === 'js'}]" @click="filterType = 'js'">JS</button>
          <button :class="['filter-tab', {active: filterType === 'css'}]" @click="filterType = 'css'">CSS</button>
          <button :class="['filter-tab', {active: filterType === 'image'}]" @click="filterType = 'image'">图片</button>
          <button :class="['filter-tab', {active: filterType === 'document'}]" @click="filterType = 'document'">文档</button>
          <button :class="['filter-tab', {active: filterType === 'other'}]" @click="filterType = 'other'">其他</button>
        </div>
        <div class="divider"></div>
        <select v-model="selectedBrowser" class="browser-select">
          <option value="chrome">Chrome</option>
          <option value="edge">Edge</option>
          <option value="firefox">Firefox</option>
        </select>
        <button @click="openBrowser" class="open-btn">打开浏览器</button>
      </div>
      <div class="toolbar-right">
        <button @click="settingsVisible = true" class="icon-btn" title="设置">⚙️</button>
        <button @click="certDialogVisible = true" class="icon-btn" title="证书">🔒</button>
        <button @click="exportData" class="icon-btn" title="导出">⬇️</button>
      </div>
    </div>

    <SettingsDrawer
      :visible="settingsVisible"
      v-model:proxyPort="proxyPort"
      v-model:selectedBrowser="selectedBrowser"
      v-model:targetUrl="targetUrl"
      :downloadPath="downloadPath"
      :proxyRunning="proxyRunning"
      @close="settingsVisible = false"
      @selectPath="selectDownloadPath"
    />

    <CertDrawer
      :visible="certDialogVisible"
      :certPath="certPath"
      @close="certDialogVisible = false"
    />

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
          <div class="col-time">{{ entry.duration ? entry.duration + 'ms' : '-' }}</div>
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

    <div v-if="selectedEntry" class="details-panel" :style="{ height: detailsHeight + 'px' }">
      <div class="resize-handle" @mousedown="startResize"></div>
      <div class="details-tabs">
        <button :class="['tab', {active: activeTab === 'headers'}]" @click="activeTab = 'headers'">标头</button>
        <button :class="['tab', {active: activeTab === 'payload'}]" @click="activeTab = 'payload'">载荷</button>
        <button :class="['tab', {active: activeTab === 'response'}]" @click="activeTab = 'response'">响应</button>
        <button :class="['tab', {active: activeTab === 'request'}]" @click="activeTab = 'request'">请求</button>
      </div>
      <div class="details-content">
        <div v-if="activeTab === 'headers'" class="headers-view">
          <div class="section">
            <div class="section-header">
              <h4>常规</h4>
              <button @click="copyGeneralInfo" class="copy-icon" title="复制">📋</button>
            </div>
            <div class="kv-list">
              <div class="kv-item"><span class="key">Request URL:</span><span class="value">{{ selectedEntry.url }}</span></div>
              <div class="kv-item"><span class="key">Request Method:</span><span class="value">{{ selectedEntry.method }}</span></div>
              <div class="kv-item" v-if="selectedEntry.status"><span class="key">Status Code:</span><span class="value">{{ selectedEntry.status }} {{ selectedEntry.statusText }}</span></div>
            </div>
          </div>
          <div class="section" v-if="selectedEntry.request.headers">
            <div class="section-header">
              <h4>请求标头</h4>
              <button @click="copyRequestHeaders" class="copy-icon" title="复制">📋</button>
            </div>
            <div class="kv-list">
              <div class="kv-item" v-for="(value, key) in selectedEntry.request.headers" :key="key">
                <span class="key">{{ key }}:</span><span class="value">{{ value }}</span>
              </div>
            </div>
          </div>
          <div class="section" v-if="selectedEntry.response.headers">
            <div class="section-header">
              <h4>响应标头</h4>
              <button @click="copyResponseHeaders" class="copy-icon" title="复制">📋</button>
            </div>
            <div class="kv-list">
              <div class="kv-item" v-for="(value, key) in selectedEntry.response.headers" :key="key">
                <span class="key">{{ key }}:</span><span class="value">{{ value }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-if="activeTab === 'payload'" class="payload-view">
          <div v-if="selectedEntry.request.body" class="payload-container">
            <div class="payload-toolbar">
              <button @click="copyPayload" class="copy-payload-btn">复制</button>
            </div>
            <pre ref="payloadBlock" class="payload-preview"><code></code></pre>
          </div>
          <div v-else class="empty">无请求体</div>
        </div>
        <div v-if="activeTab === 'request'" class="request-view">
          <div class="request-toolbar">
            <label>格式:</label>
            <select v-model="requestFormat" class="format-select">
              <option value="curl">cURL</option>
              <option value="powershell">PowerShell</option>
              <option value="fetch">Fetch</option>
            </select>
            <button @click="copyRequest" class="copy-request-btn">复制</button>
          </div>
          <pre class="request-code">{{ generateRequestCode() }}</pre>
        </div>
        <div v-if="activeTab === 'response'" class="response-view">
          <div v-if="selectedEntry.response.body" class="response-container">
            <div class="response-toolbar">
              <div class="toolbar-left-section">
                <span class="format-label">格式: <strong>{{ getFormatType(selectedEntry.response?.contentType || '') }}</strong></span>
                <div class="view-mode">
                  <label>查看方式:</label>
                  <select :value="currentViewMode" @change="handleViewModeChange" class="view-select">
                    <option value="text">文本</option>
                    <option value="json">JSON</option>
                    <option value="html">HTML</option>
                    <option value="javascript">JavaScript</option>
                    <option value="css">CSS</option>
                    <option value="image">图片</option>
                    <option value="pdf">PDF</option>
                    <option value="hex">十六进制</option>
                    <option value="base64">Base64</option>
                  </select>
                </div>
              </div>
              <button @click="saveResponse" class="save-btn">下载</button>
            </div>
            <div v-if="currentViewMode === 'image'" class="image-preview">
              <img :src="selectedEntry.url" alt="Response Image" @error="handleImageError" />
            </div>
            <div v-else-if="currentViewMode === 'pdf'" class="pdf-preview">
              <div class="pdf-actions">
                <button @click="openInBrowser" class="action-btn">在浏览器中打开</button>
                <button @click="downloadAndOpen" class="action-btn">下载并打开</button>
              </div>
              <div class="pdf-info">
                <p>PDF 文件无法在应用内预览</p>
                <p>请使用上方按钮在浏览器中打开或下载后查看</p>
                <div class="url-container">
                  <code>{{ selectedEntry.url }}</code>
                  <button @click="copyUrl" class="copy-btn" title="复制链接">📋</button>
                </div>
              </div>
            </div>
            <pre v-else-if="currentViewMode === 'hex'" class="text-preview">{{ getHexContent() }}</pre>
            <pre v-else-if="currentViewMode === 'base64'" class="text-preview">{{ getBase64Content() }}</pre>
            <pre v-else ref="codeBlock" class="text-preview"><code></code></pre>
          </div>
          <div v-else class="empty">无响应体</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { GetAllEntries, ClearCapture, StartProxyWithPort, StopProxy, IsProxyRunning, GetProxyURL, GetCACertPath, OpenInChrome, OpenInEdge, OpenInFirefox, DownloadResponse, ExportData, SelectDownloadDirectory } from '../../wailsjs/go/main/NetworkApp'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import { getDomain, getPath, getResourceType, getStatusClass, formatSize, getFormatType } from '../utils/networkUtils'
import { generateCurl, generatePowerShell, generateFetch } from '../utils/codeGenerator'
import { formatJSON, toHex, toBase64 } from '../utils/formatters'
import SettingsDrawer from './SettingsDrawer.vue'
import CertDrawer from './CertDrawer.vue'
import hljs from 'highlight.js/lib/core'
import javascript from 'highlight.js/lib/languages/javascript'
import json from 'highlight.js/lib/languages/json'
import xml from 'highlight.js/lib/languages/xml'
import css from 'highlight.js/lib/languages/css'
import plaintext from 'highlight.js/lib/languages/plaintext'
import 'highlight.js/styles/github.css'
import { html as beautifyHtml, css as beautifyCss, js as beautifyJs } from 'js-beautify'

hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('json', json)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('css', css)
hljs.registerLanguage('plaintext', plaintext)

const entries = ref<any[]>([])
const selectedEntry = ref<any>(null)
const activeTab = ref('headers')
const filterText = ref('')
const proxyRunning = ref(false)
const proxyUrl = ref('')
const targetUrl = ref('')
const selectedBrowser = ref('chrome')
const proxyPort = ref(8888)
const certDialogVisible = ref(false)
const settingsVisible = ref(false)
const certPath = ref('')
const detailsHeight = ref(400)
const isResizing = ref(false)
const viewMode = ref('')
const requestFormat = ref('curl')
const downloadPath = ref('')
const filterType = ref('all')
let refreshInterval: any = null

const filteredEntries = computed(() => {
  let result = entries.value
  
  if (filterType.value !== 'all') {
    result = result.filter(e => {
      const type = getResourceType(e)
      if (filterType.value === 'fetch') return type === 'json' || e.type === 'fetch' || e.type === 'xhr'
      if (filterType.value === 'image') return type === 'image'
      if (filterType.value === 'document') return type === 'document'
      return type === filterType.value
    })
  }
  
  if (filterText.value) {
    const filter = filterText.value.toLowerCase()
    result = result.filter(e => 
      e.url.toLowerCase().includes(filter) ||
      e.method.toLowerCase().includes(filter)
    )
  }
  
  return result
})

const currentViewMode = computed(() => {
  if (viewMode.value) return viewMode.value
  
  const ct = selectedEntry.value?.response?.contentType || ''
  if (ct.includes('image/')) return 'image'
  if (ct.includes('application/pdf')) return 'pdf'
  if (ct.includes('json')) return 'json'
  if (ct.includes('javascript')) return 'javascript'
  if (ct.includes('html')) return 'html'
  if (ct.includes('css')) return 'css'
  if (ct.includes('text/')) return 'text'
  return 'hex'
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
  if (!url) {
    url = 'http://localhost:' + (await getTestPort())
  }
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

async function getTestPort() {
  try {
    const { GetTestServerPort } = await import('../../wailsjs/go/main/App')
    return await GetTestServerPort()
  } catch {
    return 9999
  }
}

async function selectDownloadPath() {
  try {
    const path = await SelectDownloadDirectory()
    if (path) {
      downloadPath.value = path
    }
  } catch (e) {
    console.error('Select directory failed:', e)
  }
}

function selectEntry(entry: any) {
  selectedEntry.value = entry
  viewMode.value = ''
}

function handleViewModeChange(e: Event) {
  const newMode = (e.target as HTMLSelectElement).value
  viewMode.value = newMode
}

async function exportData() {
  const data = {
    entries: filteredEntries.value,
    exportTime: new Date().toISOString(),
    filter: {
      type: filterType.value,
      text: filterText.value
    },
    total: filteredEntries.value.length
  }
  
  try {
    await ExportData(JSON.stringify(data, null, 2))
  } catch (e) {
    console.error('Export failed:', e)
  }
}

function handleImageError(e: Event) {
  console.error('Image load error:', e)
}

async function saveResponse() {
  const entry = selectedEntry.value
  if (!entry?.url) {
    alert('没有响应内容')
    return
  }
  
  try {
    let filename = 'download.txt'
    try {
      const urlPath = new URL(entry.url).pathname
      filename = urlPath.split('/').pop() || 'download.txt'
    } catch {
      filename = 'download.txt'
    }
    
    await DownloadResponse(entry.url, filename)
  } catch (e) {
    console.error('Download failed:', e)
    alert('下载失败: ' + e)
  }
}

function openInBrowser() {
  const entry = selectedEntry.value
  if (!entry?.url) return
  BrowserOpenURL(entry.url)
}

async function downloadAndOpen() {
  await saveResponse()
}

async function copyUrl() {
  const entry = selectedEntry.value
  if (!entry?.url) return
  
  try {
    await navigator.clipboard.writeText(entry.url)
  } catch (e) {
    console.error('Copy failed:', e)
  }
}

function generateRequestCode() {
  const entry = selectedEntry.value
  if (!entry) return ''
  
  if (requestFormat.value === 'curl') {
    return generateCurl(entry)
  } else if (requestFormat.value === 'powershell') {
    return generatePowerShell(entry)
  } else {
    return generateFetch(entry)
  }
}

async function copyRequest() {
  const code = generateRequestCode()
  if (!code) return
  
  try {
    await navigator.clipboard.writeText(code)
  } catch (e) {
    console.error('Copy failed:', e)
  }
}

function decodeHtml(html: string): string {
  const txt = document.createElement('textarea')
  txt.innerHTML = html
  return txt.value
}

function getHexContent(): string {
  if (!selectedEntry.value?.response?.body) return ''
  return toHex(decodeHtml(selectedEntry.value.response.body))
}

function getBase64Content(): string {
  if (!selectedEntry.value?.response?.body) return ''
  return toBase64(decodeHtml(selectedEntry.value.response.body))
}

async function copyGeneralInfo() {
  const entry = selectedEntry.value
  if (!entry) return
  
  const data = {
    'Request URL': entry.url,
    'Request Method': entry.method,
    'Status Code': entry.status ? `${entry.status} ${entry.statusText}` : undefined
  }
  
  try {
    await navigator.clipboard.writeText(JSON.stringify(data, null, 2))
  } catch (e) {
    console.error('Copy failed:', e)
  }
}

async function copyRequestHeaders() {
  const entry = selectedEntry.value
  if (!entry?.request?.headers) return
  
  try {
    await navigator.clipboard.writeText(JSON.stringify(entry.request.headers, null, 2))
  } catch (e) {
    console.error('Copy failed:', e)
  }
}

async function copyResponseHeaders() {
  const entry = selectedEntry.value
  if (!entry?.response?.headers) return
  
  try {
    await navigator.clipboard.writeText(JSON.stringify(entry.response.headers, null, 2))
  } catch (e) {
    console.error('Copy failed:', e)
  }
}

const codeBlock = ref<HTMLElement | null>(null)
const payloadBlock = ref<HTMLElement | null>(null)

watch([selectedEntry, activeTab], async () => {
  if (!selectedEntry.value?.request?.body) return
  if (activeTab.value !== 'payload') return
  
  await nextTick()
  if (!payloadBlock.value) return
  
  const codeEl = payloadBlock.value.querySelector('code')
  if (!codeEl) return
  
  let decoded = decodeHtml(selectedEntry.value.request.body)
  const ct = selectedEntry.value.request?.headers?.['Content-Type'] || selectedEntry.value.request?.headers?.['content-type'] || ''
  
  let lang = 'plaintext'
  let isText = true
  
  if (ct.includes('json')) {
    lang = 'json'
    try { decoded = JSON.stringify(JSON.parse(decoded), null, 2) } catch {}
  } else if (ct.includes('text/')) {
    lang = 'plaintext'
  } else if (ct.includes('x-www-form-urlencoded')) {
    lang = 'plaintext'
  } else {
    isText = false
    decoded = toHex(decoded)
  }
  
  payloadContent.value = decoded
  codeEl.textContent = decoded
  if (isText) {
    codeEl.className = `language-${lang}`
    delete (codeEl as any).dataset.highlighted
    hljs.highlightElement(codeEl as HTMLElement)
  }
}, { flush: 'post' })

async function copyPayload() {
  if (!payloadContent.value) return
  try {
    await navigator.clipboard.writeText(payloadContent.value)
  } catch (e) {
    console.error('Copy failed:', e)
  }
}

watch([selectedEntry, currentViewMode, activeTab], async () => {
  if (!selectedEntry.value?.response?.body) return
  if (activeTab.value !== 'response') return
  if (currentViewMode.value === 'hex' || currentViewMode.value === 'base64' || currentViewMode.value === 'image' || currentViewMode.value === 'pdf') return
  
  await nextTick()
  if (!codeBlock.value) return
  
  const codeEl = codeBlock.value.querySelector('code')
  if (!codeEl) return
  
  let decoded = decodeHtml(selectedEntry.value.response.body)
  const mode = currentViewMode.value
  
  let lang = 'plaintext'
  
  if (mode === 'json') {
    lang = 'json'
    try { decoded = JSON.stringify(JSON.parse(decoded), null, 2) } catch {}
  } else if (mode === 'javascript') {
    lang = 'javascript'
    decoded = beautifyJs(decoded, { indent_size: 2 })
  } else if (mode === 'html') {
    lang = 'html'
    decoded = beautifyHtml(decoded, { indent_size: 2 })
  } else if (mode === 'css') {
    lang = 'css'
    decoded = beautifyCss(decoded, { indent_size: 2 })
  } else if (mode === 'text') {
    lang = 'plaintext'
  }
  
  codeEl.textContent = decoded
  codeEl.className = `language-${lang}`
  delete (codeEl as any).dataset.highlighted
  hljs.highlightElement(codeEl as HTMLElement)
})

const payloadContent = ref('')

function startResize(e: MouseEvent) {
  isResizing.value = true
  const startY = e.clientY
  const startHeight = detailsHeight.value
  
  const onMouseMove = (e: MouseEvent) => {
    if (!isResizing.value) return
    const delta = startY - e.clientY
    detailsHeight.value = Math.max(100, Math.min(window.innerHeight - 200, startHeight + delta))
  }
  
  const onMouseUp = () => {
    isResizing.value = false
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  }
  
  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
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
  text-align: left;
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
  min-height: 200px;
  max-height: 80vh;
  border-top: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  position: relative;
}

.resize-handle {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  cursor: ns-resize;
  background: transparent;
  z-index: 10;
}

.resize-handle:hover {
  background: #1a73e8;
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

.payload-view {
  height: 100%;
}

.payload-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.payload-toolbar {
  display: flex;
  justify-content: flex-end;
  padding: 8px;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
}

.copy-payload-btn {
  padding: 4px 12px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #5f6368;
  border-radius: 2px;
  cursor: pointer;
  font-size: 11px;
}

.copy-payload-btn:hover {
  background: #f8f9fa;
}

.payload-preview {
  flex: 1;
  margin: 0;
  padding: 12px;
  background: #f8f9fa;
  border: none;
  font-size: 12px;
  line-height: 1.6;
  overflow: auto;
  color: #333333;
  font-family: 'Consolas', 'Monaco', monospace;
  text-align: left;
}

.payload-preview code {
  display: block;
  white-space: pre;
  word-wrap: normal;
  overflow-x: auto;
  text-align: left;
}



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

.response-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.response-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
}

.toolbar-left-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.format-label {
  font-size: 11px;
  color: #5f6368;
}

.format-label strong {
  color: #1a73e8;
  font-weight: 600;
}

.view-mode {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: #5f6368;
}

.view-select {
  height: 24px;
  padding: 0 6px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #333333;
  border-radius: 2px;
  font-size: 11px;
  cursor: pointer;
}

.view-select:focus {
  outline: none;
  border-color: #1a73e8;
}

.save-btn {
  padding: 4px 12px;
  border: 1px solid #dadce0;
  background: #ffffff;
  color: #5f6368;
  border-radius: 2px;
  cursor: pointer;
  font-size: 11px;
}

.save-btn:hover {
  background: #f8f9fa;
}

.image-preview {
  flex: 1;
  overflow: auto;
  padding: 12px;
  background: #f8f9fa;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.pdf-preview {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  background: #f8f9fa;
}

.pdf-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
}

.action-btn {
  padding: 10px 20px;
  border: 1px solid #1a73e8;
  background: #1a73e8;
  color: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #1557b0;
  border-color: #1557b0;
}

.pdf-info {
  text-align: center;
  color: #5f6368;
}

.pdf-info p {
  margin: 8px 0;
  font-size: 13px;
  line-height: 1.6;
}

.url-container {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  max-width: 600px;
}

.pdf-info code {
  flex: 1;
  padding: 12px;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  font-size: 11px;
  color: #1a73e8;
  word-break: break-all;
}

.copy-btn {
  width: 36px;
  height: 36px;
  border: 1px solid #dadce0;
  background: white;
  cursor: pointer;
  border-radius: 4px;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
}

.copy-btn:hover {
  background: #f8f9fa;
  border-color: #1a73e8;
}

.json-preview,
.text-preview {
  flex: 1;
  margin: 0;
  padding: 12px;
  background: #f8f9fa;
  border: none;
  font-size: 12px;
  line-height: 1.6;
  overflow: auto;
  color: #333333;
  font-family: 'Consolas', 'Monaco', monospace;
  text-align: left;
}

.text-preview code {
  display: block;
  white-space: pre;
  word-wrap: normal;
  overflow-x: auto;
  text-align: left;
}





.empty {
  padding: 40px;
  text-align: center;
  color: #80868b;
}

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





























.filter-tabs {
  display: flex;
  gap: 4px;
}

.filter-tab {
  padding: 4px 10px;
  border: none;
  background: transparent;
  color: #5f6368;
  cursor: pointer;
  font-size: 11px;
  border-radius: 2px;
  transition: all 0.1s;
}

.filter-tab:hover {
  background: #e8eaed;
}

.filter-tab.active {
  background: #e8f0fe;
  color: #1a73e8;
  font-weight: 600;
}
</style>
