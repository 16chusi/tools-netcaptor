<template>
  <div class="node-panel">
    <div class="panel-header">组件</div>
    <div class="panel-body">
      <div
        v-for="config in nodeConfigs"
        :key="config.type"
        :class="['node-item', `node-${config.type}`]"
        @mousedown="config.type !== 'start' && config.type !== 'end' && onMouseDown($event, config)"
      >
        <span class="node-icon">{{ config.icon }}</span>
        <span class="node-label">{{ config.label }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { NODE_CONFIGS } from './nodeConfigs'
import type { NodeConfig } from '../../types/workflow'

const nodeConfigs = NODE_CONFIGS

const emit = defineEmits<{
  dragStart: [e: MouseEvent, config: NodeConfig]
}>()

function onMouseDown(e: MouseEvent, config: NodeConfig) {
  e.preventDefault()
  e.stopPropagation()
  console.log('[NodePanel] 鼠标按下:', config.label)
  emit('dragStart', e, config)
}
</script>

<style scoped>
.node-panel {
  width: 200px;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
  background: white;
}

.panel-header {
  padding: 12px 16px;
  border-bottom: 1px solid #e0e0e0;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.node-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  margin-bottom: 6px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  cursor: grab;
  background: white;
  transition: all 0.2s;
  font-size: 12px;
  user-select: none;
  -webkit-user-select: none;
}

.node-item:hover {
  border-color: #1890ff;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.node-item:active {
  cursor: grabbing;
  background: white;
}

.node-item::selection {
  background: transparent;
}

.node-item.node-start,
.node-item.node-end {
  cursor: not-allowed;
  opacity: 0.6;
}

.node-icon {
  font-size: 16px;
}

.node-label {
  flex: 1;
  font-weight: 500;
}
</style>
