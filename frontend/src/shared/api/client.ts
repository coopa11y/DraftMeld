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

export async function postMultipart<T>(path: string, form: FormData): Promise<T> {
  const response = await globalThis.fetch(new URL(`/api/v1/${path}`, window.location.origin), {
    method: "POST",
    body: form,
  });
  let body: unknown;
  try {
    body = await response.json();
  } catch {
    throw new Error(`Request failed with ${response.status}`);
  }
  if (!response.ok) {
    const message =
      typeof body === "object" && body !== null && "error" in body && typeof body.error === "string"
        ? body.error
        : `Request failed with ${response.status}`;
    throw new Error(message);
  }
  return body as T;
}
