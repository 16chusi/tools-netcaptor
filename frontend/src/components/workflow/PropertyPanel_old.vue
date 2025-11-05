<template>
  <div v-if="visible" class="property-panel">
    <div class="panel-header">
      <h4 
        @dblclick="startEditLabel" 
        :contenteditable="isEditingLabel"
        @blur="finishEditLabel"
        @keydown.enter.prevent="finishEditLabel"
        ref="labelRef"
        class="editable-label"
      >
        {{ displayLabel }}
      </h4>
      <button @click="$emit('close')" class="close-btn">✕</button>
    </div>
    <div class="panel-body">
      <!-- 动态组件渲染 -->
      <component 
        v-if="currentPropertyComponent" 
        :is="currentPropertyComponent" 
        :formData="formData"
      />
      
      <!-- 未迁移的组件保持原有结构 -->
          <small style="color: #999; font-size: 11px;">失败时自动重试，适用于网络错误、并发限制等临时问题</small>
      <!-- 所有组件已迁移完成 -->
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick, onMounted } from 'vue'
import { SelectDownloadDirectory, SelectJSONLFile, LoadJSONLFile } from '../../../wailsjs/go/main/NetworkApp'
import { propertyComponents, defaultValues } from './properties/index'

const props = defineProps<{
  visible: boolean
  nodeType?: string
  nodeLabel?: string
  nodeData?: Record<string, any>
}>()

const emit = defineEmits<{
  close: []
  save: [data: Record<string, any>]
  updateLabel: [label: string]
}>()

const formData = ref<Record<string, any>>({})
const availableKeys = ref<string[]>([])
const totalLines = ref(0)

// 动态组件计算属性
const currentPropertyComponent = computed(() => {
  return propertyComponents[props.nodeType as keyof typeof propertyComponents]
})
const aiModels = ref<any[]>([])

// 加载AI模型配置
function loadAIModels() {
  const saved = localStorage.getItem('ai-model-settings')
  if (saved) {
    const settings = JSON.parse(saved)
    aiModels.value = settings.models || []
  }
}
const isEditingLabel = ref(false)
const labelRef = ref<HTMLElement>()
const displayLabel = computed(() => formData.value.customLabel || props.nodeLabel || '组件')

let saveTimer: number | null = null
let isUpdatingFromProps = false // 添加标志防止循环

watch(formData, (newData) => {
  // 如果是从props更新的，不触发保存
  if (isUpdatingFromProps) return
  
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = window.setTimeout(() => {
    console.log('[PropertyPanel] formData 变化，自动保存:', newData)
    emit('save', newData)
  }, 300)
}, { deep: true })

