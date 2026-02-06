import { ChevronRight, Megaphone } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ChannelWithRole } from "@/lib/api";

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

function ChannelCard({ item, onClick }: { item: ChannelWithRole; onClick: () => void }) {
  const { channel, role } = item;

  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full bg-card rounded-xl border border-border p-4 flex items-center gap-3 text-left transition-colors hover:bg-accent/50 active:bg-accent"
    >
      <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
        <Megaphone className="h-6 w-6 text-primary" />
      </div>

      <div className="flex-1 min-w-0">
        <h3 className="font-medium truncate">{channel?.title}</h3>
        {channel?.username && (
          <p className="text-sm text-muted-foreground truncate">@{channel.username}</p>
        )}
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
