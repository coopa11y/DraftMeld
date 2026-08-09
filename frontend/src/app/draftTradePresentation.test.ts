import { describe, expect, it } from "vitest";
import type { DraftPickSlot, DraftSnapshot, FutureDraftPick } from "../shared/api/types";
import {
  currentDraftPicks,
  draftPickKey,
  draftPickLabel,
  draftTeamName,
  futureDraftPickLabel,
  futureDraftPicks,
  updateTradeBudget,
} from "./draftTradePresentation";

const currentPick: DraftPickSlot = {
  season: 2026,
  overallNumber: 12,
  round: 2,
  pickInRound: 2,
  originalTeamNumber: 2,
  originalTeamName: "Second Team",
  ownerTeamNumber: 1,
  ownerTeamName: "First Team",
  isUsed: false,
};

const futurePick: DraftPickSlot = {
  ...currentPick,
  season: 2027,
  overallNumber: 0,
  round: 1,
  pickInRound: 0,
};

describe("draft trade presentation", () => {
  it("separates current and future picks without mutating source slots", () => {
    const slots = [currentPick, futurePick];
    expect(currentDraftPicks(slots)).toEqual([12]);
    expect(futureDraftPicks(slots, "  makes playoffs ")).toEqual([
      expect.objectContaining({ season: 2027, round: 1, condition: "makes playoffs", conditionStatus: "" }),
    ]);
    expect(slots[1].overallNumber).toBe(0);
  });

  it("creates stable keys and human-readable labels", () => {
    expect(draftPickKey(currentPick)).toBe("2026:overall:12");
    expect(draftPickKey(futurePick)).toBe("2027:round:1:team:2");
    expect(draftPickLabel(currentPick, "redraft")).toBe("Round 2, pick 2, overall 12");
    expect(draftPickLabel(futurePick, "dynasty")).toContain("Second Team original pick");
    expect(
      futureDraftPickLabel({
        ...(futureDraftPicks([futurePick], "makes playoffs")[0] as FutureDraftPick),
        conditionStatus: "not-met",
      }),
    ).toContain("transfer void");
  });

  it("replaces one budget entry and resolves team fallback labels", () => {
    expect(updateTradeBudget([{ kind: "faab", season: 2026, amount: 10 }], "faab", 2026, 25)).toEqual([
      { kind: "faab", season: 2026, amount: 25 },
    ]);
    expect(updateTradeBudget([{ kind: "faab", season: 2026, amount: 10 }], "faab", 2026, 0)).toEqual([]);
    expect(draftTeamName({ teams: [] } as unknown as DraftSnapshot, 4)).toBe("Team 4");
  });
});
