import { useMemo, useState } from "react";
import {
  Search,
  Users,
  Eye,
  ChevronLeft,
  ChevronRight,
  ArrowUpDown,
  FileText,
  Repeat2,
  Camera,
  TrendingUp,
  TrendingDown,
  MessageCircle,
  Target,
} from "lucide-react";
import { Pie, PieChart, Cell } from "recharts";
import { ChartContainer, ChartTooltip, type ChartConfig } from "@/components/ui/chart";
import { useMarketplace } from "@/hooks/useMarketplace";
import { ChannelAvatar } from "@/components/ChannelAvatar";
import type { MarketplaceChannel, MarketplaceAdFormat } from "@/lib/api";
import { formatCompact, formatNanoTon } from "@/lib/format";
import { normalizeLangs, type LangSlice } from "@/lib/lang";

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

function Tooltip({ text, children }: { text: string; children: React.ReactNode }) {
  const [open, setOpen] = useState(false);
  return (
    <span
      className="relative inline-flex items-center"
      onPointerEnter={() => setOpen(true)}
      onPointerLeave={() => setOpen(false)}
      onClick={() => setOpen((v) => !v)}
    >
      {children}
      {open && (
        <span className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1.5 px-2 py-1 rounded bg-foreground text-background text-[10px] whitespace-nowrap z-10">
          {text}
        </span>
      )}
    </span>
  );
}

const FORMAT_ICON = {
  post: FileText,
  repost: Repeat2,
  story: Camera,
} as const;

function AdFormatBadge({ format }: { format: MarketplaceAdFormat }) {
  const Icon = FORMAT_ICON[format.format_type ?? "post"];
  const label =
    format.format_type === "post"
      ? `${format.feed_hours}+${format.top_hours}`
      : format.format_type === "repost"
        ? "Repost"
        : "Story";
  const tooltip =
    format.format_type === "post"
      ? `Post: ${format.feed_hours}h in feed, ${format.top_hours}h pinned on top`
      : format.format_type === "repost"
        ? "Repost of your message"
        : "24h story";

  return (
    <Tooltip text={tooltip}>
      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-muted text-[11px] text-muted-foreground">
        <Icon className="h-3 w-3 flex-shrink-0" />
        {label}
      </span>
    </Tooltip>
  );
}

