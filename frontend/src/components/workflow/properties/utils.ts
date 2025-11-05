import { defaultValues } from './index'

// 应用默认值的函数
export function applyDefaultValues(formData: Record<string, any>, nodeType: string) {
  // 基础默认值
  const basicDefaults = {
    selectorType: 'css',
    urlSource: 'direct',
    scrollType: 'bottom',
    openMode: 'current',
    extractKeys: '*',
    interval: 100,
    action: 'block',
    statusCode: 403,
    overwriteMode: 'skip',
    algorithm: 'sm4-ecb',
    dataFormat: 'text',
    format: 'jsonl'
  }

  // 应用基础默认值
  Object.keys(basicDefaults).forEach(key => {
    if (!formData[key]) {
      formData[key] = basicDefaults[key as keyof typeof basicDefaults]
    }
  })

  // 应用配置化的默认值
  const defaults = defaultValues[nodeType as keyof typeof defaultValues]
  if (defaults) {
    Object.keys(defaults).forEach(key => {
      if (!formData[key]) {
        formData[key] = defaults[key as keyof typeof defaults]
      }
    })
  }

  // AI组件特殊处理
  if (nodeType?.startsWith('ai_')) {
    if (formData.modelIndex === undefined) {
      formData.modelIndex = 0
    }
    if (!formData.prompt) {
      formData.prompt = ''
    }
  }
}

// 加载AI模型配置
export function loadAIModels() {
  const saved = localStorage.getItem('ai-model-settings')
  if (saved) {
    const settings = JSON.parse(saved)
    return settings.models || []
  }
  return []
}

// 防抖保存函数
export function createDebouncedSave(callback: (data: any) => void, delay = 300) {
  let timer: number | null = null
  
  return (data: any) => {
    if (timer) clearTimeout(timer)
    timer = window.setTimeout(() => {
      callback(data)
    }, delay)
  }
}
