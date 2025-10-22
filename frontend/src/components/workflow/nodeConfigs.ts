import type { NodeConfig } from '../../types/workflow'

export const NODE_CONFIGS: NodeConfig[] = [
  { type: 'start', label: '开始', icon: '▶️', color: '#52c41a', description: '工作流起点' },
  { type: 'end', label: '结束', icon: '⏹️', color: '#f5222d', description: '工作流终点' },
  { type: 'click', label: '点击元素', icon: '👆', color: '#1890ff', description: '点击页面元素' },
  { type: 'input', label: '输入文本', icon: '⌨️', color: '#1890ff', description: '输入文本' },
  { type: 'navigate', label: '导航', icon: '🌐', color: '#722ed1', description: '跳转URL' },
  { type: 'wait', label: '等待', icon: '⏱️', color: '#faad14', description: '等待时间' },
  { type: 'intercept', label: '拦截请求', icon: '🔧', color: '#13c2c2', description: '拦截请求' },
  { type: 'download', label: '下载保存', icon: '💾', color: '#52c41a', description: '下载文件' },
  { type: 'extract', label: '提取数据', icon: '📋', color: '#2f54eb', description: '提取数据' },
  { type: 'if', label: '条件判断', icon: '❓', color: '#fa8c16', description: '条件分支' },
  { type: 'for', label: '循环', icon: '🔄', color: '#eb2f96', description: '循环执行' }
]
