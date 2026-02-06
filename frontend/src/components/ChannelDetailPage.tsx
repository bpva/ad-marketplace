import { useState, useEffect } from "react";
import { toast } from "sonner";
import { ArrowLeft, Megaphone, Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import {
  type ChannelWithRole,
  type AdFormat,
  fetchAdFormats,
  updateChannelListing,
  deleteAdFormat,
} from "@/lib/api";
import { AddAdFormatSheet } from "@/components/AddAdFormatSheet";
import { getFormatDisplay } from "@/lib/adFormats";

interface ChannelDetailPageProps {
  channel: ChannelWithRole;
  onBack: () => void;
}

function formatPrice(nanoTon: number | undefined) {
  if (nanoTon === undefined || nanoTon === 0) return "0";
  const val = nanoTon / 1e9;
  return val % 1 === 0 ? val.toString() : parseFloat(val.toPrecision(10)).toString();
}

export function ChannelDetailPage({ channel, onBack }: ChannelDetailPageProps) {
  const [formats, setFormats] = useState<AdFormat[]>([]);
  const [loading, setLoading] = useState(true);
  const [isListed, setIsListed] = useState(channel.channel?.is_listed ?? false);
  const [saving, setSaving] = useState(false);
  const [showAddSheet, setShowAddSheet] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  const isOwner = channel.role === "owner";
  const channelId = channel.channel?.id;

  const loadFormats = () => {
    if (!channelId) return;
    setLoading(true);
    fetchAdFormats(channelId)
      .then((res) => setFormats(res.ad_formats ?? []))
      .catch(() => toast("Failed to load ad formats"))
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
      toast(checked ? "Channel listed" : "Channel unlisted");
    } catch {
      toast("Failed to update listing");
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
      toast("Format deleted");
    } catch {
      toast("Failed to delete format");
    } finally {
      setDeletingId(null);
      setConfirmDeleteId(null);
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
          <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
            <Megaphone className="h-6 w-6 text-primary" />
          </div>
          <div className="flex-1 min-w-0">
            <h2 className="font-semibold text-lg truncate">{channel.channel?.title}</h2>
            {channel.channel?.username && (
              <p className="text-sm text-muted-foreground truncate">@{channel.channel.username}</p>
            )}
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
                      <p className="font-medium">{formatPrice(format.price_nano_ton)} TON</p>
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
