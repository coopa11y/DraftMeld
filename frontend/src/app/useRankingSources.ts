import { useEffect, useRef, useState, type Dispatch, type FormEvent, type SetStateAction } from "react";
import { leagueToRules, updateLeague } from "../shared/api/leagues";
import {
  getConsensusRankings,
  getPlayerDirectoryStatus,
  getRankingWatchlist,
  importRankingCSV,
  importRankingPDF,
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
  PlayerDirectoryStatus,
  ProjectionSource,
  RankingSource,
  RankingSourcePreference,
  WatchlistPlayer,
} from "../shared/api/types";

interface RankingSourcesActions {
  refresh: () => Promise<void>;
  uploadPDF: (event: FormEvent<HTMLFormElement>) => Promise<void>;
  uploadRankingCSV: (name: string, file: File, mapping: Record<string, string>) => Promise<void>;
  saveWeights: (event: FormEvent<HTMLFormElement>) => Promise<void>;
  resetPreferences: () => void;
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
  const [identityIssues, setIdentityIssues] = useState<IdentityIssue[]>([]);
  const [directoryStatus, setDirectoryStatus] = useState<PlayerDirectoryStatus>(emptyDirectoryStatus);
  const [consensusMethod, setConsensusMethod] = useState<LeagueRules["consensusMethod"]>(league.consensusMethod);
  const [busy, setBusy] = useState(true);
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
        const loadedSources = await refreshRankingSources();
        const evidence = await loadRankingEvidence(league.id);
        setSources(loadedSources);
        applyEvidence(evidence, setRankings, setWatchlist, setIdentityIssues, setDirectoryStatus);
        const count = loadedSources.reduce((total, source) => total + source.recordCount, 0);
        setMessage(`Rankings refreshed. ${count} source records were normalized.`);
      },
      "Unable to refresh rankings.",
    );
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
    setMessage("All sources and default influence values restored. Choose Save preferences to apply them.");
  }

  const preferencesChanged = rankingPreferencesChanged(sources, preferences, consensusMethod, league);
  const enabledSourceCount = Object.values(preferences).filter((preference) => preference.enabled).length;
  const enabledImportedSourceCount = sources.filter(
    (source) => source.recordCount > 0 && (preferences[source.id]?.enabled ?? true),
  ).length;

  return {
    actions: {
      refresh,
      uploadPDF,
      uploadRankingCSV,
      saveWeights,
      resetPreferences,
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
    consensusMethod,
    directoryStatus,
    enabledImportedSourceCount,
    enabledSourceCount,
    error,
    identityIssues,
    league,
    message,
    pdfFile,
    preferences,
    preferencesChanged,
    projectionSources,
    rankings,
    sources,
    watchlist,
  };
}

const emptyDirectoryStatus: PlayerDirectoryStatus = { playerCount: 0, identityCount: 0, providerIdCount: 0 };

async function loadRankingWorkspace(league: League) {
  const [sources, rankings, watchlist, projectionSources, identityIssues, directoryStatus] = await Promise.all([
    listRankingSources(),
    getConsensusRankings(league.id),
    getRankingWatchlist(league.id),
    listProjectionSources(),
    listIdentityIssues(),
    getPlayerDirectoryStatus(),
  ]);
  return { sources, rankings, watchlist, projectionSources, identityIssues, directoryStatus };
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
      league.sourcePreferences[source.id] ?? { weight: source.defaultWeight, enabled: true },
    ]),
  );
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
      const current = preferences[source.id] ?? { weight: source.defaultWeight, enabled: true };
      const saved = league.sourcePreferences[source.id] ?? { weight: source.defaultWeight, enabled: true };
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

async function runBusy(
  setBusy: Dispatch<SetStateAction<boolean>>,
  setError: Dispatch<SetStateAction<string>>,
  setMessage: Dispatch<SetStateAction<string>>,
  progressMessage: string,
  operation: () => Promise<void>,
  fallbackError: string,
  rethrow = false,
) {
  setBusy(true);
  setError("");
  setMessage(progressMessage);
  try {
    await operation();
  } catch (reason) {
    setError(errorMessage(reason, fallbackError));
    setMessage("");
    if (rethrow) throw reason;
  } finally {
    setBusy(false);
  }
}

function errorMessage(reason: unknown, fallback: string) {
  return reason instanceof Error ? reason.message : fallback;
}

function pdfImportMessage(result: Awaited<ReturnType<typeof importRankingPDF>>) {
  return `${result.source.name} imported: ${result.source.recordCount} players from ${result.pageCount} page${result.pageCount === 1 ? "" : "s"}.${
    result.ocrApplied ? " Scanned pages were recognized locally with OCR; review the imported rankings carefully." : ""
  }`;
}
