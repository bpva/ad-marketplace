import { useState } from "react";
import { ArrowLeft, Check, CalendarIcon, Clock } from "lucide-react";
import { format } from "date-fns";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { ChannelAvatar } from "@/components/ChannelAvatar";
import { ChannelPost } from "@/components/TelegramMessagePreview";
import { TonPrice } from "@/components/TonPrice";
import { getFormatDisplay } from "@/lib/adFormats";
import { formatCompact } from "@/lib/format";
import { useTemplates } from "@/hooks/useTemplates";
import { createDeal, ApiError } from "@/lib/api";
import type { MarketplaceChannel, MarketplaceAdFormat, TemplateResponse } from "@/lib/api";
import { cn, notify } from "@/lib/utils";

type Step = 1 | 2 | 3 | 4;

interface CreateDealPageProps {
  channel: MarketplaceChannel;
  onBack: () => void;
}

const STEP_TITLES: Record<Step, string> = {
  1: "Select Format",
  2: "Select Template",
  3: "Schedule",
  4: "Review & Submit",
};

const MINUTES = [0, 15, 30, 45] as const;
const HOURS = Array.from({ length: 24 }, (_, i) => i);

export function CreateDealPage({ channel, onBack }: CreateDealPageProps) {
  const [step, setStep] = useState<Step>(1);
  const [selectedFormat, setSelectedFormat] = useState<MarketplaceAdFormat | null>(null);
  const [selectedTemplate, setSelectedTemplate] = useState<TemplateResponse | null>(null);
  const [selectedDate, setSelectedDate] = useState<Date | undefined>();
  const [selectedHour, setSelectedHour] = useState(12);
  const [selectedMinute, setSelectedMinute] = useState<number>(0);
  const [submitting, setSubmitting] = useState(false);

  const formats = channel.ad_formats ?? [];

  const canNext = () => {
    switch (step) {
      case 1:
        return selectedFormat != null;
      case 2:
        return selectedTemplate != null;
      case 3: {
        if (!selectedDate) return false;
        const dt = new Date(selectedDate);
        dt.setHours(selectedHour, selectedMinute, 0, 0);
        return dt > new Date();
      }
      case 4:
        return true;
    }
  };

  const handleNext = () => {
    if (step < 4) setStep((step + 1) as Step);
  };

  const handleBack = () => {
    if (step > 1) setStep((step - 1) as Step);
    else onBack();
  };

  const handleSubmit = async () => {
    if (!selectedFormat || !selectedTemplate || !selectedDate) return;
    const dt = new Date(selectedDate);
    dt.setHours(selectedHour, selectedMinute, 0, 0);

    setSubmitting(true);
    try {
      await createDeal({
        channel_id: channel.id!,
        format_type: selectedFormat.format_type!,
        is_native: selectedFormat.is_native ?? false,
        feed_hours: selectedFormat.feed_hours ?? 12,
        top_hours: selectedFormat.top_hours ?? 2,
        price_nano_ton: selectedFormat.price_nano_ton!,
        template_post_id: selectedTemplate.id!,
        scheduled_at: dt.toISOString(),
      });
      notify("Deal created!");
      onBack();
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        notify("Price has changed. Please re-select format.");
        setSelectedFormat(null);
        setStep(1);
      } else {
        notify("Failed to create deal");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="flex-1 flex flex-col p-4">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <div className="flex items-center gap-3 mb-4">
          <button
            type="button"
            onClick={handleBack}
            className="p-1.5 -ml-1 rounded-lg hover:bg-accent transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
          </button>
          <div className="flex items-center gap-1.5">
            {([1, 2, 3, 4] as Step[]).map((s) => (
              <div
                key={s}
                className={cn(
                  "rounded-full transition-all duration-300",
                  s === step
                    ? "w-2 h-2 bg-primary"
                    : s < step
                      ? "w-1.5 h-1.5 bg-primary/40"
                      : "w-1.5 h-1.5 bg-muted-foreground/20",
                )}
              />
            ))}
          </div>
          <span className="text-sm text-muted-foreground ml-auto">{step}/4</span>
        </div>

        <h2 className="text-lg font-semibold mb-4">{STEP_TITLES[step]}</h2>

        <div className="flex-1 overflow-y-auto pb-20">
          {step === 1 && (
            <StepFormat
              channel={channel}
              formats={formats}
              selected={selectedFormat}
              onSelect={setSelectedFormat}
            />
          )}
          {step === 2 && (
            <StepTemplate selected={selectedTemplate} onSelect={setSelectedTemplate} />
          )}
          {step === 3 && (
            <StepSchedule
              date={selectedDate}
              onDateChange={setSelectedDate}
              hour={selectedHour}
              onHourChange={setSelectedHour}
              minute={selectedMinute}
              onMinuteChange={setSelectedMinute}
            />
          )}
          {step === 4 && (
            <StepReview
              channel={channel}
              format={selectedFormat!}
              template={selectedTemplate!}
              date={selectedDate!}
              hour={selectedHour}
              minute={selectedMinute}
            />
          )}
        </div>

        <div
          className="fixed left-0 right-0 z-20 p-4 bg-background/95 backdrop-blur-sm border-t border-border"
          style={{ bottom: "calc(64px + var(--safe-area-inset-bottom, 0px))" }}
        >
          <div className="max-w-md mx-auto">
            {step < 4 ? (
              <Button className="w-full" disabled={!canNext()} onClick={handleNext}>
                Next
              </Button>
            ) : (
              <Button className="w-full" disabled={submitting} onClick={handleSubmit}>
                {submitting ? "Creating..." : "Create Deal"}
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function StepFormat({
  channel,
  formats,
  selected,
  onSelect,
}: {
  channel: MarketplaceChannel;
  formats: MarketplaceAdFormat[];
  selected: MarketplaceAdFormat | null;
  onSelect: (f: MarketplaceAdFormat) => void;
}) {
  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3 p-3 rounded-xl bg-card border border-border">
        <ChannelAvatar
          channelId={channel.id ?? 0}
          hasPhoto={!!channel.photo_small_url}
          className="w-10 h-10 flex-shrink-0"
        />
        <div className="min-w-0">
          <h3 className="font-medium truncate text-sm">{channel.title}</h3>
          {channel.username && (
            <span className="text-xs text-muted-foreground">@{channel.username}</span>
          )}
          {channel.subscribers != null && (
            <span className="text-xs text-muted-foreground ml-2">
              {formatCompact(channel.subscribers)} subscribers
            </span>
          )}
        </div>
      </div>

      {formats.length === 0 ? (
        <div className="text-center py-8 text-sm text-muted-foreground">
          No ad formats available
        </div>
      ) : (
        <div className="space-y-2">
          {formats.map((f, i) => {
            const display = getFormatDisplay(f.format_type, f.is_native);
            const Icon = display.icon;
            const isSelected =
              selected?.format_type === f.format_type &&
              selected?.is_native === f.is_native &&
              selected?.feed_hours === f.feed_hours &&
              selected?.top_hours === f.top_hours;

            const duration =
              f.format_type === "post"
                ? `${f.feed_hours}h feed + ${f.top_hours}h top`
                : f.format_type === "story"
                  ? "24h story"
                  : "Repost";

            return (
              <button
                key={i}
                type="button"
                onClick={() => onSelect(f)}
                className={cn(
                  "relative w-full rounded-xl border-2 p-3 text-left transition-all",
                  isSelected
                    ? "border-primary bg-primary/5"
                    : "border-border bg-card hover:border-primary/40",
                )}
              >
                {isSelected && (
                  <div className="absolute top-3 right-3">
                    <div className="w-5 h-5 rounded-full bg-primary flex items-center justify-center">
                      <Check className="h-3 w-3 text-primary-foreground" />
                    </div>
                  </div>
                )}
                <div className="flex items-center gap-3">
                  <div
                    className={cn(
                      "w-9 h-9 rounded-lg flex items-center justify-center",
                      display.bgColor,
                    )}
                  >
                    <Icon className={cn("h-4 w-4", display.color)} />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="font-medium text-sm">{display.label}</div>
                    <div className="text-xs text-muted-foreground">{duration}</div>
                  </div>
                  {f.price_nano_ton != null && (
                    <TonPrice nanoTon={f.price_nano_ton} className="text-sm" />
                  )}
                </div>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

function StepTemplate({
  selected,
  onSelect,
}: {
  selected: TemplateResponse | null;
  onSelect: (t: TemplateResponse) => void;
}) {
  const { templates, loading } = useTemplates();

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <span className="text-sm text-muted-foreground">Loading templates...</span>
      </div>
    );
  }

  if (templates.length === 0) {
    return (
      <div className="text-center py-12 space-y-2">
        <p className="text-sm text-muted-foreground">No templates yet</p>
        <p className="text-xs text-muted-foreground">
          Use <span className="font-mono bg-muted px-1 py-0.5 rounded">/add_promo</span> in the bot
          to create one
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {templates.map((t) => {
        const isSelected = selected?.id === t.id;
        const hasMedia = t.media && t.media.length > 0;
        return (
          <button
            key={t.id}
            type="button"
            onClick={() => onSelect(t)}
            className={cn(
              "relative w-full rounded-xl border-2 p-3 text-left transition-all",
              isSelected
                ? "border-primary bg-primary/5"
                : "border-border bg-card hover:border-primary/40",
            )}
          >
            {isSelected && (
              <div className="absolute top-3 right-3 z-10">
                <div className="w-5 h-5 rounded-full bg-primary flex items-center justify-center">
                  <Check className="h-3 w-3 text-primary-foreground" />
                </div>
              </div>
            )}
            <div className="flex items-start gap-3">
              <div className="flex-1 min-w-0">
                <div className="font-medium text-sm truncate">{t.name || "Untitled"}</div>
                {t.text && (
                  <div className="text-xs text-muted-foreground mt-1 line-clamp-2">{t.text}</div>
                )}
                <div className="flex items-center gap-2 mt-1.5 text-[11px] text-muted-foreground">
                  {hasMedia && <span>{t.media!.length} media</span>}
                  {t.text && <span>{t.text.length} chars</span>}
                </div>
              </div>
            </div>
          </button>
        );
      })}
    </div>
  );
}

function StepSchedule({
  date,
  onDateChange,
  hour,
  onHourChange,
  minute,
  onMinuteChange,
}: {
  date: Date | undefined;
  onDateChange: (d: Date | undefined) => void;
  hour: number;
  onHourChange: (h: number) => void;
  minute: number;
  onMinuteChange: (m: number) => void;
}) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  return (
    <div className="space-y-4">
      <div className="flex justify-center">
        <Calendar
          mode="single"
          selected={date}
          onSelect={onDateChange}
          disabled={{ before: today }}
        />
      </div>

      <div className="space-y-2">
        <div className="flex items-center gap-2 text-sm font-medium">
          <Clock className="h-4 w-4 text-muted-foreground" />
          Time
        </div>
        <div className="flex items-center gap-2">
          <select
            value={hour}
            onChange={(e) => onHourChange(Number(e.target.value))}
            className="flex-1 rounded-lg border border-border bg-card px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          >
            {HOURS.map((h) => (
              <option key={h} value={h}>
                {String(h).padStart(2, "0")}
              </option>
            ))}
          </select>
          <span className="text-lg font-medium text-muted-foreground">:</span>
          <select
            value={minute}
            onChange={(e) => onMinuteChange(Number(e.target.value))}
            className="flex-1 rounded-lg border border-border bg-card px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          >
            {MINUTES.map((m) => (
              <option key={m} value={m}>
                {String(m).padStart(2, "0")}
              </option>
            ))}
          </select>
        </div>
      </div>

      {date && (
        <div className="flex items-center gap-2 rounded-lg bg-muted/50 px-3 py-2 text-sm">
          <CalendarIcon className="h-4 w-4 text-muted-foreground" />
          <span>
            {format(date, "MMMM d, yyyy")} at {String(hour).padStart(2, "0")}:
            {String(minute).padStart(2, "0")}
          </span>
        </div>
      )}
    </div>
  );
}

function StepReview({
  channel,
  format: fmt,
  template,
  date,
  hour,
  minute,
}: {
  channel: MarketplaceChannel;
  format: MarketplaceAdFormat;
  template: TemplateResponse;
  date: Date;
  hour: number;
  minute: number;
}) {
  const display = getFormatDisplay(fmt.format_type, fmt.is_native);
  const Icon = display.icon;
  const duration =
    fmt.format_type === "post"
      ? `${fmt.feed_hours}h feed + ${fmt.top_hours}h top`
      : fmt.format_type === "story"
        ? "24h story"
        : "Repost";

  return (
    <div className="space-y-4">
      <div className="rounded-xl border border-border p-3 space-y-3">
        <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
          Channel
        </div>
        <div className="flex items-center gap-3">
          <ChannelAvatar
            channelId={channel.id ?? 0}
            hasPhoto={!!channel.photo_small_url}
            className="w-8 h-8"
          />
          <div className="min-w-0">
            <div className="font-medium text-sm truncate">{channel.title}</div>
            {channel.username && (
              <span className="text-xs text-muted-foreground">@{channel.username}</span>
            )}
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-border p-3 space-y-2">
        <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
          Format
        </div>
        <div className="flex items-center gap-3">
          <div
            className={cn("w-8 h-8 rounded-lg flex items-center justify-center", display.bgColor)}
          >
            <Icon className={cn("h-4 w-4", display.color)} />
          </div>
          <div className="flex-1 min-w-0">
            <div className="font-medium text-sm">{display.label}</div>
            <div className="text-xs text-muted-foreground">{duration}</div>
          </div>
          {fmt.price_nano_ton != null && (
            <TonPrice nanoTon={fmt.price_nano_ton} className="text-sm" />
          )}
        </div>
      </div>

      <div className="rounded-xl border border-border p-3 space-y-2">
        <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
          Schedule
        </div>
        <div className="flex items-center gap-2 text-sm">
          <CalendarIcon className="h-4 w-4 text-muted-foreground" />
          {format(date, "MMMM d, yyyy")} at {String(hour).padStart(2, "0")}:
          {String(minute).padStart(2, "0")}
        </div>
      </div>

      <div className="rounded-xl border border-border overflow-hidden">
        <div className="px-3 pt-3 pb-1">
          <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Post Preview
          </div>
        </div>
        <div className="p-3">
          <ChannelPost template={template} />
        </div>
      </div>
    </div>
  );
}
