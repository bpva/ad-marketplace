import { ChevronRight, Image, Video, FileText, Music, Mic, Film, Type } from "lucide-react";
import type { TemplateResponse } from "@/lib/api";

interface TemplateListProps {
  templates: TemplateResponse[];
  onTemplateClick?: (template: TemplateResponse) => void;
}

export function TemplateList({ templates, onTemplateClick }: TemplateListProps) {
  return (
    <div className="p-4">
      <div className="max-w-md mx-auto space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Your Templates</h2>
          <span className="text-sm text-muted-foreground">{templates.length}</span>
        </div>

        <div className="space-y-3">
          {templates.map((template) => (
            <TemplateCard
              key={template.id}
              template={template}
              onClick={() => onTemplateClick?.(template)}
            />
          ))}
        </div>

        <p className="text-xs text-muted-foreground text-center pt-4">
          Save posts via /add_promo in the bot
        </p>
      </div>
    </div>
  );
}

function MediaIcon({ mediaType }: { mediaType: string }) {
  const className = "h-5 w-5 text-muted-foreground";
  switch (mediaType) {
    case "photo":
      return <Image className={className} />;
    case "video":
      return <Video className={className} />;
    case "animation":
      return <Film className={className} />;
    case "audio":
      return <Music className={className} />;
    case "voice":
    case "video_note":
      return <Mic className={className} />;
    case "document":
      return <FileText className={className} />;
    default:
      return <FileText className={className} />;
  }
}

function formatRelativeDate(date: string): string {
  const now = new Date();
  const d = new Date(date);
  const diff = now.getTime() - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  return d.toLocaleDateString();
}

function TemplateCard({ template, onClick }: { template: TemplateResponse; onClick: () => void }) {
  const hasMedia = template.media && template.media.length > 0;
  const primaryMediaType = hasMedia ? template.media![0].media_type : null;
  const textPreview = template.text
    ? template.text.length > 80
      ? template.text.slice(0, 80) + "..."
      : template.text
    : hasMedia
      ? `${primaryMediaType} file`
      : "Empty template";

  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full bg-card rounded-xl border border-border p-4 flex items-center gap-3 text-left transition-colors hover:bg-accent/50 active:bg-accent"
    >
      <div className="w-10 h-10 rounded-lg bg-muted flex items-center justify-center flex-shrink-0">
        {hasMedia ? (
          <MediaIcon mediaType={primaryMediaType!} />
        ) : (
          <Type className="h-5 w-5 text-muted-foreground" />
        )}
      </div>

      <div className="flex-1 min-w-0">
        <p className="text-sm truncate">{textPreview}</p>
        <div className="flex items-center gap-2 mt-1">
          <span className="text-xs text-muted-foreground">
            {formatRelativeDate(template.created_at)}
          </span>
          {hasMedia && template.media!.length > 1 && (
            <span className="px-1.5 py-0.5 rounded text-xs bg-muted text-muted-foreground">
              {template.media!.length} files
            </span>
          )}
        </div>
      </div>

      <ChevronRight className="h-5 w-5 text-muted-foreground flex-shrink-0" />
    </button>
  );
}
