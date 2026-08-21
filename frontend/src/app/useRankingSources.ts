import { useEffect, useRef, useState, type Dispatch, type FormEvent, type SetStateAction } from "react";
import { leagueToRules, updateLeague } from "../shared/api/leagues";
import {
  getConsensusRankings,
  getPlayerDirectoryStatus,
  getRankingSourceRecommendations,
  getRankingWatchlist,
  importRankingCSV,
  importRankingPDF,
  listIdentityIssues,
  listProjectionSources,
  listRankingSources,
  refreshRankingSources,
  refreshRankingSource,
  refreshProjectionSources,
  refreshProjectionSource as refreshProjectionSourceRequest,
} from "../shared/api/rankings";
import type {
  ConsensusRanking,
  IdentityIssue,
  League,
  LeagueRules,
  PlayerDirectoryStatus,
  ProjectionSource,
  RankingSource,
  RankingSourceRecommendations,
  RankingSourcePreference,
  WatchlistPlayer,
} from "../shared/api/types";
import { errorMessage, pdfImportMessage, runBusy, runSourceUpdate } from "./rankingSourceOperations";

interface RankingSourcesActions {
  refresh: () => Promise<void>;
  refreshSource: (sourceId: string) => Promise<void>;
  refreshProjectionSource: (sourceId: string) => Promise<void>;
  uploadPDF: (event: FormEvent<HTMLFormElement>) => Promise<void>;
  uploadRankingCSV: (name: string, file: File, mapping: Record<string, string>) => Promise<void>;
  saveWeights: (event: FormEvent<HTMLFormElement>) => Promise<void>;
  resetPreferences: () => void;
  applyRecommendations: () => void;
  projectionImported: (source: ProjectionSource) => void;
  identityReviewed: (issueKey: string, resolution: IdentityIssue["resolution"], canonicalPlayerKey?: string) => void;
  setBusy: Dispatch<SetStateAction<boolean>>;
  setConsensusMethod: Dispatch<SetStateAction<LeagueRules["consensusMethod"]>>;
  setError: Dispatch<SetStateAction<string>>;
  setMessage: Dispatch<SetStateAction<string>>;
  setPDFFile: Dispatch<SetStateAction<File | null>>;
  setPreferences: Dispatch<SetStateAction<Record<string, RankingSourcePreference>>>;
}

export interface RankingSourcesController {
  actions: RankingSourcesActions;
  busy: boolean;
  busySourceId: string;
  consensusMethod: LeagueRules["consensusMethod"];
  directoryStatus: PlayerDirectoryStatus;
  enabledImportedSourceCount: number;
  enabledSourceCount: number;
  error: string;
  identityIssues: IdentityIssue[];
  league: League;
  message: string;
  pdfFile: File | null;
  preferences: Record<string, RankingSourcePreference>;
  preferencesChanged: boolean;
  projectionSources: ProjectionSource[];
  recommendations: RankingSourceRecommendations;
  rankings: ConsensusRanking[];
  sources: RankingSource[];
  watchlist: WatchlistPlayer[];
}

