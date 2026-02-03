import { useState, useEffect } from "react";
import WebApp from "@twa-dev/sdk";
import { useTelegramTheme } from "@/hooks/useTelegramTheme";
import { useAuth } from "@/hooks/useAuth";
import { NotInTelegram } from "@/components/NotInTelegram";
import { SettingsPage } from "@/components/SettingsPage";
import { OnboardingPage } from "@/components/OnboardingPage";
import { PublisherPage } from "@/components/PublisherPage";
import { MarketplacePage } from "@/components/MarketplacePage";
import { DealsPage } from "@/components/DealsPage";
import { BottomNav, type NavPage } from "@/components/BottomNav";
import { Toaster } from "sonner";
import { fetchProfile, updateSettings, type Profile } from "@/lib/api";
import { Target } from "lucide-react";

type Page = "channels" | "marketplace" | "deals" | "settings";

function App() {
  const { user, loading } = useAuth();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [profileLoading, setProfileLoading] = useState(true);
  const [page, setPage] = useState<Page>("channels");

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

  const handleNavigation = (navPage: NavPage) => {
    if (navPage === "profile") {
      setPage("settings");
    } else {
      setPage(navPage);
    }
  };

  const activeNavPage: NavPage = page === "settings" ? "profile" : page;
  const showNav = page !== "settings";

  const renderPage = () => {
    switch (page) {
      case "channels":
        return profile?.preferred_mode === "publisher" ? (
          <PublisherPage />
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center p-4">
            <div className="text-center">
              <div className="w-16 h-16 rounded-2xl bg-muted flex items-center justify-center mx-auto mb-4">
                <Target className="h-8 w-8 text-muted-foreground" />
              </div>
              <h1 className="text-xl font-semibold mb-2">Campaigns</h1>
              <p className="text-muted-foreground text-sm">Coming soon</p>
            </div>
          </div>
        );
      case "marketplace":
        return <MarketplacePage />;
      case "deals":
        return <DealsPage />;
      case "settings":
        return (
          <SettingsPage
            onBack={() => setPage("channels")}
            onThemeChange={(t) => setProfile((p) => (p ? { ...p, theme: t } : null))}
            onModeChange={(m) => setProfile((p) => (p ? { ...p, preferred_mode: m } : null))}
          />
        );
    }
  };

  return (
    <>
      <div className="min-h-screen flex flex-col bg-background">
        <main
          className="flex-1 flex flex-col"
          style={{
            paddingTop: "var(--total-safe-area-top, 0px)",
            paddingBottom: showNav ? "calc(64px + var(--safe-area-inset-bottom, 0px))" : 0,
          }}
        >
          {renderPage()}
        </main>
        {showNav && profile?.preferred_mode && (
          <BottomNav
            mode={profile.preferred_mode}
            activePage={activeNavPage}
            onNavigate={handleNavigation}
            userName={user?.name}
            telegramId={user?.telegram_id}
            photoUrl={WebApp.initDataUnsafe?.user?.photo_url}
          />
        )}
      </div>
      <Toaster position="bottom-center" richColors />
    </>
  );
}

export default App;
