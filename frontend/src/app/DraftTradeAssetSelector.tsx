import type { BudgetAsset, DraftPickSlot, DraftSnapshot } from "../shared/api/types";
import { draftPickGroupLabel, draftPickKey, draftPickLabel, updateTradeBudget } from "./draftTradePresentation";

interface DraftTradeAssetSelectorProps {
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

export function DraftTradeAssetSelector(props: DraftTradeAssetSelectorProps) {
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
  const selectedFuture = available.some((pick) => selected.includes(draftPickKey(pick)) && pick.overallNumber === 0);
  const balanceOptions = snapshot.budgetBalances.filter((balance) => balance.teamNumber === sourceTeam?.number);

  return (
    <fieldset className="trade-pick-selector" disabled={disabled}>
      <legend>Assets {receivingTeam} receives</legend>
      {available.length
        ? [...groups].map(([key, picks]) => (
            <fieldset key={key} className="trade-round">
              <legend>{draftPickGroupLabel(picks[0], snapshot.leagueFormat)}</legend>
              {picks.map((pick) => {
                const key = draftPickKey(pick);
                return (
                  <label key={key}>
                    <input
                      type="checkbox"
                      checked={selected.includes(key)}
                      onChange={(event) =>
                        onChange(event.target.checked ? [...selected, key] : selected.filter((item) => item !== key))
                      }
                    />
                    <span>{draftPickLabel(pick, snapshot.leagueFormat)}</span>
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
                      onBudgetsChange(
                        updateTradeBudget(budgets, balance.kind, balance.season, Number(event.target.value)),
                      )
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
