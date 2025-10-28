<template>
  <div class="flow-canvas">
    <div v-if="executionStatus" class="canvas-status">
      <span>{{ executionStatus.currentStep }}/{{ executionStatus.totalSteps }}</span>
      <span :class="'status-' + executionStatus.status">{{ getStatusText(executionStatus.status) }}</span>
    </div>
    <div ref="containerRef" class="canvas-container"></div>
  </div>
</template>

<script setup lang="ts">
import {onMounted, onUnmounted, ref, watch} from 'vue'
import {Graph} from '@antv/x6'
import {Dnd} from '@antv/x6-plugin-dnd'
import {NODE_CONFIGS} from './nodeConfigs'
import {ExecuteWorkflow, IsWebSocketRunning, StopWorkflow} from '../../../wailsjs/go/main/NetworkApp'
import {EventsOn} from '../../../wailsjs/runtime/runtime'
import {main} from '../../../wailsjs/go/models'
import {toast} from '../../utils/toast';

type WorkflowTask = main.WorkflowTask

const props = defineProps<{
  task?: any
}>()

const emit = defineEmits<{
  change: [task: any]
  selectNode: [node: any]
  graphReady: [graph: Graph]
}>()

const containerRef = ref<HTMLDivElement>()
let graph: Graph | null = null
let dnd: Dnd | null = null
const hasSelection = ref(false)
const isRunning = ref(false)
const executionStatus = ref<any>(null)

onMounted(() => {
  console.log('[FlowCanvas] 组件已挂载')
  
  // 立即初始化
  initGraph()
  setupDragDrop()
  if (props.task) {
    loadTask(props.task)
  }
  
  // 监听工作流状态
  EventsOn('workflow_status', (status: any) => {
    console.log('[FlowCanvas] 工作流状态:', status)
    executionStatus.value = status
    isRunning.value = status.status === 'running'
    
    // 高亮当前执行的节点
    if (status.currentNode && graph) {
      highlightNode(status.currentNode)
    }
  })
  
  EventsOn('workflow_error', (data: any) => {
    console.error('[FlowCanvas] 工作流错误:', data.error)
    const errorMsg = data.error || '未知错误'
    if (errorMsg.includes('超时') || errorMsg.includes('timeout')) {
      toast.error('执行超时：浏览器扩展未响应，请检查扩展是否已安装并连接')
    } else if (errorMsg.includes('未连接') || errorMsg.includes('not connected')) {
      toast.error('浏览器扩展未连接，请先安装并启用扩展')
    } else {
      toast.error('执行失败: ' + errorMsg)
    }
    isRunning.value = false
  })
})

onUnmounted(() => {
  graph?.dispose()
})

watch(() => props.task, (newTask, oldTask) => {
  if (newTask && graph && newTask.id !== oldTask?.id) {
    loadTask(newTask)
  }
}, { deep: false })

