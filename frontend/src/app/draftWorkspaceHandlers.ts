import type { Dispatch, SetStateAction } from "react";
import {
  advanceDraftSeason,
  createDraftTrade,
  deleteDraftPickTrade,
  recordDraftAction,
  resolveTradeCondition,
  resetDraftSession,
  setPlayerPreference,
  simulateToNextTurn,
  startDraftSession,
  syncSleeperDraft,
  undoDraftAction,
  undoDraftSessionReset,
} from "../shared/api/draft";
import type {
  BudgetAsset,
  DraftAction,
  DraftPickTrade,
  DraftSnapshot,
  FutureDraftPick,
  Player,
} from "../shared/api/types";
import { draftTeamName } from "./draftTradePresentation";

interface DraftWorkspaceMutationContext {
  leagueId: string;
  snapshot: DraftSnapshot | null;
  busy: boolean;
  setPendingFocus: Dispatch<SetStateAction<string | null>>;
  setAnnouncement: Dispatch<SetStateAction<string>>;
  setBusy: Dispatch<SetStateAction<boolean>>;
  setError: Dispatch<SetStateAction<string>>;
  setSelectedTeamNumber: Dispatch<SetStateAction<number>>;
  setSnapshot: Dispatch<SetStateAction<DraftSnapshot | null>>;
}

export interface DraftWorkspaceHandlers {
  action: (player: Player, action: DraftAction, cost?: number, teamNumber?: number) => Promise<void>;
  preference: (player: Player, preference: "target" | "avoid" | "") => Promise<void>;
  mock: () => Promise<void>;
  sleeperSync: (draftId: string, rosterId: number) => Promise<void>;
  undo: () => Promise<void>;
  startDraft: () => Promise<void>;
  resetDraft: (confirmation: string) => Promise<void>;
  undoReset: () => Promise<void>;
  createTrade: (
    teamOne: number,
    teamTwo: number,
    teamOneReceives: number[],
    teamTwoReceives: number[],
    teamOneFuture: FutureDraftPick[],
    teamTwoFuture: FutureDraftPick[],
    teamOnePlayers: string[],
    teamTwoPlayers: string[],
    teamOneBudgets: BudgetAsset[],
    teamTwoBudgets: BudgetAsset[],
  ) => Promise<void>;
  resolveTrade: (trade: DraftPickTrade, pick: FutureDraftPick, status: "met" | "not-met") => Promise<void>;
  advanceSeason: (season: number, draftType: DraftSnapshot["draftType"], draftOrder: number[]) => Promise<void>;
  deleteTrade: (trade: DraftPickTrade) => Promise<void>;
}

export function createDraftWorkspaceHandlers(context: DraftWorkspaceMutationContext): DraftWorkspaceHandlers {
  return {
    ...createPlayerHandlers(context),
    ...createSessionHandlers(context),
    ...createTradeHandlers(context),
  };
}

function createPlayerHandlers(context: DraftWorkspaceMutationContext) {
  return {
    action: async (player: Player, action: DraftAction, cost = 0, teamNumber = 0) => {
      if (!context.snapshot || context.busy) return;
      const index = context.snapshot.available.findIndex((candidate) => candidate.id === player.id);
      const nextFocus =
        context.snapshot.available[index + 1]?.id ?? context.snapshot.available[index - 1]?.id ?? "board";
      await runMutation(
        context,
        "The draft action failed.",
        () => recordDraftAction(context.leagueId, player.id, action, cost, teamNumber),
        (updated) => {
          context.setPendingFocus(nextFocus);
          updateSnapshot(context, updated);
          context.setAnnouncement(
            action === "draft"
              ? `${player.name} was drafted to your team. Rankings and recommendations updated.`
              : `${player.name} was marked taken by another team. Rankings and recommendations updated.`,
          );
        },
        () => {
          context.setPendingFocus(player.id);
        },
      );
    },
    preference: async (player: Player, preference: "target" | "avoid" | "") => {
      await runMutation(
        context,
        "Unable to save player preference.",
        () => setPlayerPreference(context.leagueId, player.id, preference),
        (updated) => {
          context.setSnapshot(updated);
          context.setAnnouncement(
            `${player.name} ${preference ? `was added to your ${preference} list` : "was returned to neutral"}. Recommendations updated.`,
          );
        },
      );
    },
    mock: async () => {
      await runMutation(
        context,
        "Unable to simulate opponent picks.",
        () => simulateToNextTurn(context.leagueId),
        (updated) => {
          updateSnapshot(context, updated);
          context.setAnnouncement(`Mock opponents completed. It is now pick ${updated.pickNumber}.`);
        },
      );
    },
    sleeperSync: async (draftId: string, rosterId: number) => {
      await runMutation(
        context,
        "Unable to synchronize Sleeper picks.",
        () => syncSleeperDraft(context.leagueId, draftId, rosterId),
        (result) => {
          updateSnapshot(context, result.snapshot);
          context.setAnnouncement(
            `Sleeper sync reconciled ${result.snapshot.history.length} picks: ${result.added} added, ${result.updated} changed, ${result.removed} removed${result.unmatched ? `, ${result.unmatched} unmatched` : ""}.`,
          );
        },
      );
    },
    undo: async () => {
      if (!context.snapshot?.canUndo || context.busy) return;
      const lastPick = context.snapshot.history[context.snapshot.history.length - 1];
      await runMutation(
        context,
        "Undo failed.",
        () => undoDraftAction(context.leagueId),
        (updated) => {
          context.setPendingFocus(lastPick.player.id);
          updateSnapshot(context, updated);
          context.setAnnouncement(`${lastPick.player.name} was restored to the available-player list.`);
        },
      );
    },
  };
}

