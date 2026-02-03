import { useState, useEffect, useRef } from "react";
import { toast } from "sonner";
import { Check, Megaphone, Target, Sun, Moon, Monitor } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";
import {
  fetchProfile,
  updateName,
  updateSettings,
  type Profile,
  type Theme,
  type PreferredMode,
  type Language,
} from "@/lib/api";

const isPublisher = (mode: PreferredMode | undefined) => mode === "publisher";

interface SettingsPageProps {
  onBack: () => void;
  onThemeChange: (theme: Theme) => void;
}

export function SettingsPage({ onBack, onThemeChange }: SettingsPageProps) {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [loading, setLoading] = useState(true);
  const [nameValue, setNameValue] = useState("");
  const [nameFocused, setNameFocused] = useState(false);
  const [saving, setSaving] = useState(false);
  const lastSavedName = useRef("");

  useEffect(() => {
    fetchProfile()
      .then((data) => {
        setProfile(data);
        setNameValue(data.name ?? "");
        lastSavedName.current = data.name ?? "";
      })
      .finally(() => setLoading(false));
  }, []);

  const saveName = async (name: string) => {
    const trimmed = name.trim();
    if (!trimmed || trimmed === lastSavedName.current) return;
    setSaving(true);
    try {
      await updateName(trimmed);
      setProfile((prev) => (prev ? { ...prev, name: trimmed } : null));
      lastSavedName.current = trimmed;
      toast("Name updated");
    } catch {
      toast("Failed to save name");
    } finally {
      setSaving(false);
    }
  };

  const handleNameChange = (value: string) => {
    setNameValue(value);
  };

  const handleNameFocus = () => {
    setNameFocused(true);
  };

  const handleNameBlur = () => {
    setNameFocused(false);
    saveName(nameValue);
  };

  const handleSettingChange = async (update: Parameters<typeof updateSettings>[0]) => {
    setSaving(true);
    try {
      await updateSettings(update);
      setProfile((prev) => (prev ? { ...prev, ...update } : null));
      toast("Settings saved");
    } catch {
      toast("Failed to save settings");
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-destructive">Failed to load profile</div>
      </div>
    );
  }

  return (
    <div
      className="min-h-screen bg-background p-4"
      style={{ paddingTop: "calc(var(--total-safe-area-top, 0px) + 1rem)" }}
    >
      <div className="max-w-md mx-auto space-y-6">
        <Button variant="ghost" onClick={onBack}>
          ← Back
        </Button>

        <div className="bg-card rounded-lg border border-border p-4 space-y-4">
          <h2 className="text-lg font-semibold">Profile</h2>
          <div className="space-y-2">
            <Label className="text-muted-foreground">Name</Label>
            <div className="flex items-center gap-2">
              <Input
                value={nameValue}
                onChange={(e) => handleNameChange(e.target.value)}
                onFocus={handleNameFocus}
                onBlur={handleNameBlur}
                disabled={saving}
              />
              {nameFocused && (
                <Button
                  variant="ghost"
                  size="icon"
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => saveName(nameValue)}
                >
                  <Check className="h-4 w-4" />
                </Button>
              )}
            </div>
          </div>
        </div>

        <div className="bg-card rounded-lg border border-border p-4 space-y-4">
          <h2 className="text-lg font-semibold">Settings</h2>

          <div className="relative flex rounded-lg bg-muted p-1">
            <div
              className={cn(
                "absolute inset-y-1 w-[calc(50%-4px)] rounded-md border-2 border-primary bg-background shadow-sm transition-all duration-200 ease-out",
                isPublisher(profile.preferred_mode) ? "left-1" : "left-[calc(50%+2px)]",
              )}
            />
            <button
              type="button"
              onClick={() => handleSettingChange({ preferred_mode: "publisher" })}
              disabled={saving}
              className={cn(
                "relative z-10 flex flex-1 items-center justify-center gap-2 rounded-md py-2.5 text-sm font-medium transition-colors disabled:opacity-50",
                isPublisher(profile.preferred_mode)
                  ? "text-primary"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              <Megaphone className="h-4 w-4" />
              Publisher
            </button>
            <button
              type="button"
              onClick={() => handleSettingChange({ preferred_mode: "advertiser" })}
              disabled={saving}
              className={cn(
                "relative z-10 flex flex-1 items-center justify-center gap-2 rounded-md py-2.5 text-sm font-medium transition-colors disabled:opacity-50",
                !isPublisher(profile.preferred_mode)
                  ? "text-primary"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              <Target className="h-4 w-4" />
              Advertiser
            </button>
          </div>

          <div className="space-y-2">
            <Label className="text-muted-foreground">Theme</Label>
            <div className="relative flex rounded-lg bg-muted p-1">
              <div
                className={cn(
                  "absolute inset-y-1 w-[calc(33.333%-4px)] rounded-md border-2 border-primary bg-background shadow-sm transition-all duration-200 ease-out",
                  profile.theme === "auto" && "left-1",
                  profile.theme === "light" && "left-[calc(33.333%+1px)]",
                  profile.theme === "dark" && "left-[calc(66.666%+2px)]",
                )}
              />
              {(["auto", "light", "dark"] satisfies Theme[]).map((theme) => (
                <button
                  key={theme}
                  type="button"
                  onClick={() => {
                    handleSettingChange({ theme });
                    onThemeChange(theme);
                  }}
                  disabled={saving}
                  className={cn(
                    "relative z-10 flex flex-1 items-center justify-center gap-1.5 rounded-md py-2 text-sm font-medium transition-colors disabled:opacity-50",
                    profile.theme === theme
                      ? "text-primary"
                      : "text-muted-foreground hover:text-foreground",
                  )}
                >
                  {theme === "auto" && <Monitor className="h-4 w-4" />}
                  {theme === "light" && <Sun className="h-4 w-4" />}
                  {theme === "dark" && <Moon className="h-4 w-4" />}
                  {theme.charAt(0).toUpperCase() + theme.slice(1)}
                </button>
              ))}
            </div>
          </div>

          <div className="flex items-center justify-between">
            <Label className="text-muted-foreground">Language</Label>
            <select
              value={profile.language}
              onChange={(e) => handleSettingChange({ language: e.target.value as Language })}
              disabled={saving}
              className="rounded-md border border-input bg-background px-3 py-1.5 text-sm disabled:opacity-50"
            >
              <option value="en">English</option>
              <option value="ru">Русский</option>
            </select>
          </div>

          <div className="flex items-center justify-between">
            <Label className="text-muted-foreground">Notifications</Label>
            <Switch
              checked={profile.receive_notifications}
              onCheckedChange={(checked) => handleSettingChange({ receive_notifications: checked })}
              disabled={saving}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
