import { useState } from "react";
import type { DraftAction, Player } from "../shared/api/types";

interface PlayerActionsProps {
  player: Player;
  busy: boolean;
  primary?: boolean;
  auction?: boolean;
  inflation?: number;
  onAction: (player: Player, action: DraftAction, cost?: number) => void;
}

export function PlayerActions({ player, busy, primary = false, auction = false, inflation = 1, onAction }: PlayerActionsProps) {
  const suggestedCost = Math.max(1, Math.round(player.auctionValue * inflation));
  const [cost, setCost] = useState(suggestedCost);
  return (
    <div className={`player-actions${auction ? " auction-actions" : ""}`}>
      {auction ? <label><span>Winning bid for {player.name}</span><span className="currency-input">$<input type="number" min="1" step="1" value={cost} onChange={(event) => setCost(Number(event.target.value))} disabled={busy} /></span></label> : null}
      <button className="draft-button" data-player-action={primary ? "primary" : undefined} id={primary ? `draft-${player.id}` : undefined} type="button" disabled={busy || (auction && cost < 1)} aria-label={`Draft ${player.name}, ${player.position}, to my team${auction ? ` for ${cost} dollars` : ""}`} onClick={() => onAction(player, "draft", auction ? cost : 0)}>Draft</button>
      <button className="taken-button" type="button" disabled={busy || (auction && cost < 1)} aria-label={`Mark ${player.name}, ${player.position}, as taken by another team${auction ? ` for ${cost} dollars` : ""}`} onClick={() => onAction(player, "taken", auction ? cost : 0)}>Taken</button>
    </div>
  );
}
