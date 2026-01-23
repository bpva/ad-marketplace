import { useState, useEffect } from 'react'
import WebApp from '@twa-dev/sdk'
import { authenticate, setToken, type User } from '@/lib/api'

interface AuthState {
  user: User | null
  loading: boolean
  error: string | null
}

export function useAuth(): AuthState {
  const [state, setState] = useState<AuthState>({
    user: null,
    loading: true,
    error: null,
  })

  useEffect(() => {
    const initData = WebApp.initData
    if (!initData) {
      setState({ user: null, loading: false, error: 'Not in Telegram' })
      return
    }

    authenticate(initData)
      .then((res) => {
        setToken(res.token)
        setState({ user: res.user, loading: false, error: null })
      })
      .catch((err) => {
        setState({ user: null, loading: false, error: err.message })
      })
  }, [])

  return state
}
