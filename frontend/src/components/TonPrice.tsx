import { TonIcon } from "@/components/TonIcon";
import { useTonRates } from "@/hooks/useTonRates";
import { formatTonAmount, formatFiat, nanoTonToFiat } from "@/lib/format";
import type { TonRates } from "@/lib/api";
import { cn } from "@/lib/utils";

interface TonPriceProps {
  nanoTon: number;
  showFiat?: boolean;
  fiatCurrency?: keyof TonRates;
  className?: string;
}

export function TonPrice({
  nanoTon,
  showFiat = true,
  fiatCurrency = "usd",
  className,
}: TonPriceProps) {
  const rates = useTonRates();
  const amount = formatTonAmount(nanoTon);
  const rate = rates?.[fiatCurrency as keyof TonRates] as number | undefined;
  const fiat =
    showFiat && rate ? formatFiat(nanoTonToFiat(nanoTon, rate), fiatCurrency as string) : null;

  return (
    <span
      className={cn("inline-flex items-center gap-0.5 font-bold", className)}
      style={{ fontFamily: "var(--font-price)" }}
    >
      <TonIcon className="h-[1em] w-auto shrink-0" />
      <span className="leading-none">{amount}</span>
      {fiat && <span className="leading-none font-normal opacity-60">≈ {fiat}</span>}
    </span>
  );
}
