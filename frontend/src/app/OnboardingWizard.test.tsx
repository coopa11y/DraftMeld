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

  it("imports league settings and scoring, preserves review evidence, and supports undo", async () => {
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
              settings: {
                name: "Imported League",
                teamCount: 10,
                leagueFormat: "dynasty",
                futurePickSeasons: 3,
                rosterSlots: [{ name: "WR", count: 3, positions: ["WR"], isStarting: true }],
              },
              settingMatches: [
                {
                  key: "teamCount",
                  label: "Number of teams",
                  value: "10",
                  source: "Number of teams, 10",
                  confidence: "high",
                },
              ],
              warnings: ["Review every imported value against your league settings before saving."],
            }),
            { status: 200, headers: { "Content-Type": "application/json" } },
          ),
      ),
    );
    const first = renderWizard();
    const file = new File(["Setting,Value\nNumber of teams,10\nPassing touchdowns,6"], "rules.csv", {
      type: "text/csv",
    });
    await user.upload(screen.getByLabelText("League rules PDF or CSV"), file);
    await user.click(screen.getByRole("button", { name: "Import and apply recognized settings" }));
    expect(await screen.findByText(/Applied 2 recognized values/)).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "League name" })).toHaveValue("Imported League");
    expect(screen.getByRole("spinbutton", { name: "Number of teams" })).toHaveValue(10);
    expect(screen.getByRole("table", { name: "League settings recognized from the CSV file" })).toBeInTheDocument();
    expect((await axe(first.container)).violations).toHaveLength(0);
    await user.click(screen.getByRole("button", { name: "Save and finish later" }));
    expect(first.props.onClose).toHaveBeenCalledOnce();
    cleanup();
    renderWizard();
    expect(screen.getByRole("heading", { name: "Imported values to review" })).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "League name" })).toHaveValue("Imported League");
    await user.click(screen.getByRole("button", { name: "Continue" }));
    expect(screen.getByRole("spinbutton", { name: "Points per passing touchdown" })).toHaveValue(6);
    await user.click(screen.getByRole("button", { name: "Back" }));
    await user.click(screen.getByRole("button", { name: "Undo all imported values" }));
    expect(screen.getByRole("textbox", { name: "League name" })).toHaveValue("My League");
    expect(screen.getByRole("spinbutton", { name: "Number of teams" })).toHaveValue(12);
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
