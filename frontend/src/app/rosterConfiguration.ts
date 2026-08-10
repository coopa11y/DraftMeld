import type { RosterSlot } from "../shared/api/types";
import { playerPositions } from "./leagueDefaults";

export type PlayerPosition = (typeof playerPositions)[number];

const normalizedSlotName = (slot: RosterSlot) => slot.name.trim().toUpperCase();

const isPositionSlot = (slot: RosterSlot, position: PlayerPosition) =>
  slot.isStarting &&
  normalizedSlotName(slot) === position &&
  slot.positions.length === 1 &&
  slot.positions[0] === position;

const isBenchSlot = (slot: RosterSlot) => !slot.isStarting && normalizedSlotName(slot) === "BENCH";
const isFlexSlot = (slot: RosterSlot) => slot.isStarting && normalizedSlotName(slot).includes("FLEX");

export function enabledRosterPositions(slots: RosterSlot[]): PlayerPosition[] {
  return playerPositions.filter((position) => slots.some((slot) => isPositionSlot(slot, position)));
}

export function starterCount(slots: RosterSlot[], position: PlayerPosition): number {
  return slots.filter((slot) => isPositionSlot(slot, position)).reduce((total, slot) => total + slot.count, 0);
}

export function setPositionEnabled(slots: RosterSlot[], position: PlayerPosition, enabled: boolean): RosterSlot[] {
  if (enabled) {
    if (slots.some((slot) => isPositionSlot(slot, position))) return slots;
    const insertionIndex = slots.findIndex((slot) => !slot.isStarting);
    const positionSlot: RosterSlot = { name: position, count: 1, positions: [position], isStarting: true };
    const next =
      insertionIndex < 0
        ? [...slots, positionSlot]
        : [...slots.slice(0, insertionIndex), positionSlot, ...slots.slice(insertionIndex)];
    return next.map((slot) =>
      !isBenchSlot(slot) || slot.positions.includes(position)
        ? slot
        : { ...slot, positions: [...slot.positions, position] },
    );
  }

  return slots
    .map((slot) => ({ ...slot, positions: slot.positions.filter((candidate) => candidate !== position) }))
    .filter((slot) => slot.positions.length > 0);
}

export function setStarterCount(slots: RosterSlot[], position: PlayerPosition, count: number): RosterSlot[] {
  let replaced = false;
  return slots.flatMap((slot) => {
    if (!isPositionSlot(slot, position)) return [slot];
    if (replaced) return [];
    replaced = true;
    return [{ ...slot, name: position, count }];
  });
}

export function primaryFlexIndex(slots: RosterSlot[]): number {
  return slots.findIndex(isFlexSlot);
}

export function setFlexCount(slots: RosterSlot[], count: number, eligiblePositions: PlayerPosition[]): RosterSlot[] {
  const index = primaryFlexIndex(slots);
  if (count === 0) return index < 0 ? slots : slots.filter((_, slotIndex) => slotIndex !== index);
  if (index >= 0) return slots.map((slot, slotIndex) => (slotIndex === index ? { ...slot, count } : slot));
  const insertionIndex = slots.findIndex((slot) => !slot.isStarting);
  const flex: RosterSlot = {
    name: eligiblePositions.includes("QB") ? "SUPERFLEX" : "FLEX",
    count,
    positions: eligiblePositions.length > 0 ? eligiblePositions : ["RB", "WR", "TE"],
    isStarting: true,
  };
  return insertionIndex < 0
    ? [...slots, flex]
    : [...slots.slice(0, insertionIndex), flex, ...slots.slice(insertionIndex)];
}

export function setFlexPosition(slots: RosterSlot[], position: PlayerPosition, enabled: boolean): RosterSlot[] {
  const index = primaryFlexIndex(slots);
  if (index < 0) return slots;
  const flex = slots[index];
  const positions = enabled
    ? [...flex.positions, position]
    : flex.positions.filter((candidate) => candidate !== position);
  if (positions.length === 0) return slots;
  return slots.map((slot, slotIndex) =>
    slotIndex === index ? { ...slot, name: positions.includes("QB") ? "SUPERFLEX" : "FLEX", positions } : slot,
  );
}

export function benchCount(slots: RosterSlot[]): number {
  return slots.find(isBenchSlot)?.count ?? 0;
}

export function setBenchCount(slots: RosterSlot[], count: number, positions: PlayerPosition[]): RosterSlot[] {
  const index = slots.findIndex(isBenchSlot);
  if (count === 0) return index < 0 ? slots : slots.filter((_, slotIndex) => slotIndex !== index);
  if (index >= 0) return slots.map((slot, slotIndex) => (slotIndex === index ? { ...slot, count } : slot));
  return [...slots, { name: "Bench", count, positions, isStarting: false }];
}
