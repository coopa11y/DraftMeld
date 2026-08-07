import { useEffect, useRef, useState } from "react";
import { getDraft, recordDraftAction, setPlayerPreference, simulateToNextTurn, syncSleeperDraft, undoDraftAction } from "../shared/api/draft";
import type { DraftAction, DraftSnapshot, Player } from "../shared/api/types";
import { DraftSidebar } from "./DraftSidebar";
import { PlayerBoard } from "./PlayerBoard";

interface DraftWorkspaceProps {
  leagueId: string;
}

export function DraftWorkspace({ leagueId }: DraftWorkspaceProps) {
  const [snapshot, setSnapshot] = useState<DraftSnapshot | null>(null);
  const [busy, setBusy] = useState(false);
  const [announcement, setAnnouncement] = useState("Draft board loading.");
  const [error, setError] = useState("");
  const pendingFocus = useRef<string | null>(null);
  const boardHeading = useRef<HTMLHeadingElement>(null);
  const errorAlert = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let active = true;
    setSnapshot(null);
    setError("");
    setAnnouncement("Draft board loading.");
    getDraft(leagueId)
      .then((data) => {
        if (!active) return;
        setSnapshot(data);
        setAnnouncement(`Draft board loaded. ${data.available.length} players are available.`);
      })
      .catch((reason: Error) => {
        if (active) setError(reason.message);
      });
    return () => { active = false; };
  }, [leagueId]);

  useEffect(() => {
    if (!snapshot || !pendingFocus.current) return;
    const requested = document.getElementById(`draft-${pendingFocus.current}`);
    const fallback = document.querySelector<HTMLButtonElement>("[data-player-action='primary']");
    (requested ?? fallback ?? boardHeading.current)?.focus();
    pendingFocus.current = null;
  }, [snapshot]);

  useEffect(() => {
    if (error) errorAlert.current?.focus();
  }, [error]);

  async function handleAction(player: Player, action: DraftAction, cost = 0) {
    if (!snapshot || busy) return;
    const index = snapshot.available.findIndex((candidate) => candidate.id === player.id);
    pendingFocus.current = snapshot.available[index + 1]?.id ?? snapshot.available[index - 1]?.id ?? "board";
    setBusy(true);
    setError("");
    try {
      const updated = await recordDraftAction(leagueId, player.id, action, cost);
      setSnapshot(updated);
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
    setBusy(true); setError("");
    try {
      const updated = await setPlayerPreference(leagueId, player.id, preference);
      setSnapshot(updated);
      setAnnouncement(`${player.name} ${preference ? `was added to your ${preference} list` : "was returned to neutral"}. Recommendations updated.`);
    } catch (reason) { setError(reason instanceof Error ? reason.message : "Unable to save player preference."); }
    finally { setBusy(false); }
  }

  async function handleMock() {
    if (busy) return;
    setBusy(true); setError("");
    try { const updated = await simulateToNextTurn(leagueId); setSnapshot(updated); setAnnouncement(`Mock opponents completed. It is now pick ${updated.pickNumber}.`); }
    catch (reason) { setError(reason instanceof Error ? reason.message : "Unable to simulate opponent picks."); }
    finally { setBusy(false); }
  }

  async function handleSleeperSync(draftId: string, rosterId: number) {
    if (busy) return;
    setBusy(true); setError("");
    try { const updated = await syncSleeperDraft(leagueId, draftId, rosterId); setSnapshot(updated); setAnnouncement(`Sleeper synchronization complete. ${updated.history.length} picks are recorded.`); }
    catch (reason) { setError(reason instanceof Error ? reason.message : "Unable to synchronize Sleeper picks."); }
    finally { setBusy(false); }
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
      setAnnouncement(`${lastPick.player.name} was restored to the available-player list.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Undo failed.");
    } finally {
      setBusy(false);
    }
  }

  if (!snapshot && !error) {
    return <main className="centered-status" aria-busy="true"><p role="status">Loading draft board…</p></main>;
  }

  return (
    <>
      <a className="skip-link" href="#player-board">Skip to player board</a>
      <a className="skip-link" href="#recommendations">Skip to recommendations</a>
      <a className="skip-link" href="#my-team">Skip to my team</a>

      <div className="sr-only" role="status" aria-live="polite" aria-atomic="true">{announcement}</div>
      {error ? <div className="error-banner" role="alert" tabIndex={-1} ref={errorAlert}>{error}</div> : null}

      {snapshot ? (
        <>
          <section className="workspace-status" aria-label={`Draft status. Pick ${snapshot.pickNumber}. ${snapshot.available.length} players available.`}>
            <span>{snapshot.leagueName}</span>
            <strong>Pick {snapshot.pickNumber}</strong>
            <span>{snapshot.available.length} available</span>
            <span>{snapshot.dataMode}{snapshot.projectionCount ? ` · ${snapshot.projectionCount} projections` : ""}</span>
          </section>
          <main className="draft-layout" aria-busy={busy}>
            <PlayerBoard
              snapshot={snapshot}
              busy={busy}
              headingRef={boardHeading}
              onAction={handleAction}
              onUndo={handleUndo}
              onPreference={handlePreference}
            />
            <DraftSidebar snapshot={snapshot} busy={busy} onAction={handleAction} onMock={handleMock} onSleeperSync={handleSleeperSync} />
          </main>
        </>
      ) : null}
    </>
  );
}
