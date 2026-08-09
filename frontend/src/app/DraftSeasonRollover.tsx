import { useState, type FormEvent } from "react";
import type { DraftSnapshot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";

interface DraftSeasonRolloverProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  onAdvance: (season: number, draftType: DraftSnapshot["draftType"], draftOrder: number[]) => Promise<void>;
}

export function DraftSeasonRollover({ snapshot, busy, onAdvance }: DraftSeasonRolloverProps) {
  const [open, setOpen] = useState(false);
  const [draftType, setDraftType] = useState(snapshot.draftType);
  const [draftOrder, setDraftOrder] = useState(snapshot.draftOrder);
  const [acknowledged, setAcknowledged] = useState(false);
  const nextSeason = snapshot.season + 1;
  const needsEarlyCloseConfirmation = snapshot.history.length > 0 && !snapshot.isComplete;

  function moveTeam(team: number, position: number) {
    const updated = [...draftOrder];
    const current = updated.indexOf(team);
    [updated[current], updated[position]] = [updated[position], updated[current]];
    setDraftOrder(updated);
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    await onAdvance(nextSeason, draftType, draftOrder);
    setOpen(false);
    setAcknowledged(false);
  }

  return (
    <Panel variant="board" aria-labelledby="season-rollover-title">
      <div className="trade-heading">
        <div>
          <p className="eyebrow">Dynasty lifecycle</p>
          <h3 id="season-rollover-title">Season {snapshot.season}</h3>
        </div>
        {!open ? <Button onClick={() => setOpen(true)}>Close season and start {nextSeason}</Button> : null}
      </div>
      <p className="tool-help">
        Starting a new season keeps franchise rosters, trades, and future-pick ownership while opening a clean draft
        ledger.
      </p>
      {open ? (
        <form className="season-rollover-form" onSubmit={submit}>
          <h4>Review the {nextSeason} draft setup</h4>
          <label>
            <span>Draft format</span>
            <select
              value={draftType}
              onChange={(event) => setDraftType(event.target.value as DraftSnapshot["draftType"])}
            >
              <option value="snake">Snake</option>
              <option value="linear">Linear</option>
              <option value="auction">Auction</option>
            </select>
          </label>
          {draftType !== "auction" ? (
            <fieldset>
              <legend>{nextSeason} draft order</legend>
              <p className="field-help">Each permanent franchise must appear once. Selecting a team swaps positions.</p>
              <div className="form-grid">
                {draftOrder.map((teamNumber, index) => (
                  <label key={index}>
                    <span>Pick {index + 1}</span>
                    <select value={teamNumber} onChange={(event) => moveTeam(Number(event.target.value), index)}>
                      {snapshot.teams.map((team) => (
                        <option key={team.number} value={team.number}>
                          {team.name}
                          {team.isUser ? " (you)" : ""}
                        </option>
                      ))}
                    </select>
                  </label>
                ))}
              </div>
            </fieldset>
          ) : null}
          {needsEarlyCloseConfirmation ? (
            <label className="checkbox-label season-warning">
              <input
                type="checkbox"
                checked={acknowledged}
                onChange={(event) => setAcknowledged(event.target.checked)}
              />
              <span>
                This draft is not complete. I understand that its existing picks will remain in the {snapshot.season}{" "}
                history.
              </span>
            </label>
          ) : null}
          <div className="trade-actions">
            <Button type="submit" variant="primary" disabled={busy || (needsEarlyCloseConfirmation && !acknowledged)}>
              Start {nextSeason}
            </Button>
            <Button type="button" disabled={busy} onClick={() => setOpen(false)}>
              Cancel
            </Button>
          </div>
        </form>
      ) : null}
    </Panel>
  );
}
