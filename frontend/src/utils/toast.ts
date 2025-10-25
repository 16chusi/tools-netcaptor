import { createApp, h } from 'vue'
import Toast from '../components/Toast.vue'

let toastInstance: any = null

function showToast(message: string, type: 'success' | 'error' | 'warning' | 'info' = 'info', duration = 3000) {
  if (toastInstance) {
    toastInstance.unmount()
  }

  const container = document.createElement('div')
  document.body.appendChild(container)

  const app = createApp({
    render() {
      return h(Toast, { message, type, duration })
    }
  })

  toastInstance = app.mount(container)
  toastInstance.show()

  setTimeout(() => {
    app.unmount()
    document.body.removeChild(container)
    toastInstance = null
  }, duration + 500)
}

export const toast = {
  success: (message: string, duration?: number) => showToast(message, 'success', duration),
  error: (message: string, duration?: number) => showToast(message, 'error', duration),
  warning: (message: string, duration?: number) => showToast(message, 'warning', duration),
  info: (message: string, duration?: number) => showToast(message, 'info', duration),
}
