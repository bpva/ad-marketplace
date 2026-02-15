import { useState, useEffect } from "react";
import { type TonRates, fetchTonRates } from "@/lib/api";

let cached: TonRates | null = null;
let inflight: Promise<TonRates> | null = null;

export function useTonRates(): TonRates | null {
  const [rates, setRates] = useState<TonRates | null>(cached);

  useEffect(() => {
    if (cached) return;
    if (!inflight) {
      inflight = fetchTonRates().catch((err) => {
        inflight = null;
        throw err;
      });
    }
    inflight.then((r) => {
      cached = r;
      setRates(r);
    });
  }, []);

  return rates;
}
