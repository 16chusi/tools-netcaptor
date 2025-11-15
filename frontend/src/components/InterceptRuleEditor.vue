<template>
  <div v-if="visible" class="drawer-overlay">
    <div class="drawer" @click.stop>
      <div class="drawer-header">
        <h3>{{ localRule.id ? '编辑规则' : '新建规则' }}</h3>
        <button @click="$emit('close')" class="close-btn">✕</button>
      </div>
      <div class="drawer-body">
        <div class="form-group">
          <label>规则名称</label>
          <input v-model="localRule.name" type="text" placeholder="例如: Mock API 响应">
        </div>
        
        <div class="form-group">
          <label>URL 模式</label>
          <input v-model="localRule.urlPattern" type="text" placeholder="例如: *.example.com 或 /api/*">
          <small>支持通配符 *，例如: *.example.com, /api/*, */user/*</small>
        </div>
        
        <div class="form-group">
          <label>操作类型</label>
          <select v-model="localRule.actionType">
            <option value="findReplace">内容替换</option>
            <option value="redirect">重定向</option>
            <option value="responseReplace">响应结果替换</option>
            <option value="remoteHTTP">远程HTTP处理</option>
            <option value="forwardRequest">请求转发</option>
            <option value="saveToFile">保存到文件</option>
          </select>
        </div>
        
        <!-- 内容替换 -->
        <div v-if="localRule.actionType === 'findReplace'">
          <div class="form-group">
            <label>查找</label>
            <input v-model="localRule.findText" type="text" placeholder="输入要查找的内容">
          </div>
          <div class="form-group">
            <label>替换</label>
            <input v-model="localRule.replaceText" type="text" placeholder="输入替换后的内容">
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="localRule.useRegex" type="checkbox">
              使用正则表达式
            </label>
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="localRule.replaceAll" type="checkbox">
              替换全部匹配项
            </label>
          </div>
        </div>
        
        <!-- 重定向 -->
        <div v-if="localRule.actionType === 'redirect'" class="form-group">
          <label>重定向 URL</label>
          <input v-model="localRule.redirectUrl" type="text" placeholder="例如: https://example.com/new-image.png">
          <small>请求将返回 302 重定向到此 URL</small>
        </div>
        
        <!-- 响应结果替换 -->
        <div v-if="localRule.actionType === 'responseReplace'" class="form-group">
          <label>响应内容</label>
          <textarea v-model="localRule.responseContent" rows="8" placeholder="输入完整的响应内容"></textarea>
          <small>将完全替换原始响应内容</small>
        </div>

        <!-- 远程HTTP处理 -->
        <div v-if="localRule.actionType === 'remoteHTTP'">
          <div class="form-group">
            <label>远程服务URL</label>
            <input v-model="remoteHTTP.url" type="text" placeholder="http://localhost:3000/transform">
          </div>
          <div class="form-group">
            <label>HTTP方法</label>
            <select v-model="remoteHTTP.method">
              <option value="POST">POST</option>
              <option value="GET">GET</option>
              <option value="PUT">PUT</option>
            </select>
          </div>
          <div class="form-group">
            <label>超时时间(毫秒)</label>
            <input v-model.number="remoteHTTP.timeout" type="number" placeholder="5000">
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="remoteHTTP.sendOriginal" type="checkbox">
              发送原始响应体
            </label>
            <small>勾选后直接发送原始响应，否则发送JSON格式</small>
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="remoteHTTP.useResponse" type="checkbox">
              使用远程响应替换
            </label>
            <small>勾选后用远程服务返回的内容替换原响应</small>
          </div>
          <div class="form-group">
            <label>请求体模板(可选)</label>
            <textarea v-model="remoteHTTP.bodyTemplate" rows="4" placeholder='{"url": "{{url}}", "data": "{{body}}"}'></textarea>
            <small v-text="'支持变量: {{url}}, {{body}}'"></small>
          </div>
        </div>
        
        <!-- 请求转发 -->
        <div v-if="localRule.actionType === 'forwardRequest'">
          <div class="form-group">
            <label>目标URL</label>
            <input v-model="forwardConfig.targetUrl" type="text" placeholder="http://localhost:3000">
            <small>请求将被转发到此URL</small>
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="forwardConfig.keepPath" type="checkbox">
              保持原路径
            </label>
            <small>勾选后保留原请求路径，否则使用目标URL的路径</small>
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="forwardConfig.replaceHost" type="checkbox">
              替换Host头
            </label>
            <small>勾选后将Host头替换为目标主机</small>
          </div>
          <div class="form-group">
            <label>超时时间(毫秒)</label>
            <input v-model.number="forwardConfig.timeout" type="number" placeholder="30000">
          </div>
        </div>
        
        <!-- Webhook配置 -->
        <div class="form-group">
          <label class="checkbox-label">
            <input v-model="localRule.webhookEnabled" type="checkbox">
            启用 Webhook 推送
          </label>
        </div>
        
        <div v-if="localRule.webhookEnabled">
          <div class="form-group">
            <label>Webhook URL</label>
            <input v-model="localRule.webhookUrl" type="text" placeholder="http://localhost:3000/webhook">
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="localRule.webhookSync" type="checkbox">
              同步等待响应
            </label>
            <small>勾选后会等待webhook返回，并可用返回内容替换响应</small>
          </div>
        </div>
        
        <!-- 保存到文件 -->
        <div v-if="localRule.actionType === 'saveToFile'" class="form-group">
          <label>保存路径</label>
          <input v-model="localRule.saveFilePath" type="text" placeholder="/path/to/save.jsonl">
        </div>
      </div>
      <div class="drawer-footer">
        <button @click="$emit('close')" class="btn-secondary">取消</button>
        <button @click="handleSave" class="btn-primary">保存</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { InterceptRule, RemoteHTTPConfig, ForwardConfig } from '../types/intercept'

const props = defineProps<{
  visible: boolean
  rule: InterceptRule
}>()

const emit = defineEmits<{
  close: []
  save: [rule: InterceptRule]
}>()

// 创建本地副本
const localRule = ref<InterceptRule>({ ...props.rule })

// 远程HTTP配置
const remoteHTTP = ref<RemoteHTTPConfig>({
  url: '',
  method: 'POST',
  timeout: 5000,
  sendOriginal: false,
  useResponse: true,
  bodyTemplate: ''
})

// 请求转发配置
const forwardConfig = ref<ForwardConfig>({
  targetUrl: '',
  replaceHost: true,
  timeout: 30000,
  keepPath: true
})

// 监听visible变化，只在打开时同步数据
watch(() => props.visible, (newVisible) => {
  if (newVisible) {
    localRule.value = { ...props.rule }
    
    if (props.rule.remoteHTTP) {
      remoteHTTP.value = { ...props.rule.remoteHTTP }
    } else {
      remoteHTTP.value = {
        url: '',
        method: 'POST',
        timeout: 5000,
        sendOriginal: false,
        useResponse: true,
        bodyTemplate: ''
      }
    }
    
    if (props.rule.forwardRequest) {
      forwardConfig.value = { ...props.rule.forwardRequest }
    } else {
      forwardConfig.value = {
        targetUrl: '',
        replaceHost: true,
        timeout: 30000,
        keepPath: true
      }
    }
  }
})

const handleSave = () => {
  const ruleToSave = { ...localRule.value }
  
  // 如果是remoteHTTP类型，保存配置
  if (ruleToSave.actionType === 'remoteHTTP') {
    ruleToSave.remoteHTTP = { ...remoteHTTP.value }
  }
  
  // 如果是请求转发类型，保存配置
  if (ruleToSave.actionType === 'forwardRequest') {
    ruleToSave.forwardRequest = { ...forwardConfig.value }
  }
  
  emit('save', ruleToSave)
}
</script>

<style scoped>
.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1001;
  display: flex;
  justify-content: flex-end;
}

.drawer {
  width: 600px;
  background: white;
  display: flex;
  flex-direction: column;
  box-shadow: -2px 0 8px rgba(0, 0, 0, 0.15);
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #e0e0e0;
}

.drawer-header h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  font-size: 20px;
  cursor: pointer;
  color: #666;
  border-radius: 4px;
}

.close-btn:hover {
  background: #f0f0f0;
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #333;
  text-align: left;
}

.form-group input[type="text"],
.form-group input[type="number"],
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #dadce0;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
  box-sizing: border-box;
}

.form-group textarea {
  resize: vertical;
  font-family: 'Consolas', 'Monaco', monospace;
}

.form-group small {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: #666;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: normal !important;
}

.checkbox-label input[type="checkbox"] {
  width: auto;
  cursor: pointer;
}

.drawer-footer {
  display: flex;
  gap: 12px;
  padding: 16px;
  border-top: 1px solid #e0e0e0;
}

.btn-secondary,
.btn-primary {
  flex: 1;
  padding: 10px;
  border: 1px solid #dadce0;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.btn-secondary {
  background: white;
  color: #5f6368;
}

.btn-secondary:hover {
  background: #f8f9fa;
}

.btn-primary {
  background: #1a73e8;
  color: white;
  border-color: #1a73e8;
}

.btn-primary:hover {
  background: #1557b0;
}
</style>
