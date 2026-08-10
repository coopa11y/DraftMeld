import { useState } from "react";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { axe } from "jest-axe";
import type { LeagueRules } from "../shared/api/types";
import { defaultLeagueRules } from "./leagueDefaults";
import { RosterSettings } from "./RosterSettings";

function Harness() {
  const [rules, setRules] = useState<LeagueRules>(() => defaultLeagueRules());
  return <RosterSettings rules={rules} setRules={setRules} disabled={false} />;
}

describe("RosterSettings", () => {
  it("supports simple roster changes and keeps custom slots behind the advanced disclosure", async () => {
    const user = userEvent.setup();
    const { container } = render(<Harness />);

    expect(screen.queryByRole("group", { name: "Roster slot 1" })).not.toBeInTheDocument();
    const quarterbackStarters = screen.getByRole("spinbutton", { name: "QB starters" });
    await user.clear(quarterbackStarters);
    await user.type(quarterbackStarters, "2");
    const flexSpots = screen.getByRole("spinbutton", { name: "Flex spots" });
    await user.clear(flexSpots);
    await user.type(flexSpots, "2");
    await user.click(
      within(screen.getByRole("group", { name: "Flex eligible positions" })).getByRole("checkbox", { name: "QB" }),
    );
    const benchSpots = screen.getByRole("spinbutton", { name: "Bench spots" });
    await user.clear(benchSpots);
    await user.type(benchSpots, "4");

    await user.click(screen.getByText("Advanced roster slots", { selector: "summary" }));
    const firstSlot = screen.getByRole("group", { name: "Roster slot 1" });
    const slotName = within(firstSlot).getByRole("textbox", { name: "Slot name" });
    await user.clear(slotName);
    await user.type(slotName, "Passer");
    await user.click(within(firstSlot).getByRole("checkbox", { name: "Starting lineup slot" }));
    await user.click(within(firstSlot).getByRole("checkbox", { name: "RB" }));
    await user.click(screen.getByRole("button", { name: "Add custom slot" }));
    await user.click(screen.getAllByRole("button", { name: "Remove slot" }).at(-1)!);

    expect((await axe(container)).violations).toHaveLength(0);
  });
});
