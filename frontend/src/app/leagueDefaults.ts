import type { LeagueRules, RosterSlot } from "../shared/api/types";
import { cloneLeagueRules } from "../shared/domain/leagueRules";
import { defaultScoringRules } from "./scoring";

export const playerPositions = ["QB", "RB", "WR", "TE", "K", "DST"] as const;

const defaultRoster: RosterSlot[] = [
  { name: "QB", count: 1, positions: ["QB"], isStarting: true },
  { name: "RB", count: 2, positions: ["RB"], isStarting: true },
  { name: "WR", count: 2, positions: ["WR"], isStarting: true },
  { name: "TE", count: 1, positions: ["TE"], isStarting: true },
  { name: "FLEX", count: 1, positions: ["RB", "WR", "TE"], isStarting: true },
  { name: "K", count: 1, positions: ["K"], isStarting: true },
  { name: "DST", count: 1, positions: ["DST"], isStarting: true },
  { name: "Bench", count: 6, positions: [...playerPositions], isStarting: false },
];

export function defaultLeagueRules(): LeagueRules {
  const season = new Date().getFullYear();
  return {
    name: "My League",
    teamCount: 12,
    draftPosition: 0,
    userTeamNumber: 1,
    teamNames: ["My Team", ...Array.from({ length: 11 }, () => "")],
    draftOrder: Array.from({ length: 12 }, (_, index) => index + 1),
    draftType: "snake",
    leagueFormat: "redraft",
    season,
    initialSeason: season,
    futurePickSeasons: 0,
    rookieDraftRounds: 4,
    auctionBudgetTrades: false,
    faabBudget: 100,
    faabTrades: false,
    rosterSlots: defaultRoster.map((slot) => ({ ...slot, positions: [...slot.positions] })),
    scoringRules: defaultScoringRules(),
    sourcePreferences: {},
    consensusMethod: "weighted-median",
    playerPreferences: {},
    auctionBudget: 200,
    auctionMinimumBid: 1,
    keeperBudgetSpent: 0,
    myKeeperSpend: 0,
    keeperValueRemoved: 0,
  };
}

export { cloneLeagueRules };
