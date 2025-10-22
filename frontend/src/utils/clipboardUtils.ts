export async function copyToClipboard(text: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(text)
  } catch (e) {
    console.error('Copy failed:', e)
  }
}

export async function copyJSON(data: any): Promise<void> {
  await copyToClipboard(JSON.stringify(data, null, 2))
}
