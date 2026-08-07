import { apiClient, unwrap } from "./client";
import type { ConsensusRanking, ErrorResponse, IdentityIssue, ProjectionSource, RankingPDFImport, RankingSource, WatchlistPlayer } from "./types";

export async function listRankingSources(): Promise<RankingSource[]> {
  const { data, error, response } = await apiClient.GET("/ranking-sources");
  return unwrap(data, error, response);
}

export async function refreshRankingSources(): Promise<RankingSource[]> {
  const { data, error, response } = await apiClient.POST("/ranking-sources/refresh");
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
  const response = await globalThis.fetch(new URL("/api/v1/ranking-sources/import-pdf", window.location.origin), { method: "POST", body: form });
  const body = await response.json() as RankingPDFImport | ErrorResponse;
  if (!response.ok) {
    throw new Error("error" in body ? body.error : `Request failed with ${response.status}`);
  }
  return body as RankingPDFImport;
}

export async function listProjectionSources(): Promise<ProjectionSource[]> {
  const { data, error, response } = await apiClient.GET("/projection-sources");
  return unwrap(data, error, response);
}

export async function importProjectionCSV(name: string, file: File): Promise<ProjectionSource> {
  const form = new FormData();
  form.append("name", name);
  form.append("file", file);
  const response = await globalThis.fetch(new URL("/api/v1/projection-sources/import-csv", window.location.origin), { method: "POST", body: form });
  const body = await response.json() as ProjectionSource | ErrorResponse;
  if (!response.ok) throw new Error("error" in body ? body.error : `Request failed with ${response.status}`);
  return body as ProjectionSource;
}

export async function listIdentityIssues(): Promise<IdentityIssue[]> {
  const { data, error, response } = await apiClient.GET("/ranking-identities");
  return unwrap(data, error, response);
}

export async function reviewIdentity(issueKey: string, resolution: "confirmed-separate" | "acknowledged"): Promise<void> {
  const { error, response } = await apiClient.POST("/ranking-identities/review", { body: { issueKey, resolution } });
  if (!response.ok) throw new Error((error as ErrorResponse | undefined)?.error ?? `Request failed with ${response.status}`);
}
