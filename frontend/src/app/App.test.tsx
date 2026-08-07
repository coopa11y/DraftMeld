import { axe } from "jest-axe";
import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import type {
  ConsensusRanking,
  DraftSnapshot,
  IdentityIssue,
  League,
  Player,
  RankingSource,
} from "../shared/api/types";
import { App } from "./App";

const playerIntelligence = {
  projectedPoints: 0,
  valueOverReplacement: 0,
  confidence: "demo",
  rankRange: 0,
  preference: "" as const,
  auctionValue: 0,
};

const alex: Player = {
  id: "p001",
  name: "Alex Rivers",
  nflTeam: "ATL",
  position: "RB",
  byeWeek: 12,
  overallRank: 1,
  positionRank: 1,
  adp: 1.8,
  tier: 1,
  ...playerIntelligence,
};
const jordan: Player = {
  id: "p002",
  name: "Jordan Hale",
  nflTeam: "MIN",
  position: "WR",
  byeWeek: 6,
  overallRank: 2,
  positionRank: 1,
  adp: 2.5,
  tier: 1,
  ...playerIntelligence,
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
    dataMode: "demo",
    projectionCount: 0,
    draftType: "snake",
    nextUserPick: 13,
    auctionBudget: 200,
    budgetRemaining: 0,
    auctionInflation: 1,
    auctionMinimumBid: 1,
    maximumBid: 0,
    isUserTurn: false,
    ...overrides,
  };
}

const demoLeague: League = {
  id: "demo",
  name: "Demo League",
  teamCount: 12,
  draftPosition: 1,
  draftType: "snake",
  rosterSlots: [{ name: "RB", count: 2, positions: ["RB"], isStarting: true }],
  scoringRules: { reception: 1 },
  sourcePreferences: {
    "redraft-ecr": { weight: 1, enabled: true },
    "dynasty-1qb": { weight: 0.7, enabled: true },
    "dynasty-superflex": { weight: 0.5, enabled: true },
    "expected-opportunity": { weight: 0.6, enabled: true },
    "cbs-ppr": { weight: 0.9, enabled: true },
    "espn-ppr-pdf": { weight: 0.9, enabled: true },
    "espn-dynasty-pdf": { weight: 0.6, enabled: true },
  },
  consensusMethod: "weighted-median",
  playerPreferences: {},
  auctionBudget: 200,
  auctionMinimumBid: 1,
  keeperBudgetSpent: 0,
  myKeeperSpend: 0,
  keeperValueRemoved: 0,
};
const casey: Player = {
  id: "p003",
  name: "Casey Brooks",
  nflTeam: "DET",
  position: "RB",
  byeWeek: 8,
  overallRank: 10,
  positionRank: 2,
  adp: 11.4,
  tier: 2,
  ...playerIntelligence,
};
const kicker: Player = {
  id: "p004",
  name: "Avery Cole",
  nflTeam: "DAL",
  position: "K",
  byeWeek: 10,
  overallRank: 11,
  positionRank: 1,
  adp: 145.2,
  tier: 1,
  ...playerIntelligence,
};
const defense: Player = {
  id: "p005",
  name: "Denver Defense",
  nflTeam: "DEN",
  position: "DST",
  byeWeek: 12,
  overallRank: 12,
  positionRank: 1,
  adp: 137.8,
  tier: 1,
  ...playerIntelligence,
};

