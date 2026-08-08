import { useMemo, useState } from "react";
import type { DraftPickSlot, DraftPickTrade, DraftSnapshot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";

interface DraftPickTradesProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  onCreate: (teamOne: number, teamTwo: number, teamOneReceives: number[], teamTwoReceives: number[]) => Promise<void>;
  onDelete: (trade: DraftPickTrade) => Promise<void>;
}

export function DraftPickTrades({ snapshot, busy, onCreate, onDelete }: DraftPickTradesProps) {
  const [teamOne, setTeamOne] = useState(snapshot.teams[0]?.number ?? 0);
  const [teamTwo, setTeamTwo] = useState(snapshot.teams[1]?.number ?? 0);
  const [teamOneReceives, setTeamOneReceives] = useState<number[]>([]);
  const [teamTwoReceives, setTeamTwoReceives] = useState<number[]>([]);
  const [reviewing, setReviewing] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<number | null>(null);

  const teamOneName = teamName(snapshot, teamOne);
  const teamTwoName = teamName(snapshot, teamTwo);
  const picksByNumber = useMemo(
    () => new Map(snapshot.pickSlots.map((pick) => [pick.overallNumber, pick])),
    [snapshot.pickSlots],
  );
  const resetSelection = () => {
    setTeamOneReceives([]);
    setTeamTwoReceives([]);
    setReviewing(false);
  };
  const canReview = teamOne !== teamTwo && teamOneReceives.length > 0 && teamTwoReceives.length > 0;

  async function confirmTrade() {
    try {
      await onCreate(teamOne, teamTwo, teamOneReceives, teamTwoReceives);
      resetSelection();
    } catch {
      // The workspace presents the API error and keeps this review intact for correction.
    }
  }

  return (
    <Panel variant="board" aria-labelledby="pick-trades-title">
      <div className="trade-heading">
        <div>
          <p className="eyebrow">Before or during any round</p>
          <h3 id="pick-trades-title">Draft-pick trades</h3>
        </div>
        <span>{snapshot.pickTrades.length} recorded</span>
      </div>
      <p className="tool-help">
        Trade any unused picks. Choose two teams, select what each receives, then review the complete transaction.
      </p>

      <form
        className="pick-trade-form"
        onSubmit={(event) => {
          event.preventDefault();
          if (reviewing) void confirmTrade();
          else setReviewing(true);
        }}
      >
        <fieldset disabled={busy || snapshot.isComplete || reviewing}>
          <legend>Teams in this trade</legend>
          <div className="trade-team-selectors">
            <label>
              <span>First team</span>
              <select
                value={teamOne}
                onChange={(event) => {
                  setTeamOne(Number(event.target.value));
                  resetSelection();
                }}
              >
                {snapshot.teams.map((team) => (
                  <option key={team.number} value={team.number}>
                    {team.name}
                    {team.isUser ? " (you)" : ""}
                  </option>
                ))}
              </select>
            </label>
            <label>
              <span>Second team</span>
              <select
                value={teamTwo}
                onChange={(event) => {
                  setTeamTwo(Number(event.target.value));
                  resetSelection();
                }}
              >
                {snapshot.teams.map((team) => (
                  <option key={team.number} value={team.number}>
                    {team.name}
                    {team.isUser ? " (you)" : ""}
                  </option>
                ))}
              </select>
            </label>
          </div>
          {teamOne === teamTwo ? <p className="field-error">Choose two different teams.</p> : null}
        </fieldset>

        {teamOne !== teamTwo ? (
          <div className="trade-pick-columns">
            <PickSelector
              receivingTeam={teamOneName}
              currentOwner={teamTwo}
              picks={snapshot.pickSlots}
              selected={teamOneReceives}
              disabled={busy || reviewing}
              onChange={setTeamOneReceives}
            />
            <PickSelector
              receivingTeam={teamTwoName}
              currentOwner={teamOne}
              picks={snapshot.pickSlots}
              selected={teamTwoReceives}
              disabled={busy || reviewing}
              onChange={setTeamTwoReceives}
            />
          </div>
        ) : null}

        {reviewing ? (
          <section className="trade-review" aria-labelledby="trade-review-title">
            <h4 id="trade-review-title">Review this trade</h4>
            <TradeSummary
              teamOneName={teamOneName}
              teamTwoName={teamTwoName}
              teamOneReceives={teamOneReceives}
              teamTwoReceives={teamTwoReceives}
              picksByNumber={picksByNumber}
            />
            <p>Ownership changes immediately after confirmation. Completed picks cannot be included or reversed.</p>
            <div className="trade-actions">
              <Button type="submit" disabled={busy}>
                Confirm trade
              </Button>
              <Button type="button" variant="secondary" disabled={busy} onClick={() => setReviewing(false)}>
                Edit trade
              </Button>
            </div>
          </section>
        ) : (
          <Button type="submit" disabled={busy || snapshot.isComplete || !canReview}>
            Review trade
          </Button>
        )}
      </form>

      <TradeLedger
        trades={snapshot.pickTrades}
        picksByNumber={picksByNumber}
        busy={busy}
        confirmDelete={confirmDelete}
        setConfirmDelete={setConfirmDelete}
        onDelete={onDelete}
      />
    </Panel>
  );
}

