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
          <label>选择器类型</label>
          <select v-model="formData.selectorType">
            <option value="css">CSS 选择器</option>
            <option value="xpath">XPath</option>
          </select>
        </div>
        <div class="form-item">
          <label>元素选择器</label>
          <input v-model="formData.selector" :placeholder="formData.selectorType === 'xpath' ? '//*[@id=&quot;btn&quot;]' : '#id 或 .class'" />
        </div>
        <div class="form-item">
          <label>等待元素出现(ms)</label>
          <input v-model.number="formData.waitTime" type="number" placeholder="3000" />
        </div>
      </template>

      <!-- 输入文本 -->
      <template v-if="nodeType === 'input'">
        <div class="form-item">
          <label>选择器类型</label>
          <select v-model="formData.selectorType">
            <option value="css">CSS 选择器</option>
            <option value="xpath">XPath</option>
          </select>
        </div>
        <div class="form-item">
          <label>元素选择器</label>
          <input v-model="formData.selector" :placeholder="formData.selectorType === 'xpath' ? '//*[@id=&quot;username&quot;]' : '#username'" />
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
          <label>URL来源</label>
          <select v-model="formData.urlSource">
            <option value="direct">直接输入</option>
            <option value="variable">从变量获取</option>
            <option value="template">模板拼接</option>
          </select>
        </div>
        <div class="form-item" v-if="formData.urlSource === 'direct'">
          <label>下载URL</label>
          <input v-model="formData.downloadUrl" placeholder="https://example.com/file.pdf" />
        </div>
        <div class="form-item" v-if="formData.urlSource === 'variable'">
          <label>变量名</label>
          <input v-model="formData.urlVariable" placeholder="downloadUrl" />
          <small style="color: #999; font-size: 11px;">直接输入变量名</small>
        </div>
        <div class="form-item" v-if="formData.urlSource === 'template'">
          <label>URL模板</label>
          <input v-model="formData.urlTemplate" placeholder="https://example.com/{id}/file.pdf" />
          <small style="color: #999; font-size: 11px;">使用 {变量名} 引用变量</small>
        </div>

        <div class="form-item">
          <label>保存目录</label>
          <div style="display: flex; gap: 8px;">
            <input v-model="formData.saveDirectory" placeholder="选择保存目录" readonly style="flex: 1;" />
            <button @click="selectDirectory" type="button" style="padding: 8px 12px; border: 1px solid #d9d9d9; background: white; border-radius: 4px; cursor: pointer;">选择</button>
          </div>
        </div>
      </template>

      <!-- 提取数据 -->
      <template v-if="nodeType === 'extract'">
        <div class="form-item">
          <label>选择器类型</label>
          <select v-model="formData.selectorType">
            <option value="css">CSS 选择器</option>
            <option value="xpath">XPath</option>
          </select>
        </div>
        <div class="form-item">
          <label>元素选择器</label>
          <input v-model="formData.selector" :placeholder="formData.selectorType === 'xpath' ? '//*[@class=&quot;item&quot;]' : '.item'" />
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
          <input v-model="formData.saveToVariable" placeholder="myData" />
          <small style="color: #999; font-size: 11px;">后续节点可通过 {myData} 引用</small>
        </div>
      </template>

      <!-- 条件判断 -->
      <template v-if="nodeType === 'if'">
        <div class="form-item">
          <label>条件表达式</label>
          <input v-model="formData.condition" placeholder="变量 == 值" />
        </div>
      </template>

      <!-- 滚动页面 -->
      <template v-if="nodeType === 'scroll'">
        <div class="form-item">
          <label>滚动类型</label>
          <select v-model="formData.scrollType">
            <option value="bottom">滚动到底部</option>
            <option value="times">滚动N次</option>
            <option value="distance">滚动指定距离</option>
          </select>
        </div>
        <div class="form-item" v-if="formData.scrollType === 'times'">
          <label>滚动次数</label>
          <input v-model.number="formData.times" type="number" placeholder="5" />
        </div>
        <div class="form-item" v-if="formData.scrollType === 'distance'">
          <label>滚动距离(px)</label>
          <input v-model.number="formData.distance" type="number" placeholder="1000" />
        </div>
        <div class="form-item">
          <label>每次间隔(ms)</label>
          <input v-model.number="formData.interval" type="number" placeholder="500" />
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
import { SelectDownloadDirectory } from '../../../wailsjs/go/main/NetworkApp'

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
  console.log('[PropertyPanel] watch nodeData 变化:', newData)
  if (newData) {
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

  } else {
    formData.value = { selectorType: 'css', urlSource: 'direct' }
  }
  console.log('[PropertyPanel] formData 已更新:', formData.value)
}, { immediate: true, deep: true })

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

function handleSave() {
  console.log('[PropertyPanel] 保存节点数据:', props.nodeType, formData.value)
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
