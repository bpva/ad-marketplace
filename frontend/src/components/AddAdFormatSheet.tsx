import { useState, useRef, useEffect } from "react";
import { toast } from "sonner";
import { Check, ChevronLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { addAdFormat } from "@/lib/api";
import { FORMAT_CONFIGS, type FormatConfig } from "@/lib/adFormats";

interface AddAdFormatSheetProps {
  open: boolean;
  onClose: () => void;
  channelId: number;
  onSuccess: () => void;
}

type Stage = 1 | 2 | 3;

export function AddAdFormatSheet({ open, onClose, channelId, onSuccess }: AddAdFormatSheetProps) {
  const [stage, setStage] = useState<Stage>(1);
  const [selected, setSelected] = useState<FormatConfig | null>(null);
  const [feedHours, setFeedHours] = useState<12 | 24>(12);
  const [topHours, setTopHours] = useState<2 | 4>(2);
  const [price, setPrice] = useState("");
  const [saving, setSaving] = useState(false);
  const [transitioning, setTransitioning] = useState(false);
  const [visible, setVisible] = useState(false);
  const autoAdvanceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (open) {
      requestAnimationFrame(() => setVisible(true));
    }
  }, [open]);

  const goTo = (next: Stage) => {
    setTransitioning(true);
    setTimeout(() => {
      setStage(next);
      setTransitioning(false);
    }, 150);
  };

  const handleSelect = (config: FormatConfig) => {
    if (!config.enabled) return;
    setSelected(config);
    if (autoAdvanceRef.current) clearTimeout(autoAdvanceRef.current);
    autoAdvanceRef.current = setTimeout(() => goTo(2), 200);
  };

  const handleSubmit = async () => {
    const priceNum = parseFloat(price);
    if (isNaN(priceNum) || priceNum <= 0) {
      toast("Please enter a valid price");
      return;
    }
    if (!selected) return;

    setSaving(true);
    try {
      await addAdFormat(channelId, {
        format_type: selected.formatType,
        is_native: selected.isNative,
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
    setVisible(false);
    setTimeout(() => {
      setStage(1);
      setSelected(null);
      setFeedHours(12);
      setTopHours(2);
      setPrice("");
      onClose();
    }, 250);
  };

  if (!open) return null;

  const Icon = selected?.icon;
  const titles: Record<Stage, string> = { 1: "Choose Format", 2: "Set Duration", 3: "Set Price" };

  return (
    <div className="fixed inset-0 z-50">
      <div
        className={cn(
          "absolute inset-0 bg-black/50 transition-opacity duration-250",
          visible ? "opacity-100" : "opacity-0",
        )}
        onClick={handleClose}
      />
      <div
        className={cn(
          "absolute bottom-0 left-0 right-0 mx-auto max-w-lg bg-background rounded-t-2xl pb-8 transition-transform duration-300 ease-out",
          visible ? "translate-y-0" : "translate-y-full",
        )}
      >
        <div className="w-12 h-1 bg-muted rounded-full mx-auto mt-3" />

        <div className="flex items-center justify-center gap-1.5 mt-3 mb-1">
          {([1, 2, 3] as Stage[]).map((s) => (
            <div
              key={s}
              className={cn(
                "rounded-full transition-all duration-300",
                s === stage
                  ? "w-2 h-2 bg-primary"
                  : s < stage
                    ? "w-1.5 h-1.5 bg-primary/40"
                    : "w-1.5 h-1.5 bg-muted-foreground/20",
              )}
            />
          ))}
        </div>

        <div className="px-4 pt-2 pb-4">
          <h2 className="text-lg font-semibold text-center mb-4">{titles[stage]}</h2>

          <div
            className={cn(
              "transition-all duration-150",
              transitioning ? "opacity-0 translate-y-2" : "opacity-100 translate-y-0",
            )}
          >
            {stage === 1 && (
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-3">
                  {FORMAT_CONFIGS.map((config) => {
                    const FormatIcon = config.icon;
                    const isSelected = selected?.key === config.key;
                    return (
                      <button
                        key={config.key}
                        type="button"
                        onClick={() => handleSelect(config)}
                        disabled={!config.enabled}
                        className={cn(
                          "relative rounded-xl border-2 p-4 text-left transition-all duration-150",
                          config.enabled
                            ? isSelected
                              ? "border-primary bg-primary/5 shadow-sm"
                              : "border-border bg-card hover:border-primary/40 active:scale-[0.98]"
                            : "border-border/50 bg-muted/30 opacity-50",
                        )}
                      >
                        {isSelected && (
                          <div className="absolute top-2 right-2">
                            <div className="w-5 h-5 rounded-full bg-primary flex items-center justify-center">
                              <Check className="h-3 w-3 text-primary-foreground" />
                            </div>
                          </div>
                        )}
                        {!config.enabled && (
                          <div className="absolute top-2 right-2 px-1.5 py-0.5 rounded-md bg-muted text-muted-foreground text-[10px] font-medium">
                            Soon
                          </div>
                        )}
                        <div
                          className={cn(
                            "w-10 h-10 rounded-lg flex items-center justify-center mb-2.5",
                            config.bgColor,
                          )}
                        >
                          <FormatIcon className={cn("h-5 w-5", config.color)} />
                        </div>
                        <div className="font-medium text-sm">{config.label}</div>
                        <div className="text-xs text-muted-foreground mt-0.5 leading-snug">
                          {config.description}
                        </div>
                      </button>
                    );
                  })}
                </div>
                <Button variant="ghost" className="w-full" onClick={handleClose}>
                  Cancel
                </Button>
              </div>
            )}

            {stage === 2 && selected && Icon && (
              <div className="space-y-5">
                <button
                  type="button"
                  onClick={() => goTo(1)}
                  className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-muted text-sm"
                >
                  <div
                    className={cn(
                      "w-4 h-4 rounded flex items-center justify-center",
                      selected.bgColor,
                    )}
                  >
                    <Icon className={cn("h-2.5 w-2.5", selected.color)} />
                  </div>
                  <span className="font-medium text-foreground">{selected.label}</span>
                  <ChevronLeft className="h-3 w-3 text-muted-foreground" />
                </button>

                <div className="space-y-1.5">
                  <div className="text-sm font-medium">Feed Duration</div>
                  <SegmentedToggle
                    value={feedHours}
                    options={[
                      { value: 12, label: "12 hours" },
                      { value: 24, label: "24 hours" },
                    ]}
                    onChange={(v) => setFeedHours(v as 12 | 24)}
                  />
                  <p className="text-xs text-muted-foreground">
                    How long the post stays in the channel feed
                  </p>
                </div>

                <div className="space-y-1.5">
                  <div className="text-sm font-medium">Top Duration</div>
                  <SegmentedToggle
                    value={topHours}
                    options={[
                      { value: 2, label: "2 hours" },
                      { value: 4, label: "4 hours" },
                    ]}
                    onChange={(v) => setTopHours(v as 2 | 4)}
                  />
                  <p className="text-xs text-muted-foreground">
                    How long the post is pinned at the top
                  </p>
                </div>

                <div className="flex gap-3 pt-1">
                  <Button variant="ghost" className="flex-1" onClick={() => goTo(1)}>
                    Back
                  </Button>
                  <Button className="flex-1" onClick={() => goTo(3)}>
                    Next
                  </Button>
                </div>
              </div>
            )}

            {stage === 3 && selected && Icon && (
              <div className="space-y-5">
                <button
                  type="button"
                  onClick={() => goTo(2)}
                  className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-muted text-sm"
                >
                  <div
                    className={cn(
                      "w-4 h-4 rounded flex items-center justify-center",
                      selected.bgColor,
                    )}
                  >
                    <Icon className={cn("h-2.5 w-2.5", selected.color)} />
                  </div>
                  <span className="font-medium text-foreground">{selected.label}</span>
                  <span className="text-muted-foreground mx-0.5">&middot;</span>
                  <span className="text-muted-foreground">
                    {feedHours}h feed + {topHours}h top
                  </span>
                  <ChevronLeft className="h-3 w-3 text-muted-foreground" />
                </button>

                <div className="space-y-1.5">
                  <div className="text-sm font-medium">Price</div>
                  <div className="relative">
                    <Input
                      type="number"
                      min="0"
                      placeholder="0.00"
                      value={price}
                      onChange={(e) => setPrice(e.target.value)}
                      className="pr-14"
                    />
                    <div className="absolute right-3 top-1/2 -translate-y-1/2 text-sm font-medium text-muted-foreground">
                      TON
                    </div>
                  </div>
                </div>

                <div className="flex gap-3 pt-1">
                  <Button
                    variant="ghost"
                    className="flex-1"
                    onClick={() => goTo(2)}
                    disabled={saving}
                  >
                    Back
                  </Button>
                  <Button className="flex-1" onClick={handleSubmit} disabled={saving}>
                    {saving ? "Adding..." : "Add Format"}
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function SegmentedToggle<T extends number>({
  value,
  options,
  onChange,
}: {
  value: T;
  options: { value: T; label: string }[];
  onChange: (v: T) => void;
}) {
  const activeIdx = options.findIndex((o) => o.value === value);
  return (
    <div className="relative flex rounded-lg bg-muted p-1">
      <div
        className={cn(
          "absolute inset-y-1 w-[calc(50%-4px)] rounded-md border-2 border-primary bg-background shadow-sm transition-all duration-200 ease-out",
          activeIdx === 1 ? "left-[calc(50%+2px)]" : "left-1",
        )}
      />
      {options.map((opt) => (
        <button
          key={opt.value}
          type="button"
          onClick={() => onChange(opt.value)}
          className={cn(
            "relative z-10 flex-1 py-2 text-sm font-medium transition-colors",
            opt.value === value ? "text-primary" : "text-muted-foreground",
          )}
        >
          {opt.label}
        </button>
      ))}
    </div>
  );
}
