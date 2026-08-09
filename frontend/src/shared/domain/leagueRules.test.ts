import { describe, expect, it } from "vitest";
import { defaultLeagueRules } from "../../app/leagueDefaults";
import { cloneLeagueRules } from "./leagueRules";

describe("cloneLeagueRules", () => {
  it("owns every mutable nested collection", () => {
    const original = defaultLeagueRules();
    original.sourcePreferences.example = { enabled: true, weight: 1 };
    original.playerPreferences.player = "target";
    const cloned = cloneLeagueRules(original);

    cloned.teamNames[0] = "Changed";
    cloned.draftOrder[0] = 2;
    cloned.rosterSlots[0].positions[0] = "WR";
    cloned.scoringRules.reception = 0;
    cloned.sourcePreferences.example.weight = 2;
    cloned.playerPreferences.player = "avoid";

    expect(original.teamNames[0]).toBe("My Team");
    expect(original.draftOrder[0]).toBe(1);
    expect(original.rosterSlots[0].positions[0]).toBe("QB");
    expect(original.scoringRules.reception).toBe(1);
    expect(original.sourcePreferences.example.weight).toBe(1);
    expect(original.playerPreferences.player).toBe("target");
  });
});
