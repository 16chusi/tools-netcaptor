export function generateCurl(entry: any): string {
  let curl = `curl '${entry.url}'`
  
  if (entry.method !== 'GET') {
    curl += ` \\
  -X ${entry.method}`
  }
  
  if (entry.request.headers) {
    for (const [key, value] of Object.entries(entry.request.headers)) {
      const escapedValue = String(value).replace(/'/g, "'\\''")
      curl += ` \\
  -H '${key}: ${escapedValue}'`
    }
  }
  
  if (entry.request.body) {
    const escapedBody = entry.request.body.replace(/'/g, "'\\''")
    curl += ` \\
  --data-raw '${escapedBody}'`
  }
  
  curl += ` \\
  --compressed`
  
  return curl
}

export function generatePowerShell(entry: any): string {
  let ps = `$headers = @{\n`
  
  if (entry.request.headers) {
    for (const [key, value] of Object.entries(entry.request.headers)) {
      ps += `    "${key}" = "${String(value).replace(/"/g, '`"')}";\n`
    }
  }
  ps += `}\n\n`
  
  ps += `Invoke-WebRequest -Uri "${entry.url}" `
  ps += `-Method ${entry.method} `
  ps += `-Headers $headers`
  
  if (entry.request.body) {
    const escapedBody = entry.request.body.replace(/"/g, '`"').replace(/\n/g, '`n')
    ps += ` \`\n    -Body "${escapedBody}"`
  }
  
  return ps.replace(/\\n/g, '\n')
}

export function generateFetch(entry: any): string {
  const options: any = {
    method: entry.method,
    headers: entry.request.headers || {}
  }
  
  if (entry.request.body) {
    options.body = entry.request.body
  }
  
  return `fetch('${entry.url}', ${JSON.stringify(options, null, 2)})
  .then(response => response.json())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));`
}