const rankingSources: RankingSource[] = [
  {
    id: "redraft-ecr",
    name: "Redraft expert consensus",
    description: "Current overall redraft consensus.",
    methodology: "Average expert rank",
    license: "GPL-3.0",
    projectUrl: "https://github.com/dynastyprocess/data",
    dataUrl: "https://example.test/ecr.csv",
    defaultWeight: 1,
    importMode: "download",
    role: "ranking",
    isCustom: false,
    recordCount: 500,
    publishedAt: "2026-07-31",
  },
  {
    id: "dynasty-1qb",
    name: "Dynasty market - 1 QB",
    description: "Long-term 1-QB values.",
    methodology: "Normalized player value",
    license: "GPL-3.0",
    projectUrl: "https://github.com/dynastyprocess/data",
    dataUrl: "https://example.test/1qb.csv",
    defaultWeight: 0.7,
    importMode: "download",
    role: "market",
    isCustom: false,
    recordCount: 450,
    publishedAt: "2026-07-31",
  },
  {
    id: "dynasty-superflex",
    name: "Dynasty market - Superflex",
    description: "Long-term Superflex values.",
    methodology: "Normalized Superflex value",
    license: "GPL-3.0",
    projectUrl: "https://github.com/dynastyprocess/data",
    dataUrl: "https://example.test/superflex.csv",
    defaultWeight: 0.5,
    importMode: "download",
    role: "market",
    isCustom: false,
    recordCount: 450,
    publishedAt: "2026-07-31",
  },
  {
    id: "expected-opportunity",
    name: "Expected opportunity",
    description: "Prior-season usage quality.",
    methodology: "Expected fantasy points",
    license: "CC-BY-SA-4.0",
    projectUrl: "https://github.com/ffverse/ffopportunity",
    dataUrl: "https://example.test/opportunity.csv",
    defaultWeight: 0.6,
    importMode: "download",
    role: "usage",
    isCustom: false,
    recordCount: 300,
    publishedAt: "2025",
  },
  {
    id: "cbs-ppr",
    name: "CBS Sports PPR Top 200",
    description: "Current CBS consensus.",
    methodology: "CBS expert consensus",
    license: "Proprietary; retrieved on demand",
    projectUrl: "https://www.cbssports.com/fantasy/football/rankings/",
    dataUrl: "https://www.cbssports.com/fantasy/football/rankings/",
    defaultWeight: 0.9,
    importMode: "download",
    role: "ranking",
    isCustom: false,
    recordCount: 200,
    publishedAt: "Updated today",
  },
  {
    id: "espn-ppr-pdf",
    name: "ESPN PPR Top 300 PDF",
    description: "User-supplied ESPN rankings.",
    methodology: "Overall ordinal rank",
    license: "Proprietary; user-supplied",
    projectUrl: "https://www.espn.com/fantasy/football/",
    dataUrl: "https://www.espn.com/fantasy/football/",
    defaultWeight: 0.9,
    importMode: "pdf-upload",
    role: "ranking",
    isCustom: false,
    recordCount: 0,
  },
  {
    id: "espn-dynasty-pdf",
    name: "ESPN Dynasty PDF",
    description: "User-supplied ESPN dynasty rankings.",
    methodology: "Dynasty ordinal rank",
    license: "Proprietary; user-supplied",
    projectUrl: "https://www.espn.com/fantasy/football/",
    dataUrl: "https://www.espn.com/fantasy/football/",
    defaultWeight: 0.6,
    importMode: "pdf-upload",
    role: "market",
    isCustom: false,
    recordCount: 0,
  },
];

const consensusRankings: ConsensusRanking[] = [
  {
    playerKey: "alexrivers",
    name: "Alex Rivers",
    position: "RB",
    team: "ATL",
    rank: 1,
    score: 1.5,
    sourceCount: 4,
    sourceRanks: { "redraft-ecr": 1 },
    coverage: 0.8,
    rankRange: 4,
    confidence: "high",
    method: "weighted-median",
    adp: 1.8,
    tier: 1,
  },
];

const identityIssue: IdentityIssue = {
  issueKey: "dell|WR|HOU",
  reason: "Similar names share a team and position",
  resolution: "",
  canonicalPlayerKey: "",
  candidates: [
    { playerKey: "nathanieldell", name: "Nathaniel Dell", position: "WR", team: "HOU" },
    { playerKey: "tankdell", name: "Tank Dell", position: "WR", team: "HOU" },
  ],
};

function jsonResponse(body: unknown, status = 200) {
  return Promise.resolve(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    }),
  );
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  localStorage.clear();
});

