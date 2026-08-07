import { apiClient, unwrap } from "./client";
import type { ConsensusRanking, ErrorResponse, RankingPDFImport, RankingSource } from "./types";

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