function createSessionHandlers(context: DraftWorkspaceMutationContext) {
  return {
    startDraft: async () => {
      await runMutation(
        context,
        "Unable to start the draft.",
        () => startDraftSession(context.leagueId),
        (updated) => {
          updateSnapshot(context, updated);
          context.setAnnouncement(
            `Draft started. ${draftTeamName(updated, updated.onClockTeamNumber)} is on the clock.`,
          );
        },
      );
    },
    resetDraft: async (confirmation: string) => {
      await runMutation(
        context,
        "Unable to reset the draft.",
        () => resetDraftSession(context.leagueId, confirmation),
        (updated) => {
          updateSnapshot(context, updated);
          context.setAnnouncement("Current-season draft selections were cleared. The reset can still be undone.");
        },
        undefined,
        true,
      );
    },
    undoReset: async () => {
      await runMutation(
        context,
        "Unable to restore the draft.",
        () => undoDraftSessionReset(context.leagueId),
        (updated) => {
          updateSnapshot(context, updated);
          context.setAnnouncement(`${updated.history.length} draft selections were restored.`);
        },
      );
    },
    advanceSeason: async (season: number, draftType: DraftSnapshot["draftType"], draftOrder: number[]) => {
      await runMutation(
        context,
        "Unable to start the next season.",
        () => advanceDraftSeason(context.leagueId, season, draftType, draftOrder),
        (updated) => {
          updateSnapshot(context, updated);
          context.setAnnouncement(
            `${season} is ready. Franchise rosters and traded assets carried forward to a clean draft.`,
          );
        },
        undefined,
        true,
      );
    },
  };
}

function createTradeHandlers(context: DraftWorkspaceMutationContext) {
  return {
    createTrade: async (
      teamOne: number,
      teamTwo: number,
      teamOneReceives: number[],
      teamTwoReceives: number[],
      teamOneFuture: FutureDraftPick[],
      teamTwoFuture: FutureDraftPick[],
      teamOnePlayers: string[],
      teamTwoPlayers: string[],
      teamOneBudgets: BudgetAsset[],
      teamTwoBudgets: BudgetAsset[],
    ) => {
      await runMutation(
        context,
        "Unable to save the draft asset trade.",
        () =>
          createDraftTrade(
            context.leagueId,
            teamOne,
            teamTwo,
            teamOneReceives,
            teamTwoReceives,
            teamOneFuture,
            teamTwoFuture,
            0,
            0,
            teamOnePlayers,
            teamTwoPlayers,
            teamOneBudgets,
            teamTwoBudgets,
          ),
        (updated) => {
          updateSnapshot(context, updated);
          context.setAnnouncement(
            `Trade confirmed between ${draftTeamName(updated, teamOne)} and ${draftTeamName(updated, teamTwo)} with ${assetCount(teamOneReceives, teamOneFuture, teamOnePlayers, teamOneBudgets)} and ${assetCount(teamTwoReceives, teamTwoFuture, teamTwoPlayers, teamTwoBudgets)} recorded.`,
          );
        },
        undefined,
        true,
      );
    },
    resolveTrade: async (trade: DraftPickTrade, pick: FutureDraftPick, status: "met" | "not-met") => {
      await runMutation(
        context,
        "Unable to resolve the pick condition.",
        () => resolveTradeCondition(context.leagueId, trade.id, pick, status),
        (updated) => {
          context.setSnapshot(updated);
          context.setAnnouncement(
            `The condition for the ${pick.season} round ${pick.round} pick was marked ${status === "met" ? "met" : "not met"}.`,
          );
        },
        undefined,
        true,
      );
    },
    deleteTrade: async (trade: DraftPickTrade) => {
      await runMutation(
        context,
        "Unable to reverse the draft asset trade.",
        () => deleteDraftPickTrade(context.leagueId, trade.id),
        (updated) => {
          updateSnapshot(context, updated);
          context.setAnnouncement(`Trade between ${trade.teamOneName} and ${trade.teamTwoName} was reversed.`);
        },
        undefined,
        true,
      );
    },
  };
}

async function runMutation<Result>(
  context: DraftWorkspaceMutationContext,
  fallbackError: string,
  operation: () => Promise<Result>,
  onSuccess: (result: Result) => void,
  onFailure?: () => void,
  rethrow = false,
) {
  if (context.busy) return;
  context.setBusy(true);
  context.setError("");
  try {
    onSuccess(await operation());
  } catch (reason) {
    onFailure?.();
    context.setError(reason instanceof Error ? reason.message : fallbackError);
    if (rethrow) throw reason;
  } finally {
    context.setBusy(false);
  }
}

function updateSnapshot(context: DraftWorkspaceMutationContext, snapshot: DraftSnapshot) {
  context.setSnapshot(snapshot);
  context.setSelectedTeamNumber(defaultSelectedTeam(snapshot));
}

function assetCount(current: number[], future: FutureDraftPick[], players: string[], budgets: BudgetAsset[]) {
  const count = current.length + future.length + players.length + budgets.length;
  return `${count} ${count === 1 ? "asset" : "assets"}`;
}

export function defaultSelectedTeam(snapshot: DraftSnapshot): number {
  if (snapshot.draftType !== "auction") return snapshot.onClockTeamNumber;
  return snapshot.teams.find((team) => !team.isUser)?.number ?? 0;
}
