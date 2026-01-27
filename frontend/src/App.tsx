import { useState } from 'react'
import WebApp from '@twa-dev/sdk'
import { useTelegramTheme } from '@/hooks/useTelegramTheme'
import { useAuth } from '@/hooks/useAuth'
import { NotInTelegram } from '@/components/NotInTelegram'
import { Header } from '@/components/Header'
import { SettingsPage } from '@/components/SettingsPage'
import { Button } from '@/components/ui/button'

function App() {
  useTelegramTheme()
  const { user, loading } = useAuth()
  const [page, setPage] = useState<'main' | 'settings'>('main')

  const isInTelegram = WebApp.initData !== ''

  if (!isInTelegram) {
    return <NotInTelegram />
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    )
  }

  if (page === 'settings') {
    return <SettingsPage onBack={() => setPage('main')} />
  }

  return (
    <div className="min-h-screen flex flex-col bg-background">
      <Header userId={user?.id} onSettingsClick={() => setPage('settings')} />
      <main className="flex-1 flex flex-col items-center justify-center gap-4 p-4">
        <div className="p-8 rounded-xl bg-card text-card-foreground border border-border">
          <h1 className="text-2xl font-bold">
            Welcome
          </h1>
        </div>
        <Button>Get Started</Button>
      </main>
    </div>
  )
}

export default App
