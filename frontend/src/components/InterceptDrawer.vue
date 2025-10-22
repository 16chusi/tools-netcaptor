<template>
  <div v-if="visible" class="drawer-overlay">
    <div class="drawer" @click.stop>
      <div class="drawer-header">
        <h3>拦截规则</h3>
        <button @click="$emit('close')" class="close-btn">✕</button>
      </div>
      <div class="drawer-body">
        <div v-if="rules.length === 0" class="empty">暂无拦截规则</div>
        <div v-for="rule in rules" :key="rule.id" class="rule-item">
          <div class="rule-header">
            <label class="switch">
              <input type="checkbox" :checked="rule.enabled" @change="$emit('toggle', rule.id)">
              <span class="slider"></span>
            </label>
            <span class="rule-name">{{ rule.name }}</span>
            <div class="rule-actions">
              <button @click="$emit('edit', rule)" class="action-btn">编辑</button>
              <button @click="$emit('delete', rule.id)" class="action-btn delete">删除</button>
            </div>
          </div>
          <div class="rule-info">
            <span class="label">URL:</span> {{ rule.urlPattern }}
          </div>
          <div class="rule-info">
            <span class="label">类型:</span> {{ rule.actionType === 'findReplace' ? '内容替换' : rule.actionType === 'redirect' ? '重定向' : '响应结果替换' }}
          </div>
        </div>
      </div>
      <div class="drawer-footer">
        <button @click="$emit('create')" class="footer-btn primary">➕ 新建</button>
        <button @click="$emit('import')" class="footer-btn">📥 导入</button>
        <button @click="$emit('export')" class="footer-btn">📤 导出</button>
        <button @click="$emit('close')" class="footer-btn">✕ 关闭</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { InterceptRule } from '../types/intercept'

defineProps<{
  visible: boolean
  rules: InterceptRule[]
}>()

defineEmits(['close', 'toggle', 'edit', 'delete', 'import', 'export', 'create'])
</script>

<style scoped>
.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  justify-content: flex-end;
}

.drawer {
  width: 500px;
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

.empty {
  text-align: center;
  color: #999;
  padding: 40px;
}

.rule-item {
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 12px;
}

.rule-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
  min-width: 0;
}

.switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 20px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  transition: 0.3s;
  border-radius: 20px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 14px;
  width: 14px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: #1a73e8;
}

input:checked + .slider:before {
  transform: translateX(20px);
}

.rule-name {
  flex: 1;
  font-weight: 600;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.rule-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.action-btn {
  padding: 4px 12px;
  border: 1px solid #dadce0;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  color: #5f6368;
}

.action-btn:hover {
  background: #f8f9fa;
}

.action-btn.delete {
  color: #d32f2f;
  border-color: #d32f2f;
}

.rule-info {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rule-info .label {
  font-weight: 600;
  color: #333;
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
</style>
