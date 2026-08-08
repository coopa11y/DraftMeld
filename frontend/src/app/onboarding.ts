import type { LeagueRuleImport, LeagueRules } from "../shared/api/types";
import { cloneLeagueRules, defaultLeagueRules } from "./leagueDefaults";

export const ONBOARDING_COMPLETE_KEY = "draftmeld.onboarding.complete.v1";
export const ONBOARDING_DRAFT_KEY = "draftmeld.onboarding.draft.v2";
const LEGACY_ONBOARDING_DRAFT_KEY = "draftmeld.onboarding.draft.v1";

export interface OnboardingImportReview {
  result: LeagueRuleImport;
  beforeRules: LeagueRules;
}

export interface OnboardingDraft {
  version: 2;
  step: number;
  rules: LeagueRules;
  importReview?: OnboardingImportReview;
  updatedAt: string;
}

export function onboardingComplete(): boolean {
  return localStorage.getItem(ONBOARDING_COMPLETE_KEY) === "true";
}

export function loadOnboardingDraft(): OnboardingDraft | null {
  const stored = localStorage.getItem(ONBOARDING_DRAFT_KEY) ?? localStorage.getItem(LEGACY_ONBOARDING_DRAFT_KEY);
  if (!stored) return null;
  try {
    const parsed = JSON.parse(stored) as Partial<OnboardingDraft> & { version?: number };
    if (![1, 2].includes(parsed.version ?? 0) || typeof parsed.step !== "number" || !validRules(parsed.rules))
      return null;
    return {
      version: 2,
      step: Math.max(0, Math.min(2, parsed.step)),
      rules: cloneLeagueRules(parsed.rules),
      importReview: validImportReview(parsed.importReview)
        ? { result: parsed.importReview.result, beforeRules: cloneLeagueRules(parsed.importReview.beforeRules) }
        : undefined,
      updatedAt: typeof parsed.updatedAt === "string" ? parsed.updatedAt : new Date().toISOString(),
    };
  } catch {
    return null;
  }
}

export function saveOnboardingDraft(
  step: number,
  rules: LeagueRules,
  importReview?: OnboardingImportReview,
): OnboardingDraft {
  const draft: OnboardingDraft = {
    version: 2,
    step,
    rules: cloneLeagueRules(rules),
    importReview: importReview
      ? { result: importReview.result, beforeRules: cloneLeagueRules(importReview.beforeRules) }
      : undefined,
    updatedAt: new Date().toISOString(),
  };
  localStorage.setItem(ONBOARDING_DRAFT_KEY, JSON.stringify(draft));
  return draft;
}

export function newOnboardingDraft(): OnboardingDraft {
  return { version: 2, step: 0, rules: defaultLeagueRules(), updatedAt: new Date().toISOString() };
}

export function completeOnboarding(): void {
  localStorage.setItem(ONBOARDING_COMPLETE_KEY, "true");
  localStorage.removeItem(ONBOARDING_DRAFT_KEY);
  localStorage.removeItem(LEGACY_ONBOARDING_DRAFT_KEY);
}

function validImportReview(value: unknown): value is OnboardingImportReview {
  if (typeof value !== "object" || value === null) return false;
  const candidate = value as Partial<OnboardingImportReview>;
  return (
    validRules(candidate.beforeRules) &&
    typeof candidate.result === "object" &&
    candidate.result !== null &&
    Array.isArray(candidate.result.matches) &&
    Array.isArray(candidate.result.settingMatches)
  );
}

function validRules(value: unknown): value is LeagueRules {
  if (typeof value !== "object" || value === null) return false;
  const candidate = value as Partial<LeagueRules>;
  return (
    typeof candidate.name === "string" &&
    typeof candidate.teamCount === "number" &&
    typeof candidate.scoringRules === "object" &&
    Array.isArray(candidate.rosterSlots)
  );
}
