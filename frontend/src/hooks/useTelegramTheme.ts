import { useEffect } from 'react'
import WebApp from '@twa-dev/sdk'

export function useTelegramTheme() {
  useEffect(() => {
    const applyTheme = () => {
      const isDark = WebApp.colorScheme === 'dark'
      document.documentElement.classList.toggle('dark', isDark)
    }

    applyTheme()
    WebApp.onEvent('themeChanged', applyTheme)
    WebApp.ready()
    WebApp.expand()

    return () => {
      WebApp.offEvent('themeChanged', applyTheme)
    }
  }, [])
}
