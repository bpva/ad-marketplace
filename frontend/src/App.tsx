import WebApp from '@twa-dev/sdk'
import { useTelegramTheme } from '@/hooks/useTelegramTheme'
import { NotInTelegram } from '@/components/NotInTelegram'
import { Header } from '@/components/Header'
import { Button } from '@/components/ui/button'

function App() {
  useTelegramTheme()

  const isInTelegram = WebApp.initData !== ''

  if (!isInTelegram) {
    return <NotInTelegram />
  }

  return (
    <div className="min-h-screen flex flex-col bg-background">
      <Header />
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
