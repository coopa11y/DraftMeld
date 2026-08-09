import { useMemo, useState } from "react";
import type { BudgetAsset, DraftPickTrade, DraftSnapshot, FutureDraftPick } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";
import {
  currentDraftPicks,
  draftPickKey,
  draftTeamName,
  draftTradeHelp,
  futureDraftPicks,
  selectedAssetLabels,
  selectedDraftSlots,
} from "./draftTradePresentation";
import { DraftTradeAssetSelector } from "./DraftTradeAssetSelector";
import { DraftTradeLedger } from "./DraftTradeLedger";
import { DraftTeamSelect, DraftTradeSummary } from "./DraftTradeShared";

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
    () => new Map(snapshot.pickSlots.map((pick) => [draftPickKey(pick), pick])),
    [snapshot.pickSlots],
  );
  const playersByID = useMemo(
    () => new Map(snapshot.teams.flatMap((team) => team.roster).map((player) => [player.id, player])),
    [snapshot.teams],
  );
  const teamOneSlots = selectedDraftSlots(teamOneSelected, slotsByKey);
  const teamTwoSlots = selectedDraftSlots(teamTwoSelected, slotsByKey);
  const firstName = draftTeamName(snapshot, teamOne);
  const secondName = draftTeamName(snapshot, teamTwo);
  const firstAssets = selectedAssetLabels(
    teamOneSlots,
    teamOnePlayers,
    teamOneBudgets,
    playersByID,
    snapshot.leagueFormat,
    teamOneCondition,
  );
  const secondAssets = selectedAssetLabels(
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
        currentDraftPicks(teamOneSlots),
        currentDraftPicks(teamTwoSlots),
        futureDraftPicks(teamOneSlots, teamOneCondition),
        futureDraftPicks(teamTwoSlots, teamTwoCondition),
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
      <p className="tool-help">{draftTradeHelp(snapshot)}</p>

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
            <DraftTeamSelect
              label="First team"
              value={teamOne}
              snapshot={snapshot}
              onChange={(team) => {
                setTeamOne(team);
                resetSelection();
              }}
            />
            <DraftTeamSelect
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
            <DraftTradeAssetSelector
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
            <DraftTradeAssetSelector
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
            <DraftTradeSummary
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

      <DraftTradeLedger
        snapshot={snapshot}
        playersByID={playersByID}
        ledgerSeason={ledgerSeason}
        onSeasonChange={setLedgerSeason}
        busy={busy}
        confirmDelete={confirmDelete}
        onConfirmDeleteChange={setConfirmDelete}
        onDelete={onDelete}
        onResolve={onResolve}
      />
    </Panel>
  );
}
