import ClickProperty from './ClickProperty.vue'
import InputProperty from './InputProperty.vue'
import NavigateProperty from './NavigateProperty.vue'
import WaitProperty from './WaitProperty.vue'
import ScrollProperty from './ScrollProperty.vue'
import ExtractProperty from './ExtractProperty.vue'
import DownloadProperty from './DownloadProperty.vue'
import ScreenshotProperty from './ScreenshotProperty.vue'
import IfProperty from './IfProperty.vue'
import ForProperty from './ForProperty.vue'
import JSONLProperty from './JSONLProperty.vue'
import CollectProperty from './CollectProperty.vue'
import DecryptProperty from './DecryptProperty.vue'
import AIExtractProperty from './AIExtractProperty.vue'
import AIAnalyzeProperty from './AIAnalyzeProperty.vue'
import AITransformProperty from './AITransformProperty.vue'

// 组件映射表
export const propertyComponents = {
  click: ClickProperty,
  input: InputProperty,
  navigate: NavigateProperty,
  wait: WaitProperty,
  scroll: ScrollProperty,
  extract: ExtractProperty,
  download: DownloadProperty,
  screenshot: ScreenshotProperty,
  if: IfProperty,
  for: ForProperty,
  jsonl_reader: JSONLProperty,
  collect: CollectProperty,
  decrypt: DecryptProperty,
  ai_analyze_content: AIAnalyzeProperty,
  ai_extract_data: AIExtractProperty,
  ai_transform_data: AITransformProperty,
}

// 默认值配置
export const defaultValues = {
  click: { selectorType: 'css', waitTime: 3000 },
  input: { selectorType: 'css' },
  navigate: { openMode: 'current', deviceMode: 'desktop' },
  scroll: { scrollType: 'bottom', interval: 500 },
  extract: { saveToVariable: 'myData', selectorType: 'css', attribute: 'text' },
  download: { urlSource: 'direct' },
  screenshot: { format: 'png', captureType: 'viewport', quality: '80', filenameTemplate: 'screenshot_{timestamp}' },
  if: { operator: '==', truePort: 'right', falsePort: 'bottom' },
  for: { variable: 'index', interval: 500 },
  jsonl_reader: { saveToVariable: 'data', extractKeys: '*', interval: 100 },
  collect: { format: 'jsonl' },
  decrypt: { saveToVariable: 'decryptedData', algorithm: 'sm4-ecb' },
  ai_analyze_content: { 
    saveToVariable: 'analysisResult', 
    retryCount: 3, 
    retryDelay: 2,
    timeout: 100,
    thinkingMode: 'enabled',
    topP: 0.9,
    temperature: 0.7,
    maxTokens: 2000
  },
  ai_extract_data: { 
    saveToVariable: 'extractedData', 
    outputFormat: 'json', 
    contentType: 'text',
    retryCount: 3, 
    retryDelay: 2, 
    timeout: 100,
    thinkingMode: 'enabled',
    topP: 0.9,
    temperature: 0.7,
    maxTokens: 2000
  },
  ai_transform_data: { 
    saveToVariable: 'transformedData', 
    outputFormat: 'json', 
    dataSource: 'current', 
    retryCount: 3, 
    retryDelay: 2,
    timeout: 100,
    thinkingMode: 'enabled',
    topP: 0.9,
    temperature: 0.7,
    maxTokens: 2000
  },
}
