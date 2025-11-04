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
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 #btn-{index} 或 [data-id='{data.id}']</div>
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
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 #user-{index} 或 [data-name='{data.name}']</div>
        </div>
        <div class="form-item">
          <label>输入内容</label>
          <input v-model="formData.text" placeholder="要输入的文本" />
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {data.name} 或 用户: {data.username}</div>
        </div>
      </template>

      <!-- 导航 -->
      <template v-if="nodeType === 'navigate'">
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">目标URL</label>
            <input v-model="formData.url" placeholder="https://example.com 或 {url}" style="flex: 1;" />
          </div>
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {url} 或 {data.website}</div>
        </div>
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">打开方式</label>
            <select v-model="formData.openMode" style="flex: 1;">
              <option value="current">当前窗口</option>
              <option value="new">新窗口</option>
            </select>
          </div>
        </div>
      </template>

      <!-- 等待 -->
      <template v-if="nodeType === 'wait'">
        <div class="form-item">
          <label>等待时间(ms)</label>
          <input v-model.number="formData.duration" type="number" placeholder="1000" />
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
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {data.fileUrl} 或 {downloadLink}</div>
        </div>
        <div class="form-item" v-if="formData.urlSource === 'variable'">
          <label>变量名</label>
          <input v-model="formData.urlVariable" placeholder="downloadUrl" />
          <small style="color: #999; font-size: 11px;">直接输入变量名</small>
        </div>
        <div class="form-item" v-if="formData.urlSource === 'template'">
          <label>URL模板</label>
          <input v-model="formData.urlTemplate" placeholder="https://example.com/{id}/file.pdf" />
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {data.id} 或 {index}</div>
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
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 .item-{index} 或 [data-page='{page}']</div>
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
          <label>左值</label>
          <input v-model="formData.leftValue" placeholder="{变量} 或 固定值" />
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {data.status} 或 {index}</div>
        </div>
        <div class="form-item">
          <label>运算符</label>
          <select v-model="formData.operator">
            <option value="==">等于 (==)</option>
            <option value="!=">不等于 (!=)</option>
            <option value=">">大于 (>)</option>
            <option value="<">小于 (<)</option>
            <option value=">=">大于等于 (>=)</option>
            <option value="<=">小于等于 (<=)</option>
            <option value="contains">包含 (contains)</option>
            <option value="notContains">不包含 (not contains)</option>
            <option value="startsWith">开头是 (starts with)</option>
            <option value="endsWith">结尾是 (ends with)</option>
          </select>
        </div>
        <div class="form-item">
          <label>右值</label>
          <input v-model="formData.rightValue" placeholder="{变量} 或 固定值" />
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {data.expected} 或 "完成"</div>
        </div>
        <div class="form-item">
          <label>True 分支端口</label>
          <select v-model="formData.truePort">
            <option value="top">上</option>
            <option value="right">右</option>
            <option value="bottom">下</option>
            <option value="left">左</option>
          </select>
        </div>
        <div class="form-item">
          <label>False 分支端口</label>
          <select v-model="formData.falsePort">
            <option value="top">上</option>
            <option value="right">右</option>
            <option value="bottom">下</option>
            <option value="left">左</option>
          </select>
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
          <input v-model="formData.count" placeholder="10 或 {maxCount}" />
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {data.totalPages} 或 {maxCount}</div>
        </div>
        <div class="form-item">
          <label>循环变量</label>
          <input v-model="formData.variable" placeholder="index" />
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">循环体内可通过 {index} 引用当前索引（从1开始）</small>
        </div>
        <div class="form-item">
          <label>循环间隔(ms)</label>
          <input v-model.number="formData.interval" type="number" placeholder="500" />
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">每次循环之间的等待时间</small>
        </div>
      </template>

      <!-- 拦截请求 -->
      <template v-if="nodeType === 'intercept_request'">
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">URL匹配</label>
            <input v-model="formData.urlPattern" placeholder="*/api/login" style="flex: 1;" />
          </div>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">支持通配符 *，如 */api/*</small>
        </div>
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">拦截动作</label>
            <select v-model="formData.action" style="flex: 1;">
              <option value="capture">捕获数据</option>
              <option value="block">阻断请求</option>
              <option value="mock">Mock响应</option>
              <option value="redirect">重定向</option>
              <option value="download">下载保存</option>
            </select>
          </div>
        </div>
        <div class="form-item" v-if="formData.action === 'capture'">
          <label>数据格式</label>
          <select v-model="formData.dataFormat">
            <option value="text">文本 (Text)</option>
            <option value="json">JSON</option>
            <option value="hex">十六进制 (Hex)</option>
          </select>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">选择数据转换格式</small>
        </div>
        <div class="form-item" v-if="formData.action === 'capture'">
          <label>保存到变量</label>
          <input v-model="formData.saveToVariable" placeholder="responseData" />
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">后续节点可通过 {responseData} 引用</small>
        </div>
        <div class="form-item" v-if="formData.action === 'block'">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">状态码</label>
            <input v-model.number="formData.statusCode" type="number" placeholder="403" style="flex: 1;" />
          </div>
        </div>
        <div class="form-item" v-if="formData.action === 'mock'">
          <label>Mock响应内容</label>
          <textarea v-model="formData.mockResponse" placeholder='{"success":true}' style="width: 100%; min-height: 100px; padding: 8px; border: 1px solid #d9d9d9; border-radius: 4px; font-family: monospace; font-size: 12px;"></textarea>
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {"data": "{data.result}", "status": "{status}"}</div>
        </div>
        <div class="form-item" v-if="formData.action === 'redirect'">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">重定向URL</label>
            <input v-model="formData.redirectUrl" placeholder="https://example.com/api" style="flex: 1;" />
          </div>
          <div class="variable-hint">💡 支持变量: {变量名} 或 {变量名.字段名}，如 {data.newUrl} 或 https://api.com/{data.id}</div>
        </div>
        <div class="form-item" v-if="formData.action === 'download'">
          <label>保存目录</label>
          <div style="display: flex; gap: 8px;">
            <input v-model="formData.saveDirectory" placeholder="选择保存目录" readonly style="flex: 1;" />
            <button @click="selectDirectory" type="button" style="padding: 8px 12px; border: 1px solid #d9d9d9; background: white; border-radius: 4px; cursor: pointer;">选择</button>
          </div>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">拦截的请求将被下载到此目录</small>
        </div>
        <div class="form-item" v-if="formData.action === 'download'">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">文件扩展名</label>
            <input v-model="formData.fileExtension" placeholder="json" style="flex: 1;" />
          </div>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">如 json、txt、html、pdf、png 等，留空则自动检测</small>
        </div>
      </template>

      <!-- 下载已捕获响应 -->
      <template v-if="nodeType === 'download_captured'">
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">URL匹配</label>
            <input v-model="formData.urlPattern" placeholder="*/api/data*" style="flex: 1;" />
          </div>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">支持通配符 *，匹配已捕获的请求</small>
        </div>
        <div class="form-item">
          <label>保存目录</label>
          <div style="display: flex; gap: 8px;">
            <input v-model="formData.saveDirectory" placeholder="选择保存目录" readonly style="flex: 1;" />
            <button @click="selectDirectory" type="button" style="padding: 8px 12px; border: 1px solid #d9d9d9; background: white; border-radius: 4px; cursor: pointer;">选择</button>
          </div>
        </div>
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">文件扩展名</label>
            <input v-model="formData.fileExtension" placeholder="json" style="flex: 1;" />
          </div>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">如 json、txt、html 等，留空则自动检测</small>
        </div>
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">重复文件</label>
            <select v-model="formData.overwriteMode" style="flex: 1;">
              <option value="skip">跳过</option>
              <option value="overwrite">覆盖</option>
            </select>
          </div>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">相同内容的文件处理方式</small>
        </div>
      </template>

      <!-- 数据收集器 -->
      <template v-if="nodeType === 'collect'">
        <div class="form-item">
          <label>数据来源</label>
          <input v-model="formData.dataVariable" placeholder="变量名，如 data" />
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">输入变量名，不需要 {}</small>
        </div>
        <div class="form-item">
          <label>保存文件</label>
          <div style="display: flex; gap: 8px;">
            <input v-model="formData.filePath" placeholder="选择文件" readonly style="flex: 1;" />
            <button @click="selectCollectFile" type="button" style="padding: 8px 12px; border: 1px solid #d9d9d9; background: white; border-radius: 4px; cursor: pointer;">选择</button>
          </div>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">数据将追加到此文件</small>
        </div>
        <div class="form-item">
          <label>数据格式</label>
          <select v-model="formData.format">
            <option value="jsonl">JSONL (每行一个JSON)</option>
            <option value="text">文本 (每行追加)</option>
          </select>
        </div>
      </template>

      <!-- 解密 -->
      <template v-if="nodeType === 'decrypt'">
        <div class="form-item">
          <label>数据来源</label>
          <input v-model="formData.dataVariable" placeholder="变量名，如 data" />
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">输入变量名，不需要 {}</small>
        </div>
        <div class="form-item">
          <label>解密算法</label>
          <select v-model="formData.algorithm">
            <option value="sm4-ecb">SM4-ECB</option>
            <option value="sm4-cbc">SM4-CBC</option>
          </select>
        </div>
        <div class="form-item">
          <label>密钥 (Hex)</label>
          <input v-model="formData.key" placeholder="46696e32416e63304571753245727934" />
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">16进制格式的密钥</small>
        </div>
        <div class="form-item" v-if="formData.algorithm === 'sm4-cbc'">
          <label>IV向量 (Hex)</label>
          <input v-model="formData.iv" placeholder="可选，留空使用零向量" />
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">16进制格式的初始化向量</small>
        </div>
        <div class="form-item">
          <label>保存到变量</label>
          <input v-model="formData.saveToVariable" placeholder="decryptedData" />
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">后续节点可通过 {decryptedData} 引用</small>
        </div>
      </template>

      <!-- JSONL读取器 -->
      <template v-if="nodeType === 'jsonl_reader'">
        <div class="form-item">
          <label>JSONL文件</label>
          <div style="display: flex; gap: 8px;">
            <input v-model="formData.filePath" placeholder="选择JSONL文件" readonly style="flex: 1;" />
            <button @click="selectJSONLFile" type="button" style="padding: 8px 12px; border: 1px solid #d9d9d9; background: white; border-radius: 4px; cursor: pointer;">选择</button>
          </div>
        </div>
        <div class="form-item" v-if="formData.filePath">
          <button @click="loadJSONLKeys" type="button" style="width: 100%; padding: 8px; border: 1px solid #1890ff; background: white; color: #1890ff; border-radius: 4px; cursor: pointer;">加载可用字段</button>
        </div>
        <div class="form-item" v-if="availableKeys.length > 0">
          <label>可用字段</label>
          <div style="padding: 8px; background: #f5f5f5; border-radius: 4px; font-size: 12px; color: #666;">
            {{ availableKeys.join(', ') }}
          </div>
        </div>
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">提取字段</label>
            <input v-model="formData.extractKeys" placeholder="*" style="flex: 1;" />
          </div>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">多个字段用逗号分隔，* 表示全部</small>
        </div>
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">保存变量</label>
            <input v-model="formData.saveToVariable" placeholder="data" style="flex: 1;" />
          </div>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">后续节点可通过 {data.fieldName} 引用</small>
        </div>
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">循环间隔</label>
            <input v-model.number="formData.interval" type="number" placeholder="100" style="flex: 1;" />
            <span style="color: #999; font-size: 12px;">ms</span>
          </div>
        </div>
        <div class="form-item">
          <div style="display: flex; align-items: center; gap: 8px;">
            <label style="margin: 0; min-width: 80px;">最大次数</label>
            <input v-model.number="formData.maxCount" type="number" :placeholder="totalLines > 0 ? String(totalLines) : '全部'" style="flex: 1;" />
          </div>
          <small style="color: #999; font-size: 11px; display: block; margin-top: 4px;">留空则处理全部数据</small>
        </div>
        <div class="form-item" v-if="formData.filePath && totalLines > 0">
          <div style="padding: 8px; background: #e6f7ff; border: 1px solid #91d5ff; border-radius: 4px; font-size: 12px; color: #1890ff;">
            文件共 {{ totalLines }} 行数据
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import { SelectDownloadDirectory, SelectJSONLFile, LoadJSONLFile } from '../../../wailsjs/go/main/NetworkApp'

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
</style>
