import { axe } from "jest-axe";
import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
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
  sourceWeights: {
    "redraft-ecr": 1, "dynasty-1qb": 0.7, "dynasty-superflex": 0.5,
    "expected-opportunity": 0.6, "cbs-ppr": 0.9, "espn-ppr-pdf": 0.9, "espn-dynasty-pdf": 0.6,
  },
};
const casey: Player = {
  id: "p003", name: "Casey Brooks", nflTeam: "DET", position: "RB",
  byeWeek: 8, overallRank: 10, positionRank: 2, adp: 11.4, tier: 2,
};
const kicker: Player = {
  id: "p004", name: "Avery Cole", nflTeam: "DAL", position: "K",
  byeWeek: 10, overallRank: 11, positionRank: 1, adp: 145.2, tier: 1,
};
const defense: Player = {
  id: "p005", name: "Denver Defense", nflTeam: "DEN", position: "DST",
  byeWeek: 12, overallRank: 12, positionRank: 1, adp: 137.8, tier: 1,
};

const rankingSources: RankingSource[] = [
  { id: "redraft-ecr", name: "Redraft expert consensus", description: "Current overall redraft consensus.", methodology: "Average expert rank", license: "GPL-3.0", projectUrl: "https://github.com/dynastyprocess/data", dataUrl: "https://example.test/ecr.csv", defaultWeight: 1, importMode: "download", recordCount: 500, publishedAt: "2026-07-31" },
  { id: "dynasty-1qb", name: "Dynasty market - 1 QB", description: "Long-term 1-QB values.", methodology: "Normalized player value", license: "GPL-3.0", projectUrl: "https://github.com/dynastyprocess/data", dataUrl: "https://example.test/1qb.csv", defaultWeight: 0.7, importMode: "download", recordCount: 450, publishedAt: "2026-07-31" },
  { id: "dynasty-superflex", name: "Dynasty market - Superflex", description: "Long-term Superflex values.", methodology: "Normalized Superflex value", license: "GPL-3.0", projectUrl: "https://github.com/dynastyprocess/data", dataUrl: "https://example.test/superflex.csv", defaultWeight: 0.5, importMode: "download", recordCount: 450, publishedAt: "2026-07-31" },
  { id: "expected-opportunity", name: "Expected opportunity", description: "Prior-season usage quality.", methodology: "Expected fantasy points", license: "CC-BY-SA-4.0", projectUrl: "https://github.com/ffverse/ffopportunity", dataUrl: "https://example.test/opportunity.csv", defaultWeight: 0.6, importMode: "download", recordCount: 300, publishedAt: "2025" },
  { id: "cbs-ppr", name: "CBS Sports PPR Top 200", description: "Current CBS consensus.", methodology: "CBS expert consensus", license: "Proprietary; retrieved on demand", projectUrl: "https://www.cbssports.com/fantasy/football/rankings/", dataUrl: "https://www.cbssports.com/fantasy/football/rankings/", defaultWeight: 0.9, importMode: "download", recordCount: 200, publishedAt: "Updated today" },
  { id: "espn-ppr-pdf", name: "ESPN PPR Top 300 PDF", description: "User-supplied ESPN rankings.", methodology: "Overall ordinal rank", license: "Proprietary; user-supplied", projectUrl: "https://www.espn.com/fantasy/football/", dataUrl: "https://www.espn.com/fantasy/football/", defaultWeight: 0.9, importMode: "pdf-upload", recordCount: 0 },
  { id: "espn-dynasty-pdf", name: "ESPN Dynasty PDF", description: "User-supplied ESPN dynasty rankings.", methodology: "Dynasty ordinal rank", license: "Proprietary; user-supplied", projectUrl: "https://www.espn.com/fantasy/football/", dataUrl: "https://www.espn.com/fantasy/football/", defaultWeight: 0.6, importMode: "pdf-upload", recordCount: 0 },
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

  it("shows a position-only board sorted and labeled by position rank", async () => {
    vi.stubGlobal("fetch", vi.fn((input: RequestInfo | URL) => {
      const url = input instanceof Request ? input.url : input.toString();
      return jsonResponse(new URL(url).pathname.endsWith("/leagues") ? [demoLeague] : snapshot({ available: [casey, jordan, defense, kicker, alex] }));
    }));
    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByRole("button", { name: "RB" }));

    const table = screen.getByRole("table", { name: "Available RB players sorted by RB rank" });
    expect(within(table).getByRole("columnheader", { name: "RB rank" })).toBeInTheDocument();
    expect(within(table).queryByText("Jordan Hale")).not.toBeInTheDocument();
    expect(within(table).getAllByRole("rowheader").map((cell) => cell.textContent)).toEqual([
      "Alex RiversATL, bye week 12",
      "Casey BrooksDET, bye week 8",
    ]);
    expect(screen.getByText("Showing 2 available RB players of 5 total players.")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "DST" }));
    const defenseTable = screen.getByRole("table", { name: "Available DST players sorted by DST rank" });
    expect(within(defenseTable).getByRole("columnheader", { name: "DST rank" })).toBeInTheDocument();
    expect(within(defenseTable).getByText("Denver Defense")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "K" })).toBeInTheDocument();
  });

  it("shows source provenance and refreshes the accessible consensus preview", async () => {
    let savedWeights: Record<string, number> | undefined;
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const request = input instanceof Request ? input : new Request(input, init);
      const path = new URL(request.url).pathname;
      if (path.endsWith("/leagues")) return jsonResponse([demoLeague]);
      if (path.endsWith("/leagues/demo") && request.method === "PUT") {
        const rules = await request.clone().json() as Omit<League, "id">;
        savedWeights = rules.sourceWeights;
        return jsonResponse({ id: "demo", ...rules });
      }
      if (path.endsWith("/ranking-sources/import-pdf") && request.method === "POST") return jsonResponse({ source: { ...rankingSources[5], recordCount: 245, publishedAt: "2026-08-02" }, pageCount: 1 }, 201);
      if (path.endsWith("/ranking-sources/refresh") && request.method === "POST") return jsonResponse(rankingSources);
      if (path.endsWith("/ranking-sources")) return jsonResponse(rankingSources.map((source) => ({ ...source, recordCount: 0, publishedAt: undefined })));
      if (path.endsWith("/rankings")) return jsonResponse(consensusRankings);
      return jsonResponse(snapshot());
    });
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    const { container } = render(<App />);

    await user.click(await screen.findByRole("button", { name: "Ranking sources" }));
    expect(await screen.findByRole("heading", { name: "Redraft expert consensus" })).toBeInTheDocument();
    expect(screen.getByText("CC-BY-SA-4.0")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: /View source website/ })).toHaveLength(7);
    expect(screen.getByRole("heading", { name: "Platform connector status" })).toBeInTheDocument();
    expect(screen.getByText("The public overall draft table still contains the prior-season board.")).toBeInTheDocument();

    const cbsWeight = screen.getByRole("spinbutton", { name: "CBS Sports PPR Top 200 influence" });
    await user.clear(cbsWeight);
    await user.type(cbsWeight, "3");
    await user.click(screen.getByRole("button", { name: "Save weights" }));
    expect(await screen.findByText("Ranking influence saved for Demo League. Every imported source remains included.")).toBeInTheDocument();
    expect(savedWeights?.["cbs-ppr"]).toBe(3);
    expect(Object.values(savedWeights ?? {}).every((weight) => weight > 0)).toBe(true);

    const pdf = new File(["%PDF-test"], "espn-rankings.pdf", { type: "application/pdf" });
    await user.upload(screen.getByLabelText("Import a ranking PDF"), pdf);
    await user.click(screen.getByRole("button", { name: "Import PDF" }));
    expect(await screen.findByText("ESPN PPR Top 300 PDF imported: 245 players from 1 page.")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Refresh all sources" }));
    expect(await screen.findByRole("table", { name: "Top 25 blended player rankings" })).toBeInTheDocument();
    expect(screen.getByText("Rankings refreshed. 1900 source records were normalized.")).toBeInTheDocument();
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
