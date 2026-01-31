import { Header } from "@/components/Header";
import { ChannelEmptyState } from "@/components/ChannelEmptyState";
import { ChannelList } from "@/components/ChannelList";
import { useChannels } from "@/hooks/useChannels";

interface PublisherPageProps {
  userName?: string;
  onSettingsClick: () => void;
}

export function PublisherPage({ userName, onSettingsClick }: PublisherPageProps) {
  const { channels, loading, error, refetch } = useChannels();

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4 bg-background p-4">
        <div className="text-destructive">Failed to load channels</div>
        <button onClick={refetch} className="text-sm text-primary hover:underline">
          Try again
        </button>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col bg-background">
      <Header userName={userName} onSettingsClick={onSettingsClick} />
      <main className="flex-1 flex flex-col">
        {channels.length === 0 ? <ChannelEmptyState /> : <ChannelList channels={channels} />}
      </main>
    </div>
  );
}
