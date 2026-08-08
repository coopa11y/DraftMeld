import { useEffect, useRef, useState } from "react";
import {
  createDraftPickTrade,
  deleteDraftPickTrade,
  getDraft,
  recordDraftAction,
  setPlayerPreference,
  simulateToNextTurn,
  syncSleeperDraft,
  undoDraftAction,
} from "../shared/api/draft";
import type { DraftAction, DraftPickTrade, DraftSnapshot, FutureDraftPick, Player } from "../shared/api/types";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { DraftSidebar } from "./DraftSidebar";
import { DraftOverview } from "./DraftOverview";
import { PlayerBoard } from "./PlayerBoard";

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

  async function handleCreateTrade(
    teamOne: number,
    teamTwo: number,
    teamOneReceives: number[],
    teamTwoReceives: number[],
    teamOneFuture: FutureDraftPick[],
    teamTwoFuture: FutureDraftPick[],
    teamOneBudget: number,
    teamTwoBudget: number,
  ) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const updated = await createDraftPickTrade(
        leagueId,
        teamOne,
        teamTwo,
        teamOneReceives,
        teamTwoReceives,
        teamOneFuture,
        teamTwoFuture,
        teamOneBudget,
        teamTwoBudget,
      );
      setSnapshot(updated);
      setSelectedTeamNumber(defaultSelectedTeam(updated));
      setAnnouncement(
        `Trade confirmed between ${teamName(updated, teamOne)} and ${teamName(updated, teamTwo)} with ${assetCount(teamOneReceives, teamOneFuture, teamOneBudget)} and ${assetCount(teamTwoReceives, teamTwoFuture, teamTwoBudget)} recorded.`,
      );
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to save the draft asset trade.");
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
          aria-label={`Draft status. Pick ${snapshot.pickNumber}. ${snapshot.available.length} players available.`}
        >
          <span>{snapshot.leagueName}</span>
          <strong>{snapshot.isComplete ? "Draft complete" : `Pick ${snapshot.pickNumber}`}</strong>
          {onClockTeam ? <span>{onClockTeam.name} on the clock</span> : null}
          <span>{snapshot.available.length} available</span>
          <span>
            {snapshot.dataMode}
            {snapshot.projectionCount ? ` · ${snapshot.projectionCount} projections` : ""}
          </span>
        </section>
        <div className="draft-layout" aria-busy={busy}>
          <PlayerBoard
            snapshot={snapshot}
            busy={busy}
            headingRef={boardHeading}
            selectedTeamNumber={selectedTeamNumber}
            selectedTeamIsUser={selectedTeamIsUser}
            onAction={handleAction}
            onUndo={handleUndo}
            onPreference={handlePreference}
          />
          <DraftSidebar
            snapshot={snapshot}
            busy={busy}
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
        />
      </main>
    </>
  );
}

function defaultSelectedTeam(snapshot: DraftSnapshot): number {
  if (snapshot.draftType !== "auction") return snapshot.onClockTeamNumber;
  return snapshot.teams.find((team) => !team.isUser)?.number ?? 0;
}

function teamName(snapshot: DraftSnapshot, teamNumber: number) {
  return snapshot.teams.find((team) => team.number === teamNumber)?.name ?? `Team ${teamNumber}`;
}

function assetCount(current: number[], future: FutureDraftPick[], budget: number) {
  const count = current.length + future.length + (budget > 0 ? 1 : 0);
  return `${count} ${count === 1 ? "asset" : "assets"}`;
}
