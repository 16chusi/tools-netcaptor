export function formatJSON(text: string): string {
  try {
    return JSON.stringify(JSON.parse(text), null, 2)
  } catch {
    return text
  }
}

export function toHex(text: string): string {
  let result = ''
  for (let i = 0; i < text.length; i++) {
    const hex = text.charCodeAt(i).toString(16).padStart(2, '0')
    result += hex + ' '
    if ((i + 1) % 16 === 0) result += '\n'
  }
  return result
}

export function toBase64(text: string): string {
  try {
    return btoa(text)
  } catch {
    return text
  }
}
