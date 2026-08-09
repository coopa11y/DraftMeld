import type { DraftPickTrade, DraftSnapshot, FutureDraftPick, Player } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import {
  futureDraftPickKey,
  futureDraftPickLabel,
  pendingTradeConditions,
  storedTradeAssetLabels,
} from "./draftTradePresentation";
import { DraftTradeSummary } from "./DraftTradeShared";

interface DraftTradeLedgerProps {
  snapshot: DraftSnapshot;
  playersByID: Map<string, Player>;
  ledgerSeason: string;
  onSeasonChange: (season: string) => void;
  busy: boolean;
  confirmDelete: number | null;
  onConfirmDeleteChange: (id: number | null) => void;
  onDelete: (trade: DraftPickTrade) => Promise<void>;
  onResolve: (trade: DraftPickTrade, pick: FutureDraftPick, status: "met" | "not-met") => Promise<void>;
}

export function DraftTradeLedger(props: DraftTradeLedgerProps) {
  const {
    snapshot,
    playersByID,
    ledgerSeason,
    onSeasonChange,
    busy,
    confirmDelete,
    onConfirmDeleteChange,
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
            <select value={ledgerSeason} onChange={(event) => onSeasonChange(event.target.value)}>
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
              <DraftTradeSummary
                firstName={trade.teamOneName}
                secondName={trade.teamTwoName}
                firstAssets={storedTradeAssetLabels(trade, true, currentSlots, playersByID, snapshot.leagueFormat)}
                secondAssets={storedTradeAssetLabels(trade, false, currentSlots, playersByID, snapshot.leagueFormat)}
              />
              {pendingTradeConditions(trade).map((pick) => (
                <div
                  className="conditional-pick-actions"
                  key={futureDraftPickKey(pick)}
                  role="group"
                  aria-label={`Resolve condition for ${futureDraftPickLabel(pick)}`}
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
                    <Button type="button" disabled={busy} onClick={() => onConfirmDeleteChange(null)}>
                      Keep trade
                    </Button>
                  </div>
                </div>
              ) : (
                <Button type="button" disabled={busy} onClick={() => onConfirmDeleteChange(trade.id)}>
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
