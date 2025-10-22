export function parseRequestCookies(headers: any): Array<{ name: string; value: string }> {
  if (!headers) return []
  
  const cookieHeader = headers['Cookie'] || headers['cookie']
  if (!cookieHeader) return []
  
  return cookieHeader.split(';').map((c: string) => {
    const [name, ...valueParts] = c.trim().split('=')
    return { name, value: valueParts.join('=') }
  })
}

export function parseResponseCookies(headers: any): Array<{ name: string; value: string }> {
  if (!headers) return []
  
  const setCookieHeader = headers['Set-Cookie'] || headers['set-cookie']
  if (!setCookieHeader) return []
  
  const cookies = Array.isArray(setCookieHeader) ? setCookieHeader : [setCookieHeader]
  return cookies.map((c: string) => {
    const [nameValue] = c.split(';')
    const [name, ...valueParts] = nameValue.trim().split('=')
    return { name, value: valueParts.join('=') }
  })
}
