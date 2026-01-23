import WebApp from '@twa-dev/sdk'
import { Avatar } from '@/components/ui/avatar'

interface HeaderProps {
  userId?: string
}

export function Header({ userId }: HeaderProps) {
  const user = WebApp.initDataUnsafe.user
  const name = user?.first_name || 'Guest'
  const initials = name.charAt(0).toUpperCase()

  return (
    <header
      className="sticky top-0 z-10 w-full flex items-center justify-between px-4 pb-3 bg-background border-b border-border"
      style={{ paddingTop: 'calc(var(--total-safe-area-top, 0px) + 0.75rem)' }}
    >
      {userId && (
        <span className="text-xs font-mono text-muted-foreground truncate max-w-32">
          {userId}
        </span>
      )}
      <div className="flex items-center gap-2 ml-auto">
        <span className="text-sm font-medium text-foreground">{name}</span>
        <Avatar
          src={user?.photo_url}
          fallback={initials}
          className="w-8 h-8 text-sm"
        />
      </div>
    </header>
  )
}
