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
      { type: 'screenshot', label: '网页截图', icon: '📸', color: 'rgba(250, 140, 22, 0.15)', description: '截图保存为图片或PDF' },
      { type: 'collect', label: '数据收集器', icon: '📊', color: 'rgba(250, 140, 22, 0.15)', description: '追加数据到文件' },
      { type: 'decrypt', label: '数据解密', icon: '🔐', color: 'rgba(250, 140, 22, 0.15)', description: '解密数据' },
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
      { type: 'ai_analyze_content', label: 'AI内容分析', icon: '🤖', color: 'rgba(82, 196, 26, 0.15)', description: '理解和分析页面内容' },
      { type: 'ai_extract_data', label: 'AI数据提取', icon: '🤖', color: 'rgba(82, 196, 26, 0.15)', description: '智能分析DOM并提取数据' },
      { type: 'ai_transform_data', label: 'AI数据转换', icon: '🤖', color: 'rgba(82, 196, 26, 0.15)', description: '转换数据格式和结构' }
    ]
  },
  {
    name: 'variables',
    title: '内置变量',
    icon: '🔧',
    color: '#666666',
    nodes: [
      { type: 'var_timestamp', label: '{timestamp}', icon: '⏰', color: 'transparent', description: '完整时间戳: 20241105_190116' },
      { type: 'var_date', label: '{date}', icon: '📅', color: 'transparent', description: '日期: 20241105' },
      { type: 'var_time', label: '{time}', icon: '🕐', color: 'transparent', description: '时间: 190116' },
      { type: 'var_uuid', label: '{uuid}', icon: '🆔', color: 'transparent', description: '完整UUID: 550e8400-e29b-41d4...' },
      { type: 'var_uuid_short', label: '{uuid_short}', icon: '🔖', color: 'transparent', description: '短UUID: 550e8400' },
      { type: 'var_counter', label: '{counter}', icon: '🔢', color: 'transparent', description: '自增计数器: 1, 2, 3...' },
      { type: 'var_random', label: '{random}', icon: '🎲', color: 'transparent', description: '4位随机数: 0-9999' },
      { type: 'var_random_6', label: '{random_6}', icon: '🎯', color: 'transparent', description: '6位随机数: 000000-999999' },
      { type: 'var_title', label: '{title}', icon: '📄', color: 'transparent', description: '页面标题 (自动清理特殊字符)' },
      { type: 'var_url', label: '{url}', icon: '🌐', color: 'transparent', description: '当前页面URL' },
      { type: 'var_index', label: '{index}', icon: '🔄', color: 'transparent', description: '循环索引 (在for循环中)' }
    ]
  }
]

// 保留旧的导出以兼容
export const NODE_CONFIGS: NodeConfig[] = [
  { type: 'start', label: '开始', icon: '▶️', color: 'rgba(0, 120, 212, 0.15)', description: '工作流起点' },
  { type: 'end', label: '结束', icon: '⏹️', color: 'rgba(209, 52, 56, 0.15)', description: '工作流终点' },
  ...NODE_GROUPS.flatMap(g => g.nodes)
]
