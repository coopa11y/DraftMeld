import type { League, LeagueRules } from "./types";
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

export function leagueToRules(league: League): LeagueRules {
  const { id: _id, ...rules } = league;
  return {
    ...rules,
    rosterSlots: rules.rosterSlots.map((slot) => ({ ...slot, positions: [...slot.positions] })),
    scoringRules: { ...rules.scoringRules },
    sourcePreferences: Object.fromEntries(
      Object.entries(rules.sourcePreferences).map(([sourceId, preference]) => [sourceId, { ...preference }]),
    ),
    playerPreferences: { ...rules.playerPreferences },
  };
}
