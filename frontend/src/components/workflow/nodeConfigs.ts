import type { NodeConfig } from '../../types/workflow'

export const NODE_CONFIGS: NodeConfig[] = [
  { type: 'start', label: '开始', icon: '▶️', color: '#95de64', description: '工作流起点' },
  { type: 'end', label: '结束', icon: '⏹️', color: '#ff7875', description: '工作流终点' },
  { type: 'click', label: '点击元素', icon: '👆', color: '#91d5ff', description: '点击页面元素' },
  { type: 'input', label: '输入文本', icon: '⌨️', color: '#91d5ff', description: '输入文本' },
  { type: 'navigate', label: '导航', icon: '🌐', color: '#d3adf7', description: '跳转URL' },
  { type: 'wait', label: '等待', icon: '⏱️', color: '#ffd591', description: '等待时间' },
  { type: 'intercept', label: '拦截请求', icon: '🔧', color: '#87e8de', description: '拦截请求' },
  { type: 'download', label: '下载保存', icon: '💾', color: '#b7eb8f', description: '下载文件' },
  { type: 'extract', label: '提取数据', icon: '📋', color: '#adc6ff', description: '提取数据' },
  { type: 'if', label: '条件判断', icon: '❓', color: '#ffc069', description: '条件分支' },
  { type: 'for', label: '循环', icon: '🔄', color: '#ffadd2', description: '循环执行' }
]
