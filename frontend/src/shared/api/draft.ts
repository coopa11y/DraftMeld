import createClient from "openapi-fetch";
import type { paths } from "./generated";
import type { DraftAction, DraftSnapshot, ErrorResponse } from "./types";

const baseUrl = new URL("/api/v1", window.location.origin).toString();
const client = createClient<paths>({ baseUrl, fetch: (...args) => globalThis.fetch(...args) });

function unwrap(data: DraftSnapshot | undefined, error: ErrorResponse | undefined, response: Response) {
  if (data) return data;
  throw new Error(error?.error ?? `Request failed with ${response.status}`);
}

export async function getDraft(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await client.GET("/draft", { params: { query: { leagueId } } });
  return unwrap(data, error, response);
}

export async function recordDraftAction(
  leagueId: string,
  playerId: string,
  action: DraftAction,
): Promise<DraftSnapshot> {
  const { data, error, response } = await client.POST("/draft/actions", {
    body: { leagueId, playerId, action },
  });
  return unwrap(data, error, response);
}

export async function undoDraftAction(leagueId: string): Promise<DraftSnapshot> {
  const { data, error, response } = await client.POST("/draft/undo", { body: { leagueId } });
  return unwrap(data, error, response);
}
