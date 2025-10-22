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
    getDragNode: (node) => node.clone(),
    getDropNode: (node) => node.clone(),
    groups: [
      {
        name: 'basic',
        title: '基础组件'
      }
    ]
  })
  
  stencilRef.value.appendChild(stencil.container)
  
  // 创建节点
  const nodes = NODE_CONFIGS.filter(c => c.type !== 'start' && c.type !== 'end').map(config => {
    return props.graph!.createNode({
      width: 120,
      height: 40,
      shape: 'rect',
      label: config.label,
      attrs: {
        body: {
          fill: config.color,
          stroke: config.color,
          rx: 6,
          ry: 6
        },
        label: {
          text: config.label,
          fill: '#fff',
          fontSize: 12
        }
      },
      ports: {
        groups: {
          top: { position: 'top', attrs: { circle: { r: 4, magnet: true, stroke: '#fff', fill: config.color } } },
          bottom: { position: 'bottom', attrs: { circle: { r: 4, magnet: true, stroke: '#fff', fill: config.color } } }
        },
        items: [{ group: 'top' }, { group: 'bottom' }]
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
</style>
