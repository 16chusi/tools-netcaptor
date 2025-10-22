<template>
  <div class="devtools-network">
    <div class="toolbar">
      <div class="toolbar-left">
        <button @click="toggleProxy" :class="['icon-btn', proxyRunning ? 'recording' : '']"
                :title="proxyRunning ? '停止代理' : '启动代理'">
          <span v-if="proxyRunning" class="record-dot"></span>
          <span v-else>●</span>
        </button>
        <button @click="clearAll" class="icon-btn" title="清空">🗑️</button>
        <input v-model="filterText" placeholder="Filter" class="filter-input"/>
        <div class="divider"></div>
        <div class="filter-tabs">
          <button :class="['filter-tab', {active: filterType === 'all'}]" @click="filterType = 'all'">全部</button>
          <button :class="['filter-tab', {active: filterType === 'fetch'}]" @click="filterType = 'fetch'">Fetch/XHR
          </button>
          <button :class="['filter-tab', {active: filterType === 'js'}]" @click="filterType = 'js'">JS</button>
          <button :class="['filter-tab', {active: filterType === 'css'}]" @click="filterType = 'css'">CSS</button>
          <button :class="['filter-tab', {active: filterType === 'image'}]" @click="filterType = 'image'">图片</button>
          <button :class="['filter-tab', {active: filterType === 'document'}]" @click="filterType = 'document'">文档
          </button>
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
        <button @click="interceptVisible = true" class="icon-btn" title="拦截">🔧</button>
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

    <InterceptDrawer
        :visible="interceptVisible"
        :rules="interceptRules"
        @close="interceptVisible = false"
        @toggle="toggleRule"
        @edit="editRule"
        @delete="deleteRule"
        @import="importRules"
        @export="exportRules"
        @create="createRule"
    />

    <InterceptRuleEditor
        :visible="editorVisible"
        :initialRule="editingRule"
        @close="editorVisible = false"
        @save="saveRule"
    />

    <div class="network-table">
      <div class="table-header">
        <div class="col-name">名称</div>
        <div class="col-path">路径</div>
        <div class="col-status">状态</div>
        <div class="col-type">类型</div>
        <div class="col-size">大小</div>
        <div class="col-time">时间</div>
        <div class="col-action">操作</div>
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
          <div class="col-action">
            <button @click.stop="addInterceptRule(entry)" class="add-rule-btn" title="添加拦截规则">+拦截</button>
          </div>
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
        <button :class="['tab', {active: activeTab === 'cookies'}]" @click="activeTab = 'cookies'">Cookies</button>
      </div>
      <div class="details-content">
        <HeadersTab
            v-if="activeTab === 'headers'"
            :entry="selectedEntry"
            @copyGeneral="copyGeneralInfo"
            @copyRequestHeaders="copyRequestHeaders"
            @copyResponseHeaders="copyResponseHeaders"
        />
        <PayloadTab
            v-if="activeTab === 'payload'"
            :entry="selectedEntry"
            @copy="copyPayload"
        />
        <RequestTab
            v-if="activeTab === 'request'"
            v-model:format="requestFormat"
            :code="generateRequestCode()"
            @copy="copyRequest"
        />
        <CookiesTab
            v-if="activeTab === 'cookies'"
            :requestCookies="requestCookies"
            :responseCookies="responseCookies"
            @copyRequestCookies="copyRequestCookies"
            @copyResponseCookies="copyResponseCookies"
        />
        <ResponseTab
            v-if="activeTab === 'response'"
            :entry="selectedEntry"
            :viewMode="currentViewMode"
            :hexContent="getHexContent()"
            :base64Content="getBase64Content()"
            @changeViewMode="handleViewModeChange"
            @save="saveResponse"
            @imageError="handleImageError"
            @openInBrowser="openInBrowser"
            @downloadAndOpen="downloadAndOpen"
            @copyUrl="copyUrl"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from 'vue'
import {
  ClearCapture,
  DownloadResponse,
  ExportData,
  GetAllEntries,
  GetCACertPath,
  GetProxyURL,
  IsProxyRunning,
  OpenInChrome,
  OpenInEdge,
  OpenInFirefox,
  SelectDownloadDirectory,
  SetInterceptRules,
  StartProxyWithPort,
  StopProxy
} from '../../wailsjs/go/main/NetworkApp'
import {BrowserOpenURL} from '../../wailsjs/runtime/runtime'
import {ShowErrorDialog, ShowInfoDialog, ShowQuestionDialog} from '../../wailsjs/go/main/NetworkApp'
import {formatSize, getDomain, getFormatType, getPath, getResourceType, getStatusClass} from '../utils/networkUtils'
import {generateCurl, generateFetch, generatePowerShell} from '../utils/codeGenerator'
import {toBase64, toHex} from '../utils/formatters'
import SettingsDrawer from './SettingsDrawer.vue'
import CertDrawer from './CertDrawer.vue'
import HeadersTab from './tabs/HeadersTab.vue'
import CookiesTab from './tabs/CookiesTab.vue'
import PayloadTab from './tabs/PayloadTab.vue'
import RequestTab from './tabs/RequestTab.vue'
import ResponseTab from './tabs/ResponseTab.vue'
import InterceptDrawer from './InterceptDrawer.vue'
import InterceptRuleEditor from './InterceptRuleEditor.vue'
import { parseRequestCookies, parseResponseCookies } from '../utils/cookieUtils'
import { copyJSON, copyToClipboard } from '../utils/clipboardUtils'
import type { InterceptRule } from '../types/intercept'


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
const interceptVisible = ref(false)
const editorVisible = ref(false)
const interceptRules = ref<InterceptRule[]>([])
const editingRule = ref<InterceptRule | undefined>(undefined)
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
      ShowErrorDialog('错误', `启动代理失败: ${e}`)
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
    ShowErrorDialog('错误', `打开浏览器失败: ${e}`)
  }
}

