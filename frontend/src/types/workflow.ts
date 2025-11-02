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
  | 'extract'
  | 'if'
  | 'for'
  | 'navigate'
  | 'scroll'
  | 'jsonl_reader'
  | 'intercept_request'
  | 'download_captured'
  | 'decrypt'

export interface NodeConfig {
  type: NodeType
  label: string
  icon: string
  color: string
  description: string
}
