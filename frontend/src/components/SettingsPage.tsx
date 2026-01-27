import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { fetchProfile, updateName, updateSettings, type Profile } from '@/lib/api'

interface SettingsPageProps {
  onBack: () => void
}

export function SettingsPage({ onBack }: SettingsPageProps) {
  const [profile, setProfile] = useState<Profile | null>(null)
  const [loading, setLoading] = useState(true)
  const [editingName, setEditingName] = useState(false)
  const [nameValue, setNameValue] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    fetchProfile()
      .then((data) => {
        setProfile(data)
        setNameValue(data.name)
      })
      .finally(() => setLoading(false))
  }, [])

  const handleNameSave = async () => {
    if (!nameValue.trim()) return
    setSaving(true)
    try {
      await updateName(nameValue.trim())
      setProfile((prev) => (prev ? { ...prev, name: nameValue.trim() } : null))
      setEditingName(false)
    } finally {
      setSaving(false)
    }
  }

  const handleSettingChange = async (update: Parameters<typeof updateSettings>[0]) => {
    setSaving(true)
    try {
      await updateSettings(update)
      setProfile((prev) => (prev ? { ...prev, ...update } : null))
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    )
  }

  if (!profile) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-destructive">Failed to load profile</div>
      </div>
    )
  }

  return (
    <div
      className="min-h-screen bg-background p-4"
      style={{ paddingTop: 'calc(var(--total-safe-area-top, 0px) + 1rem)' }}
    >
      <div className="max-w-md mx-auto space-y-6">
        <Button variant="ghost" onClick={onBack}>
          ← Back
        </Button>

        <div className="bg-card rounded-lg border border-border p-4 space-y-4">
          <h2 className="text-lg font-semibold">Profile</h2>

          <div className="space-y-2">
            <label className="text-sm text-muted-foreground">Name</label>
            {editingName ? (
              <div className="flex gap-2">
                <input
                  type="text"
                  value={nameValue}
                  onChange={(e) => setNameValue(e.target.value)}
                  className="flex-1 px-3 py-2 rounded-md border border-input bg-background text-foreground"
                  autoFocus
                />
                <Button onClick={handleNameSave} disabled={saving} size="sm">
                  Save
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    setEditingName(false)
                    setNameValue(profile.name)
                  }}
                >
                  Cancel
                </Button>
              </div>
            ) : (
              <div
                className="px-3 py-2 rounded-md border border-input bg-background cursor-pointer hover:bg-accent"
                onClick={() => setEditingName(true)}
              >
                {profile.name}
              </div>
            )}
          </div>
        </div>

        <div className="bg-card rounded-lg border border-border p-4 space-y-4">
          <h2 className="text-lg font-semibold">Settings</h2>

          <div className="space-y-2">
            <label className="text-sm text-muted-foreground">Language</label>
            <div className="flex gap-2">
              <Button
                variant={profile.language === 'en' ? 'default' : 'outline'}
                onClick={() => handleSettingChange({ language: 'en' })}
                disabled={saving}
                size="sm"
              >
                English
              </Button>
              <Button
                variant={profile.language === 'ru' ? 'default' : 'outline'}
                onClick={() => handleSettingChange({ language: 'ru' })}
                disabled={saving}
                size="sm"
              >
                Русский
              </Button>
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm text-muted-foreground">Mode</label>
            <div className="flex gap-2">
              <Button
                variant={profile.preferred_mode === 'publisher' ? 'default' : 'outline'}
                onClick={() => handleSettingChange({ preferred_mode: 'publisher' })}
                disabled={saving}
                size="sm"
              >
                Publisher
              </Button>
              <Button
                variant={profile.preferred_mode === 'advertiser' ? 'default' : 'outline'}
                onClick={() => handleSettingChange({ preferred_mode: 'advertiser' })}
                disabled={saving}
                size="sm"
              >
                Advertiser
              </Button>
            </div>
          </div>

          <div className="flex items-center justify-between py-2">
            <label className="text-sm text-muted-foreground">Receive notifications</label>
            <input
              type="checkbox"
              checked={profile.receive_notifications}
              onChange={(e) =>
                handleSettingChange({ receive_notifications: e.target.checked })
              }
              disabled={saving}
              className="w-5 h-5 accent-primary"
            />
          </div>
        </div>
      </div>
    </div>
  )
}
