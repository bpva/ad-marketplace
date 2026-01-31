import { useState, useEffect } from "react";
import WebApp from "@twa-dev/sdk";
import { useTelegramTheme } from "@/hooks/useTelegramTheme";
import { useAuth } from "@/hooks/useAuth";
import { NotInTelegram } from "@/components/NotInTelegram";
import { Header } from "@/components/Header";
import { SettingsPage } from "@/components/SettingsPage";
import { OnboardingPage } from "@/components/OnboardingPage";
import { PublisherPage } from "@/components/PublisherPage";
import { Toaster } from "sonner";
import { fetchProfile, updateSettings, type Profile } from "@/lib/api";

function App() {
  const { user, loading } = useAuth();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [profileLoading, setProfileLoading] = useState(true);
  const [page, setPage] = useState<"main" | "settings">("main");

  useTelegramTheme(profile?.theme ?? "auto");

  const isInTelegram = WebApp.initData !== "" || import.meta.env.VITE_ENV === "local";

  useEffect(() => {
    if (user) {
      fetchProfile()
        .then(setProfile)
        .finally(() => setProfileLoading(false));
    }
  }, [user]);

  if (!isInTelegram) {
    return <NotInTelegram />;
  }

  if (loading || profileLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    );
  }

  const handleOnboardingComplete = async (mode: "publisher" | "advertiser") => {
    await updateSettings({ preferred_mode: mode, onboarding_finished: true });
    setProfile((prev) =>
      prev ? { ...prev, preferred_mode: mode, onboarding_finished: true } : null,
    );
  };

  if (profile && !profile.onboarding_finished) {
    return (
      <>
        <OnboardingPage onComplete={handleOnboardingComplete} />
        <Toaster position="bottom-center" richColors />
      </>
    );
  }

  return (
    <>
      {page === "settings" ? (
        <SettingsPage
          onBack={() => setPage("main")}
          onThemeChange={(t) => setProfile((p) => (p ? { ...p, theme: t } : null))}
        />
      ) : profile?.preferred_mode === "publisher" ? (
        <PublisherPage userName={user?.name} onSettingsClick={() => setPage("settings")} />
      ) : (
        <div className="min-h-screen flex flex-col bg-background">
          <Header userName={user?.name} onSettingsClick={() => setPage("settings")} />
          <main className="flex-1 flex items-center justify-center p-4">
            <div className="text-muted-foreground">Advertiser flow coming soon</div>
          </main>
        </div>
      )}
      <Toaster position="bottom-center" richColors />
    </>
  );
}

export default App;