watch(() => props.nodeData, (newData) => {
  console.log('[PropertyPanel] watch nodeData 变化:', newData)
  if (newData) {
    // 设置标志，防止触发formData的watch
    isUpdatingFromProps = true
    
    // 检查数据是否真的变化
    const currentDataStr = JSON.stringify(formData.value)
    const newDataStr = JSON.stringify(newData)
    
    if (currentDataStr !== newDataStr) {
      formData.value = { ...newData }
      // 默认值
      if (!formData.value.selectorType) {
        formData.value.selectorType = 'css'
      }
      if (!formData.value.urlSource) {
        formData.value.urlSource = 'direct'
      }
      if (!formData.value.scrollType) {
        formData.value.scrollType = 'bottom'
      }
      if (!formData.value.openMode) {
        formData.value.openMode = 'current'
      }
      if (!formData.value.extractKeys) {
        formData.value.extractKeys = '*'
      }
      if (!formData.value.interval) {
        formData.value.interval = 100
      }
      if (!formData.value.action) {
        formData.value.action = 'block'
      }
      if (!formData.value.statusCode) {
        formData.value.statusCode = 403
      }
      if (!formData.value.overwriteMode) {
        formData.value.overwriteMode = 'skip'
      }
      if (!formData.value.algorithm) {
        formData.value.algorithm = 'sm4-ecb'
      }
      if (!formData.value.dataFormat) {
        formData.value.dataFormat = 'text'
      }
      if (!formData.value.format) {
        formData.value.format = 'jsonl'
      }
      if (props.nodeType === 'if') {
        if (!formData.value.operator) {
          formData.value.operator = '=='
        }
        if (!formData.value.truePort) {
          formData.value.truePort = 'right'
        }
        if (!formData.value.falsePort) {
          formData.value.falsePort = 'left'
        }
      }
      
      // 循环组件默认值
      if (props.nodeType === 'for') {
        if (!formData.value.variable) {
          formData.value.variable = 'index'
        }
      }
      
      // JSONL读取器默认值
      if (props.nodeType === 'jsonl_reader') {
        if (!formData.value.saveToVariable) {
          formData.value.saveToVariable = 'data'
        }
      }
      
      // 使用配置化的默认值
      const defaults = defaultValues[props.nodeType as keyof typeof defaultValues]
      if (defaults) {
        Object.keys(defaults).forEach(key => {
          if (!formData.value[key]) {
            formData.value[key] = defaults[key as keyof typeof defaults]
          }
        })
      }
      
      // 其他组件默认值（保持向后兼容）
      if (props.nodeType === 'extract') {
        if (!formData.value.saveToVariable) {
          formData.value.saveToVariable = 'myData'
        }
      }
      if (props.nodeType === 'intercept_request') {
        if (!formData.value.saveToVariable) {
          formData.value.saveToVariable = 'responseData'
        }
      }
      if (props.nodeType === 'decrypt') {
        if (!formData.value.saveToVariable) {
          formData.value.saveToVariable = 'decryptedData'
        }
      }
      
      // AI组件默认值
      if (props.nodeType?.startsWith('ai_')) {
        if (formData.value.modelIndex === undefined) {
          formData.value.modelIndex = 0
        }
        if (!formData.value.prompt) {
          formData.value.prompt = ''
        }
        if (!formData.value.saveToVariable && props.nodeType !== 'ai_smart_click' && props.nodeType !== 'ai_form_fill' && props.nodeType !== 'ai_navigation') {
          // 根据组件类型设置默认变量名
          switch (props.nodeType) {
            case 'ai_extract_data':
              formData.value.saveToVariable = 'extractedData'
              break
            case 'ai_analyze_content':
              formData.value.saveToVariable = 'analysisResult'
              break
            case 'ai_validate_data':
              formData.value.saveToVariable = 'validationResult'
              break
            case 'ai_transform_data':
              formData.value.saveToVariable = 'transformedData'
              break
            default:
              formData.value.saveToVariable = 'aiResult'
          }
        }
        // 设置数据来源默认值
        if (!formData.value.dataSource && (props.nodeType === 'ai_validate_data' || props.nodeType === 'ai_transform_data' || props.nodeType === 'ai_form_fill')) {
          switch (props.nodeType) {
            case 'ai_validate_data':
              formData.value.dataSource = '{extractedData}'
              break
            case 'ai_transform_data':
              formData.value.dataSource = '{rawData}'
              break
            case 'ai_form_fill':
              formData.value.dataSource = '{formData}'
              break
          }
        }
        if (!formData.value.outputFormat && (props.nodeType === 'ai_extract_data' || props.nodeType === 'ai_transform_data')) {
          formData.value.outputFormat = 'json'
        }
        if (formData.value.waitTime === undefined && (props.nodeType === 'ai_smart_click' || props.nodeType === 'ai_navigation')) {
          formData.value.waitTime = 3000
        }
        // 重试配置默认值
        if (formData.value.retryCount === undefined) {
          formData.value.retryCount = 3
        }
        if (formData.value.retryDelay === undefined) {
          formData.value.retryDelay = 2
        }
      }
    }
    
    // 清除标志
    setTimeout(() => {
      isUpdatingFromProps = false
    }, 50)
  } else {
    isUpdatingFromProps = true
    formData.value = { selectorType: 'css', urlSource: 'direct', extractKeys: '*', interval: 100, openMode: 'current', overwriteMode: 'skip', algorithm: 'sm4-ecb', dataFormat: 'text' }
    setTimeout(() => {
      isUpdatingFromProps = false
    }, 50)
  }
  console.log('[PropertyPanel] formData 已更新:', formData.value)
}, { immediate: true })

