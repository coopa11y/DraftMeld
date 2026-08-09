import { describe, expect, it } from "vitest";
import { applyReceptionPreset, defaultScoringRules, projectionScoringFields, receptionPreset } from "./scoring";

describe("scoring configuration", () => {
  it("supports simple, premium, and independently customized categories", () => {
    const simple = applyReceptionPreset(defaultScoringRules(), "0");
    expect(simple.reception).toBe(0);
    expect(simple.tightEndReceptionBonus).toBe(0);
    expect(receptionPreset(simple)).toBe("0");

    const premium = applyReceptionPreset({ ...simple, passing300YardGame: 3 }, "te-premium");
    expect(premium.reception).toBe(1);
    expect(premium.tightEndReceptionBonus).toBe(0.5);
    expect(premium.passing300YardGame).toBe(3);
    expect(receptionPreset(premium)).toBe("te-premium");
  });

  it("keeps derived TE premium out of projection CSV mappings", () => {
    const keys = projectionScoringFields.map((field) => field.key);
    expect(keys).toContain("receivingTwoPointConversion");
    expect(keys).toContain("defensePointsAllowed35Plus");
    expect(keys).not.toContain("tightEndReceptionBonus");
  });
});
