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
        :class="['task-item', { active: selectedTaskId === task.id }]"
        @click="$emit('select', task.id)"
      >
        <div class="task-info">
          <div class="task-name">{{ task.name }}</div>
          <div class="task-meta">{{ formatDate(task.updatedAt) }}</div>
        </div>
        <button @click.stop="$emit('delete', task.id)" class="delete-btn" title="删除任务">×</button>
      </div>
      <div v-if="tasks.length === 0" class="empty">
        暂无任务,点击"新建"创建
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { WorkflowTask } from '../../types/workflow'

defineProps<{
  tasks: WorkflowTask[]
  selectedTaskId?: string
}>()

defineEmits<{
  select: [id: string]
  create: []
  delete: [id: string]
}>()

function formatDate(date: string) {
  return new Date(date).toLocaleString('zh-CN')
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
