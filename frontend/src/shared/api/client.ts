import createClient from "openapi-fetch";
import type { paths } from "./generated";
import type { ErrorResponse } from "./types";

const baseUrl = new URL("/api/v1", window.location.origin).toString();

export const apiClient = createClient<paths>({
  baseUrl,
  fetch: (...args) => globalThis.fetch(...args),
});

export function unwrap<T>(data: T | undefined, error: ErrorResponse | undefined, response: Response): T {
  if (data !== undefined) return data;
  throw new Error(error?.error ?? `Request failed with ${response.status}`);
}

export function ensureSuccess(error: ErrorResponse | undefined, response: Response): void {
  if (response.ok) return;
  throw new Error(error?.error ?? `Request failed with ${response.status}`);
}
