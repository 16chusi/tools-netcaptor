<template>
  <div class="flow-canvas">
    <div class="canvas-toolbar">
      <button @click="$emit('save')" class="toolbar-btn">💾 保存</button>
      <button @click="$emit('run')" class="toolbar-btn primary">▶️ 运行</button>
      <button @click="$emit('clear')" class="toolbar-btn">🗑️ 清空</button>
    </div>
    <div ref="containerRef" class="canvas-container"></div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { Graph } from '@antv/x6'
import { Dnd } from '@antv/x6-plugin-dnd'
import type { WorkflowTask } from '../../types/workflow'

const props = defineProps<{
  task?: WorkflowTask
}>()

const emit = defineEmits<{
  save: []
  run: []
  clear: []
  change: [task: WorkflowTask]
  selectNode: [node: any]
  graphReady: [graph: Graph]
}>()

const containerRef = ref<HTMLDivElement>()
let graph: Graph | null = null
let dnd: Dnd | null = null

onMounted(() => {
  console.log('[FlowCanvas] 组件已挂载')
  
  // 立即初始化
  initGraph()
  setupDragDrop()
  if (props.task) {
    loadTask(props.task)
  } else {
    createDefaultNodes()
  }
})

onUnmounted(() => {
  graph?.dispose()
})

watch(() => props.task, (newTask) => {
  if (newTask && graph) {
    loadTask(newTask)
  }
})

function initGraph() {
  if (!containerRef.value) return

  console.log('[FlowCanvas] 初始化画布')

  graph = new Graph({
    container: containerRef.value,
    width: containerRef.value.offsetWidth,
    height: containerRef.value.offsetHeight,
    autoResize: true,
    panning: {
      enabled: true
    },
    mousewheel: {
      enabled: true,
      modifiers: ['ctrl', 'meta']
    },
    grid: {
      size: 10,
      visible: true
    },
    connecting: {
      snap: true,
      allowBlank: false,
      allowLoop: false,
      highlight: true,
      connector: 'rounded',
      router: {
        name: 'manhattan'
      }
    },
    highlighting: {
      magnetAvailable: {
        name: 'stroke',
        args: {
          attrs: {
            fill: '#fff',
            stroke: '#1890ff'
          }
        }
      }
    }
  })

  // 监听事件
  graph.on('node:added', () => {
    emitChange()
  })

  graph.on('node:removed', () => {
    emitChange()
  })

  graph.on('edge:connected', () => {
    emitChange()
  })

  // 监听节点点击
  graph.on('node:click', ({ node }) => {
    const nodeData = {
      id: node.id,
      type: node.getData()?.type,
      label: node.attr('label/text'),
      data: node.getData()
    }
    emit('selectNode', nodeData)
  })
  
  // 初始化 Dnd
  dnd = new Dnd({
    target: graph,
    scaled: false
  })
  
  console.log('[FlowCanvas] Graph 已初始化')
  
  // 通知父组件 Graph 已就绪
  emit('graphReady', graph)
}

function setupDragDrop() {
  // 使用 X6 内置的 Dnd,不需要手动处理
  console.log('[FlowCanvas] 拖拽已就绪')
}

function handleDrop(config: any, x: number, y: number) {
  console.log('[FlowCanvas] 处理拖拽:', config, x, y)
  addNode(config.type, config.label, x, y, config.color)
}

function startDrag(nodeConfig: any, e: MouseEvent) {
  if (!dnd || !graph) {
    console.error('[FlowCanvas] Dnd 或 Graph 未初始化')
    return
  }
  
  console.log('[FlowCanvas] 开始拖拽', nodeConfig)
  
  // 创建临时节点
  const node = graph.createNode(nodeConfig)
  dnd.start(node, e)
}

defineExpose({
  dnd,
  handleDrop,
  startDrag
})

function addNode(type: string, label: string, x: number, y: number, color: string) {
  if (!graph) {
    console.error('[FlowCanvas] graph 未初始化')
    return
  }

  console.log('[FlowCanvas] 添加节点:', label, x, y)

  const node = graph.addNode({
    x,
    y,
    width: 120,
    height: 40,
    shape: 'rect',
    attrs: {
      body: {
        fill: color,
        stroke: color,
        rx: 6,
        ry: 6
      },
      label: {
        text: label,
        fill: '#fff',
        fontSize: 12
      }
    },
    ports: {
      groups: {
        top: { position: 'top', attrs: { circle: { r: 4, magnet: true, stroke: '#fff', fill: color } } },
        bottom: { position: 'bottom', attrs: { circle: { r: 4, magnet: true, stroke: '#fff', fill: color } } }
      },
      items: [{ group: 'top' }, { group: 'bottom' }]
    },
    data: { type }
  })
  
  console.log('[FlowCanvas] 节点已添加:', node.id)
  return node
}

function createDefaultNodes() {
  if (!graph) return

  console.log('[FlowCanvas] 创建默认节点')

  // 开始节点
  addNode('start', '开始', 300, 50, '#52c41a')

  // 结束节点
  addNode('end', '结束', 300, 400, '#f5222d')
  
  console.log('[FlowCanvas] 节点数量:', graph.getNodes().length)
}

function loadTask(task: WorkflowTask) {
  if (!graph) return

  graph.clearCells()

  task.nodes.forEach(node => {
    const config = { type: node.type, label: node.label, color: '#1890ff' }
    addNode(config.type, config.label, node.x, node.y, config.color)
  })
}

function emitChange() {
  if (!graph || !props.task) return

  const nodes = graph.getNodes().map(node => ({
    id: node.id,
    type: node.getData()?.type || 'unknown',
    x: node.position().x,
    y: node.position().y,
    label: node.attr('label/text') as string,
    data: node.getData()
  }))

  const edges = graph.getEdges().map(edge => ({
    id: edge.id,
    source: edge.getSourceCellId(),
    target: edge.getTargetCellId()
  }))

  emit('change', {
    ...props.task,
    nodes,
    edges,
    updatedAt: new Date().toISOString()
  })
}
</script>

<style scoped>
.flow-canvas {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}

.canvas-toolbar {
  padding: 8px 12px;
  border-bottom: 1px solid #e0e0e0;
  background: white;
  display: flex;
  gap: 8px;
}

.toolbar-btn {
  padding: 6px 12px;
  border: 1px solid #d9d9d9;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.toolbar-btn:hover {
  border-color: #1890ff;
  color: #1890ff;
}

.toolbar-btn.primary {
  background: #1890ff;
  color: white;
  border-color: #1890ff;
}

.toolbar-btn.primary:hover {
  background: #40a9ff;
}

.canvas-container {
  flex: 1;
  position: relative;
  z-index: 1;
  min-height: 0;
  width: 100%;
  height: 100%;
}
</style>
