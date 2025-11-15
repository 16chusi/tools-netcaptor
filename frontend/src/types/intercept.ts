export interface InterceptRule {
  id?: string
  name: string
  enabled: boolean
  urlPattern: string
  actionType: string
  findText?: string
  replaceText?: string
  useRegex?: boolean
  replaceAll?: boolean
  responseContent?: string
  redirectUrl?: string
  webhookUrl?: string
  webhookEnabled?: boolean
  webhookSync?: boolean
  saveToFile?: boolean
  saveFilePath?: string
  saveFormat?: string
  remoteHTTP?: RemoteHTTPConfig
  forwardRequest?: ForwardConfig
}

export interface RemoteHTTPConfig {
  url: string
  method: string
  timeout: number
  sendOriginal: boolean
  useResponse: boolean
  bodyTemplate: string
  headers?: Record<string, string>
}

export interface ForwardConfig {
  targetUrl: string
  replaceHost: boolean
  replaceHeaders?: Record<string, string>
  timeout: number
  keepPath: boolean
}
