import { createApp, h } from 'vue'
import Toast from '../components/Toast.vue'

let currentApp: any = null
let currentContainer: HTMLElement | null = null

function showToast(message: string, type: 'success' | 'error' | 'warning' | 'info' = 'info', duration = 3000) {
  if (currentApp && currentContainer) {
    currentApp.unmount()
    document.body.removeChild(currentContainer)
  }

  const container = document.createElement('div')
  document.body.appendChild(container)

  const app = createApp({
    render() {
      return h(Toast, { message, type, duration })
    }
  })

  const instance = app.mount(container) as any
  currentApp = app
  currentContainer = container
  
  if (instance?.show) {
    instance.show()
  }

  setTimeout(() => {
    if (currentApp === app) {
      app.unmount()
      if (currentContainer && document.body.contains(currentContainer)) {
        document.body.removeChild(currentContainer)
      }
      currentApp = null
      currentContainer = null
    }
  }, duration + 500)
}

export const toast = {
  success: (message: string, duration?: number) => showToast(message, 'success', duration),
  error: (message: string, duration?: number) => showToast(message, 'error', duration),
  warning: (message: string, duration?: number) => showToast(message, 'warning', duration),
  info: (message: string, duration?: number) => showToast(message, 'info', duration),
}
