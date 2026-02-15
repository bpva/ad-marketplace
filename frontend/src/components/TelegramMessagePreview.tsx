import { useState, useEffect, type ReactNode } from "react";
import { ArrowLeft, Eye, Megaphone, Send, Check } from "lucide-react";
import type { TemplateResponse } from "@/lib/api";
import { fetchPostMediaBlob, sendTemplatePreview } from "@/lib/api";

interface TelegramMessagePreviewProps {
  template: TemplateResponse;
  onBack: () => void;
}

export function TelegramMessagePreview({ template, onBack }: TelegramMessagePreviewProps) {
  const [sending, setSending] = useState(false);
  const [sent, setSent] = useState(false);

  const handleSendPreview = async () => {
    if (!template.id || sending || sent) return;
    setSending(true);
    try {
      await sendTemplatePreview(template.id);
      setSent(true);
      setTimeout(() => setSent(false), 3000);
    } finally {
      setSending(false);
    }
  };

  const channelContent = (
    <div className="flex flex-col h-full bg-muted/30">
      <ChannelHeader onBack={onBack} />
      <div className="flex-1 overflow-y-auto p-3 pt-4">
        <ChannelPost template={template} />
      </div>
    </div>
  );

  return (
    <div className="flex-1 flex flex-col items-center justify-center p-4">
      <div className="sm:hidden w-full max-w-md rounded-2xl border border-border overflow-hidden max-h-[80vh]">
        {channelContent}
      </div>

      <div className="hidden sm:block">
        <div className="mockup-phone">
          <div className="mockup-phone-camera"></div>
          <div className="mockup-phone-display">{channelContent}</div>
        </div>
      </div>

      <button
        type="button"
        onClick={handleSendPreview}
        disabled={sending || sent}
        className="mt-4 flex items-center gap-2 px-4 py-2.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors disabled:opacity-60"
      >
        {sent ? <Check className="h-4 w-4" /> : <Send className="h-4 w-4" />}
        {sent ? "Sent!" : sending ? "Sending..." : "Preview in Chat"}
      </button>
    </div>
  );
}

function ChannelHeader({ onBack }: { onBack: () => void }) {
  return (
    <div className="sticky top-0 z-10 flex items-center gap-2 px-4 pt-3 pb-2 bg-card/95 backdrop-blur-sm border-b border-border">
      <button
        type="button"
        onClick={onBack}
        className="p-1 -ml-1 rounded-lg hover:bg-accent transition-colors"
      >
        <ArrowLeft className="h-4 w-4" />
      </button>
      <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center flex-shrink-0">
        <Megaphone className="h-4 w-4 text-primary" />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold truncate leading-tight">Ad Preview</p>
        <p className="text-xs text-muted-foreground leading-tight">channel post</p>
      </div>
    </div>
  );
}

function ChannelPost({ template }: { template: TemplateResponse }) {
  const hasMedia = template.media && template.media.length > 0;
  const showCaptionAbove = hasMedia && template.media!.some((m) => m.show_caption_above_media);

  const caption = template.text ? (
    <div className="px-3 py-2">
      <RenderEntities text={template.text} entities={template.entities} />
    </div>
  ) : null;

  return (
    <div className="bg-card rounded-xl overflow-hidden shadow-sm">
      {showCaptionAbove && caption}
      {hasMedia && <MediaGallery template={template} />}
      {!showCaptionAbove && caption}
      <PostFooter />
    </div>
  );
}

function PostFooter() {
  return (
    <div className="flex items-center justify-end gap-1.5 px-3 pb-2 pt-0">
      <Eye className="h-3 w-3 text-muted-foreground" />
      <span className="text-[11px] text-muted-foreground">4.2K</span>
      <span className="text-[11px] text-muted-foreground">14:32</span>
    </div>
  );
}

interface TgEntity {
  type: string;
  offset: number;
  length: number;
  url?: string;
  language?: string;
}

