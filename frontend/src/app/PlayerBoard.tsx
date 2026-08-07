import { useMemo, useState, type RefObject } from "react";
import type { DraftAction, DraftSnapshot, Player } from "../shared/api/types";
import { PlayerActions } from "./PlayerActions";

const positions = ["Overall", "QB", "RB", "WR", "TE", "K", "DST"] as const;
type PositionFilter = (typeof positions)[number];

interface PlayerBoardProps {
  snapshot: DraftSnapshot;
  busy: boolean;
  headingRef: RefObject<HTMLHeadingElement | null>;
  onAction: (player: Player, action: DraftAction, cost?: number) => void;
  onUndo: () => void;
  onPreference: (player: Player, preference: "target" | "avoid" | "") => void;
}

function valueLabel(player: Player) {
  const value = player.adp - player.overallRank;
  if (value >= 1) return `Plus ${value.toFixed(1)}`;
  if (value <= -1) return `Minus ${Math.abs(value).toFixed(1)}`;
  return "Even";
}

export function PlayerBoard({ snapshot, busy, headingRef, onAction, onUndo, onPreference }: PlayerBoardProps) {
  const [position, setPosition] = useState<PositionFilter>("Overall");
  const [search, setSearch] = useState("");
  const visiblePlayers = useMemo(() => {
    const normalizedSearch = search.trim().toLowerCase();
    const matchingPlayers = snapshot.available.filter((player) => {
      const matchesPosition = position === "Overall" || player.position === position;
      const matchesSearch = !normalizedSearch ||
        `${player.name} ${player.nflTeam} ${player.position}`.toLowerCase().includes(normalizedSearch);
      return matchesPosition && matchesSearch;
    });
    if (position !== "Overall") {
      matchingPlayers.sort((left, right) => left.positionRank - right.positionRank || left.overallRank - right.overallRank);
    }
    return matchingPlayers;
  }, [position, search, snapshot.available]);
  const rankHeading = position === "Overall" ? "Overall rank" : `${position} rank`;
  const tableCaption = position === "Overall"
    ? "Available players sorted by overall rank"
    : `Available ${position} players sorted by ${position} rank`;
  const resultSubject = position === "Overall" ? "players" : `${position} players`;

  return (
    <section className="board-panel" id="player-board" aria-labelledby="board-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">{snapshot.leagueName}</p>
          <h1 id="board-title" ref={headingRef} tabIndex={-1}>Available players</h1>
        </div>
        <button
          className="undo-button"
          type="button"
          disabled={!snapshot.canUndo || busy}
          onClick={onUndo}
          aria-describedby="undo-help"
        >
          Undo last action
        </button>
        <span className="sr-only" id="undo-help">Restores the most recently drafted or taken player.</span>
      </div>

      <div className="board-controls">
        <fieldset>
          <legend>Rankings by position</legend>
          <div className="filter-buttons">
            {positions.map((option) => (
              <button
                type="button"
                key={option}
                aria-pressed={position === option}
                onClick={() => setPosition(option)}
              >
                {option}
              </button>
            ))}
          </div>
        </fieldset>
        <label className="search-field">
          <span>Search available players</span>
          <input
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Name, team, or position"
          />
        </label>
      </div>

      <p className="result-summary" role="status">
        Showing {visiblePlayers.length} available {resultSubject} of {snapshot.available.length} total players.
      </p>

      <div className="table-scroll" role="region" aria-label="Available player rankings" tabIndex={0}>
        <table>
          <caption>{tableCaption}</caption>
          <thead>
            <tr>
              <th scope="col">{rankHeading}</th>
              <th scope="col">Player</th>
              <th scope="col">Position</th>
              <th scope="col">Tier</th>
              {snapshot.projectionCount > 0 ? <><th scope="col">Proj.</th><th scope="col"><abbr title="Value over replacement">VOR</abbr></th></> : null}
              <th scope="col"><abbr title="Average draft position">ADP</abbr></th>
              <th scope="col">Value</th>
              <th scope="col">My list</th>
              <th scope="col">Actions</th>
            </tr>
          </thead>
          <tbody>
            {visiblePlayers.map((player) => (
              <tr key={player.id}>
                <td>{position === "Overall" ? player.overallRank : player.positionRank}</td>
                <th scope="row">
                  <span className="player-name">{player.name}</span>
                  <span className="player-meta">{player.nflTeam}, bye week {player.byeWeek}</span>
                </th>
                <td>{player.position}{player.positionRank}</td>
                <td>{player.tier}</td>
                {snapshot.projectionCount > 0 ? <><td>{player.projectedPoints.toFixed(1)}</td><td>{player.valueOverReplacement.toFixed(1)}</td></> : null}
                <td>{player.adp.toFixed(1)}</td>
                <td>
                  <span aria-hidden="true">{(player.adp - player.overallRank).toFixed(1)}</span>
                  <span className="sr-only">{valueLabel(player)}</span>
                </td>
                <td><div className="preference-actions"><button type="button" aria-pressed={player.preference === "target"} onClick={() => onPreference(player, player.preference === "target" ? "" : "target")} disabled={busy}>Target</button><button type="button" aria-pressed={player.preference === "avoid"} onClick={() => onPreference(player, player.preference === "avoid" ? "" : "avoid")} disabled={busy}>Avoid</button></div></td>
                <td><PlayerActions player={player} busy={busy} primary auction={snapshot.draftType === "auction"} inflation={snapshot.auctionInflation} onAction={onAction} /></td>
              </tr>
            ))}
          </tbody>
        </table>
        {visiblePlayers.length === 0 ? <p className="empty-state">No available players match these filters.</p> : null}
      </div>
    </section>
  );
}
