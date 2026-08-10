import type { LeagueRules } from "../shared/api/types";
import { defaultLeagueRules } from "./leagueDefaults";
import { applyReceptionPreset } from "./scoring";

export interface QuickDemoSettings {
  draftPosition: number;
  draftType: LeagueRules["draftType"];
  name: string;
  receptionPreset: "0" | "0.5" | "1";
  teamCount: number;
}

export const defaultQuickDemoSettings: QuickDemoSettings = {
  draftPosition: 1,
  draftType: "snake",
  name: "My Demo League",
  receptionPreset: "0.5",
  teamCount: 12,
};

export function buildQuickDemoRules(settings: QuickDemoSettings): LeagueRules {
  const rules = defaultLeagueRules();
  const draftPosition = settings.draftType === "auction" ? 1 : Math.min(settings.draftPosition, settings.teamCount);
  const draftOrder = Array.from({ length: settings.teamCount }, (_, index) => index + 1);
  if (draftPosition > 1)
    [draftOrder[0], draftOrder[draftPosition - 1]] = [draftOrder[draftPosition - 1], draftOrder[0]];
  return {
    ...rules,
    name: settings.name.trim(),
    teamCount: settings.teamCount,
    draftPosition,
    userTeamNumber: 1,
    teamNames: ["My Team", ...Array.from({ length: settings.teamCount - 1 }, () => "")],
    draftOrder,
    draftType: settings.draftType,
    scoringRules: applyReceptionPreset(rules.scoringRules, settings.receptionPreset),
  };
}
