import { Megaphone, Target, Store, Handshake } from "lucide-react";
import { cn } from "@/lib/utils";
import { Avatar } from "@/components/ui/avatar";
import { getAvatarUrl } from "@/lib/avatar";
import type { PreferredMode } from "@/lib/api";

export type NavPage = "channels" | "marketplace" | "deals" | "profile";

interface BottomNavProps {
  mode: PreferredMode;
  activePage: NavPage;
  onNavigate: (page: NavPage) => void;
  userName?: string;
  telegramId?: number;
  photoUrl?: string;
}

export function BottomNav({
  mode,
  activePage,
  onNavigate,
  userName,
  telegramId,
  photoUrl,
}: BottomNavProps) {
  const isPublisher = mode === "publisher";
  const avatarSrc = telegramId ? getAvatarUrl(telegramId, photoUrl) : undefined;
  const initials = (userName || "U").charAt(0).toUpperCase();

  const navItems: { id: NavPage; label: string; icon?: typeof Megaphone }[] = [
    {
      id: "channels",
      label: isPublisher ? "My Channels" : "Campaigns",
      icon: isPublisher ? Megaphone : Target,
    },
    { id: "marketplace", label: "Marketplace", icon: Store },
    { id: "deals", label: "Deals", icon: Handshake },
    { id: "profile", label: "Profile" },
  ];

  return (
    <nav
      className="fixed bottom-0 left-0 right-0 z-10 bg-background border-t border-border"
      style={{ paddingBottom: "var(--safe-area-inset-bottom, 0px)" }}
    >
      <div className="relative">
        <div
          className={cn(
            "absolute -top-3 left-1/2 -translate-x-1/2 px-2.5 py-0.5 rounded-full text-[10px] font-semibold border shadow-sm",
            "bg-background text-primary border-primary/30",
          )}
        >
          {isPublisher ? "Publisher" : "Advertiser"}
        </div>
      </div>
      <div className="flex items-stretch">
        {navItems.map((item) => {
          const isActive = activePage === item.id;
          const Icon = item.icon;
          return (
            <button
              key={item.id}
              onClick={() => onNavigate(item.id)}
              className={cn(
                "flex-1 flex flex-col items-center gap-1 py-2 pt-3 transition-colors",
                isActive ? "text-primary" : "text-muted-foreground",
              )}
            >
              {item.id === "profile" ? (
                <Avatar
                  src={avatarSrc}
                  fallback={initials}
                  className={cn("w-5 h-5 text-[10px]", isActive && "ring-2 ring-primary")}
                />
              ) : (
                Icon && <Icon className={cn("h-5 w-5", isActive && "stroke-[2.5]")} />
              )}
              <span className="text-[11px] font-medium">{item.label}</span>
            </button>
          );
        })}
      </div>
    </nav>
  );
}
