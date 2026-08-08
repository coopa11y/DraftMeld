import { useEffect, useState, type FormEvent } from "react";
import { leagueToRules } from "../shared/api/leagues";
import type { League, LeagueRules, RosterSlot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { DraftSettings, LeagueSettings, TeamSettings } from "./LeagueSetupSections";
import { ScoringSettings } from "./ScoringSettings";
import { cloneLeagueRules, defaultLeagueRules, playerPositions } from "./leagueDefaults";

interface LeagueFormProps {
  league?: League;
  initialRules?: LeagueRules;
  busy: boolean;
  onCancel: () => void;
  onChange?: (rules: LeagueRules) => void;
  onSave: (rules: LeagueRules) => Promise<void>;
}

export function LeagueForm({ league, initialRules, busy, onCancel, onChange, onSave }: LeagueFormProps) {
  const [rules, setRules] = useState<LeagueRules>(() =>
    league ? leagueToRules(league) : initialRules ? cloneLeagueRules(initialRules) : defaultLeagueRules(),
  );

  useEffect(() => {
    onChange?.(rules);
  }, [onChange, rules]);

  function updateSlot(index: number, update: Partial<RosterSlot>) {
    setRules((current) => ({
      ...current,
      rosterSlots: current.rosterSlots.map((slot, slotIndex) => (slotIndex === index ? { ...slot, ...update } : slot)),
    }));
  }

  function togglePosition(index: number, position: string) {
    const slot = rules.rosterSlots[index];
    const positions = slot.positions.includes(position)
      ? slot.positions.filter((candidate) => candidate !== position)
      : [...slot.positions, position];
    updateSlot(index, { positions });
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    await onSave(rules);
  }

  return (
    <form className="league-form" onSubmit={handleSubmit}>
      <div className="form-heading">
        <div>
          <p className="eyebrow">League setup</p>
          <h2>{league ? `Edit ${league.name}` : "Create a league"}</h2>
        </div>
        <Button onClick={onCancel} disabled={busy}>
          Cancel
        </Button>
      </div>

      <div>
        <LeagueSettings
          rules={rules}
          setRules={setRules}
          disabled={busy}
          lockSeason={Boolean(league && rules.leagueFormat === "dynasty")}
        />
        <DraftSettings rules={rules} setRules={setRules} disabled={busy} />
        <TeamSettings rules={rules} setRules={setRules} disabled={busy} />
      </div>

      <fieldset disabled={busy}>
        <legend>Roster slots</legend>
        <p className="field-help">Choose how many players can fill each slot and which positions are eligible.</p>
        <div className="roster-editor">
          {rules.rosterSlots.map((slot, index) => (
            <fieldset className="roster-slot" key={index}>
              <legend>Roster slot {index + 1}</legend>
              <FormField label="Slot name">
                <input
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
                      onChange={() => togglePosition(index, position)}
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
                  setRules({ ...rules, rosterSlots: rules.rosterSlots.filter((_, slotIndex) => slotIndex !== index) })
                }
                disabled={rules.rosterSlots.length === 1}
              >
                Remove slot
              </Button>
            </fieldset>
          ))}
        </div>
        <Button
          onClick={() =>
            setRules({
              ...rules,
              rosterSlots: [
                ...rules.rosterSlots,
                { name: "FLEX", count: 1, positions: ["RB", "WR", "TE"], isStarting: true },
              ],
            })
          }
        >
          Add roster slot
        </Button>
      </fieldset>

      <ScoringSettings rules={rules} setRules={setRules} disabled={busy} />

      <div className="form-actions">
        <Button type="submit" variant="primary" disabled={busy}>
          {busy ? "Saving…" : "Save league"}
        </Button>
        <Button onClick={onCancel} disabled={busy}>
          Cancel
        </Button>
      </div>
    </form>
  );
}
