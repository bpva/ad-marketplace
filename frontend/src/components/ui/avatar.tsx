import { useState } from 'react'
import { cn } from '@/lib/utils'

interface AvatarProps {
  src?: string
  fallback: string
  className?: string
}

export function Avatar({ src, fallback, className }: AvatarProps) {
  const [error, setError] = useState(false)

  if (!src || error) {
    return (
      <div className={cn(
        "flex items-center justify-center rounded-full bg-muted text-muted-foreground font-medium",
        className
      )}>
        {fallback}
      </div>
    )
  }

  return (
    <img
      src={src}
      alt=""
      onError={() => setError(true)}
      className={cn("rounded-full object-cover", className)}
    />
  )
}
