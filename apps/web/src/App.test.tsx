import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { axe } from "jest-axe";
import { App } from "./App";

describe("public website", () => {
  it("presents the product, open-source path, supported-build policy, and contribution path", () => {
    render(<App />);

    expect(screen.getByRole("heading", { level: 1, name: "Draft with every ranking on your side." })).toBeVisible();
    expect(screen.getByRole("heading", { name: "Your league. Your board. Your call." })).toBeVisible();
    expect(screen.getByRole("heading", { name: "Free to run. Built in the open." })).toBeVisible();
    expect(screen.getByText("docker compose up -d")).toBeVisible();
    expect(screen.getByText(/paid convenience, not feature gating/i)).toBeVisible();
    expect(screen.getByText("Coming with the first public release")).toBeVisible();
    expect(screen.getByRole("link", { name: /Contribute on GitHub/i })).toHaveAttribute(
      "href",
      "https://github.com/coopa11y/DraftMeld/blob/dev/CONTRIBUTING.md",
    );
  });

  it("opens and closes the mobile navigation with an accurate expanded state", async () => {
    const user = userEvent.setup();
    render(<App />);
    const menu = screen.getByRole("button", { name: "Menu" });

    expect(menu).toHaveAttribute("aria-expanded", "false");
    await user.click(menu);
    expect(menu).toHaveAttribute("aria-expanded", "true");
    await user.click(
      within(screen.getByRole("navigation", { name: "Primary" })).getByRole("link", { name: "Product" }),
    );
    expect(menu).toHaveAttribute("aria-expanded", "false");
  });

  it("has no automated accessibility violations", async () => {
    const { container } = render(<App />);
    const violations = (await axe(container)).violations;
    expect(violations.map(({ id, nodes }) => ({ id, targets: nodes.map((node) => node.target) }))).toEqual([]);
  });
});
