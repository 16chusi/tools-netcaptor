<template>
  <div v-if="visible" class="drawer-overlay" @click="$emit('close')">
    <div class="drawer" @click.stop>
      <div class="drawer-header">
        <h3>🔄 任务流编排</h3>
        <div class="header-controls">
          <button @click="toggleWebSocket" :class="['ws-btn', wsRunning ? 'danger' : 'success']">
            {{ wsRunning ? '⏹️ 停止服务' : '▶️ 启动服务' }}
          </button>
          <span v-if="wsRunning" @click="copyPort" class="ws-port" title="点击复制端口">端口: {{ wsPort }}</span>
          <button @click="$emit('close')" class="close-icon">✕</button>
        </div>
      </div>
      <div class="drawer-content">
        <TaskList
          :tasks="tasks"
          :selectedTaskId="selectedTaskId"
          @select="selectTask"
          @create="createTask"
          @delete="deleteTask"
          @rename="renameTask"
        />
        <div class="editor-area">
          <div class="canvas-wrapper">
            <StencilPanel v-if="currentTask" ref="stencilRef" :graph="graphInstance" />
            <FlowCanvas
              ref="canvasRef"
              :task="currentTask"
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
import { ref, computed, onMounted } from 'vue'
import { Graph } from '@antv/x6'
import TaskList from './workflow/TaskList.vue'
import StencilPanel from './workflow/StencilPanel.vue'
import FlowCanvas from './workflow/FlowCanvas.vue'
import PropertyPanel from './workflow/PropertyPanel.vue'
import { toast } from '../utils/toast'
import { SaveWorkflowTask, GetAllWorkflowTasks, DeleteWorkflowTask, StartWebSocketServer, StopWebSocketServer, GetWebSocketPort, IsWebSocketRunning } from '../../wailsjs/go/main/NetworkApp'
import { main } from '../../wailsjs/go/models'

interface WorkflowTask {
  id: string
  name: string
  description?: string
  createdAt: string
  updatedAt: string
  nodes: any[]
  edges: any[]
}

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const wsRunning = ref(false)
const wsPort = ref(0)

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

async function createTask() {
  const taskData = {
    id: Date.now().toString(),
    name: `任务 ${tasks.value.length + 1}`,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    nodes: [],
    edges: []
  }
  
  try {
    const newTask = main.WorkflowTask.createFrom(taskData)
    await SaveWorkflowTask(newTask)
    tasks.value.push(taskData)
    selectedTaskId.value = taskData.id
    toast.success('任务创建成功')
  } catch (error: any) {
    console.error('[WorkflowDrawer] 创建任务失败:', error)
    toast.error('创建失败: ' + error.message)
  }
}



async function deleteTask(id: string) {
  if (!confirm('确定删除该任务？')) return
  
  try {
    await DeleteWorkflowTask(id)
    tasks.value = tasks.value.filter(t => t.id !== id)
    if (selectedTaskId.value === id) {
      selectedTaskId.value = undefined
    }
    toast.success('删除成功')
  } catch (error: any) {
    console.error('[WorkflowDrawer] 删除失败:', error)
    toast.error('删除失败: ' + error.message)
  }
}

async function renameTask(id: string, name: string) {
  const task = tasks.value.find(t => t.id === id)
  if (!task) return
  
  task.name = name
  await autoSaveTask(task)
  toast.success('重命名成功')
}

async function onTaskChange(task: any) {
  console.log('[WorkflowDrawer] ========== onTaskChange ==========')
  console.log('[WorkflowDrawer] 更新任务:', task)
  const index = tasks.value.findIndex(t => t.id === task.id)
  if (index >= 0) {
    tasks.value[index] = { ...task }
    console.log('[WorkflowDrawer] ✓ 任务已更新到列表')
    
    // 自动保存
    await autoSaveTask(task)
  }
}

async function autoSaveTask(task: any) {
  try {
    task.updatedAt = new Date().toISOString()
    const taskToSave = main.WorkflowTask.createFrom(task)
    await SaveWorkflowTask(taskToSave)
    console.log('[WorkflowDrawer] ✓ 自动保存成功')
  } catch (error: any) {
    console.error('[WorkflowDrawer] 自动保存失败:', error)
  }
}



