import { useState } from "react";
import { TemplateList } from "@/components/TemplateList";
import { TelegramMessagePreview } from "@/components/TelegramMessagePreview";
import { useTemplates } from "@/hooks/useTemplates";
import type { TemplateResponse } from "@/lib/api";
import { FileText } from "lucide-react";

export function TemplatesPage() {
  const { templates, loading, error, refetch } = useTemplates();
  const [selectedTemplate, setSelectedTemplate] = useState<TemplateResponse | null>(null);

  if (selectedTemplate) {
    return (
      <TelegramMessagePreview
        template={selectedTemplate}
        onBack={() => {
          setSelectedTemplate(null);
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
        <div className="text-destructive">Failed to load templates</div>
        <button onClick={refetch} className="text-sm text-primary hover:underline">
          Try again
        </button>
      </div>
    );
  }

  if (templates.length === 0) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-4">
        <div className="text-center">
          <div className="w-16 h-16 rounded-2xl bg-muted flex items-center justify-center mx-auto mb-4">
            <FileText className="h-8 w-8 text-muted-foreground" />
          </div>
          <h2 className="text-lg font-semibold mb-2">No templates yet</h2>
          <p className="text-muted-foreground text-sm">Use /add_promo in the bot to save posts</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col">
      <TemplateList templates={templates} onTemplateClick={setSelectedTemplate} />
    </div>
  );
}
