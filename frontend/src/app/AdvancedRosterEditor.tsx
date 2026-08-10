import type { Dispatch, SetStateAction } from "react";
import type { LeagueRules, RosterSlot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { playerPositions } from "./leagueDefaults";

interface AdvancedRosterEditorProps {
  rules: LeagueRules;
  setRules: Dispatch<SetStateAction<LeagueRules>>;
}

export function AdvancedRosterEditor({ rules, setRules }: AdvancedRosterEditorProps) {
  const updateSlot = (index: number, update: Partial<RosterSlot>) =>
    setRules((current) => ({
      ...current,
      rosterSlots: current.rosterSlots.map((slot, slotIndex) => (slotIndex === index ? { ...slot, ...update } : slot)),
    }));

  return (
    <div className="advanced-roster-editor">
      <p className="field-help">Create unusual lineup, taxi, injured-reserve, or position-flexible slots.</p>
      {rules.rosterSlots.map((slot, index) => (
        <fieldset className="roster-slot" key={`${slot.name}-${index}`}>
          <legend>Roster slot {index + 1}</legend>
          <FormField label="Slot name">
            <input required value={slot.name} onChange={(event) => updateSlot(index, { name: event.target.value })} />
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
          <Button
            variant="dangerText"
            onClick={() =>
              setRules((current) => ({
                ...current,
                rosterSlots: current.rosterSlots.filter((_, slotIndex) => slotIndex !== index),
              }))
            }
            disabled={rules.rosterSlots.length === 1}
          >
            Remove slot
          </Button>
        </fieldset>
      ))}
      <Button
        onClick={() =>
          setRules((current) => ({
            ...current,
            rosterSlots: [
              ...current.rosterSlots,
              { name: "Custom", count: 1, positions: ["RB", "WR", "TE"], isStarting: true },
            ],
          }))
        }
      >
        Add custom slot
      </Button>
    </div>
  );
}
