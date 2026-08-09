import type { BudgetAsset, DraftAction, DraftSnapshot, FutureDraftPick, SleeperSyncResult } from "./types";
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
  teamNumber = 0,
): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/actions", {
    body: { leagueId, playerId, action, cost, teamNumber },
  });
  return unwrap(data, error, response);
}

export async function startDraftSession(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/session/start", { body: { leagueId } });
  return unwrap(data, error, response);
}

export async function resetDraftSession(leagueId: string, confirmation: string): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/session/reset", {
    body: { leagueId, confirmation },
  });
  return unwrap(data, error, response);
}

export async function undoDraftSessionReset(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/session/undo-reset", { body: { leagueId } });
  return unwrap(data, error, response);
}

export async function setPlayerPreference(
  leagueId: string,
  playerId: string,
  preference: "target" | "avoid" | "",
): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.PUT("/draft/preferences", {
    body: { leagueId, playerId, preference },
  });
  return unwrap(data, error, response);
}

export async function simulateToNextTurn(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/mock", { body: { leagueId } });
  return unwrap(data, error, response);
}

export async function syncSleeperDraft(
  leagueId: string,
  sleeperDraftId: string,
  rosterId: number,
): Promise<SleeperSyncResult> {
  const { data, error, response } = await apiClient.POST("/draft/sync/sleeper", {
    body: { leagueId, sleeperDraftId, rosterId },
  });
  return unwrap(data, error, response);
}

export async function undoDraftAction(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/undo", { body: { leagueId } });
  return unwrap(data, error, response);
}

export async function createDraftTrade(
  leagueId: string,
  teamOneNumber: number,
  teamTwoNumber: number,
  teamOneReceives: number[],
  teamTwoReceives: number[],
  teamOneFuturePicks: FutureDraftPick[],
  teamTwoFuturePicks: FutureDraftPick[],
  teamOneAuctionBudget: number,
  teamTwoAuctionBudget: number,
  teamOnePlayers: string[],
  teamTwoPlayers: string[],
  teamOneBudgets: BudgetAsset[],
  teamTwoBudgets: BudgetAsset[],
): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/trades", {
    body: {
      leagueId,
      teamOneNumber,
      teamTwoNumber,
      teamOneReceives,
      teamTwoReceives,
      teamOneFuturePicks,
      teamTwoFuturePicks,
      teamOneAuctionBudget,
      teamTwoAuctionBudget,
      teamOnePlayers,
      teamTwoPlayers,
      teamOneBudgets,
      teamTwoBudgets,
    },
  });
  return unwrap(data, error, response);
}

export async function resolveTradeCondition(
  leagueId: string,
  tradeId: number,
  pick: FutureDraftPick,
  status: "met" | "not-met",
): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.PUT("/draft/trades/{tradeId}/condition", {
    params: { path: { tradeId } },
    body: {
      leagueId,
      season: pick.season,
      round: pick.round,
      originalTeamNumber: pick.originalTeamNumber,
      status,
    },
  });
  return unwrap(data, error, response);
}

export async function advanceDraftSeason(
  leagueId: string,
  season: number,
  draftType: DraftSnapshot["draftType"],
  draftOrder: number[],
): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.POST("/draft/seasons", {
    body: { leagueId, season, draftType, draftOrder },
  });
  return unwrap(data, error, response);
}

export async function deleteDraftPickTrade(leagueId: string, tradeId: number): Promise<DraftSnapshot> {
  const { data, error, response } = await apiClient.DELETE("/draft/trades/{tradeId}", {
    params: { path: { tradeId }, query: { leagueId } },
  });
  return unwrap(data, error, response);
}
