import { useState, useEffect } from "react";
import WebApp from "@twa-dev/sdk";
import { useTelegramTheme } from "@/hooks/useTelegramTheme";
import { useAuth } from "@/hooks/useAuth";
import { NotInTelegram } from "@/components/NotInTelegram";
import { Header } from "@/components/Header";
import { SettingsPage } from "@/components/SettingsPage";
import { OnboardingPage } from "@/components/OnboardingPage";
import { Button } from "@/components/ui/button";
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
      ) : (
        <div className="min-h-screen flex flex-col bg-background">
          <Header
            userName={user?.name}
            telegramId={user?.telegram_id}
            photoUrl={WebApp.initDataUnsafe?.user?.photo_url}
            onSettingsClick={() => setPage("settings")}
          />
          <main className="flex-1 flex flex-col items-center justify-center gap-4 p-4">
            <div className="p-8 rounded-xl bg-card text-card-foreground border border-border">
              <h1 className="text-2xl font-bold">Welcome</h1>
            </div>
            <Button>Get Started</Button>
          </main>
        </div>
      )}
      <Toaster position="bottom-center" richColors />
    </>
  );
}

export default App;
