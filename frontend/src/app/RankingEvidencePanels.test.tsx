import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { axe } from "jest-axe";
import { describe, expect, it } from "vitest";
import type { ConsensusRanking } from "../shared/api/types";
import { RankingEvidencePanels } from "./RankingEvidencePanels";

describe("RankingEvidencePanels", () => {
  it("presents rich Draft Sharks evidence in collapsed accessible details", async () => {
    const ranking: ConsensusRanking = {
      playerKey: "jahmyrgibbs",
      name: "Jahmyr Gibbs",
      position: "RB",
      team: "DET",
      rank: 1,
      score: 1,
      sourceCount: 1,
      sourceRanks: { "draft-sharks-ppr-1qb": 1 },
      coverage: 1,
      rankRange: 0,
      confidence: "high",
      method: "weighted-median",
      adp: 1,
      tier: 1,
      projection: {
        sourceId: "draft-sharks-ppr-1qb",
        sourceName: "Draft Sharks - PPR, 1QB",
        profile: "PPR, 1QB",
        games: 17,
        byeWeek: 6,
        floorProjection: 231,
        consensusProjection: 273,
        sourceProjection: 285,
        ceilingProjection: 328,
        sourceValue: 100,
        injuryRisk: 54,
        scheduleStrength: 0.9,
      },
    };
    const user = userEvent.setup();
    const { container } = render(
      <RankingEvidencePanels
        enabledImportedSourceCount={1}
        leagueName="Test League"
        rankings={[ranking]}
        watchlist={[]}
      />,
    );
    const details = screen.getByText("Draft Sharks - PPR, 1QB projection details");
    await user.click(details);
    expect(screen.getByText(/231\.0 floor, 285\.0 Draft Sharks, 328\.0 ceiling/)).toBeVisible();
    expect(screen.getByText("54.0% injury risk, +0.9% schedule strength")).toBeVisible();
    expect((await axe(container)).violations).toHaveLength(0);
  });
});
