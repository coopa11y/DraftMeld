import { axe } from "jest-axe";
import type { ComponentProps } from "react";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { League } from "../shared/api/types";
import { defaultLeagueRules } from "./leagueDefaults";
import { ONBOARDING_COMPLETE_KEY, loadOnboardingDraft } from "./onboarding";
import { OnboardingWizard } from "./OnboardingWizard";

afterEach(() => {
  cleanup();
  localStorage.clear();
  vi.unstubAllGlobals();
});

function renderWizard(overrides: Partial<ComponentProps<typeof OnboardingWizard>> = {}) {
  const rules = defaultLeagueRules();
  const league: League = { id: "night-league", ...rules, name: "Night League" };
  const props = {
    onClose: vi.fn(),
    onCreate: vi.fn(async () => league),
    onOpenDraft: vi.fn(),
    onOpenRankings: vi.fn(),
    ...overrides,
  };
  return { props, ...render(<OnboardingWizard {...props} />) };
}

describe("resumable onboarding wizard", () => {
  it("saves progress, resumes the same step, and offers the full manual editor", async () => {
    const user = userEvent.setup();
    const first = renderWizard();
    const name = screen.getByRole("textbox", { name: "League name" });
    await user.clear(name);
    await user.type(name, "Night League");
    await user.click(screen.getByRole("button", { name: "Continue" }));
    expect(await screen.findByRole("heading", { name: "Scoring" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Save and finish later" }));
    expect(first.props.onClose).toHaveBeenCalledOnce();
    expect(loadOnboardingDraft()?.step).toBe(1);

    cleanup();
    renderWizard();
    expect(screen.getByRole("heading", { name: "Scoring" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Configure manually" }));
    expect(screen.getByRole("heading", { name: "Configure every league setting" })).toBeInTheDocument();
    const manualName = screen.getByRole("textbox", { name: "League name" });
    expect(manualName).toHaveValue("Night League");
    await user.clear(manualName);
    await user.type(manualName, "Saved Manual League");
    await user.click(screen.getByRole("button", { name: "Save and finish later" }));
    expect(loadOnboardingDraft()?.rules.name).toBe("Saved Manual League");
  });

  it("imports recognized CSV rules, preserves review evidence, and supports undo", async () => {
    const user = userEvent.setup();
    vi.stubGlobal(
      "fetch",
      vi.fn(
        async () =>
          new Response(
            JSON.stringify({
              fileType: "csv",
              rules: { passingTouchdown: 6 },
              matches: [
                {
                  key: "passingTouchdown",
                  label: "Passing touchdown",
                  value: 6,
                  source: "Passing touchdowns, 6",
                  confidence: "high",
                },
              ],
              warnings: ["Review every imported value against your league settings before saving."],
            }),
            { status: 200, headers: { "Content-Type": "application/json" } },
          ),
      ),
    );
    renderWizard();
    await user.click(screen.getByRole("button", { name: "Continue" }));
    const file = new File(["Statistic,Points\nPassing touchdowns,6"], "rules.csv", { type: "text/csv" });
    await user.upload(screen.getByLabelText("League rules PDF or CSV"), file);
    await user.click(screen.getByRole("button", { name: "Import and apply recognized rules" }));
    expect(await screen.findByText("Applied 1 recognized scoring values.")).toBeInTheDocument();
    expect(screen.getByRole("spinbutton", { name: "Points per passing touchdown" })).toHaveValue(6);
    expect(screen.getByRole("table", { name: "Rules recognized from the CSV file" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Undo imported values" }));
    expect(screen.getByRole("spinbutton", { name: "Points per passing touchdown" })).toHaveValue(4);
  });

  it("creates a league only after review and completes the guide", async () => {
    const user = userEvent.setup();
    const { props, container } = renderWizard();
    expect((await axe(container)).violations).toHaveLength(0);
    await user.click(screen.getByRole("button", { name: "Continue" }));
    await user.click(screen.getByRole("button", { name: "Continue" }));
    expect(screen.getByRole("heading", { name: "Review and create" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Create league" }));
    await waitFor(() => expect(props.onCreate).toHaveBeenCalledOnce());
    expect(await screen.findByRole("heading", { name: "Night League is ready" })).toBeInTheDocument();
    expect(localStorage.getItem(ONBOARDING_COMPLETE_KEY)).toBe("true");
  });
});
