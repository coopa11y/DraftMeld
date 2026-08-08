import { useMemo, useState } from "react";
import type { DraftPickSlot, DraftPickTrade, DraftSnapshot, FutureDraftPick } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";

interface DraftPickTradesProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  onCreate: (
    teamOne: number,
    teamTwo: number,
    teamOneReceives: number[],
    teamTwoReceives: number[],
    teamOneFuture: FutureDraftPick[],
    teamTwoFuture: FutureDraftPick[],
    teamOneBudget: number,
    teamTwoBudget: number,
  ) => Promise<void>;
  onDelete: (trade: DraftPickTrade) => Promise<void>;
}

export function DraftPickTrades({ snapshot, busy, onCreate, onDelete }: DraftPickTradesProps) {
  const [teamOne, setTeamOne] = useState(snapshot.teams[0]?.number ?? 0);
  const [teamTwo, setTeamTwo] = useState(snapshot.teams[1]?.number ?? 0);
  const [teamOneSelected, setTeamOneSelected] = useState<string[]>([]);
  const [teamTwoSelected, setTeamTwoSelected] = useState<string[]>([]);
  const [teamOneBudget, setTeamOneBudget] = useState(0);
  const [teamTwoBudget, setTeamTwoBudget] = useState(0);
  const [reviewing, setReviewing] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<number | null>(null);

  const slotsByKey = useMemo(
    () => new Map(snapshot.pickSlots.map((pick) => [pickKey(pick), pick])),
    [snapshot.pickSlots],
  );
  const teamOneSlots = selectedSlots(teamOneSelected, slotsByKey);
  const teamTwoSlots = selectedSlots(teamTwoSelected, slotsByKey);
  const firstName = teamName(snapshot, teamOne);
  const secondName = teamName(snapshot, teamTwo);
  const firstAssets = assetLabels(teamOneSlots, teamOneBudget, snapshot.leagueFormat);
  const secondAssets = assetLabels(teamTwoSlots, teamTwoBudget, snapshot.leagueFormat);
  const canReview = teamOne !== teamTwo && firstAssets.length + secondAssets.length > 0;

  function resetSelection() {
    setTeamOneSelected([]);
    setTeamTwoSelected([]);
    setTeamOneBudget(0);
    setTeamTwoBudget(0);
    setReviewing(false);
  }

  async function confirmTrade() {
    try {
      await onCreate(
        teamOne,
        teamTwo,
        currentPicks(teamOneSlots),
        currentPicks(teamTwoSlots),
        futurePicks(teamOneSlots),
        futurePicks(teamTwoSlots),
        teamOneBudget,
        teamTwoBudget,
      );
      resetSelection();
    } catch {
      // The workspace presents the API error and keeps this review intact for correction.
    }
  }

  return (
    <Panel variant="board" aria-labelledby="pick-trades-title">
      <div className="trade-heading">
        <div>
          <p className="eyebrow">League-rule aware</p>
          <h3 id="pick-trades-title">Draft asset trades</h3>
        </div>
        <span>{snapshot.pickTrades.length} recorded</span>
      </div>
      <p className="tool-help">{tradeHelp(snapshot)}</p>

      <form
        className="pick-trade-form"
        onSubmit={(event) => {
          event.preventDefault();
          if (reviewing) void confirmTrade();
          else setReviewing(true);
        }}
      >
        <fieldset disabled={busy || reviewing}>
          <legend>Teams in this trade</legend>
          <div className="trade-team-selectors">
            <TeamSelect
              label="First team"
              value={teamOne}
              snapshot={snapshot}
              onChange={(team) => {
                setTeamOne(team);
                resetSelection();
              }}
            />
            <TeamSelect
              label="Second team"
              value={teamTwo}
              snapshot={snapshot}
              onChange={(team) => {
                setTeamTwo(team);
                resetSelection();
              }}
            />
          </div>
          {teamOne === teamTwo ? <p className="field-error">Choose two different teams.</p> : null}
        </fieldset>

        {teamOne !== teamTwo ? (
          <div className="trade-pick-columns">
            <PickSelector
              receivingTeam={firstName}
              currentOwner={teamTwo}
              sourceTeam={snapshot.teams.find((team) => team.number === teamTwo)}
              snapshot={snapshot}
              selected={teamOneSelected}
              budget={teamOneBudget}
              disabled={busy || reviewing}
              onChange={setTeamOneSelected}
              onBudgetChange={setTeamOneBudget}
            />
            <PickSelector
              receivingTeam={secondName}
              currentOwner={teamOne}
              sourceTeam={snapshot.teams.find((team) => team.number === teamOne)}
              snapshot={snapshot}
              selected={teamTwoSelected}
              budget={teamTwoBudget}
              disabled={busy || reviewing}
              onChange={setTeamTwoSelected}
              onBudgetChange={setTeamTwoBudget}
            />
          </div>
        ) : null}

        {reviewing ? (
          <section className="trade-review" aria-labelledby="trade-review-title">
            <h4 id="trade-review-title">Review this trade</h4>
            <TradeSummary
              firstName={firstName}
              secondName={secondName}
              firstAssets={firstAssets}
              secondAssets={secondAssets}
            />
            <p>Ownership changes immediately after confirmation. Used or expired assets cannot be reversed.</p>
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
          <Button type="submit" disabled={busy || !canReview}>
            Review trade
          </Button>
        )}
      </form>

      <TradeLedger
        snapshot={snapshot}
        busy={busy}
        confirmDelete={confirmDelete}
        setConfirmDelete={setConfirmDelete}
        onDelete={onDelete}
      />
    </Panel>
  );
}

