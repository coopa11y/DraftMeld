import { axe } from "jest-axe";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import type { PlayerNewsSource } from "../shared/api/types";
import { PlayerNewsSources } from "./PlayerNewsSources";

const builtIn: PlayerNewsSource = {
  id: "espn-nfl-news",
  name: "ESPN NFL news",
  kind: "rss",
  url: "https://www.espn.com/espn/rss/nfl/news",
  attribution: "ESPN",
  enabled: true,
  builtIn: true,
  refreshMinutes: 10,
  lastRefreshedAt: "2026-08-20T12:00:00Z",
};

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

it("manages compact source settings and accessible add and remove dialogs", async () => {
  const user = userEvent.setup();
  let sources = [builtIn];
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = new URL(input instanceof Request ? input.url : String(input));
    const method = init?.method ?? (input instanceof Request ? input.method : "GET");
    if (url.pathname.endsWith("/player-news/sources") && method === "GET") {
      return jsonResponse(sources);
    }
    if (url.pathname.endsWith("/player-news/refresh")) {
      return jsonResponse({ refreshed: 1, skipped: 0, errors: [] });
    }
    if (url.pathname.endsWith("/player-news/sources") && method === "POST") {
      const created: PlayerNewsSource = {
        id: "rss-local",
        name: "Local team feed",
        kind: "rss",
        url: "https://example.com/feed.xml",
        attribution: "Example Sports",
        enabled: true,
        builtIn: false,
        refreshMinutes: 15,
      };
      sources = [...sources, created];
      return jsonResponse(created, 201);
    }
    if (url.pathname.endsWith("/espn-nfl-news") && method === "PATCH") return new Response(null, { status: 204 });
    if (url.pathname.endsWith("/rss-local") && method === "DELETE") {
      sources = sources.filter((source) => source.id !== "rss-local");
      return new Response(null, { status: 204 });
    }
    return jsonResponse({ error: "unexpected request" }, 500);
  });
  vi.stubGlobal("fetch", fetchMock);
  const onBack = vi.fn();
  const { container } = render(<PlayerNewsSources onBack={onBack} />);

  expect(await screen.findByRole("heading", { name: "Player news sources" })).toHaveFocus();
  expect(screen.getByRole("checkbox", { name: "Enable ESPN NFL news" })).toBeChecked();
  await user.click(screen.getByRole("checkbox", { name: "Enable ESPN NFL news" }));
  expect(await screen.findByText("ESPN NFL news disabled.")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "Refresh news now" }));
  expect(await screen.findByText("1 sources refreshed; 0 skipped.")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "Add RSS source" }));
  expect(screen.getByRole("textbox", { name: "Source name" })).toHaveFocus();
  await user.type(screen.getByRole("textbox", { name: "Source name" }), "Local team feed");
  await user.type(screen.getByRole("textbox", { name: "RSS or Atom URL" }), "https://example.com/feed.xml");
  await user.type(screen.getByRole("textbox", { name: "Publisher attribution" }), "Example Sports");
  await user.click(screen.getByRole("button", { name: "Add source" }));
  expect(await screen.findByRole("link", { name: "Local team feed" })).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "Remove Local team feed" }));
  expect(screen.getByRole("heading", { name: "Remove Local team feed?" })).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "Remove source" }));
  await waitFor(() => expect(screen.queryByRole("link", { name: "Local team feed" })).not.toBeInTheDocument());

  expect((await axe(container)).violations).toHaveLength(0);
  await user.click(screen.getByRole("button", { name: "Back to league" }));
  expect(onBack).toHaveBeenCalledOnce();
});

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}