export function useRankingSources(league: League, onLeagueUpdated: (league: League) => void): RankingSourcesController {
  const initialLeague = useRef(league);
  const [sources, setSources] = useState<RankingSource[]>([]);
  const [rankings, setRankings] = useState<ConsensusRanking[]>([]);
  const [preferences, setPreferences] = useState<Record<string, RankingSourcePreference>>({});
  const [watchlist, setWatchlist] = useState<WatchlistPlayer[]>([]);
  const [projectionSources, setProjectionSources] = useState<ProjectionSource[]>([]);
  const [recommendations, setRecommendations] = useState<RankingSourceRecommendations>({ profile: "", sources: [] });
  const [identityIssues, setIdentityIssues] = useState<IdentityIssue[]>([]);
  const [directoryStatus, setDirectoryStatus] = useState<PlayerDirectoryStatus>(emptyDirectoryStatus);
  const [consensusMethod, setConsensusMethod] = useState<LeagueRules["consensusMethod"]>(league.consensusMethod);
  const [busy, setBusy] = useState(true);
  const [busySourceId, setBusySourceId] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [pdfFile, setPDFFile] = useState<File | null>(null);

  useEffect(() => {
    let active = true;
    loadRankingWorkspace(initialLeague.current)
      .then((loaded) => {
        if (!active) return;
        setSources(loaded.sources);
        setRankings(loaded.rankings);
        setWatchlist(loaded.watchlist);
        setProjectionSources(loaded.projectionSources);
        setRecommendations(loaded.recommendations);
        setIdentityIssues(loaded.identityIssues);
        setDirectoryStatus(loaded.directoryStatus);
        setConsensusMethod(initialLeague.current.consensusMethod);
        setPreferences(initialPreferences(loaded.sources, initialLeague.current));
      })
      .catch((reason) => {
        if (active) setError(errorMessage(reason, "Unable to load ranking sources."));
      })
      .finally(() => {
        if (active) setBusy(false);
      });
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
    const downloadableCount = sources.filter((source) => source.importMode === "download").length;
    await runBusy(
      setBusy,
      setError,
      setMessage,
      `Downloading and normalizing ${downloadableCount} online ranking feeds. This may take a moment.`,
      async () => {
        const [loadedSources, projectionSource] = await Promise.all([
          refreshRankingSources(league.id),
          refreshProjectionSources(),
        ]);
        const evidence = await loadRankingEvidence(league.id);
        setSources(loadedSources);
        setProjectionSources((current) => [
          ...current.filter((source) => source.id !== projectionSource.id),
          projectionSource,
        ]);
        applyEvidence(evidence, setRankings, setWatchlist, setIdentityIssues, setDirectoryStatus);
        const count = loadedSources.reduce((total, source) => total + source.recordCount, 0);
        setMessage(
          `Online sources refreshed. ${count} ranking records and ${projectionSource.recordCount} offensive projections were normalized.`,
        );
      },
      "Unable to refresh rankings.",
    );
  }

  async function refreshProjectionSource(sourceId: string) {
    const source = projectionSources.find((candidate) => candidate.id === sourceId);
    if (!source || source.importMode !== "download") return;
    await runSourceUpdate(setBusySourceId, setError, setMessage, source, async () => {
      const updated = await refreshProjectionSourceRequest(sourceId);
      setProjectionSources((current) => [...current.filter((candidate) => candidate.id !== updated.id), updated]);
      setDirectoryStatus(await getPlayerDirectoryStatus());
      return `${updated.name} updated with ${updated.recordCount} offensive player projections.`;
    });
  }

  async function refreshSource(sourceId: string) {
    const source = sources.find((candidate) => candidate.id === sourceId);
    if (!source) return;
    await runSourceUpdate(setBusySourceId, setError, setMessage, source, async () => {
      const updated = await refreshRankingSource(sourceId);
      setSources((current) => current.map((candidate) => (candidate.id === updated.id ? updated : candidate)));
      applyEvidence(
        await loadRankingEvidence(league.id),
        setRankings,
        setWatchlist,
        setIdentityIssues,
        setDirectoryStatus,
      );
      return `${updated.name} updated with ${updated.recordCount} players.`;
    });
  }

  async function uploadPDF(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    if (!pdfFile) {
      setError("Choose an ESPN ranking PDF before importing.");
      return;
    }
    await runBusy(
      setBusy,
      setError,
      setMessage,
      "Reading the PDF locally, recognizing scanned pages if needed, and normalizing its player rankings.",
      async () => {
        const result = await importRankingPDF(pdfFile);
        setSources((current) => current.map((source) => (source.id === result.source.id ? result.source : source)));
        applyEvidence(
          await loadRankingEvidence(league.id),
          setRankings,
          setWatchlist,
          setIdentityIssues,
          setDirectoryStatus,
        );
        setMessage(pdfImportMessage(result));
        setPDFFile(null);
        form.reset();
      },
      "Unable to import that PDF.",
    );
  }

  async function uploadRankingCSV(name: string, file: File, mapping: Record<string, string>) {
    await runBusy(
      setBusy,
      setError,
      setMessage,
      `Importing ${name} and normalizing its player rankings.`,
      async () => {
        const imported = await importRankingCSV(name, file, mapping);
        setSources((current) => [...current.filter((source) => source.id !== imported.id), imported]);
        setPreferences((current) => ({
          ...current,
          [imported.id]: current[imported.id] ?? { weight: imported.defaultWeight, enabled: true },
        }));
        applyEvidence(
          await loadRankingEvidence(league.id),
          setRankings,
          setWatchlist,
          setIdentityIssues,
          setDirectoryStatus,
        );
        setMessage(`${imported.name} imported with ${imported.recordCount} ranked players.`);
      },
      "Unable to import ranking CSV.",
      true,
    );
  }

  async function saveWeights(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await runBusy(
      setBusy,
      setError,
      setMessage,
      `Saving ranking influence for ${league.name}.`,
      async () => {
        const updated = await updateLeague(league.id, {
          ...leagueToRules(league),
          sourcePreferences: preferences,
          consensusMethod,
        });
        onLeagueUpdated(updated);
        const [updatedRankings, updatedWatchlist] = await Promise.all([
          getConsensusRankings(updated.id),
          getRankingWatchlist(updated.id),
        ]);
        setRankings(updatedRankings);
        setWatchlist(updatedWatchlist);
        setMessage(
          `Ranking preferences saved for ${updated.name}. Excluded sources are still checked for players worth another look.`,
        );
      },
      "Unable to save ranking preferences.",
    );
  }

  function resetPreferences() {
    setPreferences(initialPreferences(sources, league));
    setMessage("Unsaved source changes discarded.");
  }

  function applyRecommendations() {
    const recommended = Object.fromEntries(
      recommendations.sources.map((source) => [source.sourceId, source.preference]),
    );
    const next = { ...preferences, ...recommended };
    const changed = sources.some((source) => {
      const current = preferences[source.id];
      const replacement = next[source.id];
      return current?.enabled !== replacement?.enabled || current?.weight !== replacement?.weight;
    });
    setPreferences(next);
    setMessage(
      changed
        ? `Recommended settings applied for ${recommendations.profile}. Choose Save preferences to keep them.`
        : `${league.name} already uses the recommended settings for ${recommendations.profile}.`,
    );
  }

  const sourceState = rankingSourceState({ sources, preferences, consensusMethod, league });

  return {
    actions: {
      refresh,
      refreshSource,
      refreshProjectionSource,
      uploadPDF,
      uploadRankingCSV,
      saveWeights,
      resetPreferences,
      applyRecommendations,
      projectionImported: (source) => {
        setProjectionSources((current) => [...current.filter((candidate) => candidate.id !== source.id), source]);
        refreshDirectoryStatus();
      },
      identityReviewed: (issueKey, resolution, canonicalPlayerKey) => {
        setIdentityIssues((current) =>
          current.map((issue) => (issue.issueKey === issueKey ? { ...issue, resolution, canonicalPlayerKey } : issue)),
        );
        refreshDirectoryStatus();
      },
      setBusy,
      setConsensusMethod,
      setError,
      setMessage,
      setPDFFile,
      setPreferences,
    },
    busy,
    busySourceId,
    consensusMethod,
    directoryStatus,
    ...sourceState,
    error,
    identityIssues,
    league,
    message,
    pdfFile,
    preferences,
    projectionSources,
    recommendations,
    rankings,
    sources,
    watchlist,
  };
}

const emptyDirectoryStatus: PlayerDirectoryStatus = { playerCount: 0, identityCount: 0, providerIdCount: 0 };

async function loadRankingWorkspace(league: League) {
  const [sources, rankings, watchlist, projectionSources, identityIssues, directoryStatus, recommendations] =
    await Promise.all([
      listRankingSources(),
      getConsensusRankings(league.id),
      getRankingWatchlist(league.id),
      listProjectionSources(),
      listIdentityIssues(),
      getPlayerDirectoryStatus(),
      getRankingSourceRecommendations(league.id),
    ]);
  return { sources, rankings, watchlist, projectionSources, identityIssues, directoryStatus, recommendations };
}

async function loadRankingEvidence(leagueId: string) {
  const [rankings, watchlist, identityIssues, directoryStatus] = await Promise.all([
    getConsensusRankings(leagueId),
    getRankingWatchlist(leagueId),
    listIdentityIssues(),
    getPlayerDirectoryStatus(),
  ]);
  return { rankings, watchlist, identityIssues, directoryStatus };
}

function initialPreferences(sources: RankingSource[], league: League) {
  return Object.fromEntries(
    sources.map((source) => [
      source.id,
      league.sourcePreferences[source.id] ?? { weight: source.defaultWeight, enabled: source.defaultEnabled },
    ]),
  );
}

function rankingSourceState(input: {
  sources: RankingSource[];
  preferences: Record<string, RankingSourcePreference>;
  consensusMethod: LeagueRules["consensusMethod"];
  league: League;
}) {
  const { sources, preferences, consensusMethod, league } = input;
  return {
    preferencesChanged: rankingPreferencesChanged(sources, preferences, consensusMethod, league),
    enabledSourceCount: Object.values(preferences).filter((preference) => preference.enabled).length,
    enabledImportedSourceCount: sources.filter(
      (source) => source.recordCount > 0 && (preferences[source.id]?.enabled ?? true),
    ).length,
  };
}

function rankingPreferencesChanged(
  sources: RankingSource[],
  preferences: Record<string, RankingSourcePreference>,
  consensusMethod: LeagueRules["consensusMethod"],
  league: League,
) {
  return (
    consensusMethod !== league.consensusMethod ||
    sources.some((source) => {
      const current = preferences[source.id] ?? { weight: source.defaultWeight, enabled: source.defaultEnabled };
      const saved = league.sourcePreferences[source.id] ?? {
        weight: source.defaultWeight,
        enabled: source.defaultEnabled,
      };
      return current.weight !== saved.weight || current.enabled !== saved.enabled;
    })
  );
}

function applyEvidence(
  evidence: Awaited<ReturnType<typeof loadRankingEvidence>>,
  setRankings: Dispatch<SetStateAction<ConsensusRanking[]>>,
  setWatchlist: Dispatch<SetStateAction<WatchlistPlayer[]>>,
  setIdentityIssues: Dispatch<SetStateAction<IdentityIssue[]>>,
  setDirectoryStatus: Dispatch<SetStateAction<PlayerDirectoryStatus>>,
) {
  setRankings(evidence.rankings);
  setWatchlist(evidence.watchlist);
  setIdentityIssues(evidence.identityIssues);
  setDirectoryStatus(evidence.directoryStatus);
}
