<template>
  <div v-if="visible" class="stencil-panel" :style="{ left: position.x + 'px', top: position.y + 'px' }">
    <div class="stencil-header" @mousedown="startDrag">
      <span>组件</span>
      <button @click="$emit('close')" class="close-btn">✕</button>
    </div>
    <div class="stencil-groups">
      <div v-for="group in groups" :key="group.name" class="group">
        <div class="group-title" @click="toggleGroup(group.name)" :style="{ color: group.color }">
          <span class="group-icon">{{ expandedGroups[group.name] ? '▼' : '▶' }}</span>
          <span class="group-emoji">{{ group.icon }}</span>
          <span>{{ group.title }}</span>
        </div>
        <div v-show="expandedGroups[group.name]" class="group-content">
          <div
            v-for="node in group.nodes"
            :key="node.type"
            class="node-item"
            :style="{ background: node.color }"
            draggable="true"
            @dragstart="onDragStart($event, node)"
          >
            <span class="node-icon">{{ node.icon }}</span>
            <span class="node-label">{{ node.label }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { NODE_GROUPS } from './nodeConfigs'
import type { NodeConfig } from '../../types/workflow'

const props = defineProps<{
  graph: any
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const groups = NODE_GROUPS
const expandedGroups = reactive<Record<string, boolean>>(
  Object.fromEntries(groups.map(g => [g.name, true]))
)

const position = ref({ x: 10, y: 10 })
const isDragging = ref(false)

function startDrag(e: MouseEvent) {
  if ((e.target as HTMLElement).closest('.close-btn')) return
  
  isDragging.value = true
  const startX = e.clientX - position.value.x
  const startY = e.clientY - position.value.y
  
  function onMouseMove(e: MouseEvent) {
    position.value.x = e.clientX - startX
    position.value.y = e.clientY - startY
  }
  
  function onMouseUp() {
    isDragging.value = false
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  }
  
  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

function toggleGroup(name: string) {
  expandedGroups[name] = !expandedGroups[name]
}

function onDragStart(event: DragEvent, config: NodeConfig) {
  if (!props.graph) return
  
  const node = props.graph.createNode({
    width: 120,
    height: 40,
    shape: 'rect',
    markup: [
      { tagName: 'rect', selector: 'body' },
      { tagName: 'text', selector: 'icon' },
      { tagName: 'text', selector: 'label' }
    ],
    attrs: {
      body: {
        fill: config.color,
        stroke: config.color,
        rx: 6,
        ry: 6
      },
      icon: {
        text: config.icon,
        fontSize: 16,
        fill: '#333',
        refX: 12,
        refY: 20,
        textAnchor: 'start'
      },
      label: {
        text: config.label,
        fill: '#333',
        fontSize: 11,
        refX: 32,
        refY: 20,
        textAnchor: 'start'
      }
    },
    ports: {
      groups: {
        top: { position: 'top', attrs: { circle: { r: 6, magnet: true, stroke: '#0078D4', strokeWidth: 2, fill: '#fff' } } },
        right: { position: 'right', attrs: { circle: { r: 6, magnet: true, stroke: '#0078D4', strokeWidth: 2, fill: '#fff' } } },
        bottom: { position: 'bottom', attrs: { circle: { r: 6, magnet: true, stroke: '#0078D4', strokeWidth: 2, fill: '#fff' } } },
        left: { position: 'left', attrs: { circle: { r: 6, magnet: true, stroke: '#0078D4', strokeWidth: 2, fill: '#fff' } } }
      },
      items: [{ group: 'top' }, { group: 'right' }, { group: 'bottom' }, { group: 'left' }]
    },
    tools: [
      {
        name: 'button-remove',
        args: {
          x: '100%',
          y: 0,
          offset: { x: -10, y: 10 }
        }
      }
    ],
    data: { type: config.type }
  })
  
  event.dataTransfer!.effectAllowed = 'move'
  event.dataTransfer!.setData('application/x6-node', JSON.stringify(node.toJSON()))
}
</script>

<style scoped>
.stencil-panel {
  position: absolute;
  width: 200px;
  max-height: calc(100vh - 200px);
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  background: white;
  user-select: none;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1001;
  display: flex;
  flex-direction: column;
}

.stencil-header {
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
  padding: 12px 16px;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-radius: 6px 6px 0 0;
  cursor: move;
  flex-shrink: 0;
}

.stencil-groups {
  padding: 8px 0;
  overflow-y: auto;
  flex: 1;
}

.close-btn {
  width: 20px;
  height: 20px;
  border: none;
  background: transparent;
  color: #666;
  cursor: pointer;
  font-size: 14px;
  border-radius: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #e8eaed;
  color: #333;
}

.group {
  margin-bottom: 4px;
}

.group-title {
  padding: 8px 16px;
  font-size: 12px;
  font-weight: 500;
  background: #fafafa;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: background 0.2s;
}

.group-title:hover {
  background: #f0f0f0;
}

.group-icon {
  font-size: 10px;
  color: #666;
}

.group-emoji {
  font-size: 14px;
}

.group-content {
  padding: 6px 8px;
}

.node-item {
  width: 140px;
  height: 36px;
  margin: 0 auto 6px;
  border-radius: 6px;
  border: 1px solid rgba(0, 0, 0, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  cursor: move;
  transition: all 0.2s;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.node-item:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  border-color: rgba(0, 0, 0, 0.12);
}

.node-icon {
  font-size: 16px;
}

.node-label {
  font-size: 12px;
  font-weight: 500;
  color: #333;
}
</style>
