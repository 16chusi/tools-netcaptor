<template>
  <div class="task-list">
    <div class="task-list-header">
      <h3>任务列表</h3>
      <button @click="$emit('create')" class="create-btn">+ 新建</button>
    </div>
    <div class="task-list-body">
      <div
        v-for="task in tasks"
        :key="task.id"
        :class="['task-item', { active: selectedTaskId === task.id, running: runningTaskId === task.id }]"
        @click="$emit('select', task.id)"
      >
        <div class="task-info">
          <input 
            v-if="editingTaskId === task.id"
            v-model="editingName"
            @click.stop
            @blur="saveTaskName(task)"
            @keyup.enter="saveTaskName(task)"
            @keyup.esc="cancelEdit"
            class="task-name-input"
            ref="nameInput"
          />
          <div v-else class="task-name" @dblclick.stop="startEdit(task)" title="双击编辑">
            {{ task.name }}
            <span v-if="runningTaskId === task.id" class="running-badge">运行中</span>
          </div>
          <div class="task-meta">{{ formatDate(task.updatedAt) }}</div>
        </div>
        <button 
          v-if="runningTaskId === task.id" 
          @click.stop="$emit('stop')" 
          class="stop-btn" 
          title="停止任务"
        >⏹️</button>
        <button 
          v-else
          @click.stop="$emit('delete', task.id)" 
          class="delete-btn" 
          title="删除任务"
        >×</button>
      </div>
      <div v-if="tasks.length === 0" class="empty">
        暂无任务,点击"新建"创建
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import type { WorkflowTask } from '../../types/workflow'

defineProps<{
  tasks: WorkflowTask[]
  selectedTaskId?: string
  runningTaskId?: string
}>()

const emit = defineEmits<{
  select: [id: string]
  create: []
  delete: [id: string]
  rename: [id: string, name: string]
  stop: []
}>()

const editingTaskId = ref<string>()
const editingName = ref('')
const nameInput = ref<HTMLInputElement>()

function formatDate(date: string) {
  return new Date(date).toLocaleString('zh-CN')
}

async function startEdit(task: WorkflowTask) {
  editingTaskId.value = task.id
  editingName.value = task.name
  await nextTick()
  nameInput.value?.focus()
  nameInput.value?.select()
}

function saveTaskName(task: WorkflowTask) {
  if (editingName.value.trim() && editingName.value !== task.name) {
    emit('rename', task.id, editingName.value.trim())
  }
  editingTaskId.value = undefined
}

function cancelEdit() {
  editingTaskId.value = undefined
}
</script>

<style scoped>
.task-list {
  width: 240px;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
  background: #fafafa;
}

.task-list-header {
  padding: 16px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-list-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.create-btn {
  padding: 4px 12px;
  border: none;
  background: #1890ff;
  color: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.create-btn:hover {
  background: #40a9ff;
}

.task-list-body {
  flex: 1;
  overflow-y: auto;
}

.task-item {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background 0.2s;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.task-info {
  flex: 1;
  min-width: 0;
}

.task-item:hover {
  background: #f0f0f0;
}

.task-item.active {
  background: #e6f7ff;
  border-left: 3px solid #1890ff;
}

.task-name {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: text;
}

.task-name-input {
  width: 100%;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
  padding: 2px 4px;
  border: 1px solid #1890ff;
  border-radius: 2px;
  outline: none;
}

.delete-btn {
  width: 20px;
  height: 20px;
  border: none;
  background: transparent;
  color: #999;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
  padding: 0;
  border-radius: 2px;
  opacity: 0;
  transition: all 0.2s;
  flex-shrink: 0;
}

.task-item:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  background: #ff4d4f;
  color: white;
}

.stop-btn {
  width: 20px;
  height: 20px;
  border: none;
  background: #ff4d4f;
  color: white;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  padding: 0;
  border-radius: 2px;
  flex-shrink: 0;
  transition: all 0.2s;
}

.stop-btn:hover {
  background: #ff7875;
}

.running-badge {
  display: inline-block;
  margin-left: 6px;
  padding: 2px 6px;
  background: #52c41a;
  color: white;
  font-size: 10px;
  border-radius: 2px;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.task-item.running {
  background: #f6ffed;
  border-left: 3px solid #52c41a;
}

.task-meta {
  font-size: 11px;
  color: #999;
}

.empty {
  padding: 40px 20px;
  text-align: center;
  color: #999;
  font-size: 12px;
}
</style>
