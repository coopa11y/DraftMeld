import { useEffect, useState } from "react";
import { leagueToRules, updateLeague } from "../shared/api/leagues";
import { getConsensusRankings, getRankingWatchlist, importRankingPDF, listRankingSources, refreshRankingSources } from "../shared/api/rankings";
import type { ConsensusRanking, League, RankingSource, RankingSourcePreference, WatchlistPlayer } from "../shared/api/types";

const refreshTimeFormatter = new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" });
const providerCoverage = [
  { provider: "CBS Sports", status: "Connected", detail: "Current public PPR Top 200 consensus." },
  { provider: "Yahoo Fantasy", status: "OAuth required", detail: "Public rankings are league-specific; full league data requires approved API access." },
  { provider: "NFL.com", status: "Held out", detail: "The public overall draft table still contains the prior-season board." },
  { provider: "ESPN", status: "PDF import", detail: "Upload a PPR Top 300 or Dynasty Cheat Sheet that you are permitted to use." },
] as const;

function formatRefreshTime(value?: string | null) {
  return value ? refreshTimeFormatter.format(new Date(value)) : "Refresh required";
}

interface RankingSourcesProps {
  league: League;
  onLeagueUpdated: (league: League) => void;
}

export function RankingSources({ league, onLeagueUpdated }: RankingSourcesProps) {
  const [sources, setSources] = useState<RankingSource[]>([]);
  const [rankings, setRankings] = useState<ConsensusRanking[]>([]);
  const [preferences, setPreferences] = useState<Record<string, RankingSourcePreference>>({});
  const [watchlist, setWatchlist] = useState<WatchlistPlayer[]>([]);
  const [busy, setBusy] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [pdfFile, setPDFFile] = useState<File | null>(null);

  useEffect(() => {
    let active = true;
    async function load() {
      try {
        const [loaded, consensus, disabledSourcePlayers] = await Promise.all([
          listRankingSources(), getConsensusRankings(league.id), getRankingWatchlist(league.id),
        ]);
        if (!active) return;
        setSources(loaded);
        setRankings(consensus);
        setWatchlist(disabledSourcePlayers);
        setPreferences(Object.fromEntries(loaded.map((source) => [source.id, league.sourcePreferences[source.id] ?? { weight: source.defaultWeight, enabled: true }])));
      } catch (reason) {
        if (active) setError(reason instanceof Error ? reason.message : "Unable to load ranking sources.");
      } finally {
        if (active) setBusy(false);
      }
    }
    void load();
    return () => { active = false; };
  }, [league.id]);

  async function refresh() {
    setBusy(true);
    setError("");
    const downloadableCount = sources.filter((source) => source.importMode === "download").length;
    setMessage(`Downloading and normalizing ${downloadableCount} online ranking feeds. This may take a moment.`);
    try {
      const loaded = await refreshRankingSources();
      setSources(loaded);
      const [consensus, disabledSourcePlayers] = await Promise.all([getConsensusRankings(league.id), getRankingWatchlist(league.id)]);
      setRankings(consensus);
      setWatchlist(disabledSourcePlayers);
      const count = loaded.reduce((total, source) => total + source.recordCount, 0);
      setMessage(`Rankings refreshed. ${count} source records were normalized.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to refresh rankings.");
      setMessage("");
    } finally {
      setBusy(false);
    }
  }

  async function uploadPDF(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    if (!pdfFile) {
      setError("Choose an ESPN ranking PDF before importing.");
      return;
    }
    setBusy(true);
    setError("");
    setMessage("Reading the PDF locally and normalizing its player rankings.");
    try {
      const result = await importRankingPDF(pdfFile);
      setSources((current) => current.map((source) => source.id === result.source.id ? result.source : source));
      const [consensus, disabledSourcePlayers] = await Promise.all([getConsensusRankings(league.id), getRankingWatchlist(league.id)]);
      setRankings(consensus);
      setWatchlist(disabledSourcePlayers);
      setMessage(`${result.source.name} imported: ${result.source.recordCount} players from ${result.pageCount} page${result.pageCount === 1 ? "" : "s"}.`);
      setPDFFile(null);
      form.reset();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to import that PDF.");
      setMessage("");
    } finally {
      setBusy(false);
    }
  }

  async function saveWeights(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    setMessage(`Saving ranking influence for ${league.name}.`);
    try {
      const updated = await updateLeague(league.id, { ...leagueToRules(league), sourcePreferences: preferences });
      onLeagueUpdated(updated);
      const [consensus, disabledSourcePlayers] = await Promise.all([getConsensusRankings(updated.id), getRankingWatchlist(updated.id)]);
      setRankings(consensus);
      setWatchlist(disabledSourcePlayers);
      setMessage(`Ranking preferences saved for ${updated.name}. Excluded sources are still checked for players worth another look.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to save ranking preferences.");
      setMessage("");
    } finally {
      setBusy(false);
    }
  }

  function resetPreferences() {
    setPreferences(Object.fromEntries(sources.map((source) => [source.id, { weight: source.defaultWeight, enabled: true }])));
    setMessage("All sources and default influence values restored. Choose Save preferences to apply them.");
  }

  const preferencesChanged = sources.some((source) => {
    const current = preferences[source.id] ?? { weight: source.defaultWeight, enabled: true };
    const saved = league.sourcePreferences[source.id] ?? { weight: source.defaultWeight, enabled: true };
    return current.weight !== saved.weight || current.enabled !== saved.enabled;
  });
  const enabledSourceCount = Object.values(preferences).filter((preference) => preference.enabled).length;
  const enabledImportedSourceCount = sources.filter((source) => source.recordCount > 0 && (preferences[source.id]?.enabled ?? true)).length;

  return (
    <main className="ranking-page" id="main-content">
      <section className="ranking-panel" aria-labelledby="ranking-sources-heading">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Ranking data</p>
            <h1 id="ranking-sources-heading">Ranking sources</h1>
            <p className="section-description">Choose which sources shape {league.name}, then tune their relative influence. Equal weights have equal pull.</p>
          </div>
          <button className="primary-button" type="button" onClick={refresh} disabled={busy}>{busy ? "Working..." : "Refresh all sources"}</button>
        </div>
        <p className="status-message" role="status" aria-live="polite">{message}</p>
        {error ? <div className="error-banner" role="alert">{error}</div> : null}
        <form className="pdf-import-form" onSubmit={uploadPDF}>
          <div>
            <label htmlFor="ranking-pdf"><strong>Import a ranking PDF</strong></label>
            <p id="ranking-pdf-help">DraftMeld detects supported ESPN PPR Top 300 and Dynasty cheat sheets. Files are processed in memory and are not retained.</p>
          </div>
          <input id="ranking-pdf" name="file" type="file" accept="application/pdf,.pdf" aria-describedby="ranking-pdf-help" onChange={(event) => setPDFFile(event.target.files?.[0] ?? null)} disabled={busy} />
          <button className="secondary-button" type="submit" disabled={busy || !pdfFile}>Import PDF</button>
        </form>
        <form className="source-weight-form" onSubmit={saveWeights}>
          <fieldset disabled={busy}>
            <legend>Source preferences for {league.name}</legend>
            <p className="field-help">Included sources shape the consensus. Excluded sources stay available for the compact "Worth another look" list. Weight 2 has twice the pull of weight 1, and matching weights have equal influence. At least one source must remain included.</p>
            <ul className="source-grid">
              {sources.map((source) => {
                const preference = preferences[source.id] ?? { weight: source.defaultWeight, enabled: true };
                return (
                  <li key={source.id}>
                    <article className={`source-card${preference.enabled ? "" : " source-card-disabled"}`}>
                    <div className="source-card-heading"><h2>{source.name}</h2><span>{source.recordCount > 0 ? `${source.recordCount} players` : "Not imported"}</span></div>
                    <p>{source.description}</p>
                    <dl>
                      <div><dt>Method</dt><dd>{source.methodology}</dd></div>
                      <div><dt>License</dt><dd>{source.license}</dd></div>
                      <div><dt>Default weight</dt><dd>{source.defaultWeight}</dd></div>
                      <div><dt>Published</dt><dd>{source.publishedAt || "Refresh required"}</dd></div>
                      <div><dt>Last refreshed</dt><dd>{formatRefreshTime(source.refreshedAt)}</dd></div>
                    </dl>
                    <label className="source-enabled-control">
                      <input type="checkbox" checked={preference.enabled} disabled={preference.enabled && enabledSourceCount === 1} onChange={(event) => {
                        setPreferences((current) => ({ ...current, [source.id]: { ...preference, enabled: event.target.checked } }));
                      }} />
                      Include {source.name} in consensus
                    </label>
                    <label className="source-weight-control" htmlFor={`source-weight-${source.id}`}>
                      <span>{source.name} influence</span>
                      <input id={`source-weight-${source.id}`} type="number" min="0.1" max="10" step="0.1" required disabled={!preference.enabled} value={preference.weight} onChange={(event) => {
                        const weight = Number(event.target.value);
                        setPreferences((current) => ({ ...current, [source.id]: { ...preference, weight } }));
                      }} />
                    </label>
                    <a href={source.projectUrl} target="_blank" rel="noreferrer">View source website<span className="sr-only"> for {source.name}</span></a>
                    </article>
                  </li>
                );
              })}
            </ul>
            <div className="source-weight-actions">
              <button className="primary-button" type="submit" disabled={!preferencesChanged || busy}>Save preferences</button>
              <button className="secondary-button" type="button" onClick={resetPreferences} disabled={busy}>Include all and restore defaults</button>
            </div>
          </fieldset>
        </form>
      </section>

      {watchlist.length > 0 ? (
        <section className="ranking-panel watchlist-panel" aria-labelledby="watchlist-heading">
          <div className="section-heading"><div><p className="eyebrow">Excluded-source signals</p><h2 id="watchlist-heading">Worth another look</h2><p className="section-description">A short list of players ranked meaningfully higher by sources you excluded from the main consensus.</p></div></div>
          <ul className="watchlist-grid">
            {watchlist.map((player) => (
              <li key={player.playerKey}>
                <article>
                  <h3>{player.name} <span>{player.position} - {player.team}</span></h3>
                  <ul>
                    {player.signals.map((signal) => (
                      <li key={signal.sourceId}>{player.consensusRank === null
                        ? `${signal.sourceName} ranks this player #${signal.sourceRank}; enabled sources do not currently rank the player.`
                        : `${signal.sourceName} ranks this player #${signal.sourceRank}, ${signal.spotsHigher} spots above consensus #${player.consensusRank}.`}</li>
                    ))}
                  </ul>
                </article>
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      <section className="ranking-panel" aria-labelledby="provider-coverage-heading">
        <div className="section-heading"><div><p className="eyebrow">Freshness safeguards</p><h2 id="provider-coverage-heading">Platform connector status</h2><p className="section-description">DraftMeld holds a provider out when its available data is stale, incomplete, or requires approval.</p></div></div>
        <ul className="coverage-list">
          {providerCoverage.map((provider) => <li key={provider.provider}><div><strong>{provider.provider}</strong><span>{provider.status}</span></div><p>{provider.detail}</p></li>)}
        </ul>
      </section>

      {rankings.length > 0 ? (
        <section className="ranking-panel" aria-labelledby="consensus-heading">
          <div className="section-heading"><div><p className="eyebrow">League-weighted preview</p><h2 id="consensus-heading">DraftMeld consensus for {league.name}</h2><p className="section-description">The blended score is the weighted average of every included, imported source that ranks that player. Lower is better.</p></div></div>
          <div className="table-scroll">
            <table aria-label="Top 25 blended player rankings">
              <caption>Top 25 blended player rankings</caption>
              <thead><tr><th>Rank</th><th>Player</th><th>Pos.</th><th>Sources</th><th>Blended score</th></tr></thead>
              <tbody>{rankings.slice(0, 25).map((player) => <tr key={player.playerKey}><td>{player.rank}</td><th scope="row"><span className="player-name">{player.name}</span><span className="player-meta">{player.team || "Team unavailable"}</span></th><td>{player.position}</td><td>{player.sourceCount} of {enabledImportedSourceCount}</td><td>{player.score.toFixed(1)}</td></tr>)}</tbody>
            </table>
          </div>
        </section>
      ) : null}
    </main>
  );
}
