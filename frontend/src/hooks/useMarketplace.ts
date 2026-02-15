import { useState, useEffect, useCallback, useRef } from "react";
import {
  fetchMarketplaceChannels,
  type MarketplaceChannel,
  type MarketplaceFilter,
  type ChannelSortBy,
  type SortOrder,
} from "@/lib/api";

interface UseMarketplaceResult {
  channels: MarketplaceChannel[];
  total: number;
  loading: boolean;
  search: string;
  setSearch: (v: string) => void;
  sortBy: ChannelSortBy;
  setSortBy: (v: ChannelSortBy) => void;
  sortOrder: SortOrder;
  setSortOrder: (v: SortOrder) => void;
  page: number;
  setPage: (v: number) => void;
  selectedCategories: string[];
  setSelectedCategories: (v: string[]) => void;
}

export function useMarketplace(): UseMarketplaceResult {
  const [channels, setChannels] = useState<MarketplaceChannel[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [sortBy, setSortBy] = useState<ChannelSortBy>("subscribers");
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc");
  const [page, setPage] = useState(1);
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      setDebouncedSearch(search);
      setPage(1);
    }, 300);
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [search]);

  useEffect(() => {
    setPage(1);
  }, [selectedCategories]);

  const load = useCallback(() => {
    setLoading(true);
    const filters: MarketplaceFilter[] = [];
    if (debouncedSearch) {
      filters.push({ name: "fulltext", value: debouncedSearch });
    }
    if (selectedCategories.length > 0) {
      filters.push({ name: "categories", value: selectedCategories });
    }
    fetchMarketplaceChannels({
      filters: filters.length ? filters : undefined,
      sort_by: sortBy,
      sort_order: sortOrder,
      page,
    })
      .then((res) => {
        setChannels(res.channels ?? []);
        setTotal(res.total ?? 0);
      })
      .catch(() => {
        setChannels([]);
        setTotal(0);
      })
      .finally(() => setLoading(false));
  }, [debouncedSearch, sortBy, sortOrder, page, selectedCategories]);

  useEffect(() => {
    load();
  }, [load]);

  return {
    channels,
    total,
    loading,
    search,
    setSearch,
    sortBy,
    setSortBy,
    sortOrder,
    setSortOrder,
    page,
    setPage,
    selectedCategories,
    setSelectedCategories,
  };
}