async function getTestPort() {
  try {
    const {GetTestServerPort} = await import('../../wailsjs/go/main/App')
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
    ShowInfoDialog('提示', '没有响应内容')
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
    ShowErrorDialog('错误', '下载失败: ' + e)
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
  if (!selectedEntry.value?.url) return
  await copyToClipboard(selectedEntry.value.url)
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
  await copyToClipboard(code)
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
  await copyJSON({
    'Request URL': entry.url,
    'Request Method': entry.method,
    'Status Code': entry.status ? `${entry.status} ${entry.statusText}` : undefined
  })
}

async function copyRequestHeaders() {
  if (!selectedEntry.value?.request?.headers) return
  await copyJSON(selectedEntry.value.request.headers)
}

async function copyResponseHeaders() {
  if (!selectedEntry.value?.response?.headers) return
  await copyJSON(selectedEntry.value.response.headers)
}

async function copyPayload() {
  if (!selectedEntry.value?.request?.body) return
  const decoded = decodeHtml(selectedEntry.value.request.body)
  await copyToClipboard(decoded)
}

const requestCookies = computed(() => parseRequestCookies(selectedEntry.value?.request?.headers))
const responseCookies = computed(() => parseResponseCookies(selectedEntry.value?.response?.headers))

async function copyRequestCookies() {
  if (requestCookies.value.length === 0) return
  const data = requestCookies.value.reduce((obj: any, c: any) => ({ ...obj, [c.name]: c.value }), {})
  await copyJSON(data)
}

async function copyResponseCookies() {
  if (responseCookies.value.length === 0) return
  const data = responseCookies.value.reduce((obj: any, c: any) => ({ ...obj, [c.name]: c.value }), {})
  await copyJSON(data)
}



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

function addInterceptRule(entry: any) {
  editingRule.value = {
    id: '',
    name: `拦截 ${entry.url}`,
    enabled: true,
    urlPattern: entry.url,
    actionType: 'findReplace',
    findText: '',
    replaceText: '',
    useRegex: false,
    replaceAll: true,
    webhookEnabled: false
  }
  editorVisible.value = true
}

function createRule() {
  editingRule.value = undefined
  editorVisible.value = true
}

function editRule(rule: InterceptRule) {
  editingRule.value = { ...rule }
  editorVisible.value = true
}

function saveRule(rule: InterceptRule) {
  const index = interceptRules.value.findIndex(r => r.id === rule.id)
  if (index >= 0) {
    interceptRules.value[index] = rule
  } else {
    interceptRules.value.push(rule)
  }
  editorVisible.value = false
  saveRulesToStorage()
}

function toggleRule(id: string) {
  const rule = interceptRules.value.find(r => r.id === id)
  if (rule) {
    rule.enabled = !rule.enabled
    saveRulesToStorage()
  }
}

async function deleteRule(id: string) {
  const result = await ShowQuestionDialog('确认删除', '确定要删除这条规则吗？')
  if (result === 'Yes') {
    interceptRules.value = interceptRules.value.filter(r => r.id !== id)
    saveRulesToStorage()
  }
}

function importRules() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = (e: any) => {
    const file = e.target.files[0]
    if (!file) return
    
    const reader = new FileReader()
    reader.onload = (e: any) => {
      try {
        const data = JSON.parse(e.target.result)
        if (data.rules && Array.isArray(data.rules)) {
          interceptRules.value = data.rules
          saveRulesToStorage()
          ShowInfoDialog('成功', '导入成功')
        }
      } catch (err) {
        ShowErrorDialog('错误', '导入失败: ' + err)
      }
    }
    reader.readAsText(file)
  }
  input.click()
}

function exportRules() {
  const data = {
    version: '1.0',
    rules: interceptRules.value
  }
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `intercept-rules-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
}

async function saveRulesToStorage() {
  localStorage.setItem('interceptRules', JSON.stringify(interceptRules.value))
  // 同步到后端
  try {
    await SetInterceptRules(interceptRules.value)
  } catch (e) {
    console.error('Failed to sync rules to backend:', e)
  }
}

function loadRulesFromStorage() {
  const stored = localStorage.getItem('interceptRules')
  if (stored) {
    try {
      interceptRules.value = JSON.parse(stored)
    } catch (e) {
      console.error('Failed to load rules:', e)
    }
  }
}

onMounted(async () => {
  loadRulesFromStorage()
  // 同步到后端
  await saveRulesToStorage()
})
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
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
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

.col-status {
  width: 60px;
}

.col-type {
  width: 80px;
}

.col-size {
  width: 80px;
  text-align: right;
}

.col-time {
  width: 80px;
  text-align: right;
}

.col-action {
  width: 80px;
  text-align: center;
}

.add-rule-btn {
  padding: 2px 8px;
  border: 1px solid #1a73e8;
  background: white;
  color: #1a73e8;
  border-radius: 3px;
  cursor: pointer;
  font-size: 10px;
  transition: all 0.1s;
}

.add-rule-btn:hover {
  background: #1a73e8;
  color: white;
}

.method {
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: 600;
  font-size: 10px;
  text-transform: uppercase;
}

.method-get {
  background: #e3f2fd;
  color: #1976d2;
}

.method-post {
  background: #e8f5e9;
  color: #388e3c;
}

.method-put {
  background: #fff3e0;
  color: #f57c00;
}

.method-delete {
  background: #ffebee;
  color: #d32f2f;
}

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
