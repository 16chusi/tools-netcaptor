export interface InterceptRule {
  id: string
  name: string
  enabled: boolean
  urlPattern: string
  actionType: string  // 改为 string 以匹配后端类型
  findText?: string
  replaceText?: string
  useRegex?: boolean
  replaceAll?: boolean
  responseContent?: string
  redirectUrl?: string
  webhookUrl?: string
  webhookEnabled?: boolean
  saveToFile?: boolean
  saveFilePath?: string
  saveFormat?: string  // 改为 string
}
