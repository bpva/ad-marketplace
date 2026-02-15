import { ChevronRight, Users } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ChannelWithRole } from "@/lib/api";
import { ChannelAvatar } from "@/components/ChannelAvatar";
import { formatCompact } from "@/lib/format";

interface ChannelListProps {
  channels: ChannelWithRole[];
  onChannelClick?: (channel: ChannelWithRole) => void;
}

export function ChannelList({ channels, onChannelClick }: ChannelListProps) {
  return (
    <div className="p-4">
      <div className="max-w-md mx-auto space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Your Channels</h2>
          <span className="text-sm text-muted-foreground">{channels.length}</span>
        </div>

        <div className="space-y-3">
          {channels.map((item) => (
            <ChannelCard
              key={item.channel?.id}
              item={item}
              onClick={() => onChannelClick?.(item)}
            />
          ))}
        </div>

        <p className="text-xs text-muted-foreground text-center pt-4">
          Add more channels by making @adxchange_bot an admin
        </p>
      </div>
    </div>
  );
}

type ChannelStatus = "active" | "no-formats" | "no-stats" | "unlisted";

function getChannelStatus(item: ChannelWithRole): ChannelStatus {
  if (!item.channel?.is_listed) return "unlisted";
  if (!item.channel.ad_formats?.length) return "no-formats";
  if (!item.channel.has_stats) return "no-stats";
  return "active";
}

const statusConfig: Record<ChannelStatus, { color: string; glow: string; hint: string }> = {
  active: {
    color: "bg-emerald-400",
    glow: "shadow-[0_0_6px_2px_rgba(52,211,153,0.5)]",
    hint: "Listed on marketplace",
  },
  "no-formats": {
    color: "bg-amber-400",
    glow: "shadow-[0_0_6px_2px_rgba(251,191,36,0.5)]",
    hint: "Add ad formats to receive offers",
  },
  "no-stats": {
    color: "bg-amber-400",
    glow: "shadow-[0_0_6px_2px_rgba(251,191,36,0.5)]",
    hint: "Listed without verified stats",
  },
  unlisted: {
    color: "bg-zinc-400 dark:bg-zinc-500",
    glow: "",
    hint: "Unlisted, not showing in marketplace",
  },
};

function ChannelCard({ item, onClick }: { item: ChannelWithRole; onClick: () => void }) {
  const { channel, role } = item;
  const status = getChannelStatus(item);
  const { color, glow, hint } = statusConfig[status];

  return (
    <button
      type="button"
      onClick={onClick}
      className="relative w-full bg-card rounded-xl border border-border p-4 flex items-center gap-3 text-left transition-colors hover:bg-accent/50 active:bg-accent"
    >
      <div className="group absolute top-2.5 right-2.5 z-10">
        <div className={cn("h-2 w-2 rounded-full", color, glow)} />
        <div className="pointer-events-none absolute right-0 top-full mt-1 w-max max-w-[200px] rounded-md bg-popover px-2.5 py-1.5 text-xs text-popover-foreground shadow-md border border-border opacity-0 group-hover:opacity-100 transition-opacity">
          {hint}
        </div>
      </div>

      <ChannelAvatar
        channelId={channel?.id ?? 0}
        photoUrl={channel?.photo_small_url}
        className="w-12 h-12 flex-shrink-0"
      />

      <div className="flex-1 min-w-0">
        <h3 className="font-medium truncate">{channel?.title}</h3>
        {channel?.username && (
          <p className="text-sm text-muted-foreground truncate">@{channel.username}</p>
        )}
        <div className="flex items-center gap-3 mt-1">
          {channel?.subscribers != null && (
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <Users className="h-3 w-3" />
              {formatCompact(channel.subscribers)}
            </span>
          )}
        </div>
      </div>

      <div
        className={cn(
          "px-2 py-1 rounded-md text-xs font-medium",
          role === "owner" ? "bg-primary/10 text-primary" : "bg-muted text-muted-foreground",
        )}
      >
        {role === "owner" ? "Owner" : "Manager"}
      </div>

      <ChevronRight className="h-5 w-5 text-muted-foreground flex-shrink-0" />
    </button>
  );
}