onMounted(() => {
  loadAIModels()
})

async function selectDirectory() {
  try {
    const dir = await SelectDownloadDirectory()
    if (dir) {
      formData.value.saveDirectory = dir
    }
  } catch (error: any) {
    console.error('[PropertyPanel] 选择目录失败:', error)
  }
}

async function selectJSONLFile() {
  try {
    const file = await SelectJSONLFile()
    if (file) {
      formData.value.filePath = file
      availableKeys.value = []
      totalLines.value = 0
    }
  } catch (error: any) {
    console.error('[PropertyPanel] 选择文件失败:', error)
  }
}

async function loadJSONLKeys() {
  if (!formData.value.filePath) return
  try {
    const result = await LoadJSONLFile(formData.value.filePath)
    availableKeys.value = result.keys || []
    totalLines.value = result.totalLines || 0
  } catch (error: any) {
    console.error('[PropertyPanel] 加载文件失败:', error)
    alert('加载文件失败: ' + error)
  }
}

async function selectCollectFile() {
  try {
    const file = await SelectJSONLFile()
    if (file) {
      formData.value.filePath = file
    }
  } catch (error: any) {
    console.error('[PropertyPanel] 选择文件失败:', error)
  }
}

function startEditLabel() {
  isEditingLabel.value = true
  nextTick(() => {
    if (labelRef.value) {
      labelRef.value.focus()
      const range = document.createRange()
      range.selectNodeContents(labelRef.value)
      const sel = window.getSelection()
      sel?.removeAllRanges()
      sel?.addRange(range)
    }
  })
}

function finishEditLabel() {
  if (!isEditingLabel.value) return
  isEditingLabel.value = false
  const newLabel = labelRef.value?.textContent?.trim()
  if (newLabel && newLabel !== displayLabel.value) {
    formData.value.customLabel = newLabel
    emit('updateLabel', newLabel)
  }
}
</script>

<style scoped>
@import './properties/common.css';
.property-panel {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 320px;
  background: white;
  border-left: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
  box-shadow: -2px 0 8px rgba(0,0,0,0.1);
  z-index: 100;
  pointer-events: auto;
  animation: slideInRight 0.2s ease-out;
}

@keyframes slideInRight {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

.panel-header {
  padding: 16px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-header h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.editable-label {
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  transition: background 0.2s;
}

.editable-label:hover {
  background: rgba(0, 0, 0, 0.05);
}

.editable-label[contenteditable="true"] {
  outline: 2px solid #1890ff;
  background: white;
  cursor: text;
}

.close-btn {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 16px;
  color: #666;
  border-radius: 4px;
}

.close-btn:hover {
  background: #f0f0f0;
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  padding-bottom: 16px;
}

.form-item {
  margin-bottom: 16px;
}

.form-item label {
  display: block;
  margin-bottom: 6px;
  font-size: 12px;
  font-weight: 500;
  color: #333;
  text-align: left;
}

.form-item input,
.form-item select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 13px;
  text-align: left;
}

.form-item input:focus,
.form-item select:focus {
  outline: none;
  border-color: #1890ff;
}

.variable-hint {
  font-size: 11px;
  color: #666;
  margin-top: 4px;
  line-height: 1.4;
  background: #f8f9fa;
  padding: 4px 8px;
  border-radius: 3px;
  border-left: 3px solid #1890ff;
  text-align: left;
}

/* AI组件专用样式 */
.ai-prompt-textarea {
  width: 100% !important;
  padding: 12px !important;
  border: 1px solid #d9d9d9 !important;
  border-radius: 6px !important;
  font-size: 13px !important;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace !important;
  line-height: 1.5 !important;
  resize: vertical !important;
  min-height: 80px !important;
  background: #fafafa !important;
  transition: all 0.2s !important;
  box-sizing: border-box !important;
}

.ai-prompt-textarea:focus {
  border-color: #40a9ff !important;
  background: white !important;
  box-shadow: 0 0 0 2px rgba(64, 169, 255, 0.1) !important;
  outline: none !important;
}

.ai-prompt-textarea::placeholder {
  color: #999 !important;
  font-style: italic !important;
}
</style>
