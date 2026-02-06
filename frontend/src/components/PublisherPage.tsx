import { useState } from "react";
import { ChannelEmptyState } from "@/components/ChannelEmptyState";
import { ChannelList } from "@/components/ChannelList";
import { ChannelDetailPage } from "@/components/ChannelDetailPage";
import { useChannels } from "@/hooks/useChannels";
import type { ChannelWithRole } from "@/lib/api";

export function PublisherPage() {
  const { channels, loading, error, refetch } = useChannels();
  const [selectedChannel, setSelectedChannel] = useState<ChannelWithRole | null>(null);

  if (selectedChannel) {
    return (
      <ChannelDetailPage
        channel={selectedChannel}
        onBack={() => {
          setSelectedChannel(null);
          refetch();
        }}
      />
    );
  }

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center gap-4 p-4">
        <div className="text-destructive">Failed to load channels</div>
        <button onClick={refetch} className="text-sm text-primary hover:underline">
          Try again
        </button>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col">
      {channels.length === 0 ? (
        <ChannelEmptyState />
      ) : (
        <ChannelList channels={channels} onChannelClick={setSelectedChannel} />
      )}
    </div>
  );
}
