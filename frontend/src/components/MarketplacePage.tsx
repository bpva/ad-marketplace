import { Search, Users, Eye, ChevronLeft, ChevronRight, ArrowUpDown } from "lucide-react";
import { useMarketplace } from "@/hooks/useMarketplace";
import { ChannelAvatar } from "@/components/ChannelAvatar";
import type { MarketplaceChannel } from "@/lib/api";

function formatCompact(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, "")}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1).replace(/\.0$/, "")}K`;
  return n.toString();
}

function formatNanoTon(n: number): string {
  const ton = n / 1_000_000_000;
  if (ton >= 1) return `${ton.toFixed(ton % 1 === 0 ? 0 : 1)} TON`;
  return `${(ton * 1000).toFixed(0)}m TON`;
}

const PAGE_SIZE = 10;

export function MarketplacePage() {
  const {
    channels,
    total,
    loading,
    search,
    setSearch,
    sortBy,
    setSortBy,
    sortOrder,
    setSortOrder,
    page,
    setPage,
  } = useMarketplace();
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div className="flex-1 flex flex-col p-4">
      <div className="max-w-md mx-auto w-full space-y-4">
        <h1 className="text-lg font-semibold">Marketplace</h1>

        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search channels..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-lg border border-border bg-card pl-9 pr-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>

        <div className="flex items-center gap-2">
          <ArrowUpDown className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
          <div className="flex gap-1.5">
            {(["subscribers", "views"] as const).map((field) => (
              <button
                key={field}
                type="button"
                onClick={() => {
                  if (sortBy === field) {
                    setSortOrder(sortOrder === "desc" ? "asc" : "desc");
                  } else {
                    setSortBy(field);
                    setSortOrder("desc");
                  }
                }}
                className={`px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
                  sortBy === field
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground hover:bg-accent"
                }`}
              >
                {field === "subscribers" ? "Subscribers" : "Views"}
                {sortBy === field && (sortOrder === "desc" ? " ↓" : " ↑")}
              </button>
            ))}
          </div>
          <span className="text-xs text-muted-foreground ml-auto">{total} channels</span>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-12">
            <span className="text-sm text-muted-foreground">Loading...</span>
          </div>
        ) : channels.length === 0 ? (
          <div className="flex items-center justify-center py-12">
            <span className="text-sm text-muted-foreground">No channels found</span>
          </div>
        ) : (
          <div className="space-y-3">
            {channels.map((ch) => (
              <MarketplaceCard key={ch.id} channel={ch} />
            ))}
          </div>
        )}

        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-3 pt-2">
            <button
              type="button"
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
              className="p-1.5 rounded-md bg-muted text-muted-foreground disabled:opacity-40"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <span className="text-sm text-muted-foreground">
              {page} / {totalPages}
            </span>
            <button
              type="button"
              disabled={page >= totalPages}
              onClick={() => setPage(page + 1)}
              className="p-1.5 rounded-md bg-muted text-muted-foreground disabled:opacity-40"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

function MarketplaceCard({ channel }: { channel: MarketplaceChannel }) {
  return (
    <div className="w-full bg-card rounded-xl border border-border p-4 flex items-center gap-3">
      <ChannelAvatar
        channelId={channel.id}
        photoUrl={channel.photo_small_url}
        className="w-12 h-12 flex-shrink-0"
      />

      <div className="flex-1 min-w-0">
        <h3 className="font-medium truncate">{channel.title}</h3>
        {channel.username && (
          <p className="text-sm text-muted-foreground truncate">@{channel.username}</p>
        )}
        <div className="flex items-center gap-3 mt-1">
          {channel.subscribers != null && (
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <Users className="h-3 w-3" />
              {formatCompact(channel.subscribers)}
            </span>
          )}
          {channel.avg_views != null && (
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <Eye className="h-3 w-3" />
              {formatCompact(channel.avg_views)}
            </span>
          )}
        </div>
      </div>

      <div className="text-right flex-shrink-0">
        {channel.min_price_nano_ton != null && (
          <p className="text-sm font-medium">{formatNanoTon(channel.min_price_nano_ton)}</p>
        )}
        <p className="text-xs text-muted-foreground">
          {channel.ad_format_count} format{channel.ad_format_count !== 1 ? "s" : ""}
        </p>
      </div>
    </div>
  );
}
