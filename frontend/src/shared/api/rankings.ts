import { apiClient, postMultipart, unwrap } from "./client";
import type {
  ConsensusRanking,
  ErrorResponse,
  IdentityIssue,
  PlayerDirectoryStatus,
  ProjectionSource,
  RankingPDFImport,
  RankingSource,
  RankingSourceRecommendations,
  WatchlistPlayer,
} from "./types";

export async function listRankingSources(): Promise<RankingSource[]> {
  const { data, error, response } = await apiClient.GET("/ranking-sources");
  return unwrap(data, error, response);
}

export async function getRankingSourceRecommendations(leagueId: string): Promise<RankingSourceRecommendations> {
  const { data, error, response } = await apiClient.GET("/ranking-sources/recommendations", {
    params: { query: { leagueId } },
  });
  return unwrap(data, error, response);
}

export async function refreshRankingSources(leagueId: string): Promise<RankingSource[]> {
  const { data, error, response } = await apiClient.POST("/ranking-sources/refresh", {
    params: { query: { leagueId } },
  });
  return unwrap(data, error, response);
}

export async function refreshRankingSource(sourceId: string): Promise<RankingSource> {
  const { data, error, response } = await apiClient.POST("/ranking-sources/{sourceId}/refresh", {
    params: { path: { sourceId } },
  });
  return unwrap(data, error, response);
}

export async function getConsensusRankings(leagueId: string): Promise<ConsensusRanking[]> {
  const { data, error, response } = await apiClient.GET("/rankings", {
    params: { query: { leagueId } },
  });
  return unwrap(data, error, response);
}

export async function getRankingWatchlist(leagueId: string): Promise<WatchlistPlayer[]> {
  const { data, error, response } = await apiClient.GET("/ranking-watchlist", {
    params: { query: { leagueId } },
  });
  return unwrap(data, error, response);
}

export async function importRankingPDF(file: File): Promise<RankingPDFImport> {
  const form = new FormData();
  form.append("file", file);
  return postMultipart<RankingPDFImport>("ranking-sources/import-pdf", form);
}

export async function importRankingCSV(
  name: string,
  file: File,
  mapping: Record<string, string>,
): Promise<RankingSource> {
  const form = new FormData();
  form.append("name", name);
  form.append("file", file);
  form.append("mapping", JSON.stringify(mapping));
  return postMultipart<RankingSource>("ranking-sources/import-csv", form);
}

export async function listProjectionSources(): Promise<ProjectionSource[]> {
  const { data, error, response } = await apiClient.GET("/projection-sources");
  return unwrap(data, error, response);
}

export async function refreshProjectionSources(): Promise<ProjectionSource> {
  const { data, error, response } = await apiClient.POST("/projection-sources/refresh");
  return unwrap(data, error, response);
}

export async function refreshProjectionSource(sourceId: string): Promise<ProjectionSource> {
  const { data, error, response } = await apiClient.POST("/projection-sources/{sourceId}/refresh", {
    params: { path: { sourceId } },
  });
  return unwrap(data, error, response);
}

export async function importProjectionCSV(
  name: string,
  file: File,
  mapping: Record<string, string>,
): Promise<ProjectionSource> {
  const form = new FormData();
  form.append("name", name);
  form.append("file", file);
  form.append("mapping", JSON.stringify(mapping));
  return postMultipart<ProjectionSource>("projection-sources/import-csv", form);
}

export async function listIdentityIssues(): Promise<IdentityIssue[]> {
  const { data, error, response } = await apiClient.GET("/ranking-identities");
  return unwrap(data, error, response);
}

export async function getPlayerDirectoryStatus(): Promise<PlayerDirectoryStatus> {
  const { data, error, response } = await apiClient.GET("/player-directory/status");
  return unwrap(data, error, response);
}

export async function reviewIdentity(
  issueKey: string,
  resolution: "confirmed-separate" | "acknowledged" | "merged",
  canonicalPlayerKey = "",
): Promise<void> {
  const { error, response } = await apiClient.POST("/ranking-identities/review", {
    body: { issueKey, resolution, canonicalPlayerKey },
  });
  if (!response.ok)
    throw new Error((error as ErrorResponse | undefined)?.error ?? `Request failed with ${response.status}`);
}
