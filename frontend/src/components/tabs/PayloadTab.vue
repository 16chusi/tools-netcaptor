<template>
  <div v-if="entry.request.body" class="payload-container">
    <div class="payload-toolbar">
      <button @click="$emit('copy')" class="copy-payload-btn">复制</button>
    </div>
    <pre ref="payloadBlock" class="payload-preview"><code></code></pre>
  </div>
  <div v-else class="empty">无请求体</div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { toHex } from '../../utils/formatters'
import hljs from 'highlight.js/lib/core'

const props = defineProps<{
  entry: any
}>()

defineEmits(['copy', 'update:content'])

const payloadBlock = ref<HTMLElement | null>(null)

function decodeHtml(html: string): string {
  const txt = document.createElement('textarea')
  txt.innerHTML = html
  return txt.value
}

watch(() => props.entry, async () => {
  if (!props.entry?.request?.body) return

  await nextTick()
  if (!payloadBlock.value) return

  const codeEl = payloadBlock.value.querySelector('code')
  if (!codeEl) return

  let decoded = decodeHtml(props.entry.request.body)
  const ct = props.entry.request?.headers?.['Content-Type'] || props.entry.request?.headers?.['content-type'] || ''

  let lang = 'plaintext'
  let isText = true

  if (ct.includes('json')) {
    lang = 'json'
    try {
      decoded = JSON.stringify(JSON.parse(decoded), null, 2)
    } catch {}
  } else if (ct.includes('text/')) {
    lang = 'plaintext'
  } else if (ct.includes('x-www-form-urlencoded')) {
    lang = 'plaintext'
  } else {
    isText = false
    decoded = toHex(decoded)
  }

  codeEl.textContent = decoded
  if (isText) {
    codeEl.className = `language-${lang}`
    delete (codeEl as any).dataset.highlighted
    hljs.highlightElement(codeEl as HTMLElement)
  }
}, { flush: 'post', immediate: true, deep: true })

defineExpose({ payloadBlock })
</script>

<style scoped>
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

.empty {
  padding: 40px;
  text-align: center;
  color: #80868b;
}
</style>