describe("accessible draft board", () => {
  it("creates a customized league through an accessible setup form", async () => {
    let configuredLeagues = [demoLeague];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const request = input instanceof Request ? input : new Request(input);
        const path = new URL(request.url).pathname;
        if (path.endsWith("/leagues") && request.method === "GET") return jsonResponse(configuredLeagues);
        if (path.endsWith("/leagues") && request.method === "POST") {
          const rules = (await request.clone().json()) as Omit<League, "id">;
          const created = { id: "family-league", ...rules };
          configuredLeagues = [...configuredLeagues, created];
          return jsonResponse(created, 201);
        }
        return jsonResponse(snapshot());
      }),
    );
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

  it("restores a versioned league backup without overwriting existing leagues", async () => {
    let configuredLeagues = [demoLeague];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const request = input instanceof Request ? input : new Request(input);
        const path = new URL(request.url).pathname;
        if (path.endsWith("/leagues") && request.method === "GET") return jsonResponse(configuredLeagues);
        if (path.endsWith("/leagues/import") && request.method === "POST") {
          const backup = (await request.clone().json()) as { rules: Omit<League, "id"> };
          const restored = { id: "demo-league", ...backup.rules };
          configuredLeagues = [...configuredLeagues, restored];
          return jsonResponse(restored, 201);
        }
        return jsonResponse(snapshot());
      }),
    );
    const user = userEvent.setup();
    const { container } = render(<App />);

    await user.click(await screen.findByRole("button", { name: "Manage leagues" }));
    expect(screen.getByRole("heading", { name: "Export and backup center" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Consensus rankings (CSV)" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Draft results (JSON)" })).toBeInTheDocument();

    const { id: originalLeagueId, ...rules } = demoLeague;
    const backup = new File(
      [
        JSON.stringify({
          formatVersion: 1,
          exportedAt: "2026-08-07T12:00:00Z",
          originalLeagueId,
          rules,
          recommendationPolicy: {
            baseScore: 200,
            startingNeedBonus: 24,
            adpValueThreshold: 5,
            scarcityBonus: 8,
            scarcityDropOff: 5,
            recommendationLimit: 5,
          },
        }),
      ],
      "demo-backup.json",
      { type: "application/json" },
    );
    await user.upload(screen.getByLabelText("DraftMeld backup file"), backup);
    fireEvent.submit(screen.getByRole("button", { name: "Restore as new league" }).closest("form")!);

    await waitFor(() => expect(screen.getAllByRole("heading", { name: "Demo League", level: 2 })).toHaveLength(2));
    expect(screen.getByText("Demo League was restored as a new league.")).toBeInTheDocument();
    expect(screen.getAllByText("Demo League").length).toBeGreaterThan(1);
    expect((await axe(container)).violations).toHaveLength(0);
  });

  it("loads the league selected by the application shell", async () => {
    localStorage.setItem("draftmeld.active-league.v1", "league-a");
    const leagueA = { ...demoLeague, id: "league-a", name: "League A" };
    const fetchMock = vi
      .fn()
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
    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL) => {
        const url = input instanceof Request ? input.url : input.toString();
        return jsonResponse(new URL(url).pathname.endsWith("/leagues") ? [demoLeague] : snapshot());
      }),
    );
    const { container } = render(<App />);

    const draftHeading = await screen.findByRole("heading", { name: "Available players" });
    await waitFor(() => expect(draftHeading).toHaveFocus());
    expect(screen.getByRole("table", { name: "Available players sorted by overall rank" })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Draft Alex Rivers, RB, to my team" })).toHaveLength(2);
    expect(screen.getAllByRole("button", { name: "Mark Alex Rivers, RB, as taken by another team" })).toHaveLength(2);
    expect(screen.getByRole("link", { name: "Skip to player board" })).toBeInTheDocument();

    const results = await axe(container);
    expect(results.violations).toHaveLength(0);
  });

  it("moves focus to each client-side view and keeps load errors inside main content", async () => {
    let failDraft = false;
    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL) => {
        const url = input instanceof Request ? input.url : input.toString();
        const path = new URL(url).pathname;
        if (path.endsWith("/leagues")) return jsonResponse([demoLeague]);
        if (path.endsWith("/ranking-sources")) return jsonResponse(rankingSources);
        if (
          path.endsWith("/projection-sources") ||
          path.endsWith("/ranking-identities") ||
          path.endsWith("/ranking-watchlist") ||
          path.endsWith("/rankings")
        )
          return jsonResponse([]);
        if (path.endsWith("/player-directory/status"))
          return jsonResponse({ playerCount: 0, identityCount: 0, providerIdCount: 0 });
        if (failDraft) return jsonResponse({ error: "Unable to load the draft." }, 500);
        return jsonResponse(snapshot());
      }),
    );
    const user = userEvent.setup();
    const firstRender = render(<App />);

    const rbFilter = await screen.findByRole("button", { name: "RB" });
    rbFilter.focus();
    await user.keyboard(" ");
    expect(rbFilter).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("table", { name: "Available RB players sorted by RB rank" })).toBeInTheDocument();

    const rankingNavigation = screen.getByRole("button", { name: "Ranking sources" });
    rankingNavigation.focus();
    await user.keyboard("{Enter}");
    const rankingHeading = await screen.findByRole("heading", { name: "Ranking sources" });
    await waitFor(() => expect(rankingHeading).toHaveFocus());
    expect(screen.getByRole("region", { name: "Player identity review" })).toBeInTheDocument();
    expect(screen.getByLabelText("Canonical player directory status")).toHaveTextContent("Players0");
    await user.click(screen.getByRole("button", { name: "Manage leagues" }));
    const leaguesHeading = await screen.findByRole("heading", { name: "Your leagues" });
    await waitFor(() => expect(leaguesHeading).toHaveFocus());

    firstRender.unmount();
    failDraft = true;
    render(<App />);
    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("Unable to load the draft.");
    expect(alert.closest("main")).not.toBeNull();
  });

  it("shows a position-only board sorted and labeled by position rank", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL) => {
        const url = input instanceof Request ? input.url : input.toString();
        return jsonResponse(
          new URL(url).pathname.endsWith("/leagues")
            ? [demoLeague]
            : snapshot({ available: [casey, jordan, defense, kicker, alex] }),
        );
      }),
    );
    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByRole("button", { name: "RB" }));

    const table = screen.getByRole("table", { name: "Available RB players sorted by RB rank" });
    expect(within(table).getByRole("columnheader", { name: "RB rank" })).toBeInTheDocument();
    expect(within(table).queryByText("Jordan Hale")).not.toBeInTheDocument();
    expect(
      within(table)
        .getAllByRole("rowheader")
        .map((cell) => cell.textContent),
    ).toEqual(["Alex RiversATL, bye week 12", "Casey BrooksDET, bye week 8"]);
    expect(screen.getByText("Showing 2 available RB players of 5 total players.")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "DST" }));
    const defenseTable = screen.getByRole("table", { name: "Available DST players sorted by DST rank" });
    expect(within(defenseTable).getByRole("columnheader", { name: "DST rank" })).toBeInTheDocument();
    expect(within(defenseTable).getByText("Denver Defense")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "K" })).toBeInTheDocument();
  });

  it("persists targets and can simulate opponents to the next turn", async () => {
    const targetedAlex = { ...alex, preference: "target" as const };
    const targeted = snapshot({
      available: [targetedAlex, jordan],
      recommendations: [{ player: targetedAlex, score: 250, reasons: ["Marked as one of your targets"] }],
    });
    const simulated = snapshot({
      pickNumber: 24,
      available: [targetedAlex],
      recommendations: [{ player: targetedAlex, score: 250, reasons: ["Marked as one of your targets"] }],
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = input instanceof Request ? input : new Request(input, init);
        const path = new URL(request.url).pathname;
        if (path.endsWith("/leagues")) return jsonResponse([demoLeague]);
        if (path.endsWith("/draft/preferences")) return jsonResponse(targeted);
        if (path.endsWith("/draft/mock")) return jsonResponse(simulated);
        return jsonResponse(snapshot());
      }),
    );
    const user = userEvent.setup();
    render(<App />);
    const targetButtons = await screen.findAllByRole("button", { name: "Target" });
    await user.click(targetButtons[0]);
    expect(
      await screen.findByText("Alex Rivers was added to your target list. Recommendations updated."),
    ).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Target" })[0]).toHaveAttribute("aria-pressed", "true");
    await user.click(screen.getByRole("button", { name: /Simulate to my next turn/ }));
    expect(await screen.findByText("Mock opponents completed. It is now pick 24.")).toBeInTheDocument();
  });

  it("shows source provenance and refreshes the accessible consensus preview", async () => {
    let savedPreferences: League["sourcePreferences"] | undefined;
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const request = input instanceof Request ? input : new Request(input, init);
      const path = new URL(request.url).pathname;
      if (path.endsWith("/leagues")) return jsonResponse([demoLeague]);
      if (path.endsWith("/leagues/demo") && request.method === "PUT") {
        const rules = (await request.clone().json()) as Omit<League, "id">;
        savedPreferences = rules.sourcePreferences;
        return jsonResponse({ id: "demo", ...rules });
      }
      if (path.endsWith("/ranking-sources/import-pdf") && request.method === "POST")
        return jsonResponse(
          { source: { ...rankingSources[5], recordCount: 245, publishedAt: "2026-08-02" }, pageCount: 1 },
          201,
        );
      if (path.endsWith("/ranking-sources/refresh") && request.method === "POST") return jsonResponse(rankingSources);
      if (path.endsWith("/ranking-sources"))
        return jsonResponse(rankingSources.map((source) => ({ ...source, recordCount: 0, publishedAt: undefined })));
      if (path.endsWith("/projection-sources")) return jsonResponse([]);
      if (path.endsWith("/ranking-identities")) return jsonResponse([]);
      if (path.endsWith("/player-directory/status"))
        return jsonResponse({ playerCount: 12, identityCount: 12, providerIdCount: 0 });
      if (path.endsWith("/ranking-watchlist"))
        return jsonResponse(
          savedPreferences?.["cbs-ppr"]?.enabled === false
            ? [
                {
                  playerKey: "alexrivers",
                  name: "Alex Rivers",
                  position: "RB",
                  team: "ATL",
                  consensusRank: 42,
                  signals: [
                    { sourceId: "cbs-ppr", sourceName: "CBS Sports PPR Top 200", sourceRank: 11, spotsHigher: 31 },
                  ],
                },
              ]
            : [],
        );
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
    expect(
      screen.getByText("The public overall draft table still contains the prior-season board."),
    ).toBeInTheDocument();

    const cbsWeight = screen.getByRole("spinbutton", { name: "CBS Sports PPR Top 200 influence" });
    await user.clear(cbsWeight);
    await user.type(cbsWeight, "1");
    await user.click(screen.getByRole("checkbox", { name: "Include CBS Sports PPR Top 200 in consensus" }));
    await user.click(screen.getByRole("button", { name: "Save preferences" }));
    expect(
      await screen.findByText(
        "Ranking preferences saved for Demo League. Excluded sources are still checked for players worth another look.",
      ),
    ).toBeInTheDocument();
    expect(savedPreferences?.["cbs-ppr"]).toEqual({ weight: 1, enabled: false });
    expect(savedPreferences?.["redraft-ecr"]).toEqual({ weight: 1, enabled: true });
    expect(await screen.findByRole("heading", { name: "Worth another look" })).toBeInTheDocument();
    expect(
      screen.getByText("CBS Sports PPR Top 200 ranks this player #11, 31 spots above consensus #42."),
    ).toBeInTheDocument();

    const pdf = new File(["%PDF-test"], "espn-rankings.pdf", { type: "application/pdf" });
    await user.upload(screen.getByLabelText("Import a ranking PDF"), pdf);
    await user.click(screen.getByRole("button", { name: "Import PDF" }));
    expect(await screen.findByText("ESPN PPR Top 300 PDF imported: 245 players from 1 page.")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Refresh all sources" }));
    expect(await screen.findByRole("table", { name: "Top 25 blended player rankings" })).toBeInTheDocument();
    expect(screen.getByText("Rankings refreshed. 1900 source records were normalized.")).toBeInTheDocument();
    expect((await axe(container)).violations).toHaveLength(0);
  });

  it("maps projection columns and merges player aliases", async () => {
    class CompatibleFormData {
      private readonly values = new Map<string, string | File>();

      append(name: string, value: string | File) {
        this.values.set(name, value);
      }

      get(name: string) {
        return this.values.get(name) ?? null;
      }
    }

    vi.stubGlobal("FormData", CompatibleFormData);
    let importedRankingMapping: Record<string, string> | undefined;
    let importedMapping: Record<string, string> | undefined;
    let identityReview: Record<string, string> | undefined;
    const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const path = new URL(input instanceof Request ? input.url : input.toString()).pathname;
      if (path.endsWith("/ranking-sources/import-csv")) {
        const form = init?.body as FormData;
        importedRankingMapping = JSON.parse(String(form.get("mapping"))) as Record<string, string>;
        return jsonResponse(
          {
            id: "custom-marcus-board",
            name: "Marcus board",
            description: "Private ranking CSV uploaded by the user.",
            methodology: "User-supplied ordinal player ranking",
            license: "Private user data",
            projectUrl: "",
            dataUrl: "",
            defaultWeight: 1,
            importMode: "csv-upload",
            role: "ranking",
            isCustom: true,
            recordCount: 1,
            refreshedAt: new Date().toISOString(),
            publishedAt: "Private CSV import",
          },
          201,
        );
      }
      if (path.endsWith("/projection-sources/import-csv")) {
        const form = init?.body as FormData;
        importedMapping = JSON.parse(String(form.get("mapping"))) as Record<string, string>;
        return jsonResponse(
          { id: "mapped", name: "Mapped model", recordCount: 1, importedAt: new Date().toISOString() },
          201,
        );
      }
      const request = input instanceof Request ? input : new Request(input, init);
      if (path.endsWith("/leagues")) return jsonResponse([demoLeague]);
      if (path.endsWith("/ranking-sources")) return jsonResponse(rankingSources);
      if (path.endsWith("/projection-sources")) return jsonResponse([]);
      if (path.endsWith("/ranking-identities/review")) {
        identityReview = (await request.json()) as Record<string, string>;
        return new Response(null, { status: 204 });
      }
      if (path.endsWith("/ranking-identities")) return jsonResponse([identityIssue]);
      if (path.endsWith("/player-directory/status"))
        return jsonResponse({ playerCount: 2, identityCount: 2, providerIdCount: 1 });
      if (path.endsWith("/ranking-watchlist")) return jsonResponse([]);
      if (path.endsWith("/rankings")) return jsonResponse(consensusRankings);
      return jsonResponse(snapshot());
    });
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole("button", { name: "Ranking sources" }));

    await user.type(screen.getByRole("textbox", { name: "Ranking source name" }), "Marcus board");
    await user.upload(
      screen.getByLabelText("Ranking CSV"),
      new File(["RK,Player,POS,TM\n1,Alex Rivers,RB,ATL\n"], "rankings.csv", { type: "text/csv" }),
    );
    await user.click(screen.getByRole("button", { name: "Import rankings" }));
    await waitFor(() =>
      expect(importedRankingMapping).toMatchObject({ rank: "RK", name: "Player", position: "POS", team: "TM" }),
    );
    expect(await screen.findByText("Marcus board imported with 1 ranked players.")).toBeInTheDocument();
    expect(screen.getByText("Private upload")).toBeInTheDocument();

    await user.type(screen.getByRole("textbox", { name: "Projection source name" }), "Mapped model");
    await user.upload(
      screen.getByLabelText("Projection CSV"),
      new File(["Player Full Name,Pos,Tm,Rec Total\nAlex Rivers,RB,ATL,72\n"], "mapped.csv", { type: "text/csv" }),
    );
    expect(await screen.findByRole("combobox", { name: "Player name" })).toHaveValue("Player Full Name");
    await user.click(screen.getByText("Map optional scoring columns"));
    await user.selectOptions(screen.getByRole("combobox", { name: "Receptions" }), "Rec Total");
    const importButton = screen.getByRole("button", { name: "Import projections" });
    expect(importButton).toBeEnabled();
    await user.click(importButton);
    await waitFor(() =>
      expect(
        fetchMock.mock.calls.some(([input]) =>
          new URL(input instanceof Request ? input.url : input.toString()).pathname.endsWith(
            "/projection-sources/import-csv",
          ),
        ),
      ).toBe(true),
    );
    await waitFor(() =>
      expect(importedMapping).toMatchObject({
        name: "Player Full Name",
        position: "Pos",
        team: "Tm",
        reception: "Rec Total",
      }),
    );
    expect(await screen.findByText("Mapped model imported with 1 granular player projections.")).toBeInTheDocument();

    await user.selectOptions(screen.getByRole("combobox", { name: "Canonical player" }), "tankdell");
    await user.click(screen.getByRole("button", { name: "Merge aliases" }));
    await waitFor(() =>
      expect(identityReview).toMatchObject({
        issueKey: "dell|WR|HOU",
        resolution: "merged",
        canonicalPlayerKey: "tankdell",
      }),
    );
    expect(screen.getByText("Merged")).toBeInTheDocument();
  });

  it("shows Sleeper reconciliation counts after a read-only sync", async () => {
    const reconciled = snapshot({
      pickNumber: 3,
      history: [
        { eventId: 1, number: 1, action: "draft", player: alex, createdAt: new Date().toISOString(), cost: 0 },
        { eventId: 2, number: 2, action: "taken", player: jordan, createdAt: new Date().toISOString(), cost: 0 },
      ],
      myTeam: [alex],
      available: [],
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = input instanceof Request ? input : new Request(input, init);
        const path = new URL(request.url).pathname;
        if (path.endsWith("/leagues")) return jsonResponse([demoLeague]);
        if (path.endsWith("/draft/sync/sleeper"))
          return jsonResponse({ snapshot: reconciled, added: 1, updated: 1, removed: 2, unmatched: 1 });
        return jsonResponse(snapshot());
      }),
    );
    const user = userEvent.setup();
    render(<App />);
    await user.type(await screen.findByRole("textbox", { name: "Draft ID" }), "draft-123");
    await user.click(screen.getByRole("button", { name: "Sync picks" }));
    expect(
      await screen.findByText("Sleeper sync reconciled 2 picks: 1 added, 1 changed, 2 removed, 1 unmatched."),
    ).toBeInTheDocument();
    expect(screen.getByText("Alex Rivers, ATL")).toBeInTheDocument();
  });

  it("announces a draft, updates the team, and moves focus to the next available player", async () => {
    const drafted = snapshot({
      pickNumber: 2,
      available: [jordan],
      myTeam: [alex],
      history: [{ eventId: 1, number: 1, action: "draft", player: alex, createdAt: new Date().toISOString(), cost: 0 }],
      recommendations: [{ player: jordan, score: 219, reasons: ["Fills an open starting roster need"] }],
      canUndo: true,
    });
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => jsonResponse([demoLeague]))
      .mockImplementationOnce(() => jsonResponse(snapshot()))
      .mockImplementationOnce(() => jsonResponse(drafted));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    render(<App />);

    const draftButtons = await screen.findAllByRole("button", { name: "Draft Alex Rivers, RB, to my team" });
    await user.click(draftButtons[0]);

    expect(await screen.findByText("Alex Rivers, ATL")).toBeInTheDocument();
    expect(
      screen.getByText("Alex Rivers was drafted to your team. Rankings and recommendations updated."),
    ).toBeInTheDocument();
    await waitFor(() =>
      expect(screen.getAllByRole("button", { name: "Draft Jordan Hale, WR, to my team" })[0]).toHaveFocus(),
    );
  });

  it("restores the last player when undo is selected", async () => {
    const drafted = snapshot({
      pickNumber: 2,
      available: [jordan],
      myTeam: [alex],
      history: [{ eventId: 1, number: 1, action: "draft", player: alex, createdAt: new Date().toISOString(), cost: 0 }],
      recommendations: [{ player: jordan, score: 219, reasons: ["Fills an open starting roster need"] }],
      canUndo: true,
    });
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => jsonResponse([demoLeague]))
      .mockImplementationOnce(() => jsonResponse(drafted))
      .mockImplementationOnce(() => jsonResponse(snapshot()));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByRole("button", { name: "Undo last action" }));

    expect(await screen.findByText("Alex Rivers was restored to the available-player list.")).toBeInTheDocument();
    await waitFor(() =>
      expect(screen.getAllByRole("button", { name: "Draft Alex Rivers, RB, to my team" })[0]).toHaveFocus(),
    );
  });
});
