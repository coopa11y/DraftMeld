import type { DraftAction, DraftSnapshot } from "./types";
import { apiClient, unwrap } from "./client";

export async function getDraft(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.GET("/draft", { params: { query: { leagueId } } });
  return unwrap(data, error, response);
}

export async function recordDraftAction(
  leagueId: string,
  playerId: string,
  action: DraftAction,
): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/actions", {
    body: { leagueId, playerId, action },
  });
  return unwrap(data, error, response);
}

export async function undoDraftAction(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/undo", { body: { leagueId } });
  return unwrap(data, error, response);
}
