import { beforeEach, describe, expect, it } from "vitest";
import {
  completeOnboarding,
  loadOnboardingDraft,
  newOnboardingDraft,
  onboardingComplete,
  saveOnboardingDraft,
} from "./onboarding";

describe("onboarding persistence", () => {
  beforeEach(() => localStorage.clear());

  it("saves and restores a versioned partial setup", () => {
    const draft = newOnboardingDraft();
    draft.rules.name = "Saved League";
    saveOnboardingDraft(2, draft.rules);
    const restored = loadOnboardingDraft();
    expect(restored?.step).toBe(2);
    expect(restored?.rules.name).toBe("Saved League");
  });

  it("clears the draft when onboarding is complete", () => {
    const draft = newOnboardingDraft();
    saveOnboardingDraft(1, draft.rules);
    completeOnboarding();
    expect(onboardingComplete()).toBe(true);
    expect(loadOnboardingDraft()).toBeNull();
  });

  it("ignores unsupported saved data", () => {
    localStorage.setItem("draftmeld.onboarding.draft.v1", JSON.stringify({ version: 99, step: 1 }));
    expect(loadOnboardingDraft()).toBeNull();
  });
});
