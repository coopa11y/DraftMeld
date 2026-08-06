import type { DraftAction, DraftSnapshot } from "./types";

const leagueId = "demo";

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  const body: unknown = await response.json();
  if (!response.ok) {
    const message = typeof body === "object" && body !== null && "error" in body && typeof body.error === "string"
      ? body.error
      : `Request failed with ${response.status}`;
    throw new Error(message);
  }
  return body as T;
}

export function getDraft(): Promise<DraftSnapshot> {
  return request(`/api/v1/draft?leagueId=${leagueId}`);
}

export function recordDraftAction(playerId: string, action: DraftAction): Promise<DraftSnapshot> {
  return request("/api/v1/draft/actions", {
    method: "POST",
    body: JSON.stringify({ leagueId, playerId, action }),
  });
}

export function undoDraftAction(): Promise<DraftSnapshot> {
  return request("/api/v1/draft/undo", {
    method: "POST",
    body: JSON.stringify({ leagueId }),
  });
}
