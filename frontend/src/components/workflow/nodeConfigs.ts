import type { NodeConfig } from '../../types/workflow'

export interface NodeGroup {
  name: string
  title: string
  icon: string
  color: string
  nodes: NodeConfig[]
}

export const NODE_GROUPS: NodeGroup[] = [
  {
    name: 'basic',
    title: '基础操作',
    icon: '🖱️',
    color: '#1890ff', // 蓝色主题
    nodes: [
      { type: 'click', label: '点击元素', icon: '🖱️', color: 'rgba(24, 144, 255, 0.15)', description: '点击页面元素' },
      { type: 'input', label: '输入文本', icon: '⌨️', color: 'rgba(24, 144, 255, 0.15)', description: '输入文本' },
      { type: 'navigate', label: '页面导航', icon: '🌐', color: 'rgba(24, 144, 255, 0.15)', description: '跳转URL' },
      { type: 'wait', label: '等待延时', icon: '⏱️', color: 'rgba(24, 144, 255, 0.15)', description: '等待时间' },
      { type: 'scroll', label: '页面滚动', icon: '📜', color: 'rgba(24, 144, 255, 0.15)', description: '滚动页面' }
    ]
  },
  {
    name: 'data',
    title: '数据处理',
    icon: '📊',
    color: '#fa8c16', // 橙色主题
    nodes: [
      { type: 'extract', label: '提取数据', icon: '📊', color: 'rgba(250, 140, 22, 0.15)', description: '从网页DOM中获取内容' },
      { type: 'download', label: '文件下载', icon: '💾', color: 'rgba(250, 140, 22, 0.15)', description: '下载文件' },
      { type: 'download_captured', label: '下载已捕获响应', icon: '💾', color: 'rgba(250, 140, 22, 0.15)', description: '下载已捕获的响应' },
      { type: 'collect', label: '数据收集器', icon: '📊', color: 'rgba(250, 140, 22, 0.15)', description: '追加数据到文件' },
      { type: 'decrypt', label: '数据解密', icon: '🔐', color: 'rgba(250, 140, 22, 0.15)', description: '解密数据' },
      { type: 'intercept_request', label: '请求拦截', icon: '🔍', color: 'rgba(250, 140, 22, 0.15)', description: '拦截和修改HTTP请求' },
      { type: 'jsonl_reader', label: 'JSONL读取器', icon: '📄', color: 'rgba(250, 140, 22, 0.15)', description: '读取JSONL文件并循环' }
    ]
  },
  {
    name: 'control',
    title: '流程控制',
    icon: '🔀',
    color: '#722ed1', // 紫色主题
    nodes: [
      { type: 'if', label: '条件判断', icon: '🔀', color: 'rgba(114, 46, 209, 0.15)', description: '条件分支' },
      { type: 'for', label: '循环控制', icon: '🔄', color: 'rgba(114, 46, 209, 0.15)', description: '循环执行' }
    ]
  },
  {
    name: 'ai',
    title: 'AI功能',
    icon: '🤖',
    color: '#52c41a', // 绿色主题（与AI属性面板一致）
    nodes: [
      { type: 'ai_extract_data', label: 'AI数据提取', icon: '🤖', color: 'rgba(82, 196, 26, 0.15)', description: '智能分析DOM并提取数据' },
      { type: 'ai_smart_click', label: 'AI智能点击', icon: '🤖', color: 'rgba(82, 196, 26, 0.15)', description: '智能识别并点击元素' },
      { type: 'ai_analyze_content', label: 'AI内容分析', icon: '🤖', color: 'rgba(82, 196, 26, 0.15)', description: '理解和分析页面内容' },
      { type: 'ai_validate_data', label: 'AI数据验证', icon: '🤖', color: 'rgba(82, 196, 26, 0.15)', description: '验证数据格式和内容' },
      { type: 'ai_transform_data', label: 'AI数据转换', icon: '🤖', color: 'rgba(82, 196, 26, 0.15)', description: '转换数据格式和结构' },
      { type: 'ai_form_fill', label: 'AI表单填写', icon: '🤖', color: 'rgba(82, 196, 26, 0.15)', description: '智能理解并填写表单' },
      { type: 'ai_navigation', label: 'AI智能导航', icon: '🤖', color: 'rgba(82, 196, 26, 0.15)', description: '根据描述智能导航' }
    ]
  }
]

// 保留旧的导出以兼容
export const NODE_CONFIGS: NodeConfig[] = [
  { type: 'start', label: '开始', icon: '▶️', color: 'rgba(0, 120, 212, 0.15)', description: '工作流起点' },
  { type: 'end', label: '结束', icon: '⏹️', color: 'rgba(209, 52, 56, 0.15)', description: '工作流终点' },
  ...NODE_GROUPS.flatMap(g => g.nodes)
]
