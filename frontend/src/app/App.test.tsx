import { axe } from "jest-axe";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ConsensusRanking, DraftSnapshot, League, Player, RankingSource } from "../shared/api/types";
import { App } from "./App";

const alex: Player = {
  id: "p001", name: "Alex Rivers", nflTeam: "ATL", position: "RB",
  byeWeek: 12, overallRank: 1, positionRank: 1, adp: 1.8, tier: 1,
};
const jordan: Player = {
  id: "p002", name: "Jordan Hale", nflTeam: "MIN", position: "WR",
  byeWeek: 6, overallRank: 2, positionRank: 1, adp: 2.5, tier: 1,
};

function snapshot(overrides: Partial<DraftSnapshot> = {}): DraftSnapshot {
  return {
    leagueId: "demo",
    leagueName: "Demo League",
    pickNumber: 1,
    available: [alex, jordan],
    myTeam: [],
    history: [],
    recommendations: [
      { player: alex, score: 220, reasons: ["Fills an open starting roster need"] },
      { player: jordan, score: 219, reasons: ["Best available league-adjusted value"] },
    ],
    canUndo: false,
    ...overrides,
  };
}

const demoLeague: League = {
  id: "demo", name: "Demo League", teamCount: 12, draftPosition: 1, draftType: "snake",
  rosterSlots: [{ name: "RB", count: 2, positions: ["RB"], isStarting: true }],
  scoringRules: { reception: 1 },
};

const rankingSources: RankingSource[] = [
  { id: "redraft-ecr", name: "Redraft expert consensus", description: "Current overall redraft consensus.", methodology: "Average expert rank", license: "GPL-3.0", projectUrl: "https://github.com/dynastyprocess/data", dataUrl: "https://example.test/ecr.csv", defaultWeight: 1, recordCount: 500, publishedAt: "2026-07-31" },
  { id: "dynasty-1qb", name: "Dynasty market - 1 QB", description: "Long-term 1-QB values.", methodology: "Normalized player value", license: "GPL-3.0", projectUrl: "https://github.com/dynastyprocess/data", dataUrl: "https://example.test/1qb.csv", defaultWeight: 0.7, recordCount: 450, publishedAt: "2026-07-31" },
  { id: "dynasty-superflex", name: "Dynasty market - Superflex", description: "Long-term Superflex values.", methodology: "Normalized Superflex value", license: "GPL-3.0", projectUrl: "https://github.com/dynastyprocess/data", dataUrl: "https://example.test/superflex.csv", defaultWeight: 0.5, recordCount: 450, publishedAt: "2026-07-31" },
  { id: "expected-opportunity", name: "Expected opportunity", description: "Prior-season usage quality.", methodology: "Expected fantasy points", license: "CC-BY-SA-4.0", projectUrl: "https://github.com/ffverse/ffopportunity", dataUrl: "https://example.test/opportunity.csv", defaultWeight: 0.6, recordCount: 300, publishedAt: "2025" },
];

const consensusRankings: ConsensusRanking[] = [
  { playerKey: "alexrivers", name: "Alex Rivers", position: "RB", team: "ATL", rank: 1, score: 1.5, sourceCount: 4, sourceRanks: { "redraft-ecr": 1 } },
];

function jsonResponse(body: unknown, status = 200) {
  return Promise.resolve(new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  }));
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  localStorage.clear();
});

