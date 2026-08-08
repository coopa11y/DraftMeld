import type { League, LeagueBackup, LeagueRules } from "./types";
import { apiClient, ensureSuccess, unwrap } from "./client";

export async function listLeagues(): Promise<League[]> {
  const { data, error, response } = await apiClient.GET("/leagues");
  return unwrap(data, error, response);
}

export async function createLeague(rules: LeagueRules): Promise<League> {
  const { data, error, response } = await apiClient.POST("/leagues", { body: rules });
  return unwrap(data, error, response);
}

export async function updateLeague(id: string, rules: LeagueRules): Promise<League> {
  const { data, error, response } = await apiClient.PUT("/leagues/{leagueId}", {
    params: { path: { leagueId: id } },
    body: rules,
  });
  return unwrap(data, error, response);
}

export async function duplicateLeague(id: string): Promise<League> {
  const { data, error, response } = await apiClient.POST("/leagues/{leagueId}/duplicate", {
    params: { path: { leagueId: id } },
  });
  return unwrap(data, error, response);
}

export async function deleteLeague(id: string): Promise<void> {
  const { error, response } = await apiClient.DELETE("/leagues/{leagueId}", {
    params: { path: { leagueId: id } },
  });
  ensureSuccess(error, response);
}

export async function importLeagueBackup(backup: LeagueBackup): Promise<League> {
  const { data, error, response } = await apiClient.POST("/leagues/import", { body: backup });
  return unwrap(data, error, response);
}

export type LeagueExportKind = "backup" | "rankings.csv" | "draft.csv" | "draft.json";

export async function downloadLeagueExport(id: string, kind: LeagueExportKind): Promise<string> {
  const path =
    kind === "backup"
      ? `/leagues/${encodeURIComponent(id)}/backup`
      : `/leagues/${encodeURIComponent(id)}/exports/${kind}`;
  const response = await fetch(`/api/v1${path}`, {
    headers: { Accept: kind.endsWith(".csv") ? "text/csv" : "application/json" },
  });
  if (!response.ok) {
    let message = `Download failed with status ${response.status}.`;
    try {
      const payload = (await response.json()) as { error?: string };
      if (payload.error) message = payload.error;
    } catch {
      // Keep the status-based fallback when an intermediary returns a non-JSON error page.
    }
    throw new Error(message);
  }
  const blob = await response.blob();
  const disposition = response.headers.get("Content-Disposition") ?? "";
  const filename = disposition.match(/filename="([^"]+)"/i)?.[1] ?? `draftmeld-${id}-${kind}`;
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  window.setTimeout(() => URL.revokeObjectURL(url), 0);
  return filename;
}

export function leagueToRules(league: League): LeagueRules {
  const { id: _id, ...rules } = league;
  return {
    ...rules,
    rosterSlots: rules.rosterSlots.map((slot) => ({ ...slot, positions: [...slot.positions] })),
    teamNames: [...rules.teamNames],
    scoringRules: { ...rules.scoringRules },
    sourcePreferences: Object.fromEntries(
      Object.entries(rules.sourcePreferences).map(([sourceId, preference]) => [sourceId, { ...preference }]),
    ),
    playerPreferences: { ...rules.playerPreferences },
  };
}
