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
const teamNames = ["My Team", ...Array.from({ length: 11 }, (_, index) => `Team ${index + 2}`)];

function snapshot(overrides: Partial<DraftSnapshot> = {}): DraftSnapshot {
  return {
    leagueId: "demo",
    leagueName: "Demo League",
    pickNumber: 1,
    available: [alex, jordan],
    myTeam: [],
    teams: teamNames.map((name, index) => ({
      number: index + 1,
      name,
      isUser: index === 0,
      roster: [],
      auctionBudgetRemaining: 200,
    })),
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
    totalPicks: 24,
    isComplete: false,
    onClockTeamNumber: 2,
    pickSlots: Array.from({ length: 24 }, (_, index) => {
      const overallNumber = index + 1;
      const round = Math.floor(index / 12) + 1;
      const pickInRound = (index % 12) + 1;
      const originalTeamNumber = round % 2 === 1 ? pickInRound : 13 - pickInRound;
      return {
        season: 2026,
        overallNumber,
        round,
        pickInRound,
        originalTeamNumber,
        originalTeamName: teamNames[originalTeamNumber - 1],
        ownerTeamNumber: originalTeamNumber,
        ownerTeamName: teamNames[originalTeamNumber - 1],
        isUsed: false,
      };
    }),
    pickTrades: [],
    leagueFormat: "redraft",
    season: 2026,
    auctionBudgetTrades: false,
    faabTrades: false,
    budgetBalances: [],
    draftOrder: Array.from({ length: 12 }, (_, index) => index + 1),
    userTeamNumber: 1,
    sessionStatus: "in-progress",
    canReset: true,
    canUndoReset: false,
    ...overrides,
  };
}

