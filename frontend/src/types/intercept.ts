export interface InterceptRule {
  id: string
  name: string
  enabled: boolean
  urlPattern: string
  actionType: 'findReplace' | 'redirect' | 'responseReplace'
  findText?: string
  replaceText?: string
  useRegex?: boolean
  replaceAll?: boolean
  responseContent?: string
  redirectUrl?: string
  webhookUrl?: string
  webhookEnabled?: boolean
}
