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
      
      <!-- 所有组件已迁移完成 -->
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import { propertyComponents } from './properties/index'
import { applyDefaultValues, createDebouncedSave } from './properties/utils'

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
const isEditingLabel = ref(false)
const labelRef = ref<HTMLElement>()

// 动态组件计算属性
const currentPropertyComponent = computed(() => {
  return propertyComponents[props.nodeType as keyof typeof propertyComponents]
})

const displayLabel = computed(() => formData.value.customLabel || props.nodeLabel || '组件')

// 防抖保存
const debouncedSave = createDebouncedSave((data) => {
  console.log('[PropertyPanel] formData 变化，自动保存:', data)
  emit('save', data)
})

let isUpdatingFromProps = false

watch(formData, (newData) => {
  if (isUpdatingFromProps) return
  debouncedSave(newData)
}, { deep: true })

watch(() => props.nodeData, (newData) => {
  if (newData) {
    isUpdatingFromProps = true
    
    const currentDataStr = JSON.stringify(formData.value)
    const newDataStr = JSON.stringify(newData)
    
    if (currentDataStr !== newDataStr) {
      formData.value = { ...newData }
      
      // 应用默认值
      if (props.nodeType) {
        applyDefaultValues(formData.value, props.nodeType)
      }
    }
    
    nextTick(() => {
      isUpdatingFromProps = false
    })
  }
}, { immediate: true })

// 标签编辑功能
const startEditLabel = () => {
  isEditingLabel.value = true
  nextTick(() => {
    if (labelRef.value) {
      labelRef.value.focus()
      const range = document.createRange()
      range.selectNodeContents(labelRef.value)
      const selection = window.getSelection()
      selection?.removeAllRanges()
      selection?.addRange(range)
    }
  })
}

const finishEditLabel = () => {
  if (isEditingLabel.value && labelRef.value) {
    const newLabel = labelRef.value.textContent?.trim() || ''
    formData.value.customLabel = newLabel
    emit('updateLabel', newLabel)
    isEditingLabel.value = false
  }
}
</script>

<style scoped>
@import './properties/common.css';

.property-panel {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 320px;
  background: white;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  background: #fafafa;
}

.panel-header h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.editable-label {
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  transition: background-color 0.2s;
}

.editable-label:hover {
  background-color: #f0f0f0;
}

.editable-label[contenteditable="true"] {
  background-color: #fff;
  border: 1px solid #1890ff;
  outline: none;
}

.close-btn {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: #999;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f5f5f5;
  color: #333;
}

.panel-body {
  padding: 16px;
  overflow-y: auto;
  flex: 1;
}
</style>