const demoLeague: League = {
  id: "demo",
  name: "Demo League",
  teamCount: 12,
  draftPosition: 1,
  userTeamNumber: 1,
  teamNames,
  draftOrder: Array.from({ length: 12 }, (_, index) => index + 1),
  draftType: "snake",
  leagueFormat: "redraft",
  season: 2026,
  initialSeason: 2026,
  futurePickSeasons: 0,
  rookieDraftRounds: 4,
  auctionBudgetTrades: false,
  faabBudget: 100,
  faabTrades: false,
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
  it("offers non-blocking first-run guidance and a resumable setup entry point", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const request = input instanceof Request ? input : new Request(input);
        return new URL(request.url).pathname.endsWith("/leagues")
          ? jsonResponse([demoLeague])
          : jsonResponse(snapshot({ sessionStatus: "not-started" }));
      }),
    );
    const user = userEvent.setup();
    render(<App />);
    expect(await screen.findByText("Draft not started")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Set up your league" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Start setup" }));
    expect(screen.getByRole("heading", { name: "Set up your draft workspace" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Continue" }));
    await user.click(screen.getByRole("button", { name: "Save and finish later" }));
    const savedNotice = screen.getByText("Setup progress saved. Resume the setup guide whenever you are ready.");
    expect(savedNotice).toHaveAttribute("role", "status");
    await waitFor(() => expect(savedNotice).toHaveFocus());
    expect(screen.getAllByRole("button", { name: "Resume setup" }).length).toBeGreaterThan(0);
    expect(screen.getByText("Draft not started")).toBeInTheDocument();
    await user.click(screen.getAllByRole("button", { name: "Resume setup" })[0]);
    expect(screen.queryByText("Setup progress saved. Resume the setup guide whenever you are ready.")).toBeNull();
    expect(await screen.findByRole("heading", { name: "Scoring" })).toHaveFocus();
  });

  it("starts, safely resets, and restores a draft session", async () => {
    let draft = snapshot({ sessionStatus: "not-started", canReset: false, canUndoReset: false });
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const request = input instanceof Request ? input : new Request(input);
      const path = new URL(request.url).pathname;
      if (path.endsWith("/leagues")) return jsonResponse([demoLeague]);
      if (path.endsWith("/draft/session/start")) {
        draft = snapshot({ sessionStatus: "in-progress", canReset: true, canUndoReset: false });
      } else if (path.endsWith("/draft/session/reset")) {
        draft = snapshot({ sessionStatus: "not-started", canReset: false, canUndoReset: true });
      } else if (path.endsWith("/draft/session/undo-reset")) {
        draft = snapshot({ sessionStatus: "in-progress", canReset: true, canUndoReset: false });
      }
      return jsonResponse(draft);
    });
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    const { container } = render(<App />);

    expect(await screen.findByText("Draft not started")).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Mark Alex Rivers, RB, as taken by another team" })[0]).toBeDisabled();
    expect(screen.getByText("12 teams · 24 selections")).toBeInTheDocument();
    expect((await axe(container)).violations).toHaveLength(0);

    await user.click(screen.getByRole("button", { name: "Start draft" }));
    expect(await screen.findByText("In progress")).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Mark Alex Rivers, RB, as taken by another team" })[0]).toBeEnabled();

    await user.click(screen.getByRole("button", { name: "Reset current draft" }));
    expect(screen.getByRole("heading", { name: "Reset the current draft?" })).toBeInTheDocument();
    const confirmReset = screen.getByRole("button", { name: "Reset current draft" });
    expect(confirmReset).toBeDisabled();
    await user.click(screen.getByRole("checkbox", { name: /I understand/ }));
    await user.type(screen.getByRole("textbox", { name: "Type Demo League to confirm" }), "Demo League");
    expect(confirmReset).toBeEnabled();
    expect((await axe(container)).violations).toHaveLength(0);
    await user.click(confirmReset);

    expect(await screen.findByRole("button", { name: "Undo last reset" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Undo last reset" }));
    expect(await screen.findByText("In progress")).toBeInTheDocument();
    expect(
      fetchMock.mock.calls.some(([input]) => {
        const request = input instanceof Request ? input : new Request(input);
        return new URL(request.url).pathname.endsWith("/draft/session/reset");
      }),
    ).toBe(true);
  });

  it("creates a customized league through an accessible setup form", async () => {
    let configuredLeagues = [demoLeague];
    let submittedRules: Omit<League, "id"> | undefined;
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const request = input instanceof Request ? input : new Request(input);
        const path = new URL(request.url).pathname;
        if (path.endsWith("/leagues") && request.method === "GET") return jsonResponse(configuredLeagues);
        if (path.endsWith("/leagues") && request.method === "POST") {
          const rules = (await request.clone().json()) as Omit<League, "id">;
          submittedRules = rules;
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
    expect(screen.getByRole("group", { name: "League settings" })).toBeInTheDocument();
    expect(screen.getByRole("group", { name: "Draft settings" })).toBeInTheDocument();
    expect(screen.getByRole("group", { name: "Team settings" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "League format" })).toHaveValue("redraft");
    await user.selectOptions(screen.getByRole("combobox", { name: "Reception scoring preset" }), "te-premium");
    expect(screen.getByRole("spinbutton", { name: "Points per tight end reception bonus" })).toHaveValue(0.5);
    await user.click(screen.getByText("Kicking", { selector: "summary" }));
    const allFieldGoals = screen.getByRole("spinbutton", { name: "Points per field goal made (any distance)" });
    const longFieldGoals = screen.getByRole("spinbutton", { name: "Points per field goal made, 50+ yards" });
    await user.clear(allFieldGoals);
    await user.type(allFieldGoals, "0");
    await user.clear(longFieldGoals);
    await user.type(longFieldGoals, "5");
    expect(screen.queryByRole("combobox", { name: "Future pick seasons" })).not.toBeInTheDocument();
    await user.selectOptions(screen.getByRole("combobox", { name: "League format" }), "dynasty");
    expect(screen.getByRole("combobox", { name: "Future pick seasons" })).toBeInTheDocument();
    await user.selectOptions(screen.getByRole("combobox", { name: "League format" }), "redraft");
    expect(screen.queryByRole("combobox", { name: "Future pick seasons" })).not.toBeInTheDocument();
    const name = screen.getByRole("textbox", { name: "League name" });
    await user.clear(name);
    await user.type(name, "Family League");
    await user.selectOptions(screen.getByRole("combobox", { name: "Your draft position" }), "7");
    expect(screen.getByRole("textbox", { name: "Franchise 1 (your team)" })).toHaveValue("My Team");
    expect(screen.getByRole("textbox", { name: "Franchise 7" })).toHaveValue("");
    expect((await axe(container)).violations).toHaveLength(0);
    await user.click(screen.getByRole("button", { name: "Save league" }));

    expect(await screen.findByRole("heading", { name: "Family League" })).toBeInTheDocument();
    expect(screen.getByText("Family League was saved.")).toBeInTheDocument();
    expect(submittedRules?.draftPosition).toBe(7);
    expect(submittedRules?.teamNames[0]).toBe("My Team");
    expect(submittedRules?.draftOrder[6]).toBe(1);
    expect(submittedRules?.scoringRules.tightEndReceptionBonus).toBe(0.5);
    expect(submittedRules?.scoringRules.fieldGoalMade).toBe(0);
    expect(submittedRules?.scoringRules.fieldGoal50Plus).toBe(5);
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
      isUserTurn: true,
      onClockTeamNumber: 1,
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
          {
            source: { ...rankingSources[5], recordCount: 245, publishedAt: "2026-08-02" },
            pageCount: 1,
            ocrApplied: true,
          },
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
    expect(
      await screen.findByText(
        "ESPN PPR Top 300 PDF imported: 245 players from 1 page. Scanned pages were recognized locally with OCR; review the imported rankings carefully.",
      ),
    ).toBeInTheDocument();

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
        {
          eventId: 1,
          number: 1,
          action: "draft",
          player: alex,
          createdAt: new Date().toISOString(),
          cost: 0,
          teamNumber: 1,
          teamName: "My Team",
        },
        {
          eventId: 2,
          number: 2,
          action: "taken",
          player: jordan,
          createdAt: new Date().toISOString(),
          cost: 0,
          teamNumber: 2,
          teamName: "Team 2",
        },
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
      history: [
        {
          eventId: 1,
          number: 1,
          action: "draft",
          player: alex,
          createdAt: new Date().toISOString(),
          cost: 0,
          teamNumber: 1,
          teamName: "My Team",
        },
      ],
      recommendations: [{ player: jordan, score: 219, reasons: ["Fills an open starting roster need"] }],
      canUndo: true,
    });
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => jsonResponse([demoLeague]))
      .mockImplementationOnce(() => jsonResponse(snapshot({ isUserTurn: true, onClockTeamNumber: 1 })))
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
      expect(
        screen.getAllByRole("button", { name: "Mark Jordan Hale, WR, as taken by another team" })[0],
      ).toHaveFocus(),
    );
  });

  it("assigns a traded pick to the selected team", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => jsonResponse([demoLeague]))
      .mockImplementationOnce(() => jsonResponse(snapshot()))
      .mockImplementationOnce(() => jsonResponse(snapshot({ pickNumber: 3, onClockTeamNumber: 3 })));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    render(<App />);

    const pickOwner = await screen.findByRole("combobox", { name: /Owner of pick 1/ });
    await user.selectOptions(pickOwner, "1");
    expect(screen.getByText("Traded from Team 2 to My Team.")).toBeInTheDocument();
    await user.click(screen.getAllByRole("button", { name: "Draft Alex Rivers, RB, to my team" })[0]);

    const request = fetchMock.mock.calls[2][0] as Request;
    expect(await request.clone().json()).toMatchObject({ action: "draft", playerId: "p001", teamNumber: 1 });
  });

  it("creates and reviews an accessible multi-pick trade for any round", async () => {
    const tradedSlots = snapshot().pickSlots.map((slot) => {
      if (slot.overallNumber === 2) return { ...slot, ownerTeamNumber: 1, ownerTeamName: "My Team" };
      if (slot.overallNumber === 1 || slot.overallNumber === 24)
        return { ...slot, ownerTeamNumber: 2, ownerTeamName: "Team 2" };
      return slot;
    });
    const traded = snapshot({
      onClockTeamNumber: 2,
      pickSlots: tradedSlots,
      pickTrades: [
        {
          id: 1,
          leagueId: "demo",
          teamOneNumber: 1,
          teamOneName: "My Team",
          teamTwoNumber: 2,
          teamTwoName: "Team 2",
          teamOneReceives: [2],
          teamTwoReceives: [1, 24],
          teamOneFuturePicks: [],
          teamTwoFuturePicks: [],
          teamOneAuctionBudget: 0,
          teamTwoAuctionBudget: 0,
          teamOnePlayers: [],
          teamTwoPlayers: [],
          teamOneBudgets: [],
          teamTwoBudgets: [],
          season: 2026,
          createdAt: new Date().toISOString(),
        },
      ],
    });
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => jsonResponse([demoLeague]))
      .mockImplementationOnce(() => jsonResponse(snapshot()))
      .mockImplementationOnce(() => jsonResponse(traded, 201));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    const { container } = render(<App />);

    expect(await screen.findByRole("heading", { name: "Draft asset trades" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "First team" })).toHaveValue("1");
    expect(screen.getByRole("combobox", { name: "Second team" })).toHaveValue("2");
    for (const name of ["Round 1, pick 2, overall 2", "Round 1, pick 1, overall 1", "Round 2, pick 12, overall 24"]) {
      const pick = screen.getByRole("checkbox", { name });
      pick.focus();
      await user.keyboard(" ");
      expect(pick).toBeChecked();
    }
    const reviewButton = screen.getByRole("button", { name: "Review trade" });
    reviewButton.focus();
    await user.keyboard("{Enter}");

    const review = screen.getByRole("region", { name: "Review this trade" });
    expect(within(review).getByText("Round 1, pick 2, overall 2")).toBeInTheDocument();
    expect(within(review).getByText("Round 1, pick 1, overall 1; Round 2, pick 12, overall 24")).toBeInTheDocument();
    expect((await axe(container)).violations).toHaveLength(0);
    const confirmButton = within(review).getByRole("button", { name: "Confirm trade" });
    confirmButton.focus();
    await user.keyboard("{Enter}");

    expect(
      await screen.findByText("Trade confirmed between My Team and Team 2 with 1 asset and 2 assets recorded."),
    ).toBeInTheDocument();
    const request = fetchMock.mock.calls[2][0] as Request;
    expect(await request.clone().json()).toEqual({
      leagueId: "demo",
      teamOneNumber: 1,
      teamTwoNumber: 2,
      teamOneReceives: [2],
      teamTwoReceives: [1, 24],
      teamOneFuturePicks: [],
      teamTwoFuturePicks: [],
      teamOneAuctionBudget: 0,
      teamTwoAuctionBudget: 0,
      teamOnePlayers: [],
      teamTwoPlayers: [],
      teamOneBudgets: [],
      teamTwoBudgets: [],
    });
    expect(screen.getAllByText("Traded to Team 2")).toHaveLength(2);
  });

  it("shows dynasty future picks and auction budget only when league rules allow them", async () => {
    const futureSlot = {
      season: 2027,
      overallNumber: 0,
      round: 1,
      pickInRound: 0,
      originalTeamNumber: 2,
      originalTeamName: "Team 2",
      ownerTeamNumber: 2,
      ownerTeamName: "Team 2",
      isUsed: false,
    };
    const dynastyAuction = snapshot({
      draftType: "auction",
      leagueFormat: "dynasty",
      auctionBudgetTrades: true,
      budgetBalances: snapshot()
        .teams.slice(0, 2)
        .map((team) => ({
          teamNumber: team.number,
          season: 2026,
          kind: "auction" as const,
          remaining: 200,
        })),
      pickSlots: [futureSlot],
      teams: snapshot().teams.map((team) => ({
        ...team,
        auctionBudgetRemaining: 200,
        roster: team.number === 1 ? [alex] : team.number === 2 ? [jordan] : [],
      })),
    });
    const traded = {
      ...dynastyAuction,
      pickSlots: [{ ...futureSlot, ownerTeamNumber: 1, ownerTeamName: "My Team" }],
      pickTrades: [
        {
          id: 1,
          leagueId: "demo",
          teamOneNumber: 1,
          teamOneName: "My Team",
          teamTwoNumber: 2,
          teamTwoName: "Team 2",
          teamOneReceives: [],
          teamTwoReceives: [],
          teamOneFuturePicks: [
            {
              season: 2027,
              round: 1,
              originalTeamNumber: 2,
              originalTeamName: "Team 2",
              condition: "Jordan plays eight games",
              conditionStatus: "pending",
            },
          ],
          teamTwoFuturePicks: [],
          teamOneAuctionBudget: 0,
          teamTwoAuctionBudget: 25,
          teamOnePlayers: ["p002"],
          teamTwoPlayers: [],
          teamOneBudgets: [],
          teamTwoBudgets: [{ kind: "auction" as const, season: 2026, amount: 25 }],
          season: 2026,
          createdAt: new Date().toISOString(),
        },
      ],
      teams: dynastyAuction.teams.map((team) =>
        team.number === 1
          ? { ...team, auctionBudgetRemaining: 175 }
          : team.number === 2
            ? { ...team, auctionBudgetRemaining: 225 }
            : team,
      ),
    };
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() =>
        jsonResponse([
          {
            ...demoLeague,
            draftType: "auction",
            leagueFormat: "dynasty",
            futurePickSeasons: 2,
            auctionBudgetTrades: true,
          },
        ]),
      )
      .mockImplementationOnce(() => jsonResponse(dynastyAuction))
      .mockImplementationOnce(() => jsonResponse(traded, 201));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    const { container } = render(<App />);

    expect(
      await screen.findByText(
        "This dynasty league tracks players, unused 2026 assets, future rookie picks, and enabled budgets across seasons.",
      ),
    ).toBeInTheDocument();
    const futurePick = screen.getByRole("checkbox", { name: "2027, round 1, Team 2 original pick" });
    futurePick.focus();
    await user.keyboard(" ");
    await user.type(
      screen.getByRole("textbox", { name: "Condition for selected future picks received by My Team (optional)" }),
      "Jordan plays eight games",
    );
    await user.click(screen.getByText("Players from Team 2").closest("summary")!);
    await user.click(screen.getByRole("checkbox", { name: "Jordan Hale, WR" }));
    await user.click(screen.getAllByText("Auction and FAAB budgets")[1].closest("summary")!);
    const budget = screen.getByRole("spinbutton", { name: "2026 auction budget from My Team" });
    await user.clear(budget);
    await user.type(budget, "25");
    await user.click(screen.getByRole("button", { name: "Review trade" }));
    expect(screen.getByText("$25 2026 auction budget")).toBeInTheDocument();
    expect(screen.getByText(/condition: Jordan plays eight games/)).toBeInTheDocument();
    expect((await axe(container)).violations).toHaveLength(0);
    await user.click(screen.getByRole("button", { name: "Confirm trade" }));

    const request = fetchMock.mock.calls[2][0] as Request;
    expect(await request.clone().json()).toMatchObject({
      teamOneFuturePicks: [{ season: 2027, round: 1, originalTeamNumber: 2 }],
      teamTwoAuctionBudget: 0,
      teamTwoBudgets: [{ kind: "auction", season: 2026, amount: 25 }],
      teamOneReceives: [],
      teamTwoReceives: [],
      teamOnePlayers: ["p002"],
    });
    expect(
      await screen.findByText("Trade confirmed between My Team and Team 2 with 2 assets and 1 asset recorded."),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Mark condition met" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "Trade season" })).toHaveValue("all");
  });

  it("guides an accessible dynasty season rollover with a new franchise draft order", async () => {
    const current = snapshot({
      leagueFormat: "dynasty",
      history: [
        {
          eventId: 1,
          number: 1,
          action: "taken",
          player: alex,
          createdAt: new Date().toISOString(),
          cost: 0,
          teamNumber: 1,
          teamName: "My Team",
        },
      ],
    });
    const next = snapshot({
      leagueFormat: "dynasty",
      season: 2027,
      draftType: "linear",
      draftOrder: [2, 1, ...Array.from({ length: 10 }, (_, index) => index + 3)],
      history: [],
      pickNumber: 1,
      onClockTeamNumber: 2,
    });
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => jsonResponse([{ ...demoLeague, leagueFormat: "dynasty", futurePickSeasons: 2 }]))
      .mockImplementationOnce(() => jsonResponse(current))
      .mockImplementationOnce(() => jsonResponse(next));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    const { container } = render(<App />);

    const lifecycle = await screen.findByRole("region", { name: "Season 2026" });
    await user.click(within(lifecycle).getByRole("button", { name: "Close season and start 2027" }));
    await user.selectOptions(within(lifecycle).getByRole("combobox", { name: "Draft format" }), "linear");
    await user.selectOptions(within(lifecycle).getByRole("combobox", { name: "Pick 1" }), "2");
    const start = within(lifecycle).getByRole("button", { name: "Start 2027" });
    expect(start).toBeDisabled();
    await user.click(
      within(lifecycle).getByRole("checkbox", {
        name: "This draft is not complete. I understand that its existing picks will remain in the 2026 history.",
      }),
    );
    expect((await axe(container)).violations).toHaveLength(0);
    await user.click(start);

    const request = fetchMock.mock.calls[2][0] as Request;
    expect(await request.clone().json()).toMatchObject({
      leagueId: "demo",
      season: 2027,
      draftType: "linear",
      draftOrder: [2, 1, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12],
    });
    expect(
      await screen.findByText("2027 is ready. Franchise rosters and traded assets carried forward to a clean draft."),
    ).toBeInTheDocument();
  });

  it("restores the last player when undo is selected", async () => {
    const drafted = snapshot({
      pickNumber: 2,
      available: [jordan],
      myTeam: [alex],
      history: [
        {
          eventId: 1,
          number: 1,
          action: "draft",
          player: alex,
          createdAt: new Date().toISOString(),
          cost: 0,
          teamNumber: 1,
          teamName: "My Team",
        },
      ],
      recommendations: [{ player: jordan, score: 219, reasons: ["Fills an open starting roster need"] }],
      canUndo: true,
    });
    const fetchMock = vi
      .fn()
      .mockImplementationOnce(() => jsonResponse([demoLeague]))
      .mockImplementationOnce(() => jsonResponse(drafted))
      .mockImplementationOnce(() => jsonResponse(snapshot({ isUserTurn: true, onClockTeamNumber: 1 })));
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
