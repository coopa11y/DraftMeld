import { useEffect, useRef, useState } from "react";
import {
  advanceDraftSeason,
  createDraftTrade,
  deleteDraftPickTrade,
  getDraft,
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
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { DraftSidebar } from "./DraftSidebar";
import { DraftOverview } from "./DraftOverview";
import { PlayerBoard } from "./PlayerBoard";
import { draftTeamName } from "./draftTradePresentation";

interface DraftWorkspaceProps {
  leagueId: string;
}

export function DraftWorkspace({ leagueId }: DraftWorkspaceProps) {
  const [snapshot, setSnapshot] = useState<DraftSnapshot | null>(null);
  const [busy, setBusy] = useState(false);
  const [announcement, setAnnouncement] = useState("Draft board loading.");
  const [error, setError] = useState("");
  const [selectedTeamNumber, setSelectedTeamNumber] = useState(0);
  const pendingFocus = useRef<string | null>(null);
  const boardHeading = useViewHeadingFocus<HTMLHeadingElement>(snapshot !== null);
  const errorAlert = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let active = true;
    getDraft(leagueId)
      .then((data) => {
        if (!active) return;
        setSnapshot(data);
        setSelectedTeamNumber(defaultSelectedTeam(data));
        setAnnouncement(`Draft board loaded. ${data.available.length} players are available.`);
      })
      .catch((reason: Error) => {
        if (active) setError(reason.message);
      });
    return () => {
      active = false;
    };
  }, [leagueId]);

  useEffect(() => {
    if (!snapshot || !pendingFocus.current) return;
    const requested = document.getElementById(`draft-${pendingFocus.current}`);
    const fallback = document.querySelector<HTMLButtonElement>("[data-player-action='primary']");
    (requested ?? fallback ?? boardHeading.current)?.focus();
    pendingFocus.current = null;
  }, [boardHeading, snapshot]);

  useEffect(() => {
    if (error) errorAlert.current?.focus();
  }, [error]);

  async function handleAction(player: Player, action: DraftAction, cost = 0, teamNumber = 0) {
    if (!snapshot || busy) return;
    const index = snapshot.available.findIndex((candidate) => candidate.id === player.id);
    pendingFocus.current = snapshot.available[index + 1]?.id ?? snapshot.available[index - 1]?.id ?? "board";
    setBusy(true);
    setError("");
    try {
      const updated = await recordDraftAction(leagueId, player.id, action, cost, teamNumber);
      setSnapshot(updated);
      setSelectedTeamNumber(defaultSelectedTeam(updated));
      setAnnouncement(
        action === "draft"
          ? `${player.name} was drafted to your team. Rankings and recommendations updated.`
          : `${player.name} was marked taken by another team. Rankings and recommendations updated.`,
      );
    } catch (reason) {
      pendingFocus.current = player.id;
      setError(reason instanceof Error ? reason.message : "The draft action failed.");
    } finally {
      setBusy(false);
    }
  }

  async function handlePreference(player: Player, preference: "target" | "avoid" | "") {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const updated = await setPlayerPreference(leagueId, player.id, preference);
      setSnapshot(updated);
      setAnnouncement(
        `${player.name} ${preference ? `was added to your ${preference} list` : "was returned to neutral"}. Recommendations updated.`,
      );
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to save player preference.");
    } finally {
      setBusy(false);
    }
  }

  async function handleMock() {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const updated = await simulateToNextTurn(leagueId);
      setSnapshot(updated);
      setSelectedTeamNumber(defaultSelectedTeam(updated));
      setAnnouncement(`Mock opponents completed. It is now pick ${updated.pickNumber}.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to simulate opponent picks.");
    } finally {
      setBusy(false);
    }
  }

  async function handleSleeperSync(draftId: string, rosterId: number) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const result = await syncSleeperDraft(leagueId, draftId, rosterId);
      setSnapshot(result.snapshot);
      setSelectedTeamNumber(defaultSelectedTeam(result.snapshot));
      setAnnouncement(
        `Sleeper sync reconciled ${result.snapshot.history.length} picks: ${result.added} added, ${result.updated} changed, ${result.removed} removed${result.unmatched ? `, ${result.unmatched} unmatched` : ""}.`,
      );
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to synchronize Sleeper picks.");
    } finally {
      setBusy(false);
    }
  }

  async function handleUndo() {
    if (!snapshot?.canUndo || busy) return;
    const lastPick = snapshot.history[snapshot.history.length - 1];
    pendingFocus.current = lastPick.player.id;
    setBusy(true);
    setError("");
    try {
      const updated = await undoDraftAction(leagueId);
      setSnapshot(updated);
      setSelectedTeamNumber(defaultSelectedTeam(updated));
      setAnnouncement(`${lastPick.player.name} was restored to the available-player list.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Undo failed.");
    } finally {
      setBusy(false);
    }
  }

  async function handleStartDraft() {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const updated = await startDraftSession(leagueId);
      setSnapshot(updated);
      setSelectedTeamNumber(defaultSelectedTeam(updated));
      setAnnouncement(`Draft started. ${draftTeamName(updated, updated.onClockTeamNumber)} is on the clock.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to start the draft.");
    } finally {
      setBusy(false);
    }
  }

  async function handleResetDraft(confirmation: string) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const updated = await resetDraftSession(leagueId, confirmation);
      setSnapshot(updated);
      setSelectedTeamNumber(defaultSelectedTeam(updated));
      setAnnouncement("Current-season draft selections were cleared. The reset can still be undone.");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to reset the draft.");
      throw reason;
    } finally {
      setBusy(false);
    }
  }

  async function handleUndoReset() {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const updated = await undoDraftSessionReset(leagueId);
      setSnapshot(updated);
      setSelectedTeamNumber(defaultSelectedTeam(updated));
      setAnnouncement(`${updated.history.length} draft selections were restored.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to restore the draft.");
    } finally {
      setBusy(false);
    }
  }

  async function handleCreateTrade(
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
  ) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const updated = await createDraftTrade(
        leagueId,
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
      );
      setSnapshot(updated);
      setSelectedTeamNumber(defaultSelectedTeam(updated));
      setAnnouncement(
        `Trade confirmed between ${draftTeamName(updated, teamOne)} and ${draftTeamName(updated, teamTwo)} with ${assetCount(teamOneReceives, teamOneFuture, teamOnePlayers, teamOneBudgets)} and ${assetCount(teamTwoReceives, teamTwoFuture, teamTwoPlayers, teamTwoBudgets)} recorded.`,
      );
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to save the draft asset trade.");
      throw reason;
    } finally {
      setBusy(false);
    }
  }

  async function handleResolveTrade(trade: DraftPickTrade, pick: FutureDraftPick, status: "met" | "not-met") {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const updated = await resolveTradeCondition(leagueId, trade.id, pick, status);
      setSnapshot(updated);
      setAnnouncement(
        `The condition for the ${pick.season} round ${pick.round} pick was marked ${status === "met" ? "met" : "not met"}.`,
      );
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to resolve the pick condition.");
      throw reason;
    } finally {
      setBusy(false);
    }
  }

  async function handleAdvanceSeason(season: number, draftType: DraftSnapshot["draftType"], draftOrder: number[]) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const updated = await advanceDraftSeason(leagueId, season, draftType, draftOrder);
      setSnapshot(updated);
      setSelectedTeamNumber(defaultSelectedTeam(updated));
      setAnnouncement(`${season} is ready. Franchise rosters and traded assets carried forward to a clean draft.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to start the next season.");
      throw reason;
    } finally {
      setBusy(false);
    }
  }

  async function handleDeleteTrade(trade: DraftPickTrade) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const updated = await deleteDraftPickTrade(leagueId, trade.id);
      setSnapshot(updated);
      setSelectedTeamNumber(defaultSelectedTeam(updated));
      setAnnouncement(`Trade between ${trade.teamOneName} and ${trade.teamTwoName} was reversed.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to reverse the draft asset trade.");
      throw reason;
    } finally {
      setBusy(false);
    }
  }

  if (!snapshot) {
    return (
      <main className="centered-status" aria-busy={!error}>
        {error ? (
          <StatusMessage tone="error" tabIndex={-1} ref={errorAlert}>
            {error}
          </StatusMessage>
        ) : (
          <StatusMessage>Loading draft board…</StatusMessage>
        )}
      </main>
    );
  }

  const onClockTeam = snapshot.teams.find((team) => team.number === selectedTeamNumber);
  const selectedTeamIsUser = onClockTeam?.isUser ?? false;
  const draftActive = snapshot.sessionStatus === "in-progress";

  return (
    <>
      <a className="skip-link" href="#player-board">
        Skip to player board
      </a>
      <a className="skip-link" href="#recommendations">
        Skip to recommendations
      </a>
      <a className="skip-link" href="#my-team">
        Skip to my team
      </a>

      <StatusMessage visuallyHidden aria-atomic="true">
        {announcement}
      </StatusMessage>
      <main id="main-content">
        {error ? (
          <StatusMessage tone="error" tabIndex={-1} ref={errorAlert}>
            {error}
          </StatusMessage>
        ) : null}
        <section
          className="workspace-status"
          aria-label={`Draft status. ${sessionStatusLabel(snapshot)}. ${snapshot.available.length} players available.`}
        >
          <span>{snapshot.leagueName}</span>
          <strong>{sessionStatusLabel(snapshot)}</strong>
          {draftActive && onClockTeam ? <span>{onClockTeam.name} on the clock</span> : null}
          <span>{snapshot.available.length} available</span>
          <span>
            {snapshot.dataMode}
            {snapshot.projectionCount ? ` · ${snapshot.projectionCount} projections` : ""}
          </span>
        </section>
        <div className="draft-layout" aria-busy={busy}>
          <PlayerBoard
            snapshot={snapshot}
            busy={busy || !draftActive}
            headingRef={boardHeading}
            selectedTeamNumber={selectedTeamNumber}
            selectedTeamIsUser={selectedTeamIsUser}
            onAction={handleAction}
            onUndo={handleUndo}
            onPreference={handlePreference}
          />
          <DraftSidebar
            snapshot={snapshot}
            busy={busy || !draftActive}
            selectedTeamNumber={selectedTeamNumber}
            selectedTeamIsUser={selectedTeamIsUser}
            onSelectedTeamChange={setSelectedTeamNumber}
            onAction={handleAction}
            onMock={handleMock}
            onSleeperSync={handleSleeperSync}
          />
        </div>
        <DraftOverview
          snapshot={snapshot}
          busy={busy}
          onCreateTrade={handleCreateTrade}
          onDeleteTrade={handleDeleteTrade}
          onResolveTrade={handleResolveTrade}
          onAdvanceSeason={handleAdvanceSeason}
          onStartDraft={handleStartDraft}
          onResetDraft={handleResetDraft}
          onUndoReset={handleUndoReset}
        />
      </main>
    </>
  );
}

function sessionStatusLabel(snapshot: DraftSnapshot) {
  if (snapshot.sessionStatus === "not-started") return "Draft not started";
  if (snapshot.sessionStatus === "complete") return "Draft complete";
  return `Pick ${snapshot.pickNumber}`;
}

function defaultSelectedTeam(snapshot: DraftSnapshot): number {
  if (snapshot.draftType !== "auction") return snapshot.onClockTeamNumber;
  return snapshot.teams.find((team) => !team.isUser)?.number ?? 0;
}

function assetCount(current: number[], future: FutureDraftPick[], players: string[], budgets: BudgetAsset[]) {
  const count = current.length + future.length + players.length + budgets.length;
  return `${count} ${count === 1 ? "asset" : "assets"}`;
}
