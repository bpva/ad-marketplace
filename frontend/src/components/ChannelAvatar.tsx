import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";
import { getChannelAvatarUrl } from "@/lib/avatar";
import { fetchChannelPhotoBlob } from "@/lib/api";

interface ChannelAvatarProps {
  channelId: number;
  hasPhoto?: boolean;
  className?: string;
}

export function ChannelAvatar({ channelId, hasPhoto, className }: ChannelAvatarProps) {
  const [src, setSrc] = useState(() => getChannelAvatarUrl(channelId));

  useEffect(() => {
    if (!hasPhoto) return;
    let revoke: string | undefined;
    fetchChannelPhotoBlob(channelId)
      .then((url) => {
        revoke = url;
        setSrc(url);
      })
      .catch(() => {});
    return () => {
      if (revoke) URL.revokeObjectURL(revoke);
    };
  }, [channelId, hasPhoto]);

  return <img src={src} alt="" className={cn("rounded-full object-cover", className)} />;
}
