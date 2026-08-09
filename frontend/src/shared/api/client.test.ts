import { afterEach, describe, expect, it, vi } from "vitest";
import { ensureSuccess, postMultipart, unwrap } from "./client";

afterEach(() => vi.unstubAllGlobals());

describe("API response helpers", () => {
  it("returns successful data and accepts successful empty responses", () => {
    const response = new Response(null, { status: 204 });
    expect(unwrap({ id: "demo" }, undefined, response)).toEqual({ id: "demo" });
    expect(() => ensureSuccess(undefined, response)).not.toThrow();
  });

  it("prefers API errors and falls back to the HTTP status", () => {
    const response = new Response(null, { status: 422 });
    expect(() => unwrap(undefined, { error: "Invalid league" }, response)).toThrow("Invalid league");
    expect(() => ensureSuccess(undefined, response)).toThrow("Request failed with 422");
  });

  it("handles multipart success, structured failure, and non-JSON failure", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ count: 2 }), { status: 201 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ error: "Unsupported PDF" }), { status: 415 }))
      .mockResolvedValueOnce(new Response("gateway failure", { status: 502 }));
    vi.stubGlobal("fetch", fetchMock);
    const form = new FormData();

    await expect(postMultipart<{ count: number }>("rankings/import", form)).resolves.toEqual({ count: 2 });
    await expect(postMultipart("rankings/import", form)).rejects.toThrow("Unsupported PDF");
    await expect(postMultipart("rankings/import", form)).rejects.toThrow("Request failed with 502");
  });
});
