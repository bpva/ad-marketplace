import { useState, useEffect, useRef } from "react";
import { ArrowLeft, Plus, Users, X, ChevronDown, Tag } from "lucide-react";
import { ChannelAvatar } from "@/components/ChannelAvatar";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { cn, notify } from "@/lib/utils";
import {
  type ChannelWithRole,
  type AdFormat,
  fetchAdFormats,
  updateChannelListing,
  deleteAdFormat,
  updateCategories,
} from "@/lib/api";
import { AddAdFormatSheet } from "@/components/AddAdFormatSheet";
import { getFormatDisplay } from "@/lib/adFormats";
import { formatCompact } from "@/lib/format";
import { TonPrice } from "@/components/TonPrice";
import { ALL_CATEGORIES } from "@/lib/categories";

interface ChannelDetailPageProps {
  channel: ChannelWithRole;
  onBack: () => void;
}

export function ChannelDetailPage({ channel, onBack }: ChannelDetailPageProps) {
  const [formats, setFormats] = useState<AdFormat[]>([]);
  const [loading, setLoading] = useState(true);
  const [isListed, setIsListed] = useState(channel.channel?.is_listed ?? false);
  const [saving, setSaving] = useState(false);
  const [showAddSheet, setShowAddSheet] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [categories, setCategories] = useState<string[]>(
    () => channel.channel?.categories?.map((c) => c.slug ?? "").filter(Boolean) ?? [],
  );
  const [savingCategories, setSavingCategories] = useState(false);

  const isOwner = channel.role === "owner";
  const channelId = channel.channel?.id;

  const loadFormats = () => {
    if (!channelId) return;
    setLoading(true);
    fetchAdFormats(channelId)
      .then((res) => setFormats(res.ad_formats ?? []))
      .catch(() => notify("Failed to load ad formats"))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadFormats();
  }, [channelId]);

  const handleListingChange = async (checked: boolean) => {
    if (!channelId) return;
    setSaving(true);
    try {
      await updateChannelListing(channelId, checked);
      setIsListed(checked);
      notify(checked ? "Channel listed" : "Channel unlisted");
    } catch {
      notify("Failed to update listing");
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteFormat = async (formatId: string) => {
    if (!channelId) return;
    setDeletingId(formatId);
    try {
      await deleteAdFormat(channelId, formatId);
      setFormats((prev) => prev.filter((f) => f.id !== formatId));
      notify("Format deleted");
    } catch {
      notify("Failed to delete format");
    } finally {
      setDeletingId(null);
      setConfirmDeleteId(null);
    }
  };

  const handleCategoryToggle = async (slug: string) => {
    if (!channelId) return;
    const next = categories.includes(slug)
      ? categories.filter((s) => s !== slug)
      : [...categories, slug];
    if (next.length > 3) {
      notify("Max 3 categories");
      return;
    }
    setSavingCategories(true);
    try {
      await updateCategories(channelId, next);
      setCategories(next);
    } catch {
      notify("Failed to update categories");
    } finally {
      setSavingCategories(false);
    }
  };

  return (
    <div
      className="min-h-screen bg-background p-4"
      style={{ paddingTop: "calc(var(--total-safe-area-top, 0px) + 1rem)" }}
    >
      <div className="max-w-md mx-auto space-y-4">
        <div className="flex items-center justify-between">
          <Button variant="ghost" size="sm" onClick={onBack}>
            <ArrowLeft className="h-4 w-4 mr-1" />
            Back
          </Button>
          <div
            className={cn(
              "px-2 py-1 rounded-md text-xs font-medium",
              isOwner ? "bg-primary/10 text-primary" : "bg-muted text-muted-foreground",
            )}
          >
            {isOwner ? "Owner" : "Manager"}
          </div>
        </div>

        <div className="bg-card rounded-xl border border-border p-4 flex items-center gap-3">
          <ChannelAvatar
            channelId={channel.channel?.id ?? 0}
            hasPhoto={!!channel.channel?.photo_small_url}
            className="w-12 h-12 flex-shrink-0"
          />
          <div className="flex-1 min-w-0">
            <h2 className="font-semibold text-lg truncate">{channel.channel?.title}</h2>
            {channel.channel?.username && (
              <p className="text-sm text-muted-foreground truncate">@{channel.channel.username}</p>
            )}
            <div className="flex items-center gap-3 mt-1">
              {channel.channel?.subscribers != null && (
                <span className="flex items-center gap-1 text-xs text-muted-foreground">
                  <Users className="h-3 w-3" />
                  {formatCompact(channel.channel.subscribers)}
                </span>
              )}
            </div>
          </div>
        </div>

        {isOwner && (
          <div className="bg-card rounded-lg border border-border p-4">
            <div className="flex items-center justify-between">
              <div className="space-y-1">
                <Label className="font-medium">Visibility</Label>
                <p className="text-sm text-muted-foreground">Listed in marketplace</p>
              </div>
              <Switch checked={isListed} onCheckedChange={handleListingChange} disabled={saving} />
            </div>
          </div>
        )}

        {isOwner && (
          <CategoryEditor
            selected={categories}
            onToggle={handleCategoryToggle}
            saving={savingCategories}
          />
        )}

        <div className="bg-card rounded-lg border border-border p-4 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Ad Formats</h3>
            {isOwner && (
              <Button variant="ghost" size="icon" onClick={() => setShowAddSheet(true)}>
                <Plus className="h-5 w-5" />
              </Button>
            )}
          </div>

          {loading ? (
            <div className="text-center text-muted-foreground py-4">Loading...</div>
          ) : formats.length === 0 ? (
            <div className="text-center text-muted-foreground py-4">
              No ad formats configured
              {isOwner && (
                <p className="text-sm mt-1">Add formats to list your channel in the marketplace</p>
              )}
            </div>
          ) : (
            <div className="space-y-3">
              {formats.map((format) => {
                const display = getFormatDisplay(format.format_type, format.is_native);
                const FormatIcon = display.icon;
                const isConfirming = confirmDeleteId === format.id;
                return (
                  <div
                    key={format.id}
                    className={cn(
                      "bg-muted/50 rounded-lg p-3 flex items-start justify-between gap-2 transition-all duration-150",
                      isConfirming && "ring-2 ring-destructive/20",
                    )}
                  >
                    <div className="space-y-1 min-w-0">
                      <div className="flex items-center gap-1.5">
                        <FormatIcon className={cn("h-4 w-4 flex-shrink-0", display.color)} />
                        <span className="text-sm font-medium">{display.label}</span>
                      </div>
                      <p className="text-sm text-muted-foreground">
                        {format.feed_hours}h feed + {format.top_hours}h top
                      </p>
                      <TonPrice nanoTon={format.price_nano_ton ?? 0} className="text-sm" />
                    </div>
                    {isOwner && (
                      <div className="flex-shrink-0">
                        {isConfirming ? (
                          <div className="flex gap-1.5">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-7 px-2 text-xs"
                              onClick={() => setConfirmDeleteId(null)}
                              disabled={deletingId === format.id}
                            >
                              Cancel
                            </Button>
                            <Button
                              variant="destructive"
                              size="sm"
                              className="h-7 px-2 text-xs"
                              onClick={() => format.id && handleDeleteFormat(format.id)}
                              disabled={deletingId === format.id}
                            >
                              {deletingId === format.id ? "..." : "Delete"}
                            </Button>
                          </div>
                        ) : (
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-muted-foreground hover:text-destructive"
                            onClick={() => format.id && setConfirmDeleteId(format.id)}
                          >
                            <X className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {channelId && (
        <AddAdFormatSheet
          open={showAddSheet}
          onClose={() => setShowAddSheet(false)}
          channelId={channelId}
          onSuccess={loadFormats}
        />
      )}
    </div>
  );
}

function CategoryEditor({
  selected,
  onToggle,
  saving,
}: {
  selected: string[];
  onToggle: (slug: string) => void;
  saving: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [filterText, setFilterText] = useState("");
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
        setFilterText("");
      }
    }
    if (open) document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open]);

  const filtered = filterText
    ? ALL_CATEGORIES.filter((c) => c.displayName.toLowerCase().includes(filterText.toLowerCase()))
    : ALL_CATEGORIES;

  return (
    <div className="bg-card rounded-lg border border-border p-4 space-y-3">
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <h3 className="font-medium">Categories</h3>
          <p className="text-sm text-muted-foreground">Up to 3 categories</p>
        </div>
      </div>

      {selected.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {selected.map((slug) => {
            const cat = ALL_CATEGORIES.find((c) => c.slug === slug);
            return (
              <span
                key={slug}
                className="inline-flex items-center gap-1 px-2 py-1 rounded-full bg-primary/10 text-primary text-xs"
              >
                {cat?.displayName ?? slug}
                <button
                  type="button"
                  onClick={() => onToggle(slug)}
                  disabled={saving}
                  className="hover:text-primary/70 disabled:opacity-50"
                >
                  <X className="h-3 w-3" />
                </button>
              </span>
            );
          })}
        </div>
      )}

      <div ref={ref} className="relative">
        <button
          type="button"
          onClick={() => setOpen(!open)}
          disabled={saving}
          className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg border border-border text-sm text-muted-foreground hover:bg-accent transition-colors disabled:opacity-50"
        >
          <Tag className="h-3.5 w-3.5" />
          {selected.length >= 3 ? "Max reached" : "Add category"}
          <ChevronDown className={`h-3.5 w-3.5 transition-transform ${open ? "rotate-180" : ""}`} />
        </button>

        {open && (
          <div className="absolute z-20 mt-1 w-64 rounded-lg border border-border bg-card shadow-lg">
            <div className="p-2 border-b border-border">
              <input
                type="text"
                placeholder="Filter categories..."
                value={filterText}
                onChange={(e) => setFilterText(e.target.value)}
                className="w-full rounded-md border border-border bg-background px-2 py-1 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
                autoFocus
              />
            </div>
            <div className="max-h-56 overflow-y-auto p-1">
              {filtered.length === 0 ? (
                <div className="px-2 py-3 text-center text-xs text-muted-foreground">
                  No categories found
                </div>
              ) : (
                filtered.map((cat) => {
                  const isSelected = selected.includes(cat.slug);
                  const disabled = !isSelected && selected.length >= 3;
                  return (
                    <button
                      key={cat.slug}
                      type="button"
                      onClick={() => onToggle(cat.slug)}
                      disabled={disabled || saving}
                      className={`w-full text-left px-2 py-1.5 rounded-md text-sm transition-colors ${
                        isSelected
                          ? "bg-primary/10 text-primary font-medium"
                          : disabled
                            ? "text-muted-foreground/50 cursor-not-allowed"
                            : "text-foreground hover:bg-accent"
                      }`}
                    >
                      {cat.displayName}
                    </button>
                  );
                })
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
