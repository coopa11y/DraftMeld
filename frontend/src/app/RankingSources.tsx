import { useEffect, useRef, useState } from "react";
import { leagueToRules, updateLeague } from "../shared/api/leagues";
import {
  getConsensusRankings,
  getRankingWatchlist,
  getPlayerDirectoryStatus,
  importRankingPDF,
  importRankingCSV,
  listIdentityIssues,
  listProjectionSources,
  listRankingSources,
  refreshRankingSources,
} from "../shared/api/rankings";
import type {
  ConsensusRanking,
  IdentityIssue,
  League,
  LeagueRules,
  ProjectionSource,
  PlayerDirectoryStatus,
  RankingSource,
  RankingSourcePreference,
  WatchlistPlayer,
} from "../shared/api/types";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { IdentityReviewQueue } from "./IdentityReviewQueue";
import { ProjectionImport } from "./ProjectionImport";
import { RankingEvidencePanels } from "./RankingEvidencePanels";
import { RankingCsvImport } from "./RankingCsvImport";
import { RankingSourcePreferences } from "./RankingSourcePreferences";

interface RankingSourcesProps {
  league: League;
  onLeagueUpdated: (league: League) => void;
}

export function RankingSources({ league, onLeagueUpdated }: RankingSourcesProps) {
  const heading = useViewHeadingFocus<HTMLHeadingElement>();
  const initialLeague = useRef(league);
  const [sources, setSources] = useState<RankingSource[]>([]);
  const [rankings, setRankings] = useState<ConsensusRanking[]>([]);
  const [preferences, setPreferences] = useState<Record<string, RankingSourcePreference>>({});
  const [watchlist, setWatchlist] = useState<WatchlistPlayer[]>([]);
  const [projectionSources, setProjectionSources] = useState<ProjectionSource[]>([]);
  const [identityIssues, setIdentityIssues] = useState<IdentityIssue[]>([]);
  const [directoryStatus, setDirectoryStatus] = useState<PlayerDirectoryStatus>({
    playerCount: 0,
    identityCount: 0,
    providerIdCount: 0,
  });
  const [consensusMethod, setConsensusMethod] = useState<LeagueRules["consensusMethod"]>(league.consensusMethod);
  const [busy, setBusy] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [pdfFile, setPDFFile] = useState<File | null>(null);

  useEffect(() => {
    let active = true;
    async function load() {
      try {
        const selectedLeague = initialLeague.current;
        const [loaded, consensus, disabledSourcePlayers, projections, identities, playerDirectory] = await Promise.all([
          listRankingSources(),
          getConsensusRankings(selectedLeague.id),
          getRankingWatchlist(selectedLeague.id),
          listProjectionSources(),
          listIdentityIssues(),
          getPlayerDirectoryStatus(),
        ]);
        if (!active) return;
        setSources(loaded);
        setRankings(consensus);
        setWatchlist(disabledSourcePlayers);
        setProjectionSources(projections);
        setIdentityIssues(identities);
        setDirectoryStatus(playerDirectory);
        setConsensusMethod(selectedLeague.consensusMethod);
        setPreferences(
          Object.fromEntries(
            loaded.map((source) => [
              source.id,
              selectedLeague.sourcePreferences[source.id] ?? { weight: source.defaultWeight, enabled: true },
            ]),
          ),
        );
      } catch (reason) {
        if (active) setError(reason instanceof Error ? reason.message : "Unable to load ranking sources.");
      } finally {
        if (active) setBusy(false);
      }
    }
    void load();
    return () => {
      active = false;
    };
  }, []);

  function refreshDirectoryStatus() {
    void getPlayerDirectoryStatus()
      .then(setDirectoryStatus)
      .catch(() => setError("Unable to refresh the player directory status."));
  }

  async function refresh() {
    setBusy(true);
    setError("");
    const downloadableCount = sources.filter((source) => source.importMode === "download").length;
    setMessage(`Downloading and normalizing ${downloadableCount} online ranking feeds. This may take a moment.`);
    try {
      const loaded = await refreshRankingSources();
      setSources(loaded);
      const [consensus, disabledSourcePlayers, identities, playerDirectory] = await Promise.all([
        getConsensusRankings(league.id),
        getRankingWatchlist(league.id),
        listIdentityIssues(),
        getPlayerDirectoryStatus(),
      ]);
      setRankings(consensus);
      setWatchlist(disabledSourcePlayers);
      setIdentityIssues(identities);
      setDirectoryStatus(playerDirectory);
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
      setSources((current) => current.map((source) => (source.id === result.source.id ? result.source : source)));
      const [consensus, disabledSourcePlayers, identities, playerDirectory] = await Promise.all([
        getConsensusRankings(league.id),
        getRankingWatchlist(league.id),
        listIdentityIssues(),
        getPlayerDirectoryStatus(),
      ]);
      setRankings(consensus);
      setWatchlist(disabledSourcePlayers);
      setIdentityIssues(identities);
      setDirectoryStatus(playerDirectory);
      setMessage(
        `${result.source.name} imported: ${result.source.recordCount} players from ${result.pageCount} page${result.pageCount === 1 ? "" : "s"}.`,
      );
      setPDFFile(null);
      form.reset();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to import that PDF.");
      setMessage("");
    } finally {
      setBusy(false);
    }
  }

  async function uploadRankingCSV(name: string, file: File, mapping: Record<string, string>) {
    setBusy(true);
    setError("");
    setMessage(`Importing ${name} and normalizing its player rankings.`);
    try {
      const imported = await importRankingCSV(name, file, mapping);
      setSources((current) => [...current.filter((source) => source.id !== imported.id), imported]);
      setPreferences((current) => ({
        ...current,
        [imported.id]: current[imported.id] ?? { weight: imported.defaultWeight, enabled: true },
      }));
      const [consensus, disabledSourcePlayers, identities, playerDirectory] = await Promise.all([
        getConsensusRankings(league.id),
        getRankingWatchlist(league.id),
        listIdentityIssues(),
        getPlayerDirectoryStatus(),
      ]);
      setRankings(consensus);
      setWatchlist(disabledSourcePlayers);
      setIdentityIssues(identities);
      setDirectoryStatus(playerDirectory);
      setMessage(`${imported.name} imported with ${imported.recordCount} ranked players.`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to import ranking CSV.");
      setMessage("");
      throw reason;
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
      const updated = await updateLeague(league.id, {
        ...leagueToRules(league),
        sourcePreferences: preferences,
        consensusMethod,
      });
      onLeagueUpdated(updated);
      const [consensus, disabledSourcePlayers] = await Promise.all([
        getConsensusRankings(updated.id),
        getRankingWatchlist(updated.id),
      ]);
      setRankings(consensus);
      setWatchlist(disabledSourcePlayers);
      setMessage(
        `Ranking preferences saved for ${updated.name}. Excluded sources are still checked for players worth another look.`,
      );
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Unable to save ranking preferences.");
      setMessage("");
    } finally {
      setBusy(false);
    }
  }

  function resetPreferences() {
    setPreferences(
      Object.fromEntries(sources.map((source) => [source.id, { weight: source.defaultWeight, enabled: true }])),
    );
    setMessage("All sources and default influence values restored. Choose Save preferences to apply them.");
  }

  const preferencesChanged =
    consensusMethod !== league.consensusMethod ||
    sources.some((source) => {
      const current = preferences[source.id] ?? { weight: source.defaultWeight, enabled: true };
      const saved = league.sourcePreferences[source.id] ?? { weight: source.defaultWeight, enabled: true };
      return current.weight !== saved.weight || current.enabled !== saved.enabled;
    });
  const enabledSourceCount = Object.values(preferences).filter((preference) => preference.enabled).length;
  const enabledImportedSourceCount = sources.filter(
    (source) => source.recordCount > 0 && (preferences[source.id]?.enabled ?? true),
  ).length;

  return (
    <main className="ranking-page" id="main-content" aria-busy={busy}>
      <Panel variant="ranking" aria-labelledby="ranking-sources-heading">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Ranking data</p>
            <h1 id="ranking-sources-heading" ref={heading} tabIndex={-1}>
              Ranking sources
            </h1>
            <p className="section-description">
              Choose which sources shape {league.name}, then tune their relative influence. Equal weights have equal
              pull.
            </p>
          </div>
          <Button variant="primary" onClick={refresh} disabled={busy}>
            {busy ? "Working..." : "Refresh all sources"}
          </Button>
        </div>
        <StatusMessage>{message}</StatusMessage>
        {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
        <form className="pdf-import-form" onSubmit={uploadPDF}>
          <div>
            <label htmlFor="ranking-pdf">
              <strong>Import a ranking PDF</strong>
            </label>
            <p id="ranking-pdf-help">
              DraftMeld detects supported ESPN PPR Top 300 and Dynasty cheat sheets. Files are processed in memory and
              are not retained.
            </p>
          </div>
          <input
            id="ranking-pdf"
            name="file"
            type="file"
            accept="application/pdf,.pdf"
            aria-describedby="ranking-pdf-help"
            onChange={(event) => setPDFFile(event.target.files?.[0] ?? null)}
            disabled={busy}
          />
          <Button type="submit" disabled={busy || !pdfFile}>
            Import PDF
          </Button>
        </form>
        <RankingSourcePreferences
          busy={busy}
          consensusMethod={consensusMethod}
          enabledSourceCount={enabledSourceCount}
          leagueName={league.name}
          preferences={preferences}
          preferencesChanged={preferencesChanged}
          sources={sources}
          onConsensusMethodChange={setConsensusMethod}
          onPreferencesChange={setPreferences}
          onReset={resetPreferences}
          onSubmit={saveWeights}
        />
      </Panel>

      <RankingCsvImport busy={busy} sources={sources} onImport={uploadRankingCSV} />

      <ProjectionImport
        busy={busy}
        sources={projectionSources}
        onBusyChange={setBusy}
        onImported={(source) => {
          setProjectionSources((current) => [...current.filter((candidate) => candidate.id !== source.id), source]);
          refreshDirectoryStatus();
        }}
        onMessage={setMessage}
        onError={setError}
      />

      <IdentityReviewQueue
        busy={busy}
        issues={identityIssues}
        status={directoryStatus}
        onBusyChange={setBusy}
        onReviewed={(issueKey, resolution, canonicalPlayerKey) => {
          setIdentityIssues((current) =>
            current.map((issue) =>
              issue.issueKey === issueKey ? { ...issue, resolution, canonicalPlayerKey } : issue,
            ),
          );
          refreshDirectoryStatus();
        }}
        onError={setError}
      />

      <RankingEvidencePanels
        enabledImportedSourceCount={enabledImportedSourceCount}
        leagueName={league.name}
        rankings={rankings}
        watchlist={watchlist}
      />
    </main>
  );
}
