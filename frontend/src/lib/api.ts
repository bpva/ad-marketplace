import type { components } from "@/types/api";

const API_URL = import.meta.env.VITE_API_URL || "";

let token: string | null = null;

export function setToken(t: string) {
  token = t;
}

export function getToken(): string | null {
  return token;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    throw new Error(`API error: ${res.status}`);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return res.json();
}

export type User = components["schemas"]["UserResponse"];
type AuthResponse = components["schemas"]["AuthResponse"];

export async function authenticate(initData: string): Promise<AuthResponse> {
  return request<AuthResponse>("/api/v1/auth", {
    method: "POST",
    body: JSON.stringify({ init_data: initData }),
  });
}

export async function fetchMe(): Promise<User> {
  return request<User>("/api/v1/me");
}

export type Profile = components["schemas"]["ProfileResponse"];
export type UpdateSettingsRequest = components["schemas"]["UpdateSettingsRequest"];
export type Language = components["schemas"]["Language"];
export type Theme = components["schemas"]["Theme"];
export type PreferredMode = components["schemas"]["PreferredMode"];

export async function fetchProfile(): Promise<Profile> {
  return request<Profile>("/api/v1/user/profile");
}

export async function updateName(name: string): Promise<void> {
  await request("/api/v1/user/name", {
    method: "PATCH",
    body: JSON.stringify({ name }),
  });
}

export async function updateSettings(settings: UpdateSettingsRequest): Promise<void> {
  await request("/api/v1/user/settings", {
    method: "PATCH",
    body: JSON.stringify(settings),
  });
}

export async function linkWallet(address: string): Promise<void> {
  await request("/api/v1/user/wallet", {
    method: "PUT",
    body: JSON.stringify({ address }),
  });
}

export async function unlinkWallet(): Promise<void> {
  await request("/api/v1/user/wallet", {
    method: "DELETE",
  });
}

export type ChannelsResponse = components["schemas"]["ChannelsResponse"];
export type ChannelWithRole = components["schemas"]["ChannelWithRoleResponse"];
export type Channel = components["schemas"]["ChannelResponse"];
export type ChannelRoleType = components["schemas"]["ChannelRoleType"];

export async function fetchChannels(): Promise<ChannelsResponse> {
  return request<ChannelsResponse>("/api/v1/channels");
}

export type AdFormat = components["schemas"]["AdFormatResponse"];
export type AdFormatsResponse = components["schemas"]["AdFormatsResponse"];
export type AddAdFormatRequest = components["schemas"]["AddAdFormatRequest"];
export type AdFormatType = components["schemas"]["AdFormatType"];
export type UpdateListingRequest = components["schemas"]["UpdateListingRequest"];

export async function fetchAdFormats(channelId: number): Promise<AdFormatsResponse> {
  return request<AdFormatsResponse>(`/api/v1/channels/${channelId}/ad-formats`);
}

export async function addAdFormat(channelId: number, req: AddAdFormatRequest): Promise<void> {
  await request(`/api/v1/channels/${channelId}/ad-formats`, {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export async function deleteAdFormat(channelId: number, formatId: string): Promise<void> {
  await request(`/api/v1/channels/${channelId}/ad-formats/${formatId}`, {
    method: "DELETE",
  });
}

export async function updateChannelListing(channelId: number, isListed: boolean): Promise<void> {
  await request(`/api/v1/channels/${channelId}/listing`, {
    method: "PATCH",
    body: JSON.stringify({ is_listed: isListed }),
  });
}

export type MarketplaceChannel = components["schemas"]["MarketplaceChannel"];
export type MarketplaceChannelsResponse = components["schemas"]["MarketplaceChannelsResponse"];
export type MarketplaceAdFormat = components["schemas"]["AdFormat"];
export type MarketplaceFilter = components["schemas"]["MarketplaceFilter"];
export type ChannelSortBy = components["schemas"]["ChannelSortBy"];
export type SortOrder = components["schemas"]["SortOrder"];
export type CategoryResponse = components["schemas"]["CategoryResponse"];

export type TemplatesResponse = components["schemas"]["TemplatesResponse"];
export type TemplateResponse = components["schemas"]["TemplateResponse"];

export async function fetchTemplates(): Promise<TemplatesResponse> {
  return request<TemplatesResponse>("/api/v1/posts");
}

export async function sendTemplatePreview(postID: string): Promise<void> {
  await request(`/api/v1/posts/${postID}/preview`, { method: "POST" });
}

export async function fetchPostMediaBlob(postID: string): Promise<string> {
  const res = await fetch(`${API_URL}/api/v1/posts/${postID}/media`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (!res.ok) throw new Error(`Media fetch failed: ${res.status}`);
  return URL.createObjectURL(await res.blob());
}

export async function fetchChannelPhotoBlob(
  channelId: number,
  size: "small" | "big" = "small",
): Promise<string> {
  const res = await fetch(`${API_URL}/api/v1/channels/${channelId}/photo?size=${size}`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  if (!res.ok) throw new Error(`Channel photo fetch failed: ${res.status}`);
  return URL.createObjectURL(await res.blob());
}

export type TonRates = components["schemas"]["TonRatesResponse"];

export async function fetchTonRates(): Promise<TonRates> {
  return request<TonRates>("/api/v1/ton-rates");
}

export async function updateCategories(channelId: number, categories: string[]): Promise<void> {
  await request(`/api/v1/channels/${channelId}/categories`, {
    method: "PATCH",
    body: JSON.stringify({ categories }),
  });
}

export async function fetchMarketplaceChannels(params: {
  filters?: MarketplaceFilter[];
  sort_by?: ChannelSortBy;
  sort_order?: SortOrder;
  page?: number;
}): Promise<MarketplaceChannelsResponse> {
  const body: Record<string, unknown> = {};
  if (params.filters?.length) body.filters = params.filters;
  if (params.sort_by) body.sort_by = params.sort_by;
  if (params.sort_order) body.sort_order = params.sort_order;
  if (params.page && params.page > 1) body.page = params.page;
  return request<MarketplaceChannelsResponse>("/api/v1/mp/channels", {
    method: "POST",
    body: JSON.stringify(body),
  });
}
