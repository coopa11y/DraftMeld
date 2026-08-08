import { describe, expect, it } from "vitest";
import type { LeagueRuleImport } from "../shared/api/types";
import { applyLeagueRuleImport } from "./applyLeagueRuleImport";
import { defaultLeagueRules } from "./leagueDefaults";

describe("applyLeagueRuleImport", () => {
  it("applies compatible league-wide settings while preserving user-specific choices", () => {
    const current = defaultLeagueRules();
    current.draftPosition = 8;
    current.userTeamNumber = 8;
    const imported: LeagueRuleImport = {
      fileType: "csv",
      rules: { passingTouchdown: 6 },
      matches: [],
      settings: {
        name: "Saturday League",
        teamCount: 10,
        leagueFormat: "dynasty",
        draftType: "auction",
        futurePickSeasons: 3,
        faabBudget: 125,
        faabTrades: true,
        rosterSlots: [
          { name: "WR", count: 3, positions: ["WR"], isStarting: true },
          { name: "K", count: 0, positions: ["K"], isStarting: true },
        ],
      },
      settingMatches: [],
      warnings: [],
    };

    const result = applyLeagueRuleImport(current, imported);
    expect(result).toMatchObject({
      name: "Saturday League",
      teamCount: 10,
      draftPosition: 8,
      userTeamNumber: 8,
      leagueFormat: "dynasty",
      draftType: "auction",
      futurePickSeasons: 3,
      faabBudget: 125,
      faabTrades: true,
    });
    expect(result.scoringRules.passingTouchdown).toBe(6);
    expect(result.rosterSlots.find((slot) => slot.name === "WR")?.count).toBe(3);
    expect(result.rosterSlots.some((slot) => slot.name === "K")).toBe(false);
  });

  it("keeps redraft-only settings internally consistent", () => {
    const current = defaultLeagueRules();
    current.leagueFormat = "dynasty";
    current.futurePickSeasons = 3;
    current.faabTrades = true;
    const imported: LeagueRuleImport = {
      fileType: "pdf",
      rules: {},
      matches: [],
      settings: { leagueFormat: "redraft" },
      settingMatches: [],
      warnings: [],
    };
    const result = applyLeagueRuleImport(current, imported);
    expect(result.futurePickSeasons).toBe(0);
    expect(result.faabTrades).toBe(false);
  });
});
