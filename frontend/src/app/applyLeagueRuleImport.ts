import type { LeagueRuleImport, LeagueRules, RosterSlot } from "../shared/api/types";
import { cloneLeagueRules } from "./leagueDefaults";

const rosterOrder = new Map(
  ["QB", "RB", "WR", "TE", "FLEX", "SUPERFLEX", "K", "DST", "BENCH", "IR"].map((name, index) => [name, index]),
);

export function applyLeagueRuleImport(current: LeagueRules, imported: LeagueRuleImport): LeagueRules {
  const next = cloneLeagueRules(current);
  const settings = imported.settings;

  if (settings.name) next.name = settings.name;
  if (settings.teamCount !== undefined) resizeLeague(next, settings.teamCount);
  if (settings.draftType) next.draftType = settings.draftType;
  if (settings.leagueFormat) next.leagueFormat = settings.leagueFormat;
  if (settings.futurePickSeasons !== undefined) next.futurePickSeasons = settings.futurePickSeasons;
  if (settings.rookieDraftRounds !== undefined) next.rookieDraftRounds = settings.rookieDraftRounds;
  if (settings.auctionBudget !== undefined) next.auctionBudget = settings.auctionBudget;
  if (settings.auctionMinimumBid !== undefined) next.auctionMinimumBid = settings.auctionMinimumBid;
  if (settings.auctionBudgetTrades !== undefined) next.auctionBudgetTrades = settings.auctionBudgetTrades;
  if (settings.faabBudget !== undefined) next.faabBudget = settings.faabBudget;
  if (settings.faabTrades !== undefined) next.faabTrades = settings.faabTrades;
  if (settings.rosterSlots?.length) next.rosterSlots = mergeRosterSlots(next.rosterSlots, settings.rosterSlots);

  next.scoringRules = { ...next.scoringRules, ...imported.rules };
  if (next.leagueFormat === "redraft") {
    next.futurePickSeasons = 0;
    next.faabTrades = false;
  } else {
    next.futurePickSeasons = Math.max(1, next.futurePickSeasons);
  }
  if (next.faabBudget <= 0) next.faabTrades = false;
  if (next.draftType !== "auction") next.auctionBudgetTrades = false;
  next.auctionMinimumBid = Math.min(next.auctionMinimumBid, next.auctionBudget);
  return next;
}

function resizeLeague(rules: LeagueRules, teamCount: number) {
  rules.teamCount = teamCount;
  rules.userTeamNumber = Math.min(rules.userTeamNumber, teamCount);
  rules.draftPosition = Math.min(rules.draftPosition, teamCount);
  rules.teamNames = Array.from({ length: teamCount }, (_, index) => rules.teamNames[index] ?? "");
  rules.draftOrder = Array.from({ length: teamCount }, (_, index) => index + 1);
}

function mergeRosterSlots(current: RosterSlot[], imported: NonNullable<LeagueRuleImport["settings"]["rosterSlots"]>) {
  const importedNames = new Set(imported.map((slot) => slot.name.toLowerCase()));
  const retained = current.filter((slot) => !importedNames.has(slot.name.toLowerCase()));
  const enabled = imported
    .filter((slot) => slot.count > 0)
    .map((slot) => ({ ...slot, positions: [...slot.positions] }));
  return [...enabled, ...retained].sort(
    (left, right) =>
      (rosterOrder.get(left.name.toUpperCase()) ?? 100) - (rosterOrder.get(right.name.toUpperCase()) ?? 100),
  );
}
