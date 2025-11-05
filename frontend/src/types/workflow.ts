export interface WorkflowTask {
  id: string
  name: string
  description?: string
  createdAt: string
  updatedAt: string
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
}

export interface WorkflowNode {
  id: string
  type: NodeType
  x: number
  y: number
  label: string
  data?: Record<string, any>
}

export interface WorkflowEdge {
  id: string
  source: string
  target: string
  label?: string
}

export type NodeType = 
  | 'start'
  | 'end'
  | 'click'
  | 'input'
  | 'wait'
  | 'download'
  | 'screenshot'
  | 'extract'
  | 'if'
  | 'for'
  | 'navigate'
  | 'scroll'
  | 'jsonl_reader'
  | 'collect'
  | 'decrypt'
  | 'ai_extract_data'
  | 'ai_analyze_content'
  | 'ai_validate_data'
  | 'ai_transform_data'
  | 'ai_smart_click'
  | 'ai_form_fill'
  | 'ai_navigation'
  | 'var_timestamp'
  | 'var_date'
  | 'var_time'
  | 'var_uuid'
  | 'var_uuid_short'
  | 'var_counter'
  | 'var_random'
  | 'var_random_6'
  | 'var_title'
  | 'var_url'
  | 'var_index'

export interface NodeConfig {
  type: NodeType
  label: string
  icon: string
  color: string
  description: string
}
