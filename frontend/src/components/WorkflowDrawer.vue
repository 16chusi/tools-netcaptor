<template>
  <div v-if="visible" class="drawer-overlay" @click="$emit('close')">
    <div class="drawer" @click.stop>
      <div class="drawer-header">
        <h3>🔄 任务流编排</h3>
        <button @click="$emit('close')" class="close-icon">✕</button>
      </div>
      <div class="drawer-content">
        <TaskList
          :tasks="tasks"
          :selectedTaskId="selectedTaskId"
          @select="selectTask"
          @create="createTask"
        />
        <div class="editor-area">
          <div class="canvas-wrapper">
            <StencilPanel v-if="currentTask" ref="stencilRef" :graph="graphInstance" />
            <FlowCanvas
              ref="canvasRef"
              :task="currentTask"
              @save="saveTask"
              @run="runTask"
              @clear="clearCanvas"
              @change="onTaskChange"
              @selectNode="onSelectNode"
              @graphReady="onGraphReady"
            />
            <PropertyPanel
              :visible="propertyVisible"
              :nodeType="selectedNode?.type"
              :nodeLabel="selectedNode?.label"
              :nodeData="selectedNode?.data"
              @close="propertyVisible = false"
              @save="onSaveNodeData"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Graph } from '@antv/x6'
import TaskList from './workflow/TaskList.vue'
import StencilPanel from './workflow/StencilPanel.vue'
import FlowCanvas from './workflow/FlowCanvas.vue'
import PropertyPanel from './workflow/PropertyPanel.vue'
import type { WorkflowTask } from '../types/workflow'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const tasks = ref<WorkflowTask[]>([])
const selectedTaskId = ref<string>()
const propertyVisible = ref(false)
const selectedNode = ref<any>(null)
const canvasRef = ref<any>(null)
const graphInstance = ref<any>(null)
const stencilRef = ref<any>(null)

function onGraphReady(graph: any) {
  graphInstance.value = graph
  console.log('[WorkflowDrawer] Graph 已就绪')
  
  // 初始化 Stencil
  setTimeout(() => {
    if (stencilRef.value?.initStencil) {
      stencilRef.value.initStencil()
    }
  }, 200)
}

const currentTask = computed(() => 
  tasks.value.find(t => t.id === selectedTaskId.value)
)

function selectTask(id: string) {
  selectedTaskId.value = id
}

function createTask() {
  const newTask: WorkflowTask = {
    id: Date.now().toString(),
    name: `任务 ${tasks.value.length + 1}`,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    nodes: [],
    edges: []
  }
  tasks.value.push(newTask)
  selectedTaskId.value = newTask.id
}

function saveTask() {
  if (!currentTask.value) return
  localStorage.setItem('workflow-tasks', JSON.stringify(tasks.value))
  alert('保存成功')
}

function runTask() {
  if (!currentTask.value) return
  alert('运行功能开发中...')
}

function clearCanvas() {
  if (!currentTask.value) return
  if (confirm('确定清空画布?')) {
    currentTask.value.nodes = []
    currentTask.value.edges = []
  }
}

function onTaskChange(task: WorkflowTask) {
  const index = tasks.value.findIndex(t => t.id === task.id)
  if (index >= 0) {
    tasks.value[index] = { ...task }
  }
}



function onSelectNode(node: any) {
  if (node.type === 'start' || node.type === 'end') {
    return
  }
  selectedNode.value = node
  propertyVisible.value = true
}

function onSaveNodeData(data: Record<string, any>) {
  if (!selectedNode.value) return
  selectedNode.value.data = { ...selectedNode.value.data, ...data }
  console.log('保存节点数据:', selectedNode.value)
}

// 加载任务
function loadTasks() {
  const stored = localStorage.getItem('workflow-tasks')
  if (stored) {
    try {
      tasks.value = JSON.parse(stored)
    } catch (e) {
      console.error('加载任务失败:', e)
    }
  }
}

loadTasks()
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
}

.drawer {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  width: 100%;
  height: 100%;
  background: white;
  display: flex;
  flex-direction: column;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #e0e0e0;
  background: #f8f9fa;
}

.drawer-header h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.close-icon {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: #5f6368;
  cursor: pointer;
  font-size: 20px;
  border-radius: 4px;
  transition: background 0.2s;
}

.close-icon:hover {
  background: #e8eaed;
}

.drawer-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.editor-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.canvas-wrapper {
  flex: 1;
  display: flex;
  position: relative;
  overflow: hidden;
}
</style>