function RenderEntities({ text, entities }: { text: string; entities?: unknown }) {
  if (!entities || !Array.isArray(entities) || entities.length === 0) {
    return <span className="whitespace-pre-wrap break-words">{text}</span>;
  }

  const sorted = [...(entities as TgEntity[])].sort((a, b) => a.offset - b.offset);
  const parts: ReactNode[] = [];
  let cursor = 0;

  for (let i = 0; i < sorted.length; i++) {
    const e = sorted[i];
    if (e.offset > cursor) {
      parts.push(text.substring(cursor, e.offset));
    }

    const segment = text.substring(e.offset, e.offset + e.length);
    parts.push(
      <EntitySpan key={i} entity={e}>
        {segment}
      </EntitySpan>,
    );
    cursor = e.offset + e.length;
  }

  if (cursor < text.length) {
    parts.push(text.substring(cursor));
  }

  return <span className="whitespace-pre-wrap break-words">{parts}</span>;
}

function EntitySpan({ entity, children }: { entity: TgEntity; children: ReactNode }) {
  switch (entity.type) {
    case "bold":
      return <strong>{children}</strong>;
    case "italic":
      return <em>{children}</em>;
    case "underline":
      return <u>{children}</u>;
    case "strikethrough":
      return <s>{children}</s>;
    case "code":
      return <code className="px-1 py-0.5 bg-muted rounded text-sm font-mono">{children}</code>;
    case "pre":
      return (
        <pre className="bg-muted rounded-lg p-3 my-1 overflow-x-auto text-sm font-mono">
          {entity.language && (
            <div className="text-xs text-muted-foreground mb-1">{entity.language}</div>
          )}
          <code>{children}</code>
        </pre>
      );
    case "text_link":
      return (
        <a
          href={entity.url}
          target="_blank"
          rel="noopener noreferrer"
          className="text-primary underline"
        >
          {children}
        </a>
      );
    case "url":
      return (
        <a
          href={String(children)}
          target="_blank"
          rel="noopener noreferrer"
          className="text-primary underline"
        >
          {children}
        </a>
      );
    case "mention":
      return <span className="text-primary">{children}</span>;
    case "spoiler":
      return <SpoilerSpan>{children}</SpoilerSpan>;
    case "blockquote":
      return <blockquote className="border-l-3 border-primary/40 pl-2 my-1">{children}</blockquote>;
    default:
      return <>{children}</>;
  }
}

function SpoilerSpan({ children }: { children: ReactNode }) {
  const [revealed, setRevealed] = useState(false);

  return (
    <span
      onClick={() => setRevealed(true)}
      className={
        revealed ? "" : "bg-muted-foreground text-transparent rounded cursor-pointer select-none"
      }
    >
      {children}
    </span>
  );
}

function MediaGallery({ template }: { template: TemplateResponse }) {
  const media = template.media ?? [];
  if (media.length === 0) return null;

  if (media.length === 1) {
    return (
      <div className="w-full">
        <MediaItem postID={media[0].post_id ?? ""} />
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-0.5">
      {media.map((m, i) => {
        const isFirst = i === 0 && media.length % 2 !== 0;
        return (
          <div key={m.post_id} className={isFirst ? "col-span-2" : ""}>
            <MediaItem postID={m.post_id ?? ""} />
          </div>
        );
      })}
    </div>
  );
}

function MediaItem({ postID }: { postID: string }) {
  const [blobUrl, setBlobUrl] = useState<string | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    let cancelled = false;
    fetchPostMediaBlob(postID)
      .then((url) => {
        if (!cancelled) setBlobUrl(url);
      })
      .catch(() => {
        if (!cancelled) setError(true);
      });
    return () => {
      cancelled = true;
      if (blobUrl) URL.revokeObjectURL(blobUrl);
    };
  }, [postID]);

  if (error) {
    return (
      <div className="aspect-square bg-muted flex items-center justify-center text-xs text-muted-foreground">
        Failed to load
      </div>
    );
  }

  if (!blobUrl) {
    return <div className="aspect-square bg-muted animate-pulse" />;
  }

  return <img src={blobUrl} alt="" className="w-full aspect-square object-cover" />;
}