function onSelectNode(node: any) {
  if (node.type === 'start' || node.type === 'end') {
    return
  }
  console.log('[WorkflowDrawer] ========== 选中节点 ==========')
  console.log('[WorkflowDrawer] 节点ID:', node.id)
  console.log('[WorkflowDrawer] 节点类型:', node.type)
  console.log('[WorkflowDrawer] 节点 data:', node.data)
  
  // 确保传递完整的节点数据
  selectedNode.value = {
    id: node.id,
    type: node.type,
    label: node.label,
    data: node.data || {}
  }
  
  console.log('[WorkflowDrawer] 传递给 PropertyPanel 的数据:', selectedNode.value.data)
  propertyVisible.value = true
}

function onSaveNodeData(data: Record<string, any>) {
  if (!selectedNode.value || !graphInstance.value) return
  
  console.log('[WorkflowDrawer] ========== 保存节点数据 ==========')
  console.log('[WorkflowDrawer] 节点ID:', selectedNode.value.id)
  console.log('[WorkflowDrawer] 节点类型:', selectedNode.value.type)
  console.log('[WorkflowDrawer] 保存的数据:', data)
  
  const node = graphInstance.value.getCellById(selectedNode.value.id)
  if (node) {
    const currentData = node.getData() || {}
    console.log('[WorkflowDrawer] 当前节点数据:', currentData)
    
    const updatedData = { ...currentData, ...data }
    console.log('[WorkflowDrawer] 合并后的数据:', updatedData)
    
    node.setData(updatedData)
    
    const verifyData = node.getData()
    console.log('[WorkflowDrawer] 验证保存后的数据:', verifyData)
    console.log('[WorkflowDrawer] ✓ 节点数据已更新')
    
    // 触发 change 事件
    const pos = node.position()
    node.position(pos.x + 0.01, pos.y)
    node.position(pos.x, pos.y)
  } else {
    console.error('[WorkflowDrawer] ❌ 未找到节点:', selectedNode.value.id)
  }
  
  selectedNode.value.data = { ...selectedNode.value.data, ...data }
}

// 加载任务
async function loadTasks() {
  try {
    const loadedTasks = await GetAllWorkflowTasks()
    tasks.value = loadedTasks || []
    console.log('[WorkflowDrawer] 加载了', tasks.value.length, '个任务')
  } catch (error: any) {
    console.error('[WorkflowDrawer] 加载任务失败:', error)
    toast.error('加载任务失败: ' + error.message)
  }
}

onMounted(async () => {
  loadTasks()
  await checkWebSocketStatus()
})

async function checkWebSocketStatus() {
  try {
    wsRunning.value = await IsWebSocketRunning()
    if (wsRunning.value) {
      wsPort.value = await GetWebSocketPort()
    }
  } catch (error) {
    console.error('[WorkflowDrawer] 检查 WebSocket 状态失败:', error)
  }
}

async function toggleWebSocket() {
  try {
    if (wsRunning.value) {
      await StopWebSocketServer()
      wsRunning.value = false
      wsPort.value = 0
      toast.success('WebSocket 已停止')
    } else {
      await StartWebSocketServer()
      wsRunning.value = true
      wsPort.value = await GetWebSocketPort()
      toast.success(`WebSocket 已启动 (端口: ${wsPort.value})`)
    }
  } catch (error: any) {
    console.error('[WorkflowDrawer] WebSocket 操作失败:', error)
    toast.error('WebSocket 操作失败: ' + error.message)
  }
}

async function copyPort() {
  try {
    await navigator.clipboard.writeText(wsPort.value.toString())
    toast.success('端口号已复制: ' + wsPort.value)
  } catch (error) {
    toast.error('复制失败')
  }
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

.header-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.ws-btn {
  padding: 6px 12px;
  border: 1px solid #d9d9d9;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.ws-btn.success {
  background: #52c41a;
  color: white;
  border-color: #52c41a;
}

.ws-btn.success:hover {
  background: #73d13d;
}

.ws-btn.danger {
  background: #ff4d4f;
  color: white;
  border-color: #ff4d4f;
}

.ws-btn.danger:hover {
  background: #ff7875;
}

.ws-port {
  padding: 4px 12px;
  background: #f0f0f0;
  border-radius: 4px;
  font-size: 12px;
  font-family: monospace;
  color: #52c41a;
  font-weight: 600;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s;
}

.ws-port:hover {
  background: #e6f7ff;
  color: #1890ff;
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
