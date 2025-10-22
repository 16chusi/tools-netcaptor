<template>
  <div v-if="entry.response.body" class="response-container">
    <div class="response-toolbar">
      <div class="toolbar-left-section">
        <span class="format-label">格式: <strong>{{ formatType }}</strong></span>
        <div class="view-mode">
          <label>查看方式:</label>
          <select :value="viewMode" @change="$emit('changeViewMode', $event)" class="view-select">
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
      <button @click="$emit('save')" class="save-btn">下载</button>
    </div>
    <div v-if="viewMode === 'image'" class="image-preview">
      <img :src="entry.url" alt="Response Image" @error="$emit('imageError')"/>
    </div>
    <div v-else-if="viewMode === 'pdf'" class="pdf-preview">
      <div class="pdf-actions">
        <button @click="$emit('openInBrowser')" class="action-btn">在浏览器中打开</button>
        <button @click="$emit('downloadAndOpen')" class="action-btn">下载并打开</button>
      </div>
      <div class="pdf-info">
        <p>PDF 文件无法在应用内预览</p>
        <p>请使用上方按钮在浏览器中打开或下载后查看</p>
        <div class="url-container">
          <code>{{ entry.url }}</code>
          <button @click="$emit('copyUrl')" class="copy-btn" title="复制链接">📋</button>
        </div>
      </div>
    </div>
    <pre v-else-if="viewMode === 'hex'" class="text-preview">{{ hexContent }}</pre>
    <pre v-else-if="viewMode === 'base64'" class="text-preview">{{ base64Content }}</pre>
    <pre v-else ref="codeBlock" class="text-preview"><code></code></pre>
  </div>
  <div v-else class="empty">无响应体</div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { getFormatType } from '../../utils/networkUtils'
import { toHex, toBase64 } from '../../utils/formatters'
import hljs from 'highlight.js/lib/core'
import { css as beautifyCss, html as beautifyHtml, js as beautifyJs } from 'js-beautify'

const props = defineProps<{
  entry: any
  viewMode: string
  hexContent: string
  base64Content: string
}>()

defineEmits(['changeViewMode', 'save', 'imageError', 'openInBrowser', 'downloadAndOpen', 'copyUrl'])

const codeBlock = ref<HTMLElement | null>(null)
const formatType = getFormatType(props.entry.response?.contentType || '')

function decodeHtml(html: string): string {
  const txt = document.createElement('textarea')
  txt.innerHTML = html
  return txt.value
}

watch(() => [props.entry, props.viewMode], async () => {
  if (!props.entry?.response?.body) return
  if (props.viewMode === 'hex' || props.viewMode === 'base64' || props.viewMode === 'image' || props.viewMode === 'pdf') return

  await nextTick()
  if (!codeBlock.value) return

  const codeEl = codeBlock.value.querySelector('code')
  if (!codeEl) return

  let decoded = decodeHtml(props.entry.response.body)
  const mode = props.viewMode

  let lang = 'plaintext'

  if (mode === 'json') {
    lang = 'json'
    try {
      decoded = JSON.stringify(JSON.parse(decoded), null, 2)
    } catch {}
  } else if (mode === 'javascript') {
    lang = 'javascript'
    decoded = beautifyJs(decoded, { indent_size: 2 })
  } else if (mode === 'html') {
    lang = 'html'
    decoded = beautifyHtml(decoded, { indent_size: 2 })
  } else if (mode === 'css') {
    lang = 'css'
    decoded = beautifyCss(decoded, { indent_size: 2 })
  }

  codeEl.textContent = decoded
  codeEl.className = `language-${lang}`
  delete (codeEl as any).dataset.highlighted
  hljs.highlightElement(codeEl as HTMLElement)
}, { flush: 'post', deep: true })

defineExpose({ codeBlock })
</script>

<style scoped>
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
</style>
