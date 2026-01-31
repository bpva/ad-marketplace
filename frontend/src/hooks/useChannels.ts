import { useState, useEffect, useCallback } from "react";
import { fetchChannels, type ChannelWithRole } from "@/lib/api";

interface UseChannelsResult {
  channels: ChannelWithRole[];
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export function useChannels(): UseChannelsResult {
  const [channels, setChannels] = useState<ChannelWithRole[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(() => {
    setLoading(true);
    setError(null);
    fetchChannels()
      .then((res) => setChannels(res.channels ?? []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { channels, loading, error, refetch: fetch };
}
