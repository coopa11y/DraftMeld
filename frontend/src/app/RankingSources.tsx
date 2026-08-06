import { useEffect, useState } from "react";
import { getConsensusRankings, importRankingPDF, listRankingSources, refreshRankingSources } from "../shared/api/rankings";
import type { ConsensusRanking, RankingSource } from "../shared/api/types";

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

export function RankingSources() {
  const [sources, setSources] = useState<RankingSource[]>([]);
  const [rankings, setRankings] = useState<ConsensusRanking[]>([]);
  const [busy, setBusy] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [pdfFile, setPDFFile] = useState<File | null>(null);

  useEffect(() => {
    let active = true;
    async function load() {
      try {
        const loaded = await listRankingSources();
        if (!active) return;
        setSources(loaded);
        if (loaded.some((source) => source.recordCount > 0)) {
          const consensus = await getConsensusRankings();
          if (active) setRankings(consensus);
        }
      } catch (reason) {
        if (active) setError(reason instanceof Error ? reason.message : "Unable to load ranking sources.");
      } finally {
        if (active) setBusy(false);
      }
    }
    void load();
    return () => { active = false; };
  }, []);

  async function refresh() {
    setBusy(true);
    setError("");
    const downloadableCount = sources.filter((source) => source.importMode === "download").length;
    setMessage(`Downloading and normalizing ${downloadableCount} online ranking feeds. This may take a moment.`);
    try {
      const loaded = await refreshRankingSources();
      setSources(loaded);
      setRankings(await getConsensusRankings());
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
      setRankings(await getConsensusRankings());
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

  return (
    <main className="ranking-page" id="main-content">
      <section className="ranking-panel" aria-labelledby="ranking-sources-heading">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Ranking data</p>
            <h1 id="ranking-sources-heading">Ranking sources</h1>
            <p className="section-description">Every feed keeps its method, license, source link, publication date, and default influence visible.</p>
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
        <ul className="source-grid">
          {sources.map((source) => (
            <li key={source.id}>
              <article className="source-card">
                <div className="source-card-heading"><h2>{source.name}</h2><span>{source.recordCount > 0 ? `${source.recordCount} players` : "Not imported"}</span></div>
                <p>{source.description}</p>
                <dl>
                  <div><dt>Method</dt><dd>{source.methodology}</dd></div>
                  <div><dt>License</dt><dd>{source.license}</dd></div>
                  <div><dt>Default weight</dt><dd>{source.defaultWeight}</dd></div>
                  <div><dt>Published</dt><dd>{source.publishedAt || "Refresh required"}</dd></div>
                  <div><dt>Last refreshed</dt><dd>{formatRefreshTime(source.refreshedAt)}</dd></div>
                </dl>
                <a href={source.projectUrl} target="_blank" rel="noreferrer">View source website<span className="sr-only"> for {source.name}</span></a>
              </article>
            </li>
          ))}
        </ul>
      </section>

      <section className="ranking-panel" aria-labelledby="provider-coverage-heading">
        <div className="section-heading"><div><p className="eyebrow">Freshness safeguards</p><h2 id="provider-coverage-heading">Platform connector status</h2><p className="section-description">DraftMeld holds a provider out when its available data is stale, incomplete, or requires approval.</p></div></div>
        <ul className="coverage-list">
          {providerCoverage.map((provider) => <li key={provider.provider}><div><strong>{provider.provider}</strong><span>{provider.status}</span></div><p>{provider.detail}</p></li>)}
        </ul>
      </section>

      {rankings.length > 0 ? (
        <section className="ranking-panel" aria-labelledby="consensus-heading">
          <div className="section-heading"><div><p className="eyebrow">Normalized preview</p><h2 id="consensus-heading">DraftMeld consensus</h2></div></div>
          <div className="table-scroll">
            <table aria-label="Top 25 blended player rankings">
              <caption>Top 25 blended player rankings</caption>
              <thead><tr><th>Rank</th><th>Player</th><th>Pos.</th><th>Sources</th><th>Blended score</th></tr></thead>
              <tbody>{rankings.slice(0, 25).map((player) => <tr key={player.playerKey}><td>{player.rank}</td><th scope="row"><span className="player-name">{player.name}</span><span className="player-meta">{player.team || "Team unavailable"}</span></th><td>{player.position}</td><td>{player.sourceCount} of {sources.filter((source) => source.recordCount > 0).length}</td><td>{player.score.toFixed(1)}</td></tr>)}</tbody>
            </table>
          </div>
        </section>
      ) : null}
    </main>
  );
}
