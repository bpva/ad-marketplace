const API_URL = import.meta.env.VITE_API_URL || ''

let token: string | null = null

export function setToken(t: string) {
  token = t
}

export function getToken(): string | null {
  return token
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  })

  if (!res.ok) {
    throw new Error(`API error: ${res.status}`)
  }

  return res.json()
}

export interface User {
  id: string
  telegram_id: number
  name: string
}

interface AuthResponse {
  token: string
  user: User
}

export async function authenticate(initData: string): Promise<AuthResponse> {
  return request<AuthResponse>('/api/v1/auth', {
    method: 'POST',
    body: JSON.stringify({ init_data: initData }),
  })
}

export async function fetchMe(): Promise<User> {
  return request<User>('/api/v1/me')
}