function TeamSelect({
  label,
  value,
  snapshot,
  onChange,
}: {
  label: string;
  value: number;
  snapshot: DraftSnapshot;
  onChange: (team: number) => void;
}) {
  return (
    <label>
      <span>{label}</span>
      <select value={value} onChange={(event) => onChange(Number(event.target.value))}>
        {snapshot.teams.map((team) => (
          <option key={team.number} value={team.number}>
            {team.name}
            {team.isUser ? " (you)" : ""}
          </option>
        ))}
      </select>
    </label>
  );
}

function PickSelector({
  receivingTeam,
  currentOwner,
  sourceTeam,
  snapshot,
  selected,
  budget,
  disabled,
  onChange,
  onBudgetChange,
}: {
  receivingTeam: string;
  currentOwner: number;
  sourceTeam?: DraftSnapshot["teams"][number];
  snapshot: DraftSnapshot;
  selected: string[];
  budget: number;
  disabled: boolean;
  onChange: (picks: string[]) => void;
  onBudgetChange: (amount: number) => void;
}) {
  const available = snapshot.pickSlots.filter((pick) => !pick.isUsed && pick.ownerTeamNumber === currentOwner);
  const groups = available.reduce<Map<string, DraftPickSlot[]>>((grouped, pick) => {
    const key = `${pick.season}:${pick.round}`;
    grouped.set(key, [...(grouped.get(key) ?? []), pick]);
    return grouped;
  }, new Map());
  return (
    <fieldset className="trade-pick-selector" disabled={disabled}>
      <legend>Assets {receivingTeam} receives</legend>
      {snapshot.auctionBudgetTrades ? (
        <label className="trade-budget-field">
          <span>Auction budget from {sourceTeam?.name ?? "the other team"}</span>
          <input
            type="number"
            aria-label={`Auction budget from ${sourceTeam?.name ?? "the other team"}`}
            min="0"
            max={sourceTeam?.auctionBudgetRemaining}
            step="1"
            value={budget}
            onChange={(event) => onBudgetChange(Number(event.target.value))}
          />
          <small>
            {sourceTeam?.name ?? "That team"} has ${sourceTeam?.auctionBudgetRemaining.toFixed(0) ?? "0"} available.
          </small>
        </label>
      ) : null}
      {available.length ? (
        [...groups].map(([key, picks]) => (
          <fieldset key={key} className="trade-round">
            <legend>{groupLabel(picks[0], snapshot.leagueFormat)}</legend>
            {picks.map((pick) => {
              const key = pickKey(pick);
              return (
                <label key={key}>
                  <input
                    type="checkbox"
                    checked={selected.includes(key)}
                    onChange={(event) =>
                      onChange(event.target.checked ? [...selected, key] : selected.filter((item) => item !== key))
                    }
                  />
                  <span>{pickLabel(pick, snapshot.leagueFormat)}</span>
                </label>
              );
            })}
          </fieldset>
        ))
      ) : snapshot.auctionBudgetTrades ? null : (
        <p className="empty-state">This team has no unused picks available to trade.</p>
      )}
    </fieldset>
  );
}

