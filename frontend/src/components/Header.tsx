import { Avatar } from "@/components/ui/avatar";
import { getAvatarUrl } from "@/lib/avatar";

interface HeaderProps {
  userName?: string;
  telegramId?: number;
  photoUrl?: string;
  onSettingsClick?: () => void;
}

export function Header({ userName, telegramId, photoUrl, onSettingsClick }: HeaderProps) {
  const name = userName || "Guest";
  const initials = name.charAt(0).toUpperCase();
  const avatarSrc = telegramId ? getAvatarUrl(telegramId, photoUrl) : undefined;

  return (
    <header
      className="sticky top-0 z-10 w-full flex items-center justify-end px-4 pb-3 bg-background border-b border-border"
      style={{ paddingTop: "calc(var(--total-safe-area-top, 0px) + 0.75rem)" }}
    >
      <button className="flex items-center gap-2" onClick={onSettingsClick}>
        <span className="text-sm font-medium text-foreground">{name}</span>
        <Avatar src={avatarSrc} fallback={initials} className="w-8 h-8 text-sm" />
      </button>
    </header>
  );
}
