import { describe, expect, it } from "vitest";
import { buildQuickDemoRules, defaultQuickDemoSettings } from "./quickDemoRules";

describe("quick demo league rules", () => {
  it("builds a compact league from the selected common settings", () => {
    const rules = buildQuickDemoRules({
      ...defaultQuickDemoSettings,
      name: "  Practice Night  ",
      teamCount: 10,
      draftPosition: 6,
      receptionPreset: "1",
      draftType: "linear",
    });

    expect(rules).toMatchObject({
      name: "Practice Night",
      teamCount: 10,
      draftPosition: 6,
      draftType: "linear",
      scoringRules: { reception: 1 },
    });
    expect(rules.teamNames).toHaveLength(10);
    expect(rules.draftOrder).toEqual([6, 2, 3, 4, 5, 1, 7, 8, 9, 10]);
  });

  it("uses an auction-safe draft position", () => {
    const rules = buildQuickDemoRules({
      ...defaultQuickDemoSettings,
      draftPosition: 12,
      draftType: "auction",
    });

    expect(rules.draftPosition).toBe(1);
  });
});
