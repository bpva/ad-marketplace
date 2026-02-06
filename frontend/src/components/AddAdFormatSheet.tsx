import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import { addAdFormat, type AdFormatType } from "@/lib/api";

interface AddAdFormatSheetProps {
  open: boolean;
  onClose: () => void;
  channelId: number;
  onSuccess: () => void;
}

export function AddAdFormatSheet({ open, onClose, channelId, onSuccess }: AddAdFormatSheetProps) {
  const [formatType, setFormatType] = useState<AdFormatType>("post");
  const [isNative, setIsNative] = useState(false);
  const [feedHours, setFeedHours] = useState<12 | 24>(12);
  const [topHours, setTopHours] = useState<2 | 4>(2);
  const [price, setPrice] = useState("");
  const [saving, setSaving] = useState(false);

  const handleSubmit = async () => {
    const priceNum = parseFloat(price);
    if (isNaN(priceNum) || priceNum <= 0) {
      toast("Please enter a valid price");
      return;
    }

    setSaving(true);
    try {
      await addAdFormat(channelId, {
        format_type: formatType,
        is_native: isNative,
        feed_hours: feedHours,
        top_hours: topHours,
        price_nano_ton: Math.round(priceNum * 1e9),
      });
      toast("Ad format added");
      onSuccess();
      handleClose();
    } catch {
      toast("Failed to add format");
    } finally {
      setSaving(false);
    }
  };

  const handleClose = () => {
    setFormatType("post");
    setIsNative(false);
    setFeedHours(12);
    setTopHours(2);
    setPrice("");
    onClose();
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50">
      <div className="absolute inset-0 bg-black/50" onClick={handleClose} />
      <div className="absolute bottom-0 left-0 right-0 bg-background rounded-t-2xl p-4 pb-8 animate-in slide-in-from-bottom duration-300">
        <div className="w-12 h-1 bg-muted rounded-full mx-auto mb-4" />

        <h2 className="text-lg font-semibold mb-4">Add Ad Format</h2>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label className="text-muted-foreground">Type</Label>
            <div className="flex gap-2">
              {(["post", "repost", "story"] as const).map((type) => (
                <button
                  key={type}
                  type="button"
                  onClick={() => type === "post" && setFormatType(type)}
                  disabled={type !== "post"}
                  className={cn(
                    "flex-1 py-2 px-3 rounded-lg text-sm font-medium transition-colors",
                    formatType === type
                      ? "bg-primary text-primary-foreground"
                      : type === "post"
                        ? "bg-muted text-foreground hover:bg-muted/80"
                        : "bg-muted/50 text-muted-foreground cursor-not-allowed",
                  )}
                >
                  {type.charAt(0).toUpperCase() + type.slice(1)}
                  {type !== "post" && <span className="block text-xs opacity-70">Soon</span>}
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <Label className="text-muted-foreground">Style</Label>
            <div className="relative flex rounded-lg bg-muted p-1">
              <div
                className={cn(
                  "absolute inset-y-1 w-[calc(50%-4px)] rounded-md border-2 border-primary bg-background shadow-sm transition-all duration-200 ease-out",
                  isNative ? "left-[calc(50%+2px)]" : "left-1",
                )}
              />
              <button
                type="button"
                onClick={() => setIsNative(false)}
                className={cn(
                  "relative z-10 flex-1 py-2 text-sm font-medium transition-colors",
                  !isNative ? "text-primary" : "text-muted-foreground",
                )}
              >
                Standard
              </button>
              <button
                type="button"
                onClick={() => setIsNative(true)}
                className={cn(
                  "relative z-10 flex-1 py-2 text-sm font-medium transition-colors",
                  isNative ? "text-primary" : "text-muted-foreground",
                )}
              >
                Native
              </button>
            </div>
          </div>

          <div className="space-y-2">
            <Label className="text-muted-foreground">Feed Duration</Label>
            <div className="relative flex rounded-lg bg-muted p-1">
              <div
                className={cn(
                  "absolute inset-y-1 w-[calc(50%-4px)] rounded-md border-2 border-primary bg-background shadow-sm transition-all duration-200 ease-out",
                  feedHours === 24 ? "left-[calc(50%+2px)]" : "left-1",
                )}
              />
              <button
                type="button"
                onClick={() => setFeedHours(12)}
                className={cn(
                  "relative z-10 flex-1 py-2 text-sm font-medium transition-colors",
                  feedHours === 12 ? "text-primary" : "text-muted-foreground",
                )}
              >
                12 hours
              </button>
              <button
                type="button"
                onClick={() => setFeedHours(24)}
                className={cn(
                  "relative z-10 flex-1 py-2 text-sm font-medium transition-colors",
                  feedHours === 24 ? "text-primary" : "text-muted-foreground",
                )}
              >
                24 hours
              </button>
            </div>
          </div>

          <div className="space-y-2">
            <Label className="text-muted-foreground">Top Duration</Label>
            <div className="relative flex rounded-lg bg-muted p-1">
              <div
                className={cn(
                  "absolute inset-y-1 w-[calc(50%-4px)] rounded-md border-2 border-primary bg-background shadow-sm transition-all duration-200 ease-out",
                  topHours === 4 ? "left-[calc(50%+2px)]" : "left-1",
                )}
              />
              <button
                type="button"
                onClick={() => setTopHours(2)}
                className={cn(
                  "relative z-10 flex-1 py-2 text-sm font-medium transition-colors",
                  topHours === 2 ? "text-primary" : "text-muted-foreground",
                )}
              >
                2 hours
              </button>
              <button
                type="button"
                onClick={() => setTopHours(4)}
                className={cn(
                  "relative z-10 flex-1 py-2 text-sm font-medium transition-colors",
                  topHours === 4 ? "text-primary" : "text-muted-foreground",
                )}
              >
                4 hours
              </button>
            </div>
          </div>

          <div className="space-y-2">
            <Label className="text-muted-foreground">Price (TON)</Label>
            <Input
              type="number"
              step="0.01"
              min="0"
              placeholder="0.00"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
            />
          </div>

          <div className="flex gap-3 pt-2">
            <Button variant="outline" className="flex-1" onClick={handleClose} disabled={saving}>
              Cancel
            </Button>
            <Button className="flex-1" onClick={handleSubmit} disabled={saving}>
              {saving ? "Adding..." : "Add Format"}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
