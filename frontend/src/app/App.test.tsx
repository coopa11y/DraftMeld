import { axe } from "jest-axe";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { DraftSnapshot, Player } from "../shared/api/types";
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

function jsonResponse(body: DraftSnapshot, status = 200) {
  return Promise.resolve(new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  }));
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("accessible draft board", () => {
  it("loads the league selected by the application shell", async () => {
    const fetchMock = vi.fn((_input: RequestInfo | URL) =>
      jsonResponse(snapshot({ leagueId: "league-a", leagueName: "League A" })),
    );
    vi.stubGlobal("fetch", fetchMock);
    render(<App leagueId="league-a" />);

    expect(await screen.findByText("League A")).toBeInTheDocument();
    const requestInput = fetchMock.mock.calls[0][0];
    const requestUrl = requestInput instanceof Request ? requestInput.url : requestInput.toString();
    expect(new URL(requestUrl).searchParams.get("leagueId")).toBe("league-a");
  });

  it("exposes landmarks, names every player action, and has no automatic axe violations", async () => {
    vi.stubGlobal("fetch", vi.fn(() => jsonResponse(snapshot())));
    const { container } = render(<App />);

    expect(await screen.findByRole("heading", { name: "Available players" })).toBeInTheDocument();
    expect(screen.getByRole("table", { name: "Available players sorted by overall rank" })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Draft Alex Rivers, RB, to my team" })).toHaveLength(2);
    expect(screen.getAllByRole("button", { name: "Mark Alex Rivers, RB, as taken by another team" })).toHaveLength(2);
    expect(screen.getByRole("link", { name: "Skip to player board" })).toBeInTheDocument();

    const results = await axe(container);
    expect(results.violations).toHaveLength(0);
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
