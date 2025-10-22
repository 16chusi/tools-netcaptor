export function getDomain(url: string): string {
  try {
    return new URL(url).hostname
  } catch {
    return url
  }
}

export function getPath(url: string): string {
  try {
    const u = new URL(url)
    let path = u.pathname
    if (u.search) path += u.search
    return path || '/'
  } catch {
    return url
  }
}

export function getResourceType(entry: any): string {
  const ct = entry.response?.contentType || ''
  if (ct.includes('javascript')) return 'js'
  if (ct.includes('css')) return 'css'
  if (ct.includes('html')) return 'document'
  if (ct.includes('json')) return 'json'
  if (ct.includes('image')) return 'image'
  if (ct.includes('font')) return 'font'
  return 'other'
}

export function getStatusClass(status: number): string {
  if (status >= 200 && status < 300) return 'success'
  if (status >= 300 && status < 400) return 'redirect'
  if (status >= 400 && status < 500) return 'client-error'
  return 'server-error'
}

export function formatSize(bytes: number): string {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

export function getFormatType(contentType: string): string {
  if (contentType.includes('json')) return 'JSON'
  if (contentType.includes('html')) return 'HTML'
  if (contentType.includes('javascript')) return 'JavaScript'
  if (contentType.includes('css')) return 'CSS'
  if (contentType.includes('xml')) return 'XML'
  if (contentType.includes('image/png')) return 'PNG'
  if (contentType.includes('image/jpeg') || contentType.includes('image/jpg')) return 'JPEG'
  if (contentType.includes('image/gif')) return 'GIF'
  if (contentType.includes('image/svg')) return 'SVG'
  if (contentType.includes('image/webp')) return 'WebP'
  if (contentType.includes('application/pdf')) return 'PDF'
  if (contentType.includes('text/plain')) return 'Text'
  return contentType.split('/').pop()?.toUpperCase() || 'Unknown'
}
