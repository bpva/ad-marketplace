import { useState } from "react";
import { cn } from "@/lib/utils";
import { getChannelAvatarUrl } from "@/lib/avatar";

interface ChannelAvatarProps {
  channelId: number;
  photoUrl?: string;
  className?: string;
}

export function ChannelAvatar({ channelId, photoUrl, className }: ChannelAvatarProps) {
  const [error, setError] = useState(false);
  const src = !error && photoUrl ? photoUrl : getChannelAvatarUrl(channelId);

  return (
    <img
      src={src}
      alt=""
      onError={() => setError(true)}
      className={cn("rounded-full object-cover", className)}
    />
  );
}
