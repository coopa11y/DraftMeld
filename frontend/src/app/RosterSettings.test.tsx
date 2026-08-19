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

    expect(screen.queryByRole("group", { name: "Custom roster slot 1" })).not.toBeInTheDocument();
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
    const quarterbackCountBeforeCustomEdit = quarterbackStarters.getAttribute("value");

    await user.click(screen.getByText("Advanced roster slots", { selector: "summary" }));
    expect(screen.getByText("No custom roster slots have been created.")).toBeInTheDocument();
    expect(screen.queryByRole("group", { name: "Custom roster slot 1" })).not.toBeInTheDocument();
    expect(screen.queryByDisplayValue("QB")).not.toBeInTheDocument();

    const addCustomSlot = screen.getByRole("button", { name: "Add custom slot" });
    await user.click(addCustomSlot);
    const firstSlot = screen.getByRole("group", { name: "Custom roster slot 1" });
    const slotName = within(firstSlot).getByRole("textbox", { name: "Slot name" });
    expect(addCustomSlot).toAppearBefore(firstSlot);
    expect(slotName).toHaveFocus();
    expect(screen.getByRole("status")).toHaveTextContent("Custom roster slot 1 added.");
    await user.clear(slotName);
    await user.type(slotName, "Taxi");
    await user.click(within(firstSlot).getByRole("checkbox", { name: "Starting lineup slot" }));
    await user.click(within(firstSlot).getByRole("checkbox", { name: "RB" }));
    expect(screen.getByRole("spinbutton", { name: "QB starters" })).toHaveAttribute(
      "value",
      quarterbackCountBeforeCustomEdit,
    );
    await user.click(within(firstSlot).getByRole("button", { name: "Remove slot" }));
    expect(addCustomSlot).toHaveFocus();
    expect(screen.getByRole("status")).toHaveTextContent("Custom roster slot 1 removed.");
    expect(screen.getByText("No custom roster slots have been created.")).toBeInTheDocument();

    expect((await axe(container)).violations).toHaveLength(0);
  });
});
