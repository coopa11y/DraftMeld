import type { LeagueRules } from "../shared/api/types";

export interface ScoringField {
  key: string;
  label: string;
  step: number;
  aliases?: string[];
  projection?: boolean;
  projectionLabel?: string;
  requiresPosition?: "TE";
}

export interface ScoringGroup {
  description: string;
  fields: ScoringField[];
  name: string;
  open?: boolean;
  requiresPosition?: "K" | "DST";
}

export const scoringGroups: ScoringGroup[] = [
  {
    name: "Essential scoring",
    description:
      "The common values most leagues need. The reception preset above fills in the reception values for you.",
    open: true,
    fields: [
      {
        key: "reception",
        label: "Reception",
        projectionLabel: "Receptions",
        step: 0.5,
        aliases: ["receptions", "rec"],
      },
      {
        key: "tightEndReceptionBonus",
        label: "Tight end reception bonus",
        step: 0.25,
        projection: false,
        requiresPosition: "TE",
      },
      { key: "passingYard", label: "Passing yard", step: 0.01, aliases: ["passingyards", "passyds"] },
      { key: "passingTouchdown", label: "Passing touchdown", step: 1, aliases: ["passingtouchdowns", "passtd"] },
      { key: "interception", label: "Interception thrown", step: 0.5, aliases: ["interceptions", "int"] },
      { key: "rushingYard", label: "Rushing yard", step: 0.01, aliases: ["rushingyards", "rushyds"] },
      { key: "rushingTouchdown", label: "Rushing touchdown", step: 1, aliases: ["rushingtouchdowns", "rushtd"] },
      { key: "receivingYard", label: "Receiving yard", step: 0.01, aliases: ["receivingyards", "recyds"] },
      {
        key: "receivingTouchdown",
        label: "Receiving touchdown",
        step: 1,
        aliases: ["receivingtouchdowns", "rectd"],
      },
    ],
  },
  {
    name: "Offensive bonuses and turnovers",
    description: "Optional two-point plays, yardage milestones, and fumble penalties. Leave unused rules at zero.",
    fields: [
      {
        key: "passingTwoPointConversion",
        label: "Passing two-point conversion",
        step: 0.5,
        aliases: ["passing2pt", "pass2pt"],
      },
      { key: "passing300YardGame", label: "300-yard passing game", step: 0.5, aliases: ["pass300games"] },
      { key: "passing400YardGame", label: "400-yard passing game", step: 0.5, aliases: ["pass400games"] },
      { key: "rushingTwoPointConversion", label: "Rushing two-point conversion", step: 0.5, aliases: ["rush2pt"] },
      { key: "rushing100YardGame", label: "100-yard rushing game", step: 0.5, aliases: ["rush100games"] },
      { key: "rushing200YardGame", label: "200-yard rushing game", step: 0.5, aliases: ["rush200games"] },
      {
        key: "receivingTwoPointConversion",
        label: "Receiving two-point conversion",
        step: 0.5,
        aliases: ["receiving2pt", "rec2pt"],
      },
      { key: "receiving100YardGame", label: "100-yard receiving game", step: 0.5, aliases: ["rec100games"] },
      { key: "receiving200YardGame", label: "200-yard receiving game", step: 0.5, aliases: ["rec200games"] },
      { key: "fumble", label: "Fumble", step: 0.5, aliases: ["fumbles"] },
      { key: "fumbleLost", label: "Fumble lost", step: 0.5, aliases: ["fumbleslost", "fumlost"] },
    ],
  },
  {
    name: "Kicking",
    requiresPosition: "K",
    description: "Use the general field-goal value or set it to zero and score by distance.",
    fields: [
      { key: "fieldGoalMade", label: "Field goal made (any distance)", step: 0.5 },
      { key: "fieldGoal0To39", label: "Field goal made, 0–39 yards", step: 0.5, aliases: ["fg0to39", "fg039"] },
      { key: "fieldGoal40To49", label: "Field goal made, 40–49 yards", step: 0.5, aliases: ["fg40to49", "fg4049"] },
      { key: "fieldGoal50To59", label: "Field goal made, 50–59 yards", step: 0.5, aliases: ["fg50to59", "fg5059"] },
      { key: "fieldGoal60Plus", label: "Field goal made, 60+ yards", step: 0.5, aliases: ["fg60plus", "fg60"] },
      { key: "fieldGoal50Plus", label: "Field goal made, 50+ yards", step: 0.5, aliases: ["fg50plus", "fg50"] },
      { key: "fieldGoalMissed", label: "Field goal missed", step: 0.5 },
      { key: "extraPointMade", label: "Extra point made", step: 0.5 },
      { key: "extraPointMissed", label: "Extra point missed", step: 0.5 },
    ],
  },
  {
    name: "Team defense and special teams",
    requiresPosition: "DST",
    description: "Play scoring and optional points-allowed tiers. Leave any unused category at zero.",
    fields: [
      { key: "defenseSack", label: "Sack", step: 0.5 },
      { key: "defenseInterception", label: "Interception", step: 0.5 },
      { key: "defenseFumbleRecovery", label: "Fumble recovery", step: 0.5 },
      { key: "defenseTouchdown", label: "Defense or special-teams touchdown", step: 1 },
      { key: "defenseSafety", label: "Safety", step: 0.5 },
      { key: "defenseOnePointSafety", label: "One-point safety", step: 0.5 },
      { key: "defenseBlockedKick", label: "Blocked kick", step: 0.5 },
      { key: "defenseTwoPointReturn", label: "Defensive two-point return", step: 0.5 },
      { key: "defensePointsAllowed0", label: "Game allowing 0 points", step: 0.5, aliases: ["dstpa0"] },
      { key: "defensePointsAllowed1To6", label: "Game allowing 1–6 points", step: 0.5, aliases: ["dstpa1to6"] },
      { key: "defensePointsAllowed7To13", label: "Game allowing 7–13 points", step: 0.5, aliases: ["dstpa7to13"] },
      { key: "defensePointsAllowed14To20", label: "Game allowing 14–20 points", step: 0.5, aliases: ["dstpa14to20"] },
      { key: "defensePointsAllowed21To27", label: "Game allowing 21–27 points", step: 0.5, aliases: ["dstpa21to27"] },
      { key: "defensePointsAllowed28To34", label: "Game allowing 28–34 points", step: 0.5, aliases: ["dstpa28to34"] },
      { key: "defensePointsAllowed35Plus", label: "Game allowing 35+ points", step: 0.5, aliases: ["dstpa35plus"] },
      { key: "defensePointsAllowed14To17", label: "Game allowing 14–17 points", step: 0.5, aliases: ["dstpa14to17"] },
      { key: "defensePointsAllowed35To45", label: "Game allowing 35–45 points", step: 0.5, aliases: ["dstpa35to45"] },
      { key: "defensePointsAllowed46Plus", label: "Game allowing 46+ points", step: 0.5, aliases: ["dstpa46plus"] },
      { key: "defenseYardsAllowedUnder100", label: "Game allowing fewer than 100 yards", step: 0.5 },
      { key: "defenseYardsAllowed100To199", label: "Game allowing 100–199 yards", step: 0.5 },
      { key: "defenseYardsAllowed200To299", label: "Game allowing 200–299 yards", step: 0.5 },
      { key: "defenseYardsAllowed350To399", label: "Game allowing 350–399 yards", step: 0.5 },
      { key: "defenseYardsAllowed400To449", label: "Game allowing 400–449 yards", step: 0.5 },
      { key: "defenseYardsAllowed450To499", label: "Game allowing 450–499 yards", step: 0.5 },
      { key: "defenseYardsAllowed500To549", label: "Game allowing 500–549 yards", step: 0.5 },
      { key: "defenseYardsAllowed550Plus", label: "Game allowing 550+ yards", step: 0.5 },
    ],
  },
];

