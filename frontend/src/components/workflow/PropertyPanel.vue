<template>
  <div v-if="visible" class="property-panel">
    <div class="panel-header">
      <h4>{{ nodeLabel }} 属性</h4>
      <button @click="$emit('close')" class="close-btn">✕</button>
    </div>
    <div class="panel-body">
      <!-- 点击元素 -->
      <template v-if="nodeType === 'click'">
        <div class="form-item">
          <label>元素选择器</label>
          <input v-model="formData.selector" placeholder="#id 或 .class" />
        </div>
        <div class="form-item">
          <label>等待元素出现(ms)</label>
          <input v-model.number="formData.waitTime" type="number" placeholder="3000" />
        </div>
      </template>

      <!-- 输入文本 -->
      <template v-if="nodeType === 'input'">
        <div class="form-item">
          <label>元素选择器</label>
          <input v-model="formData.selector" placeholder="#username" />
        </div>
        <div class="form-item">
          <label>输入内容</label>
          <input v-model="formData.text" placeholder="要输入的文本" />
        </div>
      </template>

      <!-- 导航 -->
      <template v-if="nodeType === 'navigate'">
        <div class="form-item">
          <label>目标URL</label>
          <input v-model="formData.url" placeholder="https://example.com" />
        </div>
      </template>

      <!-- 等待 -->
      <template v-if="nodeType === 'wait'">
        <div class="form-item">
          <label>等待时间(ms)</label>
          <input v-model.number="formData.duration" type="number" placeholder="1000" />
        </div>
      </template>

      <!-- 拦截请求 -->
      <template v-if="nodeType === 'intercept'">
        <div class="form-item">
          <label>URL匹配模式</label>
          <input v-model="formData.urlPattern" placeholder="/api/*" />
        </div>
        <div class="form-item">
          <label>操作类型</label>
          <select v-model="formData.actionType">
            <option value="log">记录日志</option>
            <option value="save">保存响应</option>
            <option value="modify">修改响应</option>
          </select>
        </div>
      </template>

      <!-- 下载保存 -->
      <template v-if="nodeType === 'download'">
        <div class="form-item">
          <label>下载URL</label>
          <input v-model="formData.downloadUrl" placeholder="文件URL" />
        </div>
        <div class="form-item">
          <label>保存文件名</label>
          <input v-model="formData.filename" placeholder="file.txt" />
        </div>
      </template>

      <!-- 提取数据 -->
      <template v-if="nodeType === 'extract'">
        <div class="form-item">
          <label>元素选择器</label>
          <input v-model="formData.selector" placeholder=".item" />
        </div>
        <div class="form-item">
          <label>提取属性</label>
          <select v-model="formData.attribute">
            <option value="text">文本内容</option>
            <option value="href">链接</option>
            <option value="src">图片源</option>
            <option value="value">表单值</option>
          </select>
        </div>
        <div class="form-item">
          <label>保存到变量</label>
          <input v-model="formData.variable" placeholder="myData" />
        </div>
      </template>

      <!-- 条件判断 -->
      <template v-if="nodeType === 'if'">
        <div class="form-item">
          <label>条件表达式</label>
          <input v-model="formData.condition" placeholder="变量 == 值" />
        </div>
      </template>

      <!-- 循环 -->
      <template v-if="nodeType === 'for'">
        <div class="form-item">
          <label>循环次数</label>
          <input v-model.number="formData.count" type="number" placeholder="10" />
        </div>
        <div class="form-item">
          <label>循环变量</label>
          <input v-model="formData.variable" placeholder="i" />
        </div>
      </template>
    </div>
    <div class="panel-footer">
      <button @click="handleSave" class="save-btn">保存</button>
      <button @click="$emit('close')" class="cancel-btn">取消</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  nodeType?: string
  nodeLabel?: string
  nodeData?: Record<string, any>
}>()

const emit = defineEmits<{
  close: []
  save: [data: Record<string, any>]
}>()

const formData = ref<Record<string, any>>({})

watch(() => props.nodeData, (newData) => {
  if (newData) {
    formData.value = { ...newData }
  } else {
    formData.value = {}
  }
}, { immediate: true })

function handleSave() {
  emit('save', formData.value)
  emit('close')
}
</script>

<style scoped>
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
}

.form-item input,
.form-item select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 13px;
}

.form-item input:focus,
.form-item select:focus {
  outline: none;
  border-color: #1890ff;
}

.panel-footer {
  padding: 16px;
  border-top: 1px solid #e0e0e0;
  display: flex;
  gap: 8px;
}

.save-btn,
.cancel-btn {
  flex: 1;
  padding: 8px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.save-btn {
  background: #1890ff;
  color: white;
}

.save-btn:hover {
  background: #40a9ff;
}

.cancel-btn {
  background: #f0f0f0;
  color: #333;
}

.cancel-btn:hover {
  background: #e0e0e0;
}
</style>
