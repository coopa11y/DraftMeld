import type { LeagueRules } from "../shared/api/types";
import { cloneLeagueRules, defaultLeagueRules } from "./leagueDefaults";

export const ONBOARDING_COMPLETE_KEY = "draftmeld.onboarding.complete.v1";
export const ONBOARDING_DRAFT_KEY = "draftmeld.onboarding.draft.v1";

export interface OnboardingDraft {
  version: 1;
  step: number;
  rules: LeagueRules;
  updatedAt: string;
}

export function onboardingComplete(): boolean {
  return localStorage.getItem(ONBOARDING_COMPLETE_KEY) === "true";
}

export function loadOnboardingDraft(): OnboardingDraft | null {
  const stored = localStorage.getItem(ONBOARDING_DRAFT_KEY);
  if (!stored) return null;
  try {
    const parsed = JSON.parse(stored) as Partial<OnboardingDraft>;
    if (parsed.version !== 1 || typeof parsed.step !== "number" || !validRules(parsed.rules)) return null;
    return {
      version: 1,
      step: Math.max(0, Math.min(2, parsed.step)),
      rules: cloneLeagueRules(parsed.rules),
      updatedAt: typeof parsed.updatedAt === "string" ? parsed.updatedAt : new Date().toISOString(),
    };
  } catch {
    return null;
  }
}

export function saveOnboardingDraft(step: number, rules: LeagueRules): OnboardingDraft {
  const draft: OnboardingDraft = {
    version: 1,
    step,
    rules: cloneLeagueRules(rules),
    updatedAt: new Date().toISOString(),
  };
  localStorage.setItem(ONBOARDING_DRAFT_KEY, JSON.stringify(draft));
  return draft;
}

export function newOnboardingDraft(): OnboardingDraft {
  return { version: 1, step: 0, rules: defaultLeagueRules(), updatedAt: new Date().toISOString() };
}

export function completeOnboarding(): void {
  localStorage.setItem(ONBOARDING_COMPLETE_KEY, "true");
  localStorage.removeItem(ONBOARDING_DRAFT_KEY);
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
