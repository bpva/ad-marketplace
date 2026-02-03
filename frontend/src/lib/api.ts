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
