<template>
  <div v-if="visible" class="drawer-overlay">
    <div class="drawer" @click.stop>
      <div class="drawer-header">
        <h3>{{ rule.id ? '编辑规则' : '新建规则' }}</h3>
        <button @click="$emit('close')" class="close-btn">✕</button>
      </div>
      <div class="drawer-body">
        <div class="form-group">
          <label>规则名称</label>
          <input v-model="rule.name" type="text" placeholder="例如: Mock API 响应">
        </div>
        
        <div class="form-group">
          <label>URL 模式</label>
          <input v-model="rule.urlPattern" type="text" placeholder="例如: *.example.com 或 /api/*">
          <small>支持通配符 *，例如: *.example.com, /api/*, */user/*</small>
        </div>
        
        <div class="form-group">
          <label>操作类型</label>
          <select v-model="rule.actionType">
            <option value="findReplace">内容替换</option>
            <option value="redirect">重定向</option>
            <option value="responseReplace">响应结果替换</option>
          </select>
        </div>
        
        <!-- 内容替换 -->
        <div v-if="rule.actionType === 'findReplace'">
          <div class="form-group">
            <label>查找</label>
            <input v-model="rule.findText" type="text" placeholder="输入要查找的内容">
          </div>
          <div class="form-group">
            <label>替换</label>
            <input v-model="rule.replaceText" type="text" placeholder="输入替换后的内容">
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="rule.useRegex" type="checkbox">
              使用正则表达式
            </label>
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input v-model="rule.replaceAll" type="checkbox">
              替换全部匹配项
            </label>
          </div>
        </div>
        
        <!-- 重定向 -->
        <div v-if="rule.actionType === 'redirect'" class="form-group">
          <label>重定向 URL</label>
          <input v-model="rule.redirectUrl" type="text" placeholder="例如: https://example.com/new-image.png">
          <small>请求将返回 302 重定向到此 URL</small>
        </div>
        
        <!-- 响应结果替换 -->
        <div v-if="rule.actionType === 'responseReplace'" class="form-group">
          <label>响应内容</label>
          <textarea v-model="rule.responseContent" rows="8" placeholder="输入完整的响应内容"></textarea>
          <small>将完全替换原始响应内容</small>
        </div>
        
        <div class="form-group">
          <label class="checkbox-label">
            <input v-model="rule.webhookEnabled" type="checkbox">
            启用 Webhook 推送
          </label>
        </div>
        
        <div v-if="rule.webhookEnabled" class="form-group">
          <label>Webhook URL</label>
          <input v-model="rule.webhookUrl" type="text" placeholder="http://localhost:3000/webhook">
        </div>
        
        <div class="examples-section">
          <div class="examples-title">💡 使用示例</div>
          
          <div class="example-item">
            <div class="example-title">示例 1: 内容替换</div>
            <div class="example-desc">操作类型: <strong>内容替换</strong></div>
            <div class="example-field">查找: <code>error</code></div>
            <div class="example-field">替换: <code>success</code></div>
            <div class="example-desc">将响应中所有 "error" 替换为 "success"</div>
          </div>
          
          <div class="example-item">
            <div class="example-title">示例 2: 正则替换</div>
            <div class="example-desc">操作类型: <strong>内容替换</strong> + 勾选正则</div>
            <div class="example-field">查找: <code>"status":\s*"\w+"</code></div>
            <div class="example-field">替换: <code>"status": "success"</code></div>
          </div>
          
          <div class="example-item">
            <div class="example-title">示例 3: 重定向</div>
            <div class="example-desc">操作类型: <strong>重定向</strong></div>
            <div class="example-field">重定向 URL: <code>https://example.com/new.png</code></div>
          </div>
          
          <div class="example-item">
            <div class="example-title">示例 4: 响应结果替换</div>
            <div class="example-desc">操作类型: <strong>响应结果替换</strong></div>
            <div class="example-code">{"status": "success", "data": []}</div>
            <div class="example-desc">完全替换整个响应内容</div>
          </div>
        </div>
      </div>
      <div class="drawer-footer">
        <button @click="$emit('close')" class="footer-btn">取消</button>
        <button @click="handleSave" class="footer-btn primary">保存</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'
import type { InterceptRule } from '../types/intercept'
import {ShowErrorDialog} from '../../wailsjs/go/main/NetworkApp'

const props = defineProps<{
  visible: boolean
  initialRule?: InterceptRule
}>()

const emit = defineEmits(['close', 'save'])

const rule = reactive<InterceptRule>({
  id: '',
  name: '',
  enabled: true,
  urlPattern: '',
  actionType: 'findReplace',
  findText: '',
  replaceText: '',
  useRegex: false,
  replaceAll: true,
  responseContent: '',
  redirectUrl: '',
  webhookUrl: '',
  webhookEnabled: false
})

watch(() => props.initialRule, (newRule) => {
  if (newRule) {
    Object.assign(rule, newRule)
  } else {
    rule.id = ''
    rule.name = ''
    rule.enabled = true
    rule.urlPattern = ''
    rule.actionType = 'findReplace'
    rule.findText = ''
    rule.replaceText = ''
    rule.useRegex = false
    rule.replaceAll = true
    rule.responseContent = ''
    rule.redirectUrl = ''
    rule.webhookUrl = ''
    rule.webhookEnabled = false
  }
}, { immediate: true })

function handleSave() {
  if (!rule.name || !rule.urlPattern) {
    ShowErrorDialog('验证错误', '请填写规则名称和 URL 模式')
    return
  }
  
  if (rule.actionType === 'findReplace' && !rule.findText) {
    ShowErrorDialog('验证错误', '请填写查找内容')
    return
  }
  
  if (rule.actionType === 'responseReplace' && !rule.responseContent) {
    ShowErrorDialog('验证错误', '请填写响应内容')
    return
  }
  
  if (rule.actionType === 'redirect' && !rule.redirectUrl) {
    ShowErrorDialog('验证错误', '请填写重定向 URL')
    return
  }
  
  if (!rule.id) {
    rule.id = 'rule-' + Date.now()
  }
  
  emit('save', { ...rule })
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
  display: flex;
  align-items: center;
  gap: 12px;
}

.form-group label {
  min-width: 100px;
  font-size: 13px;
  font-weight: 600;
  color: #333;
  text-align: left;
  flex-shrink: 0;
}

.form-group input[type="text"],
.form-group select,
.form-group textarea {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #dadce0;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #1a73e8;
}

.form-group textarea {
  resize: vertical;
  font-family: 'Consolas', 'Monaco', monospace;
}

.form-group small {
  flex-basis: 100%;
  margin-left: 112px;
  font-size: 11px;
  color: #666;
  text-align: left;
}

.form-group:has(textarea),
.form-group:has(small) {
  flex-wrap: wrap;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: normal !important;
  min-width: auto;
  text-align: left;
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

.footer-btn {
  flex: 1;
  padding: 10px;
  border: 1px solid #dadce0;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  color: #5f6368;
}

.footer-btn:hover {
  background: #f8f9fa;
}

.footer-btn.primary {
  background: #1a73e8;
  color: white;
  border-color: #1a73e8;
}

.footer-btn.primary:hover {
  background: #1557b0;
}

.examples-section {
  margin-top: 24px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 4px;
  border: 1px solid #e0e0e0;
}

.examples-title {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  margin-bottom: 12px;
}

.example-item {
  margin-bottom: 16px;
  padding: 12px;
  background: white;
  border-radius: 4px;
  border: 1px solid #e0e0e0;
  text-align: left;
}

.example-item:last-child {
  margin-bottom: 0;
}

.example-title {
  font-size: 13px;
  font-weight: 600;
  color: #1a73e8;
  margin-bottom: 6px;
  text-align: left;
}

.example-desc {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
  text-align: left;
}

.example-field {
  font-size: 12px;
  color: #333;
  margin: 4px 0;
  text-align: left;
}

.example-code {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 11px;
  background: #f1f3f4;
  padding: 8px;
  border-radius: 3px;
  margin-top: 6px;
  white-space: pre;
  color: #333;
  text-align: left;
}

.example-field code {
  background: #f1f3f4;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 11px;
  color: #d93025;
}
</style>
