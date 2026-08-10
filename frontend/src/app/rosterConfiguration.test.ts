import { describe, expect, it } from "vitest";
import type { RosterSlot } from "../shared/api/types";
import {
  benchCount,
  enabledRosterPositions,
  primaryFlexIndex,
  setBenchCount,
  setFlexCount,
  setFlexPosition,
  setPositionEnabled,
  setStarterCount,
  starterCount,
} from "./rosterConfiguration";

const slots: RosterSlot[] = [
  { name: "QB", count: 1, positions: ["QB"], isStarting: true },
  { name: "RB", count: 2, positions: ["RB"], isStarting: true },
  { name: "FLEX", count: 1, positions: ["RB", "WR"], isStarting: true },
  { name: "Bench", count: 6, positions: ["QB", "RB", "WR"], isStarting: false },
];

describe("roster configuration", () => {
  it("derives enabled positions from dedicated starting slots", () => {
    expect(enabledRosterPositions(slots)).toEqual(["QB", "RB"]);
    expect(starterCount(slots, "RB")).toBe(2);
    expect(benchCount(slots)).toBe(6);
  });

  it("enables a position before reserve slots and makes it bench eligible", () => {
    const updated = setPositionEnabled(slots, "TE", true);
    expect(enabledRosterPositions(updated)).toContain("TE");
    expect(updated.at(-1)?.positions).toContain("TE");
    expect(starterCount(updated, "TE")).toBe(1);
  });

  it("removes a disabled position from every slot without leaving empty slots", () => {
    const updated = setPositionEnabled(slots, "QB", false);
    expect(updated.some((slot) => slot.positions.includes("QB"))).toBe(false);
    expect(updated.every((slot) => slot.positions.length > 0)).toBe(true);
  });

  it("consolidates duplicate dedicated slots when their count changes", () => {
    const duplicated = [slots[0], { ...slots[0], count: 2 }, ...slots.slice(1)];
    const updated = setStarterCount(duplicated, "QB", 3);
    expect(starterCount(updated, "QB")).toBe(3);
  });

  it("adds, changes, and removes a flex slot", () => {
    const withoutFlex = slots.filter((slot) => slot.name !== "FLEX");
    const added = setFlexCount(withoutFlex, 2, ["QB", "RB"]);
    expect(added[primaryFlexIndex(added)]).toMatchObject({ name: "SUPERFLEX", count: 2 });
    const withoutQuarterback = setFlexPosition(added, "QB", false);
    expect(withoutQuarterback[primaryFlexIndex(withoutQuarterback)]).toMatchObject({ name: "FLEX", positions: ["RB"] });
    expect(setFlexCount(withoutQuarterback, 0, ["RB"])).toHaveLength(withoutQuarterback.length - 1);
  });

  it("adds and removes a bench when its count crosses zero", () => {
    const withoutBench = slots.filter((slot) => slot.isStarting);
    const added = setBenchCount(withoutBench, 4, ["QB", "RB"]);
    expect(benchCount(added)).toBe(4);
    expect(setBenchCount(added, 0, ["QB", "RB"]).every((slot) => slot.isStarting)).toBe(true);
  });

  it("leaves advanced reserve and multi-position slots out of the simple controls", () => {
    const custom = [
      ...slots,
      { name: "Taxi", count: 3, positions: ["QB", "WR"], isStarting: false },
      { name: "WR/TE", count: 1, positions: ["WR", "TE"], isStarting: true },
    ];
    const updated = setBenchCount(custom, 4, ["QB", "RB"]);
    expect(updated.find((slot) => slot.name === "Taxi")?.count).toBe(3);
    expect(primaryFlexIndex(updated)).toBe(updated.findIndex((slot) => slot.name === "FLEX"));
  });
});
