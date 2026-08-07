import type { DraftAction, DraftSnapshot, SleeperSyncResult } from "./types";
import { apiClient, unwrap } from "./client";

export async function getDraft(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.GET("/draft", { params: { query: { leagueId } } });
  return unwrap(data, error, response);
}

export async function recordDraftAction(
  leagueId: string,
  playerId: string,
  action: DraftAction,
  cost = 0,
): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/actions", {
    body: { leagueId, playerId, action, cost },
  });
  return unwrap(data, error, response);
}

export async function setPlayerPreference(leagueId: string, playerId: string, preference: "target" | "avoid" | ""): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.PUT("/draft/preferences", { body: { leagueId, playerId, preference } });
  return unwrap(data, error, response);
}

export async function simulateToNextTurn(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/mock", { body: { leagueId } });
  return unwrap(data, error, response);
}

export async function syncSleeperDraft(leagueId: string, sleeperDraftId: string, rosterId: number): Promise<SleeperSyncResult> {
  const { data, error, response } = await apiClient.POST("/draft/sync/sleeper", { body: { leagueId, sleeperDraftId, rosterId } });
  return unwrap(data, error, response);
}

export async function undoDraftAction(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/undo", { body: { leagueId } });
  return unwrap(data, error, response);
}