function LanguagePie({ langs }: { langs: LangSlice[] }) {
  const config = useMemo(() => {
    const c: ChartConfig = {};
    for (const l of langs) {
      c[l.lang] = { label: l.flag, color: l.fill };
    }
    return c;
  }, [langs]);

  return (
    <div className="inline-flex items-center gap-1.5">
      <ChartContainer
        config={config}
        className="flex-shrink-0 !aspect-square"
        style={{ height: 32, width: 32 }}
      >
        <PieChart>
          <Pie
            data={langs}
            dataKey="pct"
            nameKey="lang"
            cx="50%"
            cy="50%"
            outerRadius={15}
            innerRadius={7}
            strokeWidth={1}
            stroke="hsl(var(--background))"
            isAnimationActive={false}
          >
            {langs.map((l) => (
              <Cell key={l.lang} fill={l.fill} />
            ))}
          </Pie>
          <ChartTooltip
            content={({ active, payload }) => {
              if (!active || !payload?.length) return null;
              const d = payload[0].payload as LangSlice;
              return (
                <div className="rounded-md border border-border bg-background px-2 py-1 text-xs shadow-md">
                  {d.flag} {d.pct}%
                </div>
              );
            }}
          />
        </PieChart>
      </ChartContainer>
      <div className="flex flex-col gap-px">
        {langs.map((l) => (
          <div key={l.lang} className="flex items-center gap-1 text-[10px] leading-tight">
            <span
              className="inline-block w-1.5 h-1.5 rounded-full flex-shrink-0"
              style={{ background: l.fill }}
            />
            <span>{l.flag}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function topReactions(reactions?: Record<string, number>, limit = 3): [string, number][] {
  if (!reactions) return [];
  return Object.entries(reactions)
    .filter(([, v]) => v > 0)
    .sort((a, b) => b[1] - a[1])
    .slice(0, limit);
}

function MarketplaceCard({ channel }: { channel: MarketplaceChannel }) {
  const formats = channel.ad_formats ?? [];
  const prices = formats.map((f) => f.price_nano_ton).filter((p): p is number => p != null);
  const cheapest = prices.length ? Math.min(...prices) : null;
  const growth = channel.sub_growth_7d;
  const langs = normalizeLangs(channel.languages);
  const reactions = topReactions(channel.reactions_by_emotion);
  const storyReactions = topReactions(channel.story_reactions_by_emotion);
  const storyTotal = storyReactions.reduce((sum, [, c]) => sum + c, 0);
  const hasReactions = reactions.length > 0 || channel.avg_interactions_7d != null;

  return (
    <div className="w-full bg-card rounded-xl border border-border p-3 space-y-2">
      <div className="flex items-center gap-3">
        <ChannelAvatar
          channelId={channel.id ?? 0}
          photoUrl={channel.photo_small_url}
          className="w-10 h-10 flex-shrink-0"
        />
        <div className="min-w-0">
          <h3 className="font-medium truncate text-sm">{channel.title}</h3>
          {channel.username && (
            <p className="text-xs text-muted-foreground truncate">@{channel.username}</p>
          )}
        </div>
      </div>

      <div className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
        {channel.subscribers != null && (
          <span className="inline-flex items-center gap-1">
            <Users className="h-3 w-3" />
            {formatCompact(channel.subscribers)}
            {growth != null && growth !== 0 && (
              <span
                className={`inline-flex items-center gap-0.5 ${growth > 0 ? "text-green-500" : "text-red-500"}`}
              >
                {growth > 0 ? (
                  <TrendingUp className="h-2.5 w-2.5" />
                ) : (
                  <TrendingDown className="h-2.5 w-2.5" />
                )}
                {growth > 0 ? "+" : ""}
                {formatCompact(growth)}
              </span>
            )}
          </span>
        )}
        {channel.avg_daily_views_7d != null && (
          <Tooltip text="Avg. daily views (7d)">
            <span className="inline-flex items-center gap-1">
              <Eye className="h-3 w-3" />
              {formatCompact(channel.avg_daily_views_7d)}
            </span>
          </Tooltip>
        )}
        {channel.engagement_rate_7d != null && (
          <Tooltip text="Engagement rate: interactions / views (7d)">
            <span className="inline-flex items-center gap-1">
              <Target className="h-3 w-3" />
              {(channel.engagement_rate_7d * 100).toFixed(1)}%
            </span>
          </Tooltip>
        )}
      </div>

      {langs.length > 0 && <LanguagePie langs={langs} />}

      {hasReactions && (
        <Tooltip
          text={
            "Avg. daily interactions (7d)" +
            (storyTotal > 0
              ? `\nStory reactions: ${storyReactions.map(([e, c]) => `${e} ${formatCompact(c)}`).join(" ")}`
              : "")
          }
        >
          <div className="inline-flex flex-col gap-1.5 rounded-lg border border-border px-2.5 py-1.5">
            <div className="flex items-center justify-center gap-3 text-xs text-muted-foreground">
              {channel.avg_interactions_7d != null && (
                <span className="inline-flex items-center gap-1">
                  <MessageCircle className="h-3 w-3" />
                  {formatCompact(channel.avg_interactions_7d)}
                </span>
              )}
              {storyTotal > 0 && (
                <span className="inline-flex items-center gap-1">
                  <Camera className="h-3 w-3" />
                  {formatCompact(storyTotal)}
                </span>
              )}
            </div>
            {reactions.length > 0 && (
              <div className="flex flex-wrap justify-center gap-1">
                {reactions.map(([emoji, count]) => (
                  <span
                    key={emoji}
                    className="inline-flex items-center gap-1 rounded-full bg-muted px-1.5 py-0.5"
                  >
                    <span className="text-xs leading-none">{emoji}</span>
                    <span className="text-[10px] text-muted-foreground">
                      {formatCompact(count)}
                    </span>
                  </span>
                ))}
              </div>
            )}
          </div>
        </Tooltip>
      )}

      {formats.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {formats.map((f, i) => (
            <AdFormatBadge key={i} format={f} />
          ))}
        </div>
      )}

      {cheapest != null && (
        <button
          type="button"
          className="w-full py-2 rounded-lg bg-primary text-primary-foreground text-sm font-medium"
        >
          Place ad from {formatNanoTon(cheapest)}
        </button>
      )}
    </div>
  );
}
