import { axe } from "jest-axe";
import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it } from "vitest";
import type { Player, PlayerNewsUpdate } from "../shared/api/types";
import { PlayerNewsDetails } from "./PlayerNewsDetails";

const player: Player = {
  id: "player-1",
  name: "Jordan Hale",
  nflTeam: "MIN",
  position: "WR",
  byeWeek: 6,
  overallRank: 2,
  positionRank: 1,
  adp: 2.5,
  tier: 1,
  projectedPoints: 250,
  valueOverReplacement: 80,
  confidence: "high",
  rankRange: 2,
  preference: "",
  auctionValue: 42,
};

const update: PlayerNewsUpdate = {
  playerId: player.id,
  playerName: player.name,
  availability: {
    playerId: player.id,
    status: "Questionable",
    injury: "Hamstring",
    practiceParticipation: "Limited",
    sourceId: "sleeper-player-status",
    sourceName: "Sleeper player status",
    updatedAt: "2026-08-20T15:00:00Z",
  },
  events: [
    {
      id: "news-1",
      sourceId: "espn-nfl-news",
      sourceName: "ESPN NFL news",
      playerIds: [player.id],
      playerNames: [player.name],
      category: "injury",
      title: "Jordan Hale limited in practice",
      url: "https://www.espn.com/example",
      publishedAt: "2026-08-20T14:00:00Z",
      observedAt: "2026-08-20T14:01:00Z",
      confidence: "medium",
    },
  ],
};

afterEach(cleanup);

it("shows a concise status and moves focus into player news details", async () => {
  const user = userEvent.setup();
  const { container } = render(<PlayerNewsDetails player={player} update={update} />);

  expect(screen.getByText("Questionable · Hamstring")).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "View news details for Jordan Hale" }));

  expect(screen.getByRole("heading", { name: "Jordan Hale" })).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "Close" })).toHaveFocus();
  expect(screen.getByText("Limited")).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "Jordan Hale limited in practice" })).toHaveAttribute(
    "href",
    "https://www.espn.com/example",
  );
  expect((await axe(container)).violations).toHaveLength(0);

  await user.click(screen.getByRole("button", { name: "Close" }));
  expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
});

it("does not add a control when an active player has no stories", () => {
  render(
    <PlayerNewsDetails
      player={player}
      update={{
        playerId: player.id,
        playerName: player.name,
        availability: { ...update.availability!, status: "Active", injury: "", practiceParticipation: "" },
        events: [],
      }}
    />,
  );
  expect(screen.queryByRole("button")).not.toBeInTheDocument();
});
