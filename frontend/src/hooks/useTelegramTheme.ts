import { useEffect } from "react";
import WebApp from "@twa-dev/sdk";

type ThemePreference = "light" | "dark" | "auto";

export function useTelegramTheme(preference: ThemePreference = "auto") {
  useEffect(() => {
    const root = document.documentElement;

    const applyTheme = () => {
      let isDark: boolean;
      if (preference === "auto") {
        isDark = WebApp.colorScheme === "dark";
      } else {
        isDark = preference === "dark";
      }
      root.classList.toggle("dark", isDark);
    };

    const applySafeArea = () => {
      const safe = WebApp.safeAreaInset;
      const content = WebApp.contentSafeAreaInset;

      root.style.setProperty("--safe-area-inset-top", `${safe.top}px`);
      root.style.setProperty("--safe-area-inset-bottom", `${safe.bottom}px`);
      root.style.setProperty("--safe-area-inset-left", `${safe.left}px`);
      root.style.setProperty("--safe-area-inset-right", `${safe.right}px`);
      root.style.setProperty("--content-safe-area-inset-top", `${content.top}px`);
      root.style.setProperty("--content-safe-area-inset-bottom", `${content.bottom}px`);

      const totalTop = safe.top + content.top;
      root.style.setProperty("--total-safe-area-top", `${totalTop}px`);
    };

    applyTheme();
    applySafeArea();

    if (preference === "auto") {
      WebApp.onEvent("themeChanged", applyTheme);
    }
    WebApp.onEvent("safeAreaChanged", applySafeArea);
    WebApp.onEvent("contentSafeAreaChanged", applySafeArea);
    WebApp.ready();
    WebApp.expand();

    return () => {
      if (preference === "auto") {
        WebApp.offEvent("themeChanged", applyTheme);
      }
      WebApp.offEvent("safeAreaChanged", applySafeArea);
      WebApp.offEvent("contentSafeAreaChanged", applySafeArea);
    };
  }, [preference]);
}
