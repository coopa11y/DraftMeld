import type { LeagueRules } from "../api/types";

/** Returns a copy whose nested collections can be edited without mutating the source. */
export function cloneLeagueRules(rules: LeagueRules): LeagueRules {
  return {
    ...rules,
    teamNames: [...rules.teamNames],
    draftOrder: [...rules.draftOrder],
    rosterSlots: rules.rosterSlots.map((slot) => ({ ...slot, positions: [...slot.positions] })),
    scoringRules: { ...rules.scoringRules },
    sourcePreferences: Object.fromEntries(
      Object.entries(rules.sourcePreferences).map(([sourceId, preference]) => [sourceId, { ...preference }]),
    ),
    playerPreferences: { ...rules.playerPreferences },
  };
}
