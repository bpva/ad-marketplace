import { FileText, Sparkles, Repeat2, CircleDot, type LucideIcon } from "lucide-react";
import type { AdFormatType } from "@/lib/api";

export interface FormatConfig {
  key: string;
  label: string;
  icon: LucideIcon;
  description: string;
  formatType: AdFormatType;
  isNative: boolean;
  enabled: boolean;
  color: string;
  bgColor: string;
}

export const FORMAT_CONFIGS: FormatConfig[] = [
  {
    key: "post",
    label: "Post",
    icon: FileText,
    description: "Standard ad placed in the channel",
    formatType: "post",
    isNative: false,
    enabled: true,
    color: "text-primary",
    bgColor: "bg-primary/10",
  },
  {
    key: "native-post",
    label: "Native Post",
    icon: Sparkles,
    description: "Ad styled to match your channel's tone",
    formatType: "post",
    isNative: true,
    enabled: true,
    color: "text-green-500",
    bgColor: "bg-green-500/10",
  },
  {
    key: "repost",
    label: "Repost",
    icon: Repeat2,
    description: "Forward a post to your channel",
    formatType: "repost",
    isNative: false,
    enabled: false,
    color: "text-secondary",
    bgColor: "bg-secondary/10",
  },
  {
    key: "story",
    label: "Story",
    icon: CircleDot,
    description: "Temporary story visible for 24 hours",
    formatType: "story",
    isNative: false,
    enabled: false,
    color: "text-amber-500",
    bgColor: "bg-amber-500/10",
  },
];

export function getFormatDisplay(
  formatType: AdFormatType | undefined,
  isNative: boolean | undefined,
) {
  const config = FORMAT_CONFIGS.find(
    (c) => c.formatType === formatType && c.isNative === (isNative ?? false),
  );
  return config ?? FORMAT_CONFIGS[0];
}
