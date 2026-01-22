import WebApp from '@twa-dev/sdk'
import { useTelegramTheme } from '@/hooks/useTelegramTheme'
import { NotInTelegram } from '@/components/NotInTelegram'
import { Button } from '@/components/ui/button'

function App() {
  useTelegramTheme()

  const isInTelegram = WebApp.initData !== ''

  if (!isInTelegram) {
    return <NotInTelegram />
  }

  const user = WebApp.initDataUnsafe.user
  const name = user?.first_name || 'Guest'

  return (
    <div className="min-h-screen flex flex-col items-center justify-center gap-4 bg-background">
      <div className="p-8 rounded-xl bg-card text-card-foreground border border-border">
        <h1 className="text-2xl font-bold">
          Hello {name}
        </h1>
      </div>
      <Button>Get Started</Button>
    </div>
  )
}

export default App
