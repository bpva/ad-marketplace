// Source: https://github.com/Martinnord/get-emoji-by-language-code/blob/main/data.json
const LANG_FLAGS: Record<string, string> = {
  af: "🇳🇦",
  am: "🇪🇹",
  ar: "🇦🇪",
  ay: "🇧🇴",
  az: "🇦🇿",
  be: "🇧🇾",
  bg: "🇧🇬",
  bi: "🇻🇺",
  bn: "🇧🇩",
  bs: "🇧🇦",
  ca: "🏴󠁥󠁳󠁣󠁴󠁿",
  ch: "🇬🇺",
  cs: "🇨🇿",
  da: "🇩🇰",
  de: "🇩🇪",
  dv: "🇲🇻",
  dz: "🇧🇹",
  el: "🇬🇷",
  en: "🇬🇧",
  es: "🇪🇸",
  et: "🇪🇪",
  eu: "🇪🇸",
  fa: "🇮🇷",
  ff: "🇸🇳",
  fi: "🇫🇮",
  fj: "🇫🇯",
  fo: "🇫🇴",
  fr: "🇫🇷",
  ga: "🇮🇪",
  gl: "🇪🇸",
  gn: "🇵🇾",
  gv: "🇮🇲",
  he: "🇮🇱",
  hi: "🇮🇳",
  hr: "🇭🇷",
  ht: "🇭🇹",
  hu: "🇭🇺",
  hy: "🇦🇲",
  id: "🇮🇩",
  is: "🇮🇸",
  it: "🇮🇹",
  ja: "🇯🇵",
  ka: "🇬🇪",
  kg: "🇨🇬",
  kk: "🇰🇿",
  kl: "🇬🇱",
  km: "🇰🇭",
  ko: "🇰🇷",
  ku: "🇮🇶",
  ky: "🇰🇬",
  la: "🇻🇦",
  lb: "🇱🇺",
  ln: "🇨🇩",
  lo: "🇱🇦",
  lt: "🇱🇹",
  lu: "🇨🇩",
  lv: "🇱🇻",
  mg: "🇲🇬",
  mh: "🇲🇭",
  mi: "🇳🇿",
  mk: "🇲🇰",
  mn: "🇲🇳",
  ms: "🇲🇾",
  mt: "🇲🇹",
  my: "🇲🇲",
  na: "🇳🇷",
  nb: "🇧🇻",
  nd: "🇿🇦",
  ne: "🇳🇵",
  nl: "🇳🇱",
  nn: "🇧🇻",
  no: "🇧🇻",
  nr: "🇿🇦",
  ny: "🇲🇼",
  oc: "🇪🇸",
  pa: "🇮🇳",
  pl: "🇵🇱",
  ps: "🇵🇰",
  pt: "🇵🇹",
  qu: "🇧🇴",
  ro: "🇲🇩",
  ru: "🇷🇺",
  rw: "🇷🇼",
  rn: "🇧🇮",
  sg: "🇨🇫",
  si: "🇱🇰",
  sk: "🇸🇰",
  sl: "🇸🇮",
  sm: "🇼🇸",
  sn: "🇿🇼",
  so: "🇸🇴",
  sq: "🇦🇱",
  sr: "🇷🇸",
  ss: "🇸🇿",
  st: "🇱🇸",
  sv: "🇸🇪",
  sw: "🇹🇿",
  ta: "🇮🇳",
  tg: "🇹🇯",
  th: "🇹🇭",
  ti: "🇪🇷",
  tk: "🇹🇲",
  tn: "🇹🇳",
  to: "🇹🇴",
  tr: "🇹🇷",
  ts: "🇿🇦",
  uk: "🇺🇦",
  ur: "🇵🇰",
  uz: "🇺🇿",
  ve: "🇿🇦",
  vi: "🇻🇳",
  xh: "🇿🇦",
  zh: "🇨🇳",
  zu: "🇿🇦",
};

const PIE_FILLS = [
  "hsl(var(--primary))",
  "hsl(var(--primary) / 0.55)",
  "hsl(var(--primary) / 0.3)",
  "hsl(var(--primary) / 0.15)",
];

export type LangSlice = { lang: string; flag: string; pct: number; fill: string };

export function normalizeLangs(
  raw: { language?: string; percentage?: number }[] | undefined,
  limit = 3,
): LangSlice[] {
  if (!raw?.length) return [];
  const sorted = raw
    .filter((l) => (l.percentage ?? 0) > 0)
    .sort((a, b) => (b.percentage ?? 0) - (a.percentage ?? 0));
  const total = sorted.reduce((s, l) => s + (l.percentage ?? 0), 0);
  if (total === 0) return [];
  const top = sorted.slice(0, limit);
  const topSum = top.reduce((s, l) => s + (l.percentage ?? 0), 0);
  const otherRaw = total - topSum;
  const result = top.map((l, i) => ({
    lang: l.language ?? "",
    flag: LANG_FLAGS[l.language ?? ""] ?? l.language ?? "",
    pct: Math.round(((l.percentage ?? 0) / total) * 100),
    fill: PIE_FILLS[i],
  }));
  if (otherRaw > 0) {
    const otherPct = 100 - result.reduce((s, l) => s + l.pct, 0);
    if (otherPct > 0) {
      result.push({ lang: "other", flag: "Other", pct: otherPct, fill: PIE_FILLS[3] });
    }
  }
  const rounding = 100 - result.reduce((s, l) => s + l.pct, 0);
  if (rounding !== 0 && result.length > 0) result[0].pct += rounding;
  return result;
}
