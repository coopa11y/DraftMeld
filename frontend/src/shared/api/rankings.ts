import { apiClient, unwrap } from "./client";
import type { ConsensusRanking, RankingSource } from "./types";

export async function listRankingSources(): Promise<RankingSource[]> {
  const { data, error, response } = await apiClient.GET("/ranking-sources");
  return unwrap(data, error, response);
}

export async function refreshRankingSources(): Promise<RankingSource[]> {
  const { data, error, response } = await apiClient.POST("/ranking-sources/refresh");
  return unwrap(data, error, response);
}

export async function getConsensusRankings(): Promise<ConsensusRanking[]> {
  const { data, error, response } = await apiClient.GET("/rankings");
  return unwrap(data, error, response);
}
