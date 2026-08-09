import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { DraftPickTrade, DraftSnapshot, FutureDraftPick } from "../shared/api/types";
import { DraftTradeLedger } from "./DraftTradeLedger";

afterEach(cleanup);

const conditionalPick: FutureDraftPick = {
  season: 2027,
  round: 1,
  originalTeamNumber: 2,
  originalTeamName: "Team 2",
  condition: "Player starts eight games",
  conditionStatus: "pending",
};

const trade: DraftPickTrade = {
  id: 7,
  leagueId: "demo",
  teamOneNumber: 1,
  teamOneName: "My Team",
  teamTwoNumber: 2,
  teamTwoName: "Team 2",
  teamOneReceives: [],
  teamTwoReceives: [],
  teamOneFuturePicks: [conditionalPick],
  teamTwoFuturePicks: [],
  teamOneAuctionBudget: 0,
  teamTwoAuctionBudget: 0,
  teamOnePlayers: [],
  teamTwoPlayers: [],
  teamOneBudgets: [],
  teamTwoBudgets: [],
  season: 2026,
  createdAt: "2026-08-09T00:00:00Z",
};

function ledgerSnapshot(pickTrades: DraftPickTrade[] = [trade]) {
  return {
    leagueFormat: "dynasty",
    pickTrades,
    pickSlots: [],
  } as unknown as DraftSnapshot;
}

describe("DraftTradeLedger", () => {
  it("filters seasons and exposes both condition outcomes", async () => {
    const user = userEvent.setup();
    const onSeasonChange = vi.fn();
    const onResolve = vi.fn().mockResolvedValue(undefined);
    render(
      <DraftTradeLedger
        snapshot={ledgerSnapshot()}
        playersByID={new Map()}
        ledgerSeason="all"
        onSeasonChange={onSeasonChange}
        busy={false}
        confirmDelete={null}
        onConfirmDeleteChange={vi.fn()}
        onDelete={vi.fn()}
        onResolve={onResolve}
      />,
    );

    await user.selectOptions(screen.getByRole("combobox", { name: "Trade season" }), "2026");
    expect(onSeasonChange).toHaveBeenCalledWith("2026");
    await user.click(screen.getByRole("button", { name: "Mark condition met" }));
    await user.click(screen.getByRole("button", { name: "Mark condition not met" }));
    expect(onResolve).toHaveBeenNthCalledWith(1, trade, conditionalPick, "met");
    expect(onResolve).toHaveBeenNthCalledWith(2, trade, conditionalPick, "not-met");
  });

  it("requires confirmation before reversing and supports cancellation", async () => {
    const user = userEvent.setup();
    const onConfirmDeleteChange = vi.fn();
    const onDelete = vi.fn().mockResolvedValue(undefined);
    const { rerender } = render(
      <DraftTradeLedger
        snapshot={ledgerSnapshot()}
        playersByID={new Map()}
        ledgerSeason="all"
        onSeasonChange={vi.fn()}
        busy={false}
        confirmDelete={null}
        onConfirmDeleteChange={onConfirmDeleteChange}
        onDelete={onDelete}
        onResolve={vi.fn()}
      />,
    );
    await user.click(screen.getByRole("button", { name: "Reverse trade" }));
    expect(onConfirmDeleteChange).toHaveBeenCalledWith(7);

    rerender(
      <DraftTradeLedger
        snapshot={ledgerSnapshot()}
        playersByID={new Map()}
        ledgerSeason="all"
        onSeasonChange={vi.fn()}
        busy={false}
        confirmDelete={7}
        onConfirmDeleteChange={onConfirmDeleteChange}
        onDelete={onDelete}
        onResolve={vi.fn()}
      />,
    );
    const reversal = screen.getByRole("group", { name: "Reverse trade between My Team and Team 2" });
    await user.click(screen.getByRole("button", { name: "Keep trade" }));
    expect(onConfirmDeleteChange).toHaveBeenLastCalledWith(null);
    await user.click(screen.getByRole("button", { name: "Confirm reversal" }));
    expect(onDelete).toHaveBeenCalledWith(trade);
    expect(reversal).toBeInTheDocument();
  });

  it("announces an empty filtered ledger", () => {
    render(
      <DraftTradeLedger
        snapshot={ledgerSnapshot([])}
        playersByID={new Map()}
        ledgerSeason="all"
        onSeasonChange={vi.fn()}
        busy={false}
        confirmDelete={null}
        onConfirmDeleteChange={vi.fn()}
        onDelete={vi.fn()}
        onResolve={vi.fn()}
      />,
    );
    expect(screen.getByText("No draft trades recorded for this season filter.")).toBeInTheDocument();
  });
});
