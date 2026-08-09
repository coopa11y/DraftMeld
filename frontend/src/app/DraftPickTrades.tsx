import { useMemo, useState } from "react";
import type {
  BudgetAsset,
  DraftPickSlot,
  DraftPickTrade,
  DraftSnapshot,
  FutureDraftPick,
  Player,
} from "../shared/api/types";
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
    teamOnePlayers: string[],
    teamTwoPlayers: string[],
    teamOneBudgets: BudgetAsset[],
    teamTwoBudgets: BudgetAsset[],
  ) => Promise<void>;
  onDelete: (trade: DraftPickTrade) => Promise<void>;
  onResolve: (trade: DraftPickTrade, pick: FutureDraftPick, status: "met" | "not-met") => Promise<void>;
}

export function DraftPickTrades({ snapshot, busy, onCreate, onDelete, onResolve }: DraftPickTradesProps) {
  const [teamOne, setTeamOne] = useState(snapshot.teams[0]?.number ?? 0);
  const [teamTwo, setTeamTwo] = useState(snapshot.teams[1]?.number ?? 0);
  const [teamOneSelected, setTeamOneSelected] = useState<string[]>([]);
  const [teamTwoSelected, setTeamTwoSelected] = useState<string[]>([]);
  const [teamOnePlayers, setTeamOnePlayers] = useState<string[]>([]);
  const [teamTwoPlayers, setTeamTwoPlayers] = useState<string[]>([]);
  const [teamOneBudgets, setTeamOneBudgets] = useState<BudgetAsset[]>([]);
  const [teamTwoBudgets, setTeamTwoBudgets] = useState<BudgetAsset[]>([]);
  const [teamOneCondition, setTeamOneCondition] = useState("");
  const [teamTwoCondition, setTeamTwoCondition] = useState("");
  const [reviewing, setReviewing] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<number | null>(null);
  const [ledgerSeason, setLedgerSeason] = useState("all");

  const slotsByKey = useMemo(
    () => new Map(snapshot.pickSlots.map((pick) => [pickKey(pick), pick])),
    [snapshot.pickSlots],
  );
  const playersByID = useMemo(
    () => new Map(snapshot.teams.flatMap((team) => team.roster).map((player) => [player.id, player])),
    [snapshot.teams],
  );
  const teamOneSlots = selectedSlots(teamOneSelected, slotsByKey);
  const teamTwoSlots = selectedSlots(teamTwoSelected, slotsByKey);
  const firstName = teamName(snapshot, teamOne);
  const secondName = teamName(snapshot, teamTwo);
  const firstAssets = assetLabels(
    teamOneSlots,
    teamOnePlayers,
    teamOneBudgets,
    playersByID,
    snapshot.leagueFormat,
    teamOneCondition,
  );
  const secondAssets = assetLabels(
    teamTwoSlots,
    teamTwoPlayers,
    teamTwoBudgets,
    playersByID,
    snapshot.leagueFormat,
    teamTwoCondition,
  );
  const canReview = teamOne !== teamTwo && firstAssets.length + secondAssets.length > 0;

  function resetSelection() {
    setTeamOneSelected([]);
    setTeamTwoSelected([]);
    setTeamOnePlayers([]);
    setTeamTwoPlayers([]);
    setTeamOneBudgets([]);
    setTeamTwoBudgets([]);
    setTeamOneCondition("");
    setTeamTwoCondition("");
    setReviewing(false);
  }

  async function confirmTrade() {
    try {
      await onCreate(
        teamOne,
        teamTwo,
        currentPicks(teamOneSlots),
        currentPicks(teamTwoSlots),
        futurePicks(teamOneSlots, teamOneCondition),
        futurePicks(teamTwoSlots, teamTwoCondition),
        teamOnePlayers,
        teamTwoPlayers,
        teamOneBudgets,
        teamTwoBudgets,
      );
      resetSelection();
    } catch {
      // The workspace presents the API error and keeps the review intact for correction.
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
          else if (canReview) setReviewing(true);
        }}
      >
        <fieldset disabled={busy || reviewing}>
          <legend>Trade partners</legend>
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
            <AssetSelector
              receivingTeam={firstName}
              sourceTeam={snapshot.teams.find((team) => team.number === teamTwo)}
              snapshot={snapshot}
              selected={teamOneSelected}
              selectedPlayers={teamOnePlayers}
              budgets={teamOneBudgets}
              condition={teamOneCondition}
              disabled={busy || reviewing}
              onChange={setTeamOneSelected}
              onPlayersChange={setTeamOnePlayers}
              onBudgetsChange={setTeamOneBudgets}
              onConditionChange={setTeamOneCondition}
            />
            <AssetSelector
              receivingTeam={secondName}
              sourceTeam={snapshot.teams.find((team) => team.number === teamOne)}
              snapshot={snapshot}
              selected={teamTwoSelected}
              selectedPlayers={teamTwoPlayers}
              budgets={teamTwoBudgets}
              condition={teamTwoCondition}
              disabled={busy || reviewing}
              onChange={setTeamTwoSelected}
              onPlayersChange={setTeamTwoPlayers}
              onBudgetsChange={setTeamTwoBudgets}
              onConditionChange={setTeamTwoCondition}
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
            <p>Ownership changes immediately. Pending conditional picks stay locked until their result is recorded.</p>
            <div className="trade-actions">
              <Button type="submit" disabled={busy}>
                Confirm trade
              </Button>
              <Button type="button" onClick={() => setReviewing(false)} disabled={busy}>
                Go back and edit
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
        playersByID={playersByID}
        ledgerSeason={ledgerSeason}
        setLedgerSeason={setLedgerSeason}
        busy={busy}
        confirmDelete={confirmDelete}
        setConfirmDelete={setConfirmDelete}
        onDelete={onDelete}
        onResolve={onResolve}
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

interface AssetSelectorProps {
  receivingTeam: string;
  sourceTeam?: DraftSnapshot["teams"][number];
  snapshot: DraftSnapshot;
  selected: string[];
  selectedPlayers: string[];
  budgets: BudgetAsset[];
  condition: string;
  disabled: boolean;
  onChange: (picks: string[]) => void;
  onPlayersChange: (players: string[]) => void;
  onBudgetsChange: (budgets: BudgetAsset[]) => void;
  onConditionChange: (condition: string) => void;
}

function AssetSelector(props: AssetSelectorProps) {
  const {
    receivingTeam,
    sourceTeam,
    snapshot,
    selected,
    selectedPlayers,
    budgets,
    condition,
    disabled,
    onChange,
    onPlayersChange,
    onBudgetsChange,
    onConditionChange,
  } = props;
  const available = snapshot.pickSlots.filter((pick) => !pick.isUsed && pick.ownerTeamNumber === sourceTeam?.number);
  const groups = available.reduce<Map<string, DraftPickSlot[]>>((grouped, pick) => {
    const key = `${pick.season}:${pick.round}`;
    grouped.set(key, [...(grouped.get(key) ?? []), pick]);
    return grouped;
  }, new Map());
  const selectedFuture = available.some((pick) => selected.includes(pickKey(pick)) && pick.overallNumber === 0);
  const balanceOptions = snapshot.budgetBalances.filter((balance) => balance.teamNumber === sourceTeam?.number);

  return (
    <fieldset className="trade-pick-selector" disabled={disabled}>
      <legend>Assets {receivingTeam} receives</legend>
      {available.length
        ? [...groups].map(([key, picks]) => (
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
        : null}

      {selectedFuture ? (
        <label className="trade-condition-field">
          <span>Condition for selected future picks received by {receivingTeam} (optional)</span>
          <textarea
            aria-label={`Condition for selected future picks received by ${receivingTeam} (optional)`}
            maxLength={300}
            value={condition}
            onChange={(event) => onConditionChange(event.target.value)}
          />
          <small>Example: transfers only if the player appears in eight games.</small>
        </label>
      ) : null}

      {snapshot.leagueFormat === "dynasty" && sourceTeam?.roster.length ? (
        <details className="trade-asset-details">
          <summary>Players from {sourceTeam.name}</summary>
          <div className="trade-player-list">
            {sourceTeam.roster.map((player) => (
              <label key={player.id}>
                <input
                  type="checkbox"
                  checked={selectedPlayers.includes(player.id)}
                  onChange={(event) =>
                    onPlayersChange(
                      event.target.checked
                        ? [...selectedPlayers, player.id]
                        : selectedPlayers.filter((id) => id !== player.id),
                    )
                  }
                />
                <span>
                  {player.name}, {player.position}
                </span>
              </label>
            ))}
          </div>
        </details>
      ) : null}

      {balanceOptions.length ? (
        <details className="trade-asset-details">
          <summary>Auction and FAAB budgets</summary>
          <div className="trade-budget-grid">
            {balanceOptions.map((balance) => {
              const current =
                budgets.find((asset) => asset.kind === balance.kind && asset.season === balance.season)?.amount ?? 0;
              const label = `${balance.season} ${balance.kind === "faab" ? "FAAB" : "auction budget"} from ${sourceTeam?.name ?? "the other team"}`;
              return (
                <label key={`${balance.kind}:${balance.season}`} className="trade-budget-field">
                  <span>{label}</span>
                  <input
                    type="number"
                    aria-label={label}
                    min="0"
                    max={balance.remaining}
                    step="1"
                    value={current}
                    onChange={(event) =>
                      onBudgetsChange(updateBudget(budgets, balance.kind, balance.season, Number(event.target.value)))
                    }
                  />
                  <small>${balance.remaining.toFixed(0)} available.</small>
                </label>
              );
            })}
          </div>
        </details>
      ) : null}

      {!available.length && !sourceTeam?.roster.length && !balanceOptions.length ? (
        <p className="empty-state">This team has no eligible assets available to trade.</p>
      ) : null}
    </fieldset>
  );
}

function TradeLedger(props: {
  snapshot: DraftSnapshot;
  playersByID: Map<string, Player>;
  ledgerSeason: string;
  setLedgerSeason: (season: string) => void;
  busy: boolean;
  confirmDelete: number | null;
  setConfirmDelete: (id: number | null) => void;
  onDelete: (trade: DraftPickTrade) => Promise<void>;
  onResolve: DraftPickTradesProps["onResolve"];
}) {
  const {
    snapshot,
    playersByID,
    ledgerSeason,
    setLedgerSeason,
    busy,
    confirmDelete,
    setConfirmDelete,
    onDelete,
    onResolve,
  } = props;
  const seasons = [...new Set(snapshot.pickTrades.map((trade) => trade.season))].sort((a, b) => b - a);
  const visibleTrades =
    ledgerSeason === "all"
      ? snapshot.pickTrades
      : snapshot.pickTrades.filter((trade) => trade.season === Number(ledgerSeason));
  const currentSlots = new Map(
    snapshot.pickSlots
      .filter((slot) => slot.overallNumber > 0)
      .map((slot) => [`${slot.season}:${slot.overallNumber}`, slot]),
  );
  return (
    <section className="trade-ledger" aria-labelledby="trade-ledger-title">
      <div className="trade-heading">
        <h4 id="trade-ledger-title">Trade ledger</h4>
        {snapshot.leagueFormat === "dynasty" && seasons.length > 0 ? (
          <label>
            <span>Trade season</span>
            <select value={ledgerSeason} onChange={(event) => setLedgerSeason(event.target.value)}>
              <option value="all">All seasons</option>
              {seasons.map((season) => (
                <option key={season} value={season}>
                  {season}
                </option>
              ))}
            </select>
          </label>
        ) : null}
      </div>
      {visibleTrades.length ? (
        <ol>
          {visibleTrades.map((trade) => (
            <li key={trade.id}>
              <p className="eyebrow">Recorded in {trade.season}</p>
              <TradeSummary
                firstName={trade.teamOneName}
                secondName={trade.teamTwoName}
                firstAssets={storedTradeAssets(trade, true, currentSlots, playersByID, snapshot.leagueFormat)}
                secondAssets={storedTradeAssets(trade, false, currentSlots, playersByID, snapshot.leagueFormat)}
              />
              {pendingConditions(trade).map((pick) => (
                <div
                  className="conditional-pick-actions"
                  key={pickKeyFromFuture(pick)}
                  role="group"
                  aria-label={`Resolve condition for ${futurePickLabel(pick)}`}
                >
                  <p>
                    <strong>Pending:</strong> {pick.condition}
                  </p>
                  <Button type="button" disabled={busy} onClick={() => void onResolve(trade, pick, "met")}>
                    Mark condition met
                  </Button>
                  <Button type="button" disabled={busy} onClick={() => void onResolve(trade, pick, "not-met")}>
                    Mark condition not met
                  </Button>
                </div>
              ))}
              {confirmDelete === trade.id ? (
                <div
                  className="trade-reversal"
                  role="group"
                  aria-label={`Reverse trade between ${trade.teamOneName} and ${trade.teamTwoName}`}
                >
                  <p>Reverse this entire trade? This restores every eligible asset to its previous owner.</p>
                  <div className="trade-actions">
                    <Button type="button" variant="danger" disabled={busy} onClick={() => void onDelete(trade)}>
                      Confirm reversal
                    </Button>
                    <Button type="button" disabled={busy} onClick={() => setConfirmDelete(null)}>
                      Keep trade
                    </Button>
                  </div>
                </div>
              ) : (
                <Button type="button" disabled={busy} onClick={() => setConfirmDelete(trade.id)}>
                  Reverse trade
                </Button>
              )}
            </li>
          ))}
        </ol>
      ) : (
        <p className="empty-state">No draft trades recorded for this season filter.</p>
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

function futurePicks(slots: DraftPickSlot[], condition: string): FutureDraftPick[] {
  return slots
    .filter((pick) => pick.overallNumber === 0)
    .map((pick) => ({
      season: pick.season,
      round: pick.round,
      originalTeamNumber: pick.originalTeamNumber,
      originalTeamName: pick.originalTeamName,
      condition: condition.trim(),
      conditionStatus: "",
    }));
}

function updateBudget(budgets: BudgetAsset[], kind: BudgetAsset["kind"], season: number, amount: number) {
  const remaining = budgets.filter((asset) => asset.kind !== kind || asset.season !== season);
  return amount > 0 ? [...remaining, { kind, season, amount }] : remaining;
}

function assetLabels(
  slots: DraftPickSlot[],
  players: string[],
  budgets: BudgetAsset[],
  playersByID: Map<string, Player>,
  format: DraftSnapshot["leagueFormat"],
  condition: string,
) {
  return [
    ...slots.map((pick) => {
      const label = pickLabel(pick, format);
      return pick.overallNumber === 0 && condition.trim() ? `${label} (condition: ${condition.trim()})` : label;
    }),
    ...players.map((id) => playersByID.get(id)?.name ?? `Player ${id}`),
    ...budgets.map(budgetLabel),
  ];
}

function storedTradeAssets(
  trade: DraftPickTrade,
  first: boolean,
  currentSlots: Map<string, DraftPickSlot>,
  playersByID: Map<string, Player>,
  format: DraftSnapshot["leagueFormat"],
) {
  const current = first ? trade.teamOneReceives : trade.teamTwoReceives;
  const future = first ? trade.teamOneFuturePicks : trade.teamTwoFuturePicks;
  const players = first ? trade.teamOnePlayers : trade.teamTwoPlayers;
  const budgets = first ? trade.teamOneBudgets : trade.teamTwoBudgets;
  const labels = current.map((number) => {
    const slot = currentSlots.get(`${trade.season}:${number}`);
    return slot ? pickLabel(slot, format) : `${trade.season} overall pick ${number}`;
  });
  labels.push(...future.map(futurePickLabel));
  labels.push(...players.map((id) => playersByID.get(id)?.name ?? `Player ${id}`));
  labels.push(...budgets.map(budgetLabel));
  if (!budgets.length) {
    if (first && trade.teamOneAuctionBudget > 0)
      labels.push(`$${trade.teamOneAuctionBudget.toFixed(0)} ${trade.season} auction budget`);
    if (!first && trade.teamTwoAuctionBudget > 0)
      labels.push(`$${trade.teamTwoAuctionBudget.toFixed(0)} ${trade.season} auction budget`);
  }
  return labels;
}

function pendingConditions(trade: DraftPickTrade) {
  return [...trade.teamOneFuturePicks, ...trade.teamTwoFuturePicks].filter(
    (pick) => pick.conditionStatus === "pending",
  );
}

function futurePickLabel(pick: FutureDraftPick) {
  const base = `${pick.season}, round ${pick.round}, ${pick.originalTeamName} original pick`;
  if (!pick.condition) return base;
  const status =
    pick.conditionStatus === "not-met"
      ? "condition not met; transfer void"
      : pick.conditionStatus === "met"
        ? "condition met"
        : "condition pending";
  return `${base} (${status}: ${pick.condition})`;
}

function budgetLabel(asset: BudgetAsset) {
  return `$${asset.amount.toFixed(0)} ${asset.season} ${asset.kind === "faab" ? "FAAB" : "auction budget"}`;
}
function teamName(snapshot: DraftSnapshot, teamNumber: number) {
  return snapshot.teams.find((team) => team.number === teamNumber)?.name ?? `Team ${teamNumber}`;
}
function pickKey(pick: DraftPickSlot) {
  return pick.overallNumber > 0
    ? `${pick.season}:overall:${pick.overallNumber}`
    : `${pick.season}:round:${pick.round}:team:${pick.originalTeamNumber}`;
}
function pickKeyFromFuture(pick: FutureDraftPick) {
  return `${pick.season}:round:${pick.round}:team:${pick.originalTeamNumber}`;
}
function groupLabel(pick: DraftPickSlot, format: DraftSnapshot["leagueFormat"]) {
  return format === "dynasty" ? `${pick.season} round ${pick.round}` : `Round ${pick.round}`;
}
function pickLabel(pick: DraftPickSlot, format: DraftSnapshot["leagueFormat"]) {
  if (pick.overallNumber === 0) return `${pick.season}, round ${pick.round}, ${pick.originalTeamName} original pick`;
  const current = `Round ${pick.round}, pick ${pick.pickInRound}, overall ${pick.overallNumber}`;
  return format === "dynasty" ? `${pick.season}, ${current.toLowerCase()}` : current;
}
function tradeHelp(snapshot: DraftSnapshot) {
  if (snapshot.leagueFormat === "redraft")
    return snapshot.draftType === "auction"
      ? "This redraft league shows only auction budget enabled by its rules; future assets stay out of the way."
      : `This redraft league shows only unused ${snapshot.season} picks. Future seasons stay out of the way.`;
  return `This dynasty league tracks players, unused ${snapshot.season} assets, future rookie picks, and enabled budgets across seasons.`;
}
