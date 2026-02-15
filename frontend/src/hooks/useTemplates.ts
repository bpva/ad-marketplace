import { useState, useEffect, useCallback } from "react";
import { fetchTemplates, type TemplateResponse } from "@/lib/api";

interface UseTemplatesResult {
  templates: TemplateResponse[];
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export function useTemplates(): UseTemplatesResult {
  const [templates, setTemplates] = useState<TemplateResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(() => {
    setLoading(true);
    setError(null);
    fetchTemplates()
      .then((res) => setTemplates(res.templates ?? []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { templates, loading, error, refetch: fetch };
}
