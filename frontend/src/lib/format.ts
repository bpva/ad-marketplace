export function formatCompact(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, "")}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1).replace(/\.0$/, "")}K`;
  return n.toString();
}

export function formatNanoTon(n: number): string {
  const ton = n / 1_000_000_000;
  if (ton >= 1) return `${ton.toFixed(ton % 1 === 0 ? 0 : 1)} TON`;
  return `${(ton * 1000).toFixed(0)}m TON`;
}
