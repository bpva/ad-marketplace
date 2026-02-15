export function formatCompact(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, "")}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1).replace(/\.0$/, "")}K`;
  return n.toString();
}

export function formatTonAmount(nanoTon: number): string {
  const ton = nanoTon / 1_000_000_000;
  if (ton >= 1) {
    const s = ton.toFixed(ton % 1 === 0 ? 0 : 1);
    return s;
  }
  return `${(ton * 1000).toFixed(0)}m`;
}

const CURRENCY_SYMBOLS: Record<string, string> = {
  usd: "$",
  eur: "€",
  gbp: "£",
  rub: "₽",
};

const SUFFIX_CURRENCIES = new Set(["rub"]);

export function formatFiat(amount: number, currency: string): string {
  const sym = CURRENCY_SYMBOLS[currency] ?? currency.toUpperCase();
  let formatted: string;
  if (amount >= 1000) {
    formatted = Math.round(amount).toLocaleString("en-US");
  } else if (amount >= 0.01) {
    formatted = amount.toFixed(2);
  } else if (amount > 0) {
    formatted = amount.toFixed(3);
  } else {
    formatted = "0";
  }
  return SUFFIX_CURRENCIES.has(currency) ? `${formatted} ${sym}` : `${sym}${formatted}`;
}

export function nanoTonToFiat(nanoTon: number, rate: number): number {
  return (nanoTon / 1_000_000_000) * rate;
}

export function tonToFiat(ton: number, rate: number): number {
  return ton * rate;
}
