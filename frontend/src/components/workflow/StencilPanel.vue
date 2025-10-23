<template>
  <div ref="stencilRef" class="stencil-panel"></div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Graph } from '@antv/x6'
import { Stencil } from '@antv/x6-plugin-stencil'
import { NODE_CONFIGS } from './nodeConfigs'

const props = defineProps<{
  graph: any
}>()

const stencilRef = ref<HTMLDivElement>()
let stencil: Stencil | null = null

function initStencil() {
  if (!props.graph || !stencilRef.value) {
    console.log('[StencilPanel] graph 或 stencilRef 不存在')
    return
  }
  
  if (stencil) return
  
  console.log('[StencilPanel] 初始化 Stencil')
  
  stencil = new Stencil({
    title: '组件',
    target: props.graph,
    stencilGraphWidth: 200,
    stencilGraphHeight: 800,
    search: false,
    collapsable: false,
    validateNode: () => true,
    getDragNode: (node) => node.clone(),
    getDropNode: (node) => {
      const cloned = node.clone()
      const config = NODE_CONFIGS.find(c => c.type === node.getData()?.type)
      if (config) {
        // 保持颜色
        cloned.setAttrs({
          body: {
            fill: config.color,
            stroke: config.color,
            strokeWidth: 2,
            rx: 8,
            ry: 8
          }
        })
        // 添加连接点
        cloned.setProp('ports', {
          groups: {
            top: { 
              position: 'top',
              attrs: { 
                circle: { 
                  r: 6, 
                  magnet: true, 
                  stroke: '#1890ff', 
                  strokeWidth: 2,
                  fill: '#fff'
                } 
              } 
            },
            bottom: { 
              position: 'bottom',
              attrs: { 
                circle: { 
                  r: 6, 
                  magnet: true, 
                  stroke: '#1890ff', 
                  strokeWidth: 2,
                  fill: '#fff'
                } 
              } 
            }
          },
          items: [{ group: 'top' }, { group: 'bottom' }]
        })
        // 添加删除按钮
        cloned.addTools([
          {
            name: 'button-remove',
            args: {
              x: '100%',
              y: 0,
              offset: { x: -10, y: 10 },
            },
          },
        ])
      }
      return cloned
    },
    layoutOptions: {
      columns: 1,
      columnWidth: 180,
      rowHeight: 60,
      dx: 10,
      dy: 10
    },
    groups: [
      {
        name: 'basic',
        title: '基础组件',
        collapsable: false
      }
    ]
  })
  
  stencilRef.value.appendChild(stencil.container)
  
  // 确保 stencil 容器可以响应鼠标事件
  const stencilContainer = stencil.container as HTMLElement
  stencilContainer.style.position = 'relative'
  stencilContainer.style.zIndex = '10'
  
  // 创建节点
  const nodes = NODE_CONFIGS.filter(c => c.type !== 'start' && c.type !== 'end').map(config => {
    return props.graph!.createNode({
      width: 160,
      height: 50,
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
          strokeWidth: 2,
          rx: 8,
          ry: 8
        },
        icon: {
          text: config.icon,
          fill: '#333',
          fontSize: 20,
          refX: 20,
          refY: 25,
          textAnchor: 'middle',
          textVerticalAnchor: 'middle'
        },
        label: {
          text: config.label,
          fill: '#333',
          fontSize: 13,
          fontWeight: 500,
          refX: 50,
          refY: 25,
          textAnchor: 'start',
          textVerticalAnchor: 'middle'
        }
      },
      data: { type: config.type }
    })
  })
  
  stencil.load(nodes, 'basic')
  console.log('[StencilPanel] 节点已加载:', nodes.length)
}

onMounted(() => {
  console.log('[StencilPanel] 组件已挂载, graph:', !!props.graph)
  if (props.graph) {
    setTimeout(initStencil, 100)
  }
})

defineExpose({
  initStencil
})
</script>

<style scoped>
.stencil-panel {
  width: 200px;
  border-right: 1px solid #e0e0e0;
  background: white;
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
  position: relative;
  z-index: 10;
}

.stencil-panel :deep(.x6-widget-stencil) {
  background: white;
}

.stencil-panel :deep(.x6-widget-stencil-title) {
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
  padding: 12px 16px;
  font-size: 13px;
  font-weight: 600;
}

.stencil-panel :deep(.x6-widget-stencil-group-title) {
  background: #fafafa !important;
  padding: 8px 16px;
  font-size: 12px;
  font-weight: 500;
}

.stencil-panel :deep(.x6-graph) {
  background: white;
}

.stencil-panel :deep(.x6-graph-svg) {
  background: white;
}

.stencil-panel :deep(.x6-node) {
  cursor: move;
  user-select: none;
  pointer-events: auto;
}

.stencil-panel :deep(text) {
  user-select: none;
  pointer-events: none;
}

.stencil-panel :deep(.x6-widget-stencil-content) {
  position: relative;
  z-index: 10;
}

.stencil-panel :deep(.x6-graph-svg-stage) {
  pointer-events: auto;
}
</style>
