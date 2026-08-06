import { useEffect, useMemo, useRef, useState } from "react";
import { getDraft, recordDraftAction, undoDraftAction } from "../shared/api/draft";
import type { DraftAction, DraftSnapshot, Player } from "../shared/api/types";
import { PlayerActions } from "./PlayerActions";

const positions = ["Overall", "QB", "RB", "WR", "TE"] as const;
type PositionFilter = (typeof positions)[number];

function valueLabel(player: Player) {
  const value = player.adp - player.overallRank;
  if (value >= 1) return `Plus ${value.toFixed(1)}`;
  if (value <= -1) return `Minus ${Math.abs(value).toFixed(1)}`;
  return "Even";
}

export function App() {
  const [snapshot, setSnapshot] = useState<DraftSnapshot | null>(null);
  const [position, setPosition] = useState<PositionFilter>("Overall");
  const [search, setSearch] = useState("");
  const [busy, setBusy] = useState(false);
  const [announcement, setAnnouncement] = useState("Draft board loading.");
  const [error, setError] = useState("");
  const pendingFocus = useRef<string | null>(null);
  const boardHeading = useRef<HTMLHeadingElement>(null);
  const errorAlert = useRef<HTMLDivElement>(null);

  useEffect(() => {
    getDraft()
      .then((data) => {
        setSnapshot(data);
        setAnnouncement(`Draft board loaded. ${data.available.length} players are available.`);
      })
      .catch((reason: Error) => setError(reason.message));
  }, []);

  useEffect(() => {
    if (!snapshot || !pendingFocus.current) return;
    const requested = document.getElementById(`draft-${pendingFocus.current}`);
    const fallback = document.querySelector<HTMLButtonElement>("[data-player-action='primary']");
    (requested ?? fallback ?? boardHeading.current)?.focus();
    pendingFocus.current = null;
  }, [snapshot]);

  useEffect(() => {
    if (error) errorAlert.current?.focus();
  }, [error]);

  const visiblePlayers = useMemo(() => {
    if (!snapshot) return [];
    const normalizedSearch = search.trim().toLowerCase();
    return snapshot.available.filter((player) => {
      const matchesPosition = position === "Overall" || player.position === position;
      const matchesSearch = !normalizedSearch ||
        `${player.name} ${player.nflTeam} ${player.position}`.toLowerCase().includes(normalizedSearch);
      return matchesPosition && matchesSearch;
    });
  }, [position, search, snapshot]);

  async function handleAction(player: Player, action: DraftAction) {
    if (!snapshot || busy) return;
    const index = snapshot.available.findIndex((candidate) => candidate.id === player.id);
    pendingFocus.current = snapshot.available[index + 1]?.id ?? snapshot.available[index - 1]?.id ?? "board";
    setBusy(true);
    setError("");
    try {
      const updated = await recordDraftAction(player.id, action);
      setSnapshot(updated);
      setAnnouncement(
        action === "draft"
          ? `${player.name} was drafted to your team. Rankings and recommendations updated.`
          : `${player.name} was marked taken by another team. Rankings and recommendations updated.`,
      );
    } catch (reason) {
      pendingFocus.current = player.id;
      setError(reason instanceof Error ? reason.message : "The draft action failed.");
    } finally {
      setBusy(false);
    }
  }

  async function handleUndo() {
    if (!snapshot?.canUndo || busy) return;
    const lastPick = snapshot.history[snapshot.history.length - 1];
    pendingFocus.current = lastPick.player.id;
    setBusy(true);
    setError("");
    try {
      const updated = await undoDraftAction();
      setSnapshot(updated);
      setAnnouncement(`${lastPick.player.name} was restored to the available-player list.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Undo failed.");
    } finally {
      setBusy(false);
    }
  }

  if (!snapshot && !error) {
    return <main className="centered-status" aria-busy="true"><p role="status">Loading draft board…</p></main>;
  }

  return (
    <>
      <a className="skip-link" href="#player-board">Skip to player board</a>
      <a className="skip-link" href="#recommendations">Skip to recommendations</a>
      <a className="skip-link" href="#my-team">Skip to my team</a>

      <header className="app-header">
        <div>
          <a className="brand" href="/" aria-label="DraftMeld home">DraftMeld</a>
          <span className="version">v{__APP_VERSION__}</span>
        </div>
        {snapshot && (
          <div className="draft-status" aria-label={`Draft status. Pick ${snapshot.pickNumber}. ${snapshot.available.length} players available.`}>
            <strong>Pick {snapshot.pickNumber}</strong>
            <span>{snapshot.available.length} available</span>
          </div>
        )}
      </header>

      <div className="sr-only" role="status" aria-live="polite" aria-atomic="true">{announcement}</div>
      {error && <div className="error-banner" role="alert" tabIndex={-1} ref={errorAlert}>{error}</div>}

      {snapshot && (
        <main className="draft-layout" aria-busy={busy}>
          <section className="board-panel" id="player-board" aria-labelledby="board-title">
            <div className="section-heading">
              <div>
                <p className="eyebrow">{snapshot.leagueName}</p>
                <h1 id="board-title" ref={boardHeading} tabIndex={-1}>Available players</h1>
              </div>
              <button
                className="undo-button"
                type="button"
                disabled={!snapshot.canUndo || busy}
                onClick={handleUndo}
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
              Showing {visiblePlayers.length} of {snapshot.available.length} available players.
            </p>

            <div className="table-scroll" role="region" aria-label="Available player rankings" tabIndex={0}>
              <table>
                <caption>Available players sorted by overall rank</caption>
                <thead>
                  <tr>
                    <th scope="col">Rank</th>
                    <th scope="col">Player</th>
                    <th scope="col">Position</th>
                    <th scope="col">Tier</th>
                    <th scope="col"><abbr title="Average draft position">ADP</abbr></th>
                    <th scope="col">Value</th>
                    <th scope="col">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {visiblePlayers.map((player) => (
                    <tr key={player.id}>
                      <td>{player.overallRank}</td>
                      <th scope="row">
                        <span className="player-name">{player.name}</span>
                        <span className="player-meta">{player.nflTeam}, bye week {player.byeWeek}</span>
                      </th>
                      <td>{player.position}{player.positionRank}</td>
                      <td>{player.tier}</td>
                      <td>{player.adp.toFixed(1)}</td>
                      <td>
                        <span aria-hidden="true">{(player.adp - player.overallRank).toFixed(1)}</span>
                        <span className="sr-only">{valueLabel(player)}</span>
                      </td>
                      <td><PlayerActions player={player} busy={busy} primary onAction={handleAction} /></td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {visiblePlayers.length === 0 && <p className="empty-state">No available players match these filters.</p>}
            </div>
          </section>

          <aside className="sidebar" aria-label="Draft assistant">
            <section className="side-panel" id="recommendations" aria-labelledby="recommendations-title">
              <p className="eyebrow">Updated after every pick</p>
              <h2 id="recommendations-title">Recommended</h2>
              <ol className="recommendation-list">
                {snapshot.recommendations.map((recommendation) => (
                  <li key={recommendation.player.id}>
                    <article>
                      <h3>{recommendation.player.name} <span>{recommendation.player.position}{recommendation.player.positionRank}</span></h3>
                      <ul className="reason-list">
                        {recommendation.reasons.map((reason) => <li key={reason}>{reason}</li>)}
                      </ul>
                      <PlayerActions player={recommendation.player} busy={busy} onAction={handleAction} />
                    </article>
                  </li>
                ))}
              </ol>
            </section>

            <section className="side-panel" id="my-team" aria-labelledby="my-team-title">
              <p className="eyebrow">{snapshot.myTeam.length} players</p>
              <h2 id="my-team-title">My team</h2>
              {snapshot.myTeam.length ? (
                <ul className="team-list">
                  {snapshot.myTeam.map((player) => (
                    <li key={player.id}><strong>{player.position}</strong><span>{player.name}, {player.nflTeam}</span></li>
                  ))}
                </ul>
              ) : <p className="empty-state">No players drafted yet.</p>}
            </section>

            <section className="side-panel" aria-labelledby="history-title">
              <h2 id="history-title">Draft history</h2>
              {snapshot.history.length ? (
                <ol className="history-list">
                  {[...snapshot.history].reverse().slice(0, 8).map((pick) => (
                    <li key={pick.eventId}>
                      <strong>Pick {pick.number}:</strong> {pick.player.name} {pick.action === "draft" ? "to my team" : "taken"}
                    </li>
                  ))}
                </ol>
              ) : <p className="empty-state">No picks recorded yet.</p>}
            </section>
          </aside>
        </main>
      )}
    </>
  );
}
