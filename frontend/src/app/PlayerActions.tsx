import type { DraftAction, Player } from "../shared/api/types";

interface PlayerActionsProps {
  player: Player;
  busy: boolean;
  primary?: boolean;
  onAction: (player: Player, action: DraftAction) => void;
}

export function PlayerActions({ player, busy, primary = false, onAction }: PlayerActionsProps) {
  return (
    <div className="player-actions">
      <button
        className="draft-button"
        data-player-action={primary ? "primary" : undefined}
        id={primary ? `draft-${player.id}` : undefined}
        type="button"
        disabled={busy}
        aria-label={`Draft ${player.name}, ${player.position}, to my team`}
        onClick={() => onAction(player, "draft")}
      >
        Draft
      </button>
      <button
        className="taken-button"
        type="button"
        disabled={busy}
        aria-label={`Mark ${player.name}, ${player.position}, as taken by another team`}
        onClick={() => onAction(player, "taken")}
      >
        Taken
      </button>
    </div>
  );
}