function TradeLedger({
  snapshot,
  busy,
  confirmDelete,
  setConfirmDelete,
  onDelete,
}: {
  snapshot: DraftSnapshot;
  busy: boolean;
  confirmDelete: number | null;
  setConfirmDelete: (id: number | null) => void;
  onDelete: (trade: DraftPickTrade) => Promise<void>;
}) {
  const currentSlots = new Map(
    snapshot.pickSlots
      .filter((slot) => slot.overallNumber > 0)
      .map((slot) => [`${slot.season}:${slot.overallNumber}`, slot]),
  );
  return (
    <section className="trade-ledger" aria-labelledby="trade-ledger-title">
      <h4 id="trade-ledger-title">Trade ledger</h4>
      {snapshot.pickTrades.length ? (
        <ol>
          {snapshot.pickTrades.map((trade) => (
            <li key={trade.id}>
              <TradeSummary
                firstName={trade.teamOneName}
                secondName={trade.teamTwoName}
                firstAssets={storedTradeAssets(trade, true, currentSlots, snapshot.leagueFormat)}
                secondAssets={storedTradeAssets(trade, false, currentSlots, snapshot.leagueFormat)}
              />
              {confirmDelete === trade.id ? (
                <div
                  className="trade-reversal"
                  role="group"
                  aria-label={`Reverse trade between ${trade.teamOneName} and ${trade.teamTwoName}`}
                >
                  <p>Reverse this entire trade? This restores every eligible asset to its previous owner.</p>
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
        <p className="empty-state">No draft trades recorded.</p>
      )}
    </section>
  );
}

function TradeSummary({
  firstName,
  secondName,
  firstAssets,
  secondAssets,
}: {
  firstName: string;
  secondName: string;
  firstAssets: string[];
  secondAssets: string[];
}) {
  return (
    <dl className="trade-summary">
      <div>
        <dt>{firstName} receives</dt>
        <dd>{firstAssets.join("; ") || "External consideration not tracked in DraftMeld"}</dd>
      </div>
      <div>
        <dt>{secondName} receives</dt>
        <dd>{secondAssets.join("; ") || "External consideration not tracked in DraftMeld"}</dd>
      </div>
    </dl>
  );
}

function selectedSlots(selected: string[], indexed: Map<string, DraftPickSlot>) {
  return selected.map((key) => indexed.get(key)).filter((pick): pick is DraftPickSlot => Boolean(pick));
}

function currentPicks(slots: DraftPickSlot[]) {
  return slots.filter((pick) => pick.overallNumber > 0).map((pick) => pick.overallNumber);
}

function futurePicks(slots: DraftPickSlot[]): FutureDraftPick[] {
  return slots
    .filter((pick) => pick.overallNumber === 0)
    .map((pick) => ({
      season: pick.season,
      round: pick.round,
      originalTeamNumber: pick.originalTeamNumber,
      originalTeamName: pick.originalTeamName,
    }));
}

function assetLabels(slots: DraftPickSlot[], budget: number, format: DraftSnapshot["leagueFormat"]) {
  const labels = slots.map((pick) => pickLabel(pick, format));
  if (budget > 0) labels.push(`$${budget.toFixed(0)} auction budget`);
  return labels;
}

function storedTradeAssets(
  trade: DraftPickTrade,
  first: boolean,
  currentSlots: Map<string, DraftPickSlot>,
  format: DraftSnapshot["leagueFormat"],
) {
  const current = first ? trade.teamOneReceives : trade.teamTwoReceives;
  const future = first ? trade.teamOneFuturePicks : trade.teamTwoFuturePicks;
  const budget = first ? trade.teamOneAuctionBudget : trade.teamTwoAuctionBudget;
  const labels = current.map((number) => {
    const slot = currentSlots.get(`${trade.season}:${number}`);
    return slot ? pickLabel(slot, format) : `${trade.season} overall pick ${number}`;
  });
  labels.push(...future.map((pick) => `${pick.season}, round ${pick.round}, ${pick.originalTeamName} original pick`));
  if (budget > 0) labels.push(`$${budget.toFixed(0)} auction budget`);
  return labels;
}

function tradeHelp(snapshot: DraftSnapshot) {
  if (snapshot.leagueFormat === "redraft") {
    return snapshot.draftType === "auction"
      ? "This redraft league shows only auction budget enabled by its rules; future picks stay out of the way."
      : `This redraft league shows only unused ${snapshot.season} picks. Future seasons stay out of the way.`;
  }
  return `This dynasty league tracks unused ${snapshot.season} assets and future rookie picks across seasons.`;
}

function teamName(snapshot: DraftSnapshot, teamNumber: number) {
  return snapshot.teams.find((team) => team.number === teamNumber)?.name ?? `Team ${teamNumber}`;
}

function pickKey(pick: DraftPickSlot) {
  return pick.overallNumber > 0
    ? `${pick.season}:overall:${pick.overallNumber}`
    : `${pick.season}:round:${pick.round}:team:${pick.originalTeamNumber}`;
}

function groupLabel(pick: DraftPickSlot, format: DraftSnapshot["leagueFormat"]) {
  return format === "dynasty" ? `${pick.season} round ${pick.round}` : `Round ${pick.round}`;
}

function pickLabel(pick: DraftPickSlot, format: DraftSnapshot["leagueFormat"]) {
  if (pick.overallNumber === 0) return `${pick.season}, round ${pick.round}, ${pick.originalTeamName} original pick`;
  const current = `Round ${pick.round}, pick ${pick.pickInRound}, overall ${pick.overallNumber}`;
  return format === "dynasty" ? `${pick.season}, ${current.toLowerCase()}` : current;
}
