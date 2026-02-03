import { useState, useEffect } from "react";
import WebApp from "@twa-dev/sdk";
import { authenticate, setToken, type User } from "@/lib/api";
import { generateMockInitData } from "@/lib/mockTelegram";

interface AuthState {
  user: User | null;
  loading: boolean;
  error: string | null;
}

async function getInitData(): Promise<string> {
  if (import.meta.env.VITE_ENV === "local") {
    return generateMockInitData(import.meta.env.VITE_BOT_TOKEN);
  }
  return WebApp.initData;
}

export function useAuth(): AuthState {
  const [state, setState] = useState<AuthState>({
    user: null,
    loading: true,
    error: null,
  });

  useEffect(() => {
    getInitData().then((initData) => {
      if (!initData) {
        setState({ user: null, loading: false, error: "Not in Telegram" });
        return;
      }

      authenticate(initData)
        .then((res) => {
          if (res.token) {
            setToken(res.token);
          }
          setState({ user: res.user ?? null, loading: false, error: null });
        })
        .catch((err) => {
          setState({ user: null, loading: false, error: err.message });
        });
    });
  }, []);

  return state;
}
