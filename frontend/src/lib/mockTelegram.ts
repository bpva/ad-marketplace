import { MOCK_TG_USER } from './mockUser'

async function generateHash(data: string, botToken: string): Promise<string> {
  const encoder = new TextEncoder()
  const webAppDataKey = await crypto.subtle.importKey(
    'raw',
    encoder.encode('WebAppData'),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  )
  const secretKeyData = await crypto.subtle.sign('HMAC', webAppDataKey, encoder.encode(botToken))
  const secretKey = await crypto.subtle.importKey(
    'raw',
    secretKeyData,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  )
  const signature = await crypto.subtle.sign('HMAC', secretKey, encoder.encode(data))
  return Array.from(new Uint8Array(signature))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

export async function generateMockInitData(botToken: string): Promise<string> {
  const userJson = JSON.stringify(MOCK_TG_USER)
  const authDate = Math.floor(Date.now() / 1000)

  const params: Record<string, string> = {
    user: userJson,
    auth_date: authDate.toString(),
  }

  const dataCheckString = Object.keys(params)
    .sort()
    .map((k) => `${k}=${params[k]}`)
    .join('\n')

  const hash = await generateHash(dataCheckString, botToken)

  return new URLSearchParams({ ...params, hash }).toString()
}