function initGraph() {
  if (!containerRef.value) return

  console.log('[FlowCanvas] 初始化画布')

  graph = new Graph({
    container: containerRef.value,
    width: containerRef.value.offsetWidth,
    height: containerRef.value.offsetHeight,
    autoResize: true,
    panning: {
      enabled: true,
      modifiers: 'shift'
    },
    mousewheel: {
      enabled: true,
      modifiers: ['ctrl', 'meta']
    },
    grid: {
      size: 10,
      visible: true
    },

    highlighting: {
      magnetAvailable: {
        name: 'className',
        args: {
          className: 'available-magnet'
        }
      },
      magnetAdsorbed: {
        name: 'stroke',
        args: {
          attrs: {
            fill: '#1890ff',
            stroke: '#1890ff'
          }
        }
      },
      nodeAvailable: {
        name: 'className',
        args: {
          className: 'node-available'
        }
      }
    },
    connecting: {
      snap: {
        radius: 50
      },
      allowBlank: false,
      allowLoop: false,
      allowNode: false,
      highlight: true,
      connector: 'rounded',
      router: {
        name: 'orth'
      },
      createEdge() {
        return this.createEdge({
          attrs: {
            line: {
              stroke: '#0078D4',
              strokeWidth: 2,
              targetMarker: 'classic'
            }
          }
        })
      },
      validateConnection({ targetMagnet }) {
        return !!targetMagnet
      }
    },
    interacting: {
      nodeMovable: true
    }
  })

  // 监听事件
  graph.on('node:change:position', () => {
    emitChange()
  })

  graph.on('edge:connected', ({ edge }) => {
    console.log('[FlowCanvas] 边已连接:', edge.id, edge.getSourceCellId(), '->', edge.getTargetCellId())
    
    // 保存端口信息
    const sourcePort = edge.getSourcePortId()
    const targetPort = edge.getTargetPortId()
    edge.setData({
      sourcePort,
      targetPort
    })
    
    edge.addTools([
      {
        name: 'button-remove',
        args: { distance: '50%' },
      },
    ])
    emitChange()
  })
  
  graph.on('edge:removed', () => {
    emitChange()
  })
  
  graph.on('node:removed', () => {
    emitChange()
  })

  // 监听节点点击
  graph.on('node:click', ({ node }) => {
    const data = node.getData() || {}
    const nodeData = {
      id: node.id,
      type: data.type,
      label: node.attr('label/text'),
      data: data // 传递完整的 data 对象
    }
    console.log('[FlowCanvas] 节点点击:', nodeData)
    emit('selectNode', nodeData)
  })

  // 右键菜单
  graph.on('node:contextmenu', ({ node, e }) => {
    e.preventDefault()
    showContextMenu(node, e.clientX, e.clientY)
  })

  graph.on('edge:contextmenu', ({ edge, e }) => {
    e.preventDefault()
    showContextMenu(edge, e.clientX, e.clientY)
  })
  
  // 初始化 Dnd
  dnd = new Dnd({
    target: graph,
    scaled: false
  })
  
  console.log('[FlowCanvas] Graph 已初始化')
  
  // 调试：暴露 graph 到 window
  if (typeof window !== 'undefined') {
    (window as any).__x6_graph__ = graph
  }
  
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

  const nodeConfig: any = {
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
        fill: '#333',
        fontSize: 12
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
    data: { type } // 初始化时只设置 type，其他属性由 PropertyPanel 设置
  }

  // 只有非开始/结束节点才添加删除按钮
  if (type !== 'start' && type !== 'end') {
    nodeConfig.tools = [
      {
        name: 'button-remove',
        args: {
          x: '100%',
          y: 0,
          offset: { x: -10, y: 10 },
        },
      },
    ]
  }

  const node = graph.addNode(nodeConfig)
  
  console.log('[FlowCanvas] 节点已添加:', node.id)
  return node
}

function createDefaultNodes() {
  if (!graph) return

  console.log('[FlowCanvas] 创建默认节点')

  // 开始节点 - Windows 11 蓝色
  graph.addNode({
    x: 300,
    y: 50,
    width: 60,
    height: 60,
    shape: 'circle',
    attrs: {
      body: { fill: 'rgba(0, 120, 212, 0.15)', stroke: '#0078D4', strokeWidth: 2 },
      label: { text: '开始', fill: '#333', fontSize: 12 }
    },
    ports: {
      groups: {
        bottom: { position: 'bottom', attrs: { circle: { r: 6, magnet: true, stroke: '#0078D4', strokeWidth: 2, fill: '#fff' } } }
      },
      items: [{ group: 'bottom' }]
    },
    data: { type: 'start' }
  })

  // 结束节点 - Windows 11 红色
  graph.addNode({
    x: 300,
    y: 400,
    width: 60,
    height: 60,
    shape: 'circle',
    attrs: {
      body: { fill: 'rgba(209, 52, 56, 0.15)', stroke: '#D13438', strokeWidth: 2 },
      label: { text: '结束', fill: '#333', fontSize: 12 }
    },
    ports: {
      groups: {
        top: { position: 'top', attrs: { circle: { r: 6, magnet: true, stroke: '#0078D4', strokeWidth: 2, fill: '#fff' } } }
      },
      items: [{ group: 'top' }]
    },
    data: { type: 'end' }
  })
  
  console.log('[FlowCanvas] 节点数量:', graph.getNodes().length)
}

