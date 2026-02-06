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
    }
  };

  const formatPrice = (nanoTon: number | undefined) => {
    if (nanoTon === undefined) return "0";
    return (nanoTon / 1e9).toFixed(2);
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
              {formats.map((format) => (
                <div
                  key={format.id}
                  className="bg-muted/50 rounded-lg p-3 flex items-start justify-between gap-2"
                >
                  <div className="space-y-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="px-2 py-0.5 rounded-md text-xs font-medium bg-primary/10 text-primary capitalize">
                        {format.format_type}
                      </span>
                      <span
                        className={cn(
                          "px-2 py-0.5 rounded-md text-xs font-medium",
                          format.is_native
                            ? "bg-green-500/10 text-green-600 dark:text-green-400"
                            : "bg-muted text-muted-foreground",
                        )}
                      >
                        {format.is_native ? "Native" : "Standard"}
                      </span>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      {format.feed_hours}h feed + {format.top_hours}h top
                    </p>
                    <p className="font-medium">{formatPrice(format.price_nano_ton)} TON</p>
                  </div>
                  {isOwner && (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="flex-shrink-0 h-8 w-8 text-muted-foreground hover:text-destructive"
                      onClick={() => format.id && handleDeleteFormat(format.id)}
                      disabled={deletingId === format.id}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              ))}
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