function PickSelector({
  receivingTeam,
  currentOwner,
  picks,
  selected,
  disabled,
  onChange,
}: {
  receivingTeam: string;
  currentOwner: number;
  picks: DraftPickSlot[];
  selected: number[];
  disabled: boolean;
  onChange: (picks: number[]) => void;
}) {
  const available = picks.filter((pick) => !pick.isUsed && pick.ownerTeamNumber === currentOwner);
  const rounds = available.reduce<Map<number, DraftPickSlot[]>>((grouped, pick) => {
    grouped.set(pick.round, [...(grouped.get(pick.round) ?? []), pick]);
    return grouped;
  }, new Map());
  return (
    <fieldset className="trade-pick-selector" disabled={disabled}>
      <legend>Picks {receivingTeam} receives</legend>
      {available.length ? (
        [...rounds].map(([round, roundPicks]) => (
          <fieldset key={round} className="trade-round">
            <legend>Round {round}</legend>
            {roundPicks.map((pick) => (
              <label key={pick.overallNumber}>
                <input
                  type="checkbox"
                  checked={selected.includes(pick.overallNumber)}
                  onChange={(event) =>
                    onChange(
                      event.target.checked
                        ? [...selected, pick.overallNumber].sort((left, right) => left - right)
                        : selected.filter((number) => number !== pick.overallNumber),
                    )
                  }
                />
                <span>{pickLabel(pick)}</span>
              </label>
            ))}
          </fieldset>
        ))
      ) : (
        <p className="empty-state">This team has no unused picks available to trade.</p>
      )}
    </fieldset>
  );
}

function TradeLedger({
  trades,
  picksByNumber,
  busy,
  confirmDelete,
  setConfirmDelete,
  onDelete,
}: {
  trades: DraftPickTrade[];
  picksByNumber: Map<number, DraftPickSlot>;
  busy: boolean;
  confirmDelete: number | null;
  setConfirmDelete: (id: number | null) => void;
  onDelete: (trade: DraftPickTrade) => Promise<void>;
}) {
  return (
    <section className="trade-ledger" aria-labelledby="trade-ledger-title">
      <h4 id="trade-ledger-title">Trade ledger</h4>
      {trades.length ? (
        <ol>
          {trades.map((trade) => (
            <li key={trade.id}>
              <TradeSummary
                teamOneName={trade.teamOneName}
                teamTwoName={trade.teamTwoName}
                teamOneReceives={trade.teamOneReceives}
                teamTwoReceives={trade.teamTwoReceives}
                picksByNumber={picksByNumber}
              />
              {confirmDelete === trade.id ? (
                <div
                  className="trade-reversal"
                  role="group"
                  aria-label={`Reverse trade between ${trade.teamOneName} and ${trade.teamTwoName}`}
                >
                  <p>Reverse this entire trade? This restores every unused pick to its previous owner.</p>
                  <div className="trade-actions">
                    <Button
                      type="button"
                      variant="danger"
                      disabled={busy}
                      onClick={() =>
                        void onDelete(trade)
                          .then(() => setConfirmDelete(null))
                          .catch(() => undefined)
                      }
                    >
                      Confirm reversal
                    </Button>
                    <Button type="button" variant="secondary" disabled={busy} onClick={() => setConfirmDelete(null)}>
                      Keep trade
                    </Button>
                  </div>
                </div>
              ) : (
                <Button type="button" variant="secondary" disabled={busy} onClick={() => setConfirmDelete(trade.id)}>
                  Reverse trade
                </Button>
              )}
            </li>
          ))}
        </ol>
      ) : (
        <p className="empty-state">No draft-pick trades recorded.</p>
      )}
    </section>
  );
}

function TradeSummary({
  teamOneName,
  teamTwoName,
  teamOneReceives,
  teamTwoReceives,
  picksByNumber,
}: {
  teamOneName: string;
  teamTwoName: string;
  teamOneReceives: number[];
  teamTwoReceives: number[];
  picksByNumber: Map<number, DraftPickSlot>;
}) {
  return (
    <dl className="trade-summary">
      <div>
        <dt>{teamOneName} receives</dt>
        <dd>{pickList(teamOneReceives, picksByNumber)}</dd>
      </div>
      <div>
        <dt>{teamTwoName} receives</dt>
        <dd>{pickList(teamTwoReceives, picksByNumber)}</dd>
      </div>
    </dl>
  );
}

function teamName(snapshot: DraftSnapshot, teamNumber: number) {
  return snapshot.teams.find((team) => team.number === teamNumber)?.name ?? `Team ${teamNumber}`;
}

function pickLabel(pick: DraftPickSlot) {
  return `Round ${pick.round}, pick ${pick.pickInRound}, overall ${pick.overallNumber}`;
}

function pickList(picks: number[], indexed: Map<number, DraftPickSlot>) {
  return picks
    .map((number) => indexed.get(number))
    .filter((pick): pick is DraftPickSlot => Boolean(pick))
    .map(pickLabel)
    .join("; ");
}
