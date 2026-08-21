import { apiClient, ensureSuccess, unwrap } from "./client";
import type { AddPlayerNewsSourceRequest, PlayerNewsFeed, PlayerNewsRefreshResult, PlayerNewsSource } from "./types";

export async function getPlayerNews(): Promise<PlayerNewsFeed> {
  const { data, error, response } = await apiClient.GET("/player-news");
  return unwrap(data, error, response);
}

export async function listPlayerNewsSources(): Promise<PlayerNewsSource[]> {
  const { data, error, response } = await apiClient.GET("/player-news/sources");
  return unwrap(data, error, response);
}

export async function refreshPlayerNews(): Promise<PlayerNewsRefreshResult> {
  const { data, error } = await apiClient.POST("/player-news/refresh");
  if (data !== undefined) return data;
  if (error !== undefined) throw new Error(error.errors.join(" ") || "No player-news source could be refreshed.");
  throw new Error("No player-news source could be refreshed.");
}

export async function addPlayerNewsSource(input: AddPlayerNewsSourceRequest): Promise<PlayerNewsSource> {
  const { data, error, response } = await apiClient.POST("/player-news/sources", { body: input });
  return unwrap(data, error, response);
}

export async function setPlayerNewsSourceEnabled(sourceId: string, enabled: boolean): Promise<void> {
  const { error, response } = await apiClient.PATCH("/player-news/sources/{sourceId}", {
    params: { path: { sourceId } },
    body: { enabled },
  });
  ensureSuccess(error, response);
}

export async function deletePlayerNewsSource(sourceId: string): Promise<void> {
  const { error, response } = await apiClient.DELETE("/player-news/sources/{sourceId}", {
    params: { path: { sourceId } },
  });
  ensureSuccess(error, response);
}