function loadTask(task: WorkflowTask) {
  if (!graph) return

  graph.clearCells()

  if (task.nodes.length === 0) {
    createDefaultNodes()
  } else {
    // 加载节点
    const nodeIdMap = new Map()
    task.nodes.forEach(node => {
      let newNode
      if (node.type === 'start') {
        newNode = graph!.addNode({
          id: node.id,
          x: node.x,
          y: node.y,
          width: 60,
          height: 60,
          shape: 'circle',
          attrs: {
            body: { fill: 'rgba(0, 120, 212, 0.15)', stroke: '#0078D4', strokeWidth: 2 },
            label: { text: '开始', fill: '#333', fontSize: 12 }
          },
          ports: {
            groups: {
              bottom: { position: 'bottom', attrs: { circle: { r: 6, magnet: true, stroke: '#0078D4', strokeWidth: 2, fill: '#fff' } } }
            },
            items: [{ group: 'bottom' }]
          },
          data: { type: 'start' }
        })
      } else if (node.type === 'end') {
        newNode = graph!.addNode({
          id: node.id,
          x: node.x,
          y: node.y,
          width: 60,
          height: 60,
          shape: 'circle',
          attrs: {
            body: { fill: 'rgba(209, 52, 56, 0.15)', stroke: '#D13438', strokeWidth: 2 },
            label: { text: '结束', fill: '#333', fontSize: 12 }
          },
          ports: {
            groups: {
              top: { position: 'top', attrs: { circle: { r: 6, magnet: true, stroke: '#0078D4', strokeWidth: 2, fill: '#fff' } } }
            },
            items: [{ group: 'top' }]
          },
          data: { type: 'end' }
        })
      } else {
        const config = NODE_CONFIGS.find(c => c.type === node.type)
        if (config) {
          newNode = graph!.addNode({
            id: node.id,
            x: node.x,
            y: node.y,
            width: 120,
            height: 40,
            shape: 'rect',
            attrs: {
              body: {
                fill: config.color,
                stroke: config.color,
                rx: 6,
                ry: 6
              },
              label: {
                text: config.label,
                fill: '#333',
                fontSize: 12
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
            data: node.data || { type: node.type },
            tools: [
              {
                name: 'button-remove',
                args: {
                  x: '100%',
                  y: 0,
                  offset: { x: -10, y: 10 },
                },
              },
            ]
          })
          if (newNode && node.data) {
            newNode.setData({ type: node.type, ...node.data })
          }
        }
      }
      if (newNode) {
        nodeIdMap.set(node.id, newNode)
      }
    })
    
    // 加载边
    if (task.edges && task.edges.length > 0) {
      task.edges.forEach(edge => {
        const newEdge = graph!.addEdge({
          id: edge.id,
          source: edge.source,
          target: edge.target,
          attrs: {
            line: {
              stroke: '#0078D4',
              strokeWidth: 2,
              targetMarker: 'classic'
            }
          },
          tools: [
            {
              name: 'button-remove',
              args: { distance: '50%' },
            },
          ]
        })
        // 恢复端口信息
        if (edge.sourcePort || edge.targetPort) {
          newEdge.setData({
            sourcePort: edge.sourcePort,
            targetPort: edge.targetPort
          })
        }
      })
      console.log('[FlowCanvas] 已加载', task.edges.length, '条边')
    }
  }
}

function showContextMenu(cell: any, clientX: number, clientY: number) {
  const menu = document.createElement('div')
  menu.className = 'context-menu'
  menu.style.left = `${clientX}px`
  menu.style.top = `${clientY}px`
  menu.innerHTML = `
    <div class="menu-item" data-action="delete">🗑️ 删除</div>
  `
  
  menu.addEventListener('click', (event) => {
    const target = event.target as HTMLElement
    if (target.dataset.action === 'delete') {
      graph?.removeCell(cell)
    }
    document.body.removeChild(menu)
  })
  
  document.body.appendChild(menu)
  
  const closeMenu = () => {
    if (document.body.contains(menu)) {
      document.body.removeChild(menu)
    }
    document.removeEventListener('click', closeMenu)
  }
  
  setTimeout(() => {
    document.addEventListener('click', closeMenu)
  }, 100)
}



function highlightNode(nodeId: string) {
  if (!graph) return
  
  // 清除之前的高亮
  graph.getNodes().forEach(node => {
    node.attr('body/stroke', node.getData()?.type === 'start' ? '#0078D4' : 
                             node.getData()?.type === 'end' ? '#D13438' : 
                             node.attr('body/fill'))
    node.attr('body/strokeWidth', 1)
  })
  
  // 高亮当前节点
  const node = graph.getCellById(nodeId)
  if (node) {
    node.attr('body/stroke', '#ff4d4f')
    node.attr('body/strokeWidth', 3)
  }
}

function getStatusText(status: string): string {
  const statusMap: Record<string, string> = {
    running: '运行中',
    success: '成功',
    failed: '失败',
    stopped: '已停止'
  }
  return statusMap[status] || status
}

let isEmitting = false
function emitChange() {
  if (!graph || !props.task || isEmitting) return
  
  isEmitting = true
  setTimeout(() => {
    console.log('[FlowCanvas] ========== emitChange 开始 ==========')
    
    const nodes = graph!.getNodes().map((node, index) => {
      const nodeData = node.getData() || {}
      console.log(`[FlowCanvas] 序列化节点[${index}] - ID: ${node.id}`)
      console.log(`[FlowCanvas] 序列化节点[${index}] - getData():`, nodeData)
      
      const serializedNode = {
        id: node.id,
        type: nodeData.type || 'unknown',
        x: node.position().x,
        y: node.position().y,
        label: node.attr('label/text') as string,
        data: nodeData // 这里包含 type 和其他所有属性（如 url, selector 等）
      }
      
      console.log(`[FlowCanvas] 序列化节点[${index}] - 结果:`, serializedNode)
      return serializedNode
    })

    const edges = graph!.getEdges().map((edge, index) => {
      const edgeData = {
        id: edge.id,
        source: edge.getSourceCellId(),
        sourcePort: edge.getSourcePortId(),
        target: edge.getTargetCellId(),
        targetPort: edge.getTargetPortId()
      }
      console.log(`[FlowCanvas] 序列化边[${index}]:`, edgeData)
      return edgeData
    })
    
    console.log('[FlowCanvas] 总计序列化:', nodes.length, '个节点,', edges.length, '条边')

    const updatedTask = {
      id: props.task!.id,
      name: props.task!.name,
      description: props.task!.description,
      createdAt: props.task!.createdAt,
      updatedAt: new Date().toISOString(),
      nodes,
      edges
    }
    console.log('[FlowCanvas] 发送 change 事件:', updatedTask)
    emit('change', updatedTask)
    
    isEmitting = false
  }, 0)
}
</script>

<style scoped>
.flow-canvas {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f5f5f5;
}

.canvas-status {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 10;
  background: white;
  padding: 8px 16px;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #666;
}

.status-running {
  color: #1890ff;
  font-weight: 500;
}

.status-success {
  color: #52c41a;
  font-weight: 500;
}

.status-failed {
  color: #ff4d4f;
  font-weight: 500;
}

.status-stopped {
  color: #faad14;
  font-weight: 500;
}

.canvas-container {
  flex: 1;
  position: relative;
  z-index: 1;
  min-height: 0;
  width: 100%;
  height: 100%;
  user-select: none;
}

:deep(.context-menu) {
  position: fixed;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  z-index: 9999;
  min-width: 120px;
}

:deep(.menu-item) {
  padding: 8px 16px;
  cursor: pointer;
  font-size: 12px;
  transition: background 0.2s;
}

:deep(.menu-item:hover) {
  background: #f5f5f5;
}

:deep(.available-magnet) {
  fill: #1890ff !important;
  stroke: #1890ff !important;
  stroke-width: 3 !important;
}

:deep(.x6-node text) {
  user-select: none;
  pointer-events: none;
}

:deep(.x6-graph-svg) {
  user-select: none;
}

:deep(.x6-graph-svg-stage) {
  user-select: none;
}

:deep(.x6-port-body) {
  fill: #fff;
  stroke: #0078D4;
  stroke-width: 2;
}

:deep(.x6-port) {
  opacity: 0;
  transition: opacity 0.2s;
}

:deep(.x6-node:hover .x6-port) {
  opacity: 1;
}

:deep(.available-magnet .x6-port-body) {
  fill: #0078D4 !important;
  stroke: #fff !important;
  stroke-width: 3 !important;
  r: 8 !important;
}
</style>