describe("accessible draft board", () => {
  it("creates a customized league through an accessible setup form", async () => {
    let configuredLeagues = [demoLeague];
    vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
      const request = input instanceof Request ? input : new Request(input);
      const path = new URL(request.url).pathname;
      if (path.endsWith("/leagues") && request.method === "GET") return jsonResponse(configuredLeagues);
      if (path.endsWith("/leagues") && request.method === "POST") {
        const rules = await request.clone().json() as Omit<League, "id">;
        const created = { id: "family-league", ...rules };
        configuredLeagues = [...configuredLeagues, created];
        return jsonResponse(created, 201);
      }
      return jsonResponse(snapshot());
    }));
    const user = userEvent.setup();
    const { container } = render(<App />);

    await user.click(await screen.findByRole("button", { name: "Manage leagues" }));
    await user.click(screen.getByRole("button", { name: "Create league" }));
    const name = screen.getByRole("textbox", { name: "League name" });
    await user.clear(name);
    await user.type(name, "Family League");
    expect((await axe(container)).violations).toHaveLength(0);
    await user.click(screen.getByRole("button", { name: "Save league" }));

    expect(await screen.findByRole("heading", { name: "Family League" })).toBeInTheDocument();
    expect(screen.getByText("Family League was saved.")).toBeInTheDocument();
  });

  it("loads the league selected by the application shell", async () => {
    localStorage.setItem("draftmeld.active-league.v1", "league-a");
    const leagueA = { ...demoLeague, id: "league-a", name: "League A" };
    const fetchMock = vi.fn()
      .mockImplementationOnce(() => jsonResponse([leagueA]))
      .mockImplementationOnce(() => jsonResponse(snapshot({ leagueId: "league-a", leagueName: "League A" })));
    vi.stubGlobal("fetch", fetchMock);
    render(<App />);

    expect(await screen.findByText("League A")).toBeInTheDocument();
    const requestInput = fetchMock.mock.calls[1][0];
    const requestUrl = requestInput instanceof Request ? requestInput.url : requestInput.toString();
    expect(new URL(requestUrl).searchParams.get("leagueId")).toBe("league-a");
  });

  it("exposes landmarks, names every player action, and has no automatic axe violations", async () => {
    vi.stubGlobal("fetch", vi.fn((input: RequestInfo | URL) => {
      const url = input instanceof Request ? input.url : input.toString();
      return jsonResponse(new URL(url).pathname.endsWith("/leagues") ? [demoLeague] : snapshot());
    }));
    const { container } = render(<App />);

    expect(await screen.findByRole("heading", { name: "Available players" })).toBeInTheDocument();
    expect(screen.getByRole("table", { name: "Available players sorted by overall rank" })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Draft Alex Rivers, RB, to my team" })).toHaveLength(2);
    expect(screen.getAllByRole("button", { name: "Mark Alex Rivers, RB, as taken by another team" })).toHaveLength(2);
    expect(screen.getByRole("link", { name: "Skip to player board" })).toBeInTheDocument();

    const results = await axe(container);
    expect(results.violations).toHaveLength(0);
  });

  it("shows source provenance and refreshes the accessible consensus preview", async () => {
    vi.stubGlobal("fetch", vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const request = input instanceof Request ? input : new Request(input, init);
      const path = new URL(request.url).pathname;
      if (path.endsWith("/leagues")) return jsonResponse([demoLeague]);
      if (path.endsWith("/ranking-sources/refresh") && request.method === "POST") return jsonResponse(rankingSources);
      if (path.endsWith("/ranking-sources")) return jsonResponse(rankingSources.map((source) => ({ ...source, recordCount: 0, publishedAt: undefined })));
      if (path.endsWith("/rankings")) return jsonResponse(consensusRankings);
      return jsonResponse(snapshot());
    }));
    const user = userEvent.setup();
    const { container } = render(<App />);

    await user.click(await screen.findByRole("button", { name: "Ranking sources" }));
    expect(await screen.findByRole("heading", { name: "Redraft expert consensus" })).toBeInTheDocument();
    expect(screen.getByText("CC-BY-SA-4.0")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /View open-source project/ })).toHaveLength(4);

    await user.click(screen.getByRole("button", { name: "Refresh all sources" }));
    expect(await screen.findByRole("table", { name: "Top 25 blended player rankings" })).toBeInTheDocument();
    expect(screen.getByText("Rankings refreshed. 1700 source records were normalized.")).toBeInTheDocument();
    expect((await axe(container)).violations).toHaveLength(0);
  });

  it("announces a draft, updates the team, and moves focus to the next available player", async () => {
    const drafted = snapshot({
      pickNumber: 2,
      available: [jordan],
      myTeam: [alex],
      history: [{ eventId: 1, number: 1, action: "draft", player: alex, createdAt: new Date().toISOString() }],
      recommendations: [{ player: jordan, score: 219, reasons: ["Fills an open starting roster need"] }],
      canUndo: true,
    });
    const fetchMock = vi.fn()
      .mockImplementationOnce(() => jsonResponse([demoLeague]))
      .mockImplementationOnce(() => jsonResponse(snapshot()))
      .mockImplementationOnce(() => jsonResponse(drafted));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    render(<App />);

    const draftButtons = await screen.findAllByRole("button", { name: "Draft Alex Rivers, RB, to my team" });
    await user.click(draftButtons[0]);

    expect(await screen.findByText("Alex Rivers, ATL")).toBeInTheDocument();
    expect(screen.getByText("Alex Rivers was drafted to your team. Rankings and recommendations updated.")).toBeInTheDocument();
    await waitFor(() => expect(screen.getAllByRole("button", { name: "Draft Jordan Hale, WR, to my team" })[0]).toHaveFocus());
  });

  it("restores the last player when undo is selected", async () => {
    const drafted = snapshot({
      pickNumber: 2,
      available: [jordan],
      myTeam: [alex],
      history: [{ eventId: 1, number: 1, action: "draft", player: alex, createdAt: new Date().toISOString() }],
      recommendations: [{ player: jordan, score: 219, reasons: ["Fills an open starting roster need"] }],
      canUndo: true,
    });
    const fetchMock = vi.fn()
      .mockImplementationOnce(() => jsonResponse([demoLeague]))
      .mockImplementationOnce(() => jsonResponse(drafted))
      .mockImplementationOnce(() => jsonResponse(snapshot()));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByRole("button", { name: "Undo last action" }));

    expect(await screen.findByText("Alex Rivers was restored to the available-player list.")).toBeInTheDocument();
    await waitFor(() => expect(screen.getAllByRole("button", { name: "Draft Alex Rivers, RB, to my team" })[0]).toHaveFocus());
  });
});
