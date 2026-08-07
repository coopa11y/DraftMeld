import { useState } from "react";
import type { DraftAction, Player } from "../shared/api/types";

interface PlayerActionsProps {
  player: Player;
  busy: boolean;
  primary?: boolean;
  auction?: boolean;
  inflation?: number;
  minimumBid?: number;
  maximumBid?: number;
  onAction: (player: Player, action: DraftAction, cost?: number) => void;
}

export function PlayerActions({ player, busy, primary = false, auction = false, inflation = 1, minimumBid = 1, maximumBid = Number.MAX_SAFE_INTEGER, onAction }: PlayerActionsProps) {
  const suggestedCost = Math.min(maximumBid, Math.max(minimumBid, Math.round(player.auctionValue * inflation)));
  const [bid, setBid] = useState({ suggestedCost, value: suggestedCost });
  const cost = bid.suggestedCost === suggestedCost ? bid.value : suggestedCost;
  return (
    <div className={`player-actions${auction ? " auction-actions" : ""}`}>
      {auction ? <label><span>Winning bid for {player.name}</span><span className="currency-input">$<input type="number" min={minimumBid} max={maximumBid} step="1" value={cost} onChange={(event) => setBid({ suggestedCost, value: Number(event.target.value) })} disabled={busy} /></span></label> : null}
      <button className="draft-button" data-player-action={primary ? "primary" : undefined} id={primary ? `draft-${player.id}` : undefined} type="button" disabled={busy || (auction && (cost < minimumBid || cost > maximumBid))} aria-label={`Draft ${player.name}, ${player.position}, to my team${auction ? ` for ${cost} dollars` : ""}`} onClick={() => onAction(player, "draft", auction ? cost : 0)}>Draft</button>
      <button className="taken-button" type="button" disabled={busy || (auction && cost < minimumBid)} aria-label={`Mark ${player.name}, ${player.position}, as taken by another team${auction ? ` for ${cost} dollars` : ""}`} onClick={() => onAction(player, "taken", auction ? cost : 0)}>Taken</button>
    </div>
  );
}
