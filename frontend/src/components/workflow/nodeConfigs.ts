import type { NodeConfig } from '../../types/workflow'

// Windows 11 风格配色方案 - 浅色半透明
export const NODE_CONFIGS: NodeConfig[] = [
  { type: 'start', label: '开始', icon: '▶️', color: 'rgba(0, 120, 212, 0.15)', description: '工作流起点' },
  { type: 'end', label: '结束', icon: '⏹️', color: 'rgba(209, 52, 56, 0.15)', description: '工作流终点' },
  { type: 'click', label: '点击元素', icon: '👆', color: 'rgba(96, 205, 255, 0.15)', description: '点击页面元素' },
  { type: 'input', label: '输入文本', icon: '⌨️', color: 'rgba(140, 189, 255, 0.15)', description: '输入文本' },
  { type: 'navigate', label: '导航', icon: '🌐', color: 'rgba(180, 160, 255, 0.15)', description: '跳转URL' },
  { type: 'wait', label: '等待', icon: '⏱️', color: 'rgba(255, 200, 61, 0.15)', description: '等待时间' },
  { type: 'download', label: '下载保存', icon: '💾', color: 'rgba(16, 137, 62, 0.15)', description: '下载文件' },
  { type: 'extract', label: '提取数据', icon: '📋', color: 'rgba(94, 92, 230, 0.15)', description: '提取数据' },
  { type: 'scroll', label: '滚动页面', icon: '📜', color: 'rgba(255, 185, 0, 0.15)', description: '滚动页面' },
  { type: 'if', label: '条件判断', icon: '❓', color: 'rgba(255, 140, 0, 0.15)', description: '条件分支' },
  { type: 'for', label: '循环', icon: '🔄', color: 'rgba(227, 0, 140, 0.15)', description: '循环执行' },
  { type: 'jsonl_reader', label: 'JSONL读取器', icon: '📄', color: 'rgba(135, 100, 184, 0.15)', description: '读取JSONL文件并循环' },
  { type: 'intercept_request', label: '拦截请求', icon: '🚫', color: 'rgba(247, 99, 12, 0.15)', description: '拦截和修改HTTP请求' },
  { type: 'download_captured', label: '下载已捕获', icon: '📥', color: 'rgba(0, 153, 115, 0.15)', description: '下载已捕获的响应' },
  { type: 'decrypt', label: '解密', icon: '🔓', color: 'rgba(255, 87, 34, 0.15)', description: '解密数据' }
]
