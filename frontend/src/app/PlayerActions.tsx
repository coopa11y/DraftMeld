import { useState } from "react";
import type { DraftAction, Player } from "../shared/api/types";
import { Button } from "../shared/ui/Button";

interface PlayerActionsProps {
  player: Player;
  busy: boolean;
  primary?: boolean;
  auction?: boolean;
  inflation?: number;
  minimumBid?: number;
  maximumBid?: number;
  selectedTeamNumber?: number;
  isUserTurn?: boolean;
  isComplete?: boolean;
  onAction: (player: Player, action: DraftAction, cost?: number, teamNumber?: number) => void;
}

export function PlayerActions({
  player,
  busy,
  primary = false,
  auction = false,
  inflation = 1,
  minimumBid = 1,
  maximumBid = Number.MAX_SAFE_INTEGER,
  selectedTeamNumber = 0,
  isUserTurn = true,
  isComplete = false,
  onAction,
}: PlayerActionsProps) {
  const suggestedCost = Math.min(maximumBid, Math.max(minimumBid, Math.round(player.auctionValue * inflation)));
  const [bid, setBid] = useState({ suggestedCost, value: suggestedCost });
  const cost = bid.suggestedCost === suggestedCost ? bid.value : suggestedCost;
  const draftIsPrimary = primary && (auction || isUserTurn);
  const takenIsPrimary = primary && !auction && !isUserTurn;
  return (
    <div className={`player-actions${auction ? " auction-actions" : ""}`}>
      {auction ? (
        <label>
          <span>Winning bid for {player.name}</span>
          <span className="currency-input">
            $
            <input
              type="number"
              min={minimumBid}
              max={maximumBid}
              step="1"
              value={cost}
              onChange={(event) => setBid({ suggestedCost, value: Number(event.target.value) })}
              disabled={busy}
            />
          </span>
        </label>
      ) : null}
      <Button
        variant="primary"
        data-player-action={draftIsPrimary ? "primary" : undefined}
        id={draftIsPrimary ? `draft-${player.id}` : undefined}
        disabled={
          isComplete || busy || (!auction && !isUserTurn) || (auction && (cost < minimumBid || cost > maximumBid))
        }
        aria-label={`Draft ${player.name}, ${player.position}, to my team${auction ? ` for ${cost} dollars` : ""}`}
        onClick={() => onAction(player, "draft", auction ? cost : 0, auction ? 0 : selectedTeamNumber)}
      >
        Draft
      </Button>
      <Button
        variant="neutral"
        data-player-action={takenIsPrimary ? "primary" : undefined}
        id={takenIsPrimary ? `draft-${player.id}` : undefined}
        disabled={
          isComplete || busy || (!auction && isUserTurn) || (auction && (cost < minimumBid || selectedTeamNumber === 0))
        }
        aria-label={`Mark ${player.name}, ${player.position}, as taken by another team${auction ? ` for ${cost} dollars` : ""}`}
        onClick={() => onAction(player, "taken", auction ? cost : 0, selectedTeamNumber)}
      >
        Gone
      </Button>
    </div>
  );
}
