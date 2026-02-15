import { TonIcon } from "@/components/TonIcon";
import { useTonRates } from "@/hooks/useTonRates";
import { formatTonAmount, formatFiat, nanoTonToFiat } from "@/lib/format";
import type { TonRates } from "@/lib/api";
import { cn } from "@/lib/utils";

interface TonPriceProps {
  nanoTon: number;
  size?: "sm" | "md";
  showFiat?: boolean;
  fiatCurrency?: keyof TonRates;
  variant?: "branded" | "inherit";
  className?: string;
}

const SIZES = {
  sm: { icon: 14, value: "text-sm", fiat: "text-[11px]", gap: "gap-0.5" },
  md: {
    icon: 13,
    badge: "h-7",
    label: "text-[8px]",
    value: "text-base",
    fiat: "text-xs",
    gap: "gap-1.5",
  },
} as const;

export function TonPrice({
  nanoTon,
  size = "md",
  showFiat = true,
  fiatCurrency = "usd",
  variant = "branded",
  className,
}: TonPriceProps) {
  const rates = useTonRates();
  const s = SIZES[size];
  const amount = formatTonAmount(nanoTon);
  const rate = rates?.[fiatCurrency as keyof TonRates] as number | undefined;
  const fiat =
    showFiat && rate ? formatFiat(nanoTonToFiat(nanoTon, rate), fiatCurrency as string) : null;
  const brandColor = variant === "branded" ? "text-[#0098EA]" : "";

  return (
    <span className={cn("inline-flex items-center", s.gap, className)}>
      {"badge" in s ? (
        <span
          className={cn("inline-flex flex-col items-center justify-center leading-none", s.badge)}
        >
          <TonIcon size={s.icon} className={brandColor} />
          <span
            className={cn("font-bold uppercase tracking-tight", brandColor, s.label)}
            style={{ fontFamily: "var(--font-price)" }}
          >
            TON
          </span>
        </span>
      ) : (
        <TonIcon size={s.icon} className={brandColor} />
      )}
      <span
        className={cn("font-bold leading-none", s.value)}
        style={{ fontFamily: "var(--font-price)" }}
      >
        {amount}
      </span>
      {fiat && (
        <span
          className={cn(
            "leading-none",
            variant === "branded" ? "text-muted-foreground" : "opacity-70",
            s.fiat,
          )}
          style={{ fontFamily: "var(--font-price)" }}
        >
          ≈ {fiat}
        </span>
      )}
    </span>
  );
}
