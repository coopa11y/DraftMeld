import { StatusMessage } from "../shared/ui/StatusMessage";
import { DraftOverview } from "./DraftOverview";
import { DraftSidebar } from "./DraftSidebar";
import { PlayerBoard } from "./PlayerBoard";
import type { DraftWorkspaceController } from "./useDraftWorkspace";

interface DraftWorkspaceViewProps {
  controller: DraftWorkspaceController;
}

export function DraftWorkspaceView({ controller }: DraftWorkspaceViewProps) {
  const {
    announcement,
    boardHeading,
    busy,
    error,
    errorAlert,
    handlers,
    selectedTeamNumber,
    setSelectedTeamNumber,
    snapshot,
  } = controller;

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

  const selectedTeam = snapshot.teams.find((team) => team.number === selectedTeamNumber);
  const selectedTeamIsUser = selectedTeam?.isUser ?? false;
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
          {draftActive && selectedTeam ? <span>{selectedTeam.name} on the clock</span> : null}
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
            onAction={handlers.action}
            onUndo={handlers.undo}
            onPreference={handlers.preference}
          />
          <DraftSidebar
            snapshot={snapshot}
            busy={busy || !draftActive}
            selectedTeamNumber={selectedTeamNumber}
            selectedTeamIsUser={selectedTeamIsUser}
            onSelectedTeamChange={setSelectedTeamNumber}
            onAction={handlers.action}
            onMock={handlers.mock}
            onSleeperSync={handlers.sleeperSync}
          />
        </div>
        <DraftOverview
          snapshot={snapshot}
          busy={busy}
          onCreateTrade={handlers.createTrade}
          onDeleteTrade={handlers.deleteTrade}
          onResolveTrade={handlers.resolveTrade}
          onAdvanceSeason={handlers.advanceSeason}
          onStartDraft={handlers.startDraft}
          onResetDraft={handlers.resetDraft}
          onUndoReset={handlers.undoReset}
        />
      </main>
    </>
  );
}

function sessionStatusLabel(snapshot: NonNullable<DraftWorkspaceController["snapshot"]>) {
  if (snapshot.sessionStatus === "not-started") return "Draft not started";
  if (snapshot.sessionStatus === "complete") return "Draft complete";
  return `Pick ${snapshot.pickNumber}`;
}
