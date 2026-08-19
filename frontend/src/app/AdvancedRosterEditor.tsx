import { useRef, useState, type Dispatch, type SetStateAction } from "react";
import type { LeagueRules, RosterSlot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { playerPositions } from "./leagueDefaults";
import { isStandardRosterSlot } from "./rosterConfiguration";

interface AdvancedRosterEditorProps {
  rules: LeagueRules;
  setRules: Dispatch<SetStateAction<LeagueRules>>;
}

export function AdvancedRosterEditor({ rules, setRules }: AdvancedRosterEditorProps) {
  const addButton = useRef<HTMLButtonElement>(null);
  const pendingNewSlotFocus = useRef(false);
  const [announcement, setAnnouncement] = useState("");
  const customSlots = rules.rosterSlots
    .map((slot, index) => ({ index, slot }))
    .filter(({ slot }) => !isStandardRosterSlot(slot));

  const updateSlot = (index: number, update: Partial<RosterSlot>) =>
    setRules((current) => ({
      ...current,
      rosterSlots: current.rosterSlots.map((slot, slotIndex) => (slotIndex === index ? { ...slot, ...update } : slot)),
    }));

  function addCustomSlot() {
    pendingNewSlotFocus.current = true;
    setAnnouncement(`Custom roster slot ${customSlots.length + 1} added.`);
    setRules((current) => ({
      ...current,
      rosterSlots: [
        ...current.rosterSlots,
        { name: "Custom", count: 1, positions: ["RB", "WR", "TE"], isStarting: true },
      ],
    }));
  }

  function removeCustomSlot(index: number, customIndex: number) {
    addButton.current?.focus();
    setAnnouncement(`Custom roster slot ${customIndex + 1} removed.`);
    setRules((current) => ({
      ...current,
      rosterSlots: current.rosterSlots.filter((_, slotIndex) => slotIndex !== index),
    }));
  }

  return (
    <div className="advanced-roster-editor">
      <p className="field-help">Create unusual lineup, taxi, injured-reserve, or position-flexible slots.</p>
      <Button ref={addButton} onClick={addCustomSlot}>
        Add custom slot
      </Button>
      <StatusMessage visuallyHidden>{announcement}</StatusMessage>
      {customSlots.length === 0 ? <p>No custom roster slots have been created.</p> : null}
      {customSlots.map(({ slot, index }, customIndex) => (
        <fieldset className="roster-slot" key={`custom-slot-${index}`}>
          <legend>Custom roster slot {customIndex + 1}</legend>
          <FormField label="Slot name">
            <input
              ref={(node) => {
                if (node && customIndex === customSlots.length - 1 && pendingNewSlotFocus.current) {
                  pendingNewSlotFocus.current = false;
                  node.focus();
                }
              }}
              required
              value={slot.name}
              onChange={(event) => updateSlot(index, { name: event.target.value })}
            />
          </FormField>
          <FormField label="Count">
            <input
              type="number"
              min="1"
              max="30"
              required
              value={slot.count}
              onChange={(event) => updateSlot(index, { count: Number(event.target.value) })}
            />
          </FormField>
          <fieldset className="position-options">
            <legend>Eligible positions</legend>
            {playerPositions.map((position) => (
              <label key={position}>
                <input
                  type="checkbox"
                  checked={slot.positions.includes(position)}
                  onChange={() =>
                    updateSlot(index, {
                      positions: slot.positions.includes(position)
                        ? slot.positions.filter((candidate) => candidate !== position)
                        : [...slot.positions, position],
                    })
                  }
                />
                {position}
              </label>
            ))}
          </fieldset>
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={slot.isStarting}
              onChange={(event) => updateSlot(index, { isStarting: event.target.checked })}
            />
            Starting lineup slot
          </label>
          <Button variant="dangerText" onClick={() => removeCustomSlot(index, customIndex)}>
            Remove slot
          </Button>
        </fieldset>
      ))}
    </div>
  );
}
