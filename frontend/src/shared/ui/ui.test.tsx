import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { Button } from "./Button";
import { Dialog } from "./Dialog";
import { FormField } from "./FormField";
import { Panel } from "./Panel";
import { StatusMessage } from "./StatusMessage";

afterEach(cleanup);

describe("shared UI primitives", () => {
  it("provides consistent controls and semantics", () => {
    render(
      <Panel variant="ranking" aria-label="Settings">
        <FormField label="League name" help="Shown on the draft board."><input /></FormField>
        <Button variant="primary">Save</Button>
        <StatusMessage tone="error">Unable to save.</StatusMessage>
      </Panel>,
    );

    expect(screen.getByRole("region", { name: "Settings" })).toHaveClass("ranking-panel");
    expect(screen.getByRole("textbox", { name: "League name" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toHaveAttribute("type", "button");
    expect(screen.getByRole("alert")).toHaveTextContent("Unable to save.");
  });

  it("opens a modal dialog and handles keyboard cancellation", () => {
    const onClose = vi.fn();
    const { rerender } = render(
      <Dialog open labelledBy="confirm-title" onClose={onClose}>
        <h2 id="confirm-title">Delete league?</h2>
      </Dialog>,
    );

    const dialog = screen.getByRole("dialog", { name: "Delete league?" });
    expect(dialog).toHaveAttribute("open");
    fireEvent(dialog, new Event("cancel", { bubbles: true, cancelable: true }));
    expect(onClose).toHaveBeenCalledOnce();

    rerender(<Dialog open={false} labelledBy="confirm-title" onClose={onClose}><h2 id="confirm-title">Delete league?</h2></Dialog>);
    expect(dialog).not.toHaveAttribute("open");
  });
});
