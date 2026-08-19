import { useState } from "react";
import type { DraftSnapshot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { Panel } from "../shared/ui/Panel";
import { DraftPositionDialog } from "./DraftPositionDialog";

interface DraftSessionControlsProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  onStart: (draftPosition?: number) => Promise<void>;
  onReset: (confirmation: string) => Promise<void>;
  onUndoReset: () => Promise<void>;
}

export function DraftSessionControls({ snapshot, busy, onStart, onReset, onUndoReset }: DraftSessionControlsProps) {
  const [reviewingReset, setReviewingReset] = useState(false);
  const [understood, setUnderstood] = useState(false);
  const [confirmation, setConfirmation] = useState("");
  const [launchOpen, setLaunchOpen] = useState(false);
  const userPosition = snapshot.draftPosition;
  const canConfirmReset = understood && confirmation === snapshot.leagueName;

  function cancelReset() {
    setReviewingReset(false);
    setUnderstood(false);
    setConfirmation("");
  }

  async function submitReset() {
    if (!canConfirmReset) return;
    try {
      await onReset(confirmation);
      cancelReset();
    } catch {
      // The workspace presents the API error and leaves this review open for correction.
    }
  }

  return (
    <Panel variant="form" className="draft-session-controls" aria-labelledby="draft-session-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Current season</p>
          <h3 id="draft-session-title">Draft session</h3>
        </div>
        <strong>{sessionLabel(snapshot.sessionStatus)}</strong>
      </div>

      {snapshot.sessionStatus === "not-started" ? (
        <>
          <p>Review the setup, then start when your league is ready to make selections.</p>
          <dl className="draft-session-summary">
            <div>
              <dt>Season and format</dt>
              <dd>
                {snapshot.season} {snapshot.leagueFormat} · {snapshot.draftType}
              </dd>
            </div>
            <div>
              <dt>League size</dt>
              <dd>
                {snapshot.teams.length} teams · {snapshot.totalPicks} selections
              </dd>
            </div>
            {snapshot.draftType !== "auction" ? (
              <div>
                <dt>Your draft position</dt>
                <dd>{userPosition > 0 ? `${userPosition} of ${snapshot.teams.length}` : "Not assigned"}</dd>
              </div>
            ) : null}
          </dl>
          <div className="form-actions">
            <Button
              variant="primary"
              disabled={busy}
              onClick={() => (snapshot.draftType === "auction" ? void onStart() : setLaunchOpen(true))}
            >
              Start draft
            </Button>
            {snapshot.canUndoReset ? (
              <Button disabled={busy} onClick={() => void onUndoReset()}>
                Undo last reset
              </Button>
            ) : null}
          </div>
          {snapshot.canUndoReset ? (
            <p className="field-help">The reset can be undone until you start this draft again.</p>
          ) : null}
        </>
      ) : reviewingReset ? (
        <div className="draft-reset-confirmation">
          <h4>Reset the current draft?</h4>
          <p>
            This clears current-season selections. League and team settings, draft order, trades, prior dynasty seasons,
            and future assets stay intact.
          </p>
          <label className="confirmation-check">
            <input
              type="checkbox"
              checked={understood}
              disabled={busy}
              onChange={(event) => setUnderstood(event.target.checked)}
            />
            <span>I understand that current-season selections will be removed.</span>
          </label>
          <FormField label={`Type ${snapshot.leagueName} to confirm`} help="The name must match exactly.">
            <input
              type="text"
              autoComplete="off"
              value={confirmation}
              disabled={busy}
              onChange={(event) => setConfirmation(event.target.value)}
            />
          </FormField>
          <div className="form-actions">
            <Button variant="danger" disabled={busy || !canConfirmReset} onClick={() => void submitReset()}>
              Reset current draft
            </Button>
            <Button disabled={busy} onClick={cancelReset}>
              Cancel
            </Button>
          </div>
        </div>
      ) : (
        <div className="form-actions">
          <Button variant="dangerText" disabled={busy || !snapshot.canReset} onClick={() => setReviewingReset(true)}>
            Reset current draft
          </Button>
        </div>
      )}
      {launchOpen ? (
        <DraftPositionDialog
          busy={busy}
          kind="real"
          teamCount={snapshot.teams.length}
          initialPosition={userPosition}
          onClose={() => setLaunchOpen(false)}
          onConfirm={async (draftPosition) => {
            await onStart(draftPosition);
            setLaunchOpen(false);
          }}
        />
      ) : null}
    </Panel>
  );
}

function sessionLabel(status: DraftSnapshot["sessionStatus"]) {
  if (status === "in-progress") return "In progress";
  if (status === "complete") return "Complete";
  return "Not started";
}