export const projectionScoringFields = scoringGroups.flatMap((group) =>
  group.fields.filter((field) => field.projection !== false),
);

const commonScoringRules: LeagueRules["scoringRules"] = {
  reception: 1,
  tightEndReceptionBonus: 0,
  passingYard: 0.04,
  passingTouchdown: 4,
  interception: -2,
  passingTwoPointConversion: 2,
  passing300YardGame: 0,
  passing400YardGame: 0,
  rushingYard: 0.1,
  rushingTouchdown: 6,
  rushingTwoPointConversion: 2,
  rushing100YardGame: 0,
  rushing200YardGame: 0,
  receivingYard: 0.1,
  receivingTouchdown: 6,
  receivingTwoPointConversion: 2,
  receiving100YardGame: 0,
  receiving200YardGame: 0,
  fumble: 0,
  fumbleLost: -2,
  fieldGoalMade: 3,
  fieldGoal0To39: 0,
  fieldGoal40To49: 0,
  fieldGoal50To59: 0,
  fieldGoal60Plus: 0,
  fieldGoal50Plus: 0,
  fieldGoalMissed: 0,
  extraPointMade: 1,
  extraPointMissed: 0,
  defenseSack: 1,
  defenseInterception: 2,
  defenseFumbleRecovery: 2,
  defenseTouchdown: 6,
  defenseSafety: 2,
  defenseOnePointSafety: 0,
  defenseBlockedKick: 2,
  defenseTwoPointReturn: 2,
  defensePointsAllowed0: 10,
  defensePointsAllowed1To6: 7,
  defensePointsAllowed7To13: 4,
  defensePointsAllowed14To20: 1,
  defensePointsAllowed21To27: 0,
  defensePointsAllowed28To34: -1,
  defensePointsAllowed35Plus: -4,
  defensePointsAllowed14To17: 0,
  defensePointsAllowed35To45: 0,
  defensePointsAllowed46Plus: 0,
  defenseYardsAllowedUnder100: 0,
  defenseYardsAllowed100To199: 0,
  defenseYardsAllowed200To299: 0,
  defenseYardsAllowed350To399: 0,
  defenseYardsAllowed400To449: 0,
  defenseYardsAllowed450To499: 0,
  defenseYardsAllowed500To549: 0,
  defenseYardsAllowed550Plus: 0,
};

export function defaultScoringRules(): LeagueRules["scoringRules"] {
  return { ...commonScoringRules };
}

export function receptionPreset(scoring: LeagueRules["scoringRules"]): string {
  const reception = scoring.reception ?? 0;
  const tightEndBonus = scoring.tightEndReceptionBonus ?? 0;
  if (reception === 1 && tightEndBonus === 0.5) return "te-premium";
  if (tightEndBonus === 0 && (reception === 0 || reception === 0.5 || reception === 1)) return String(reception);
  return "custom";
}

export function applyReceptionPreset(
  scoring: LeagueRules["scoringRules"],
  preset: string,
): LeagueRules["scoringRules"] {
  if (preset === "custom") return scoring;
  if (preset === "te-premium") return { ...scoring, reception: 1, tightEndReceptionBonus: 0.5 };
  return { ...scoring, reception: Number(preset), tightEndReceptionBonus: 0 };
}
