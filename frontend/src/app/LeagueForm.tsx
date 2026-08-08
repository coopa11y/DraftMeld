import { useState, type FormEvent } from "react";
import { leagueToRules } from "../shared/api/leagues";
import type { League, LeagueRules, RosterSlot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { DraftSettings, LeagueSettings, TeamSettings } from "./LeagueSetupSections";

const playerPositions = ["QB", "RB", "WR", "TE", "K", "DST"] as const;
const scoringFields = [
  ["reception", "Points per reception", 0.5],
  ["passingYard", "Points per passing yard", 0.01],
  ["passingTouchdown", "Points per passing touchdown", 1],
  ["interception", "Points per interception", 0.5],
  ["rushingYard", "Points per rushing yard", 0.01],
  ["rushingTouchdown", "Points per rushing touchdown", 1],
  ["receivingYard", "Points per receiving yard", 0.01],
  ["receivingTouchdown", "Points per receiving touchdown", 1],
  ["fieldGoalMade", "Points per field goal made", 0.5],
  ["extraPointMade", "Points per extra point made", 0.5],
  ["defenseSack", "Points per defensive sack", 0.5],
  ["defenseInterception", "Points per defensive interception", 0.5],
  ["defenseFumbleRecovery", "Points per defensive fumble recovery", 0.5],
  ["defenseTouchdown", "Points per defensive touchdown", 1],
  ["defenseSafety", "Points per defensive safety", 0.5],
] as const;

const defaultRoster: RosterSlot[] = [
  { name: "QB", count: 1, positions: ["QB"], isStarting: true },
  { name: "RB", count: 2, positions: ["RB"], isStarting: true },
  { name: "WR", count: 2, positions: ["WR"], isStarting: true },
  { name: "TE", count: 1, positions: ["TE"], isStarting: true },
  { name: "FLEX", count: 1, positions: ["RB", "WR", "TE"], isStarting: true },
  { name: "K", count: 1, positions: ["K"], isStarting: true },
  { name: "DST", count: 1, positions: ["DST"], isStarting: true },
  { name: "Bench", count: 6, positions: [...playerPositions], isStarting: false },
];

function defaultRules(): LeagueRules {
  return {
    name: "My League",
    teamCount: 12,
    draftPosition: 1,
    teamNames: ["My Team", ...Array.from({ length: 11 }, () => "")],
    draftType: "snake",
    rosterSlots: defaultRoster.map((slot) => ({ ...slot, positions: [...slot.positions] })),
    scoringRules: {
      reception: 1,
      passingYard: 0.04,
      passingTouchdown: 4,
      interception: -2,
      rushingYard: 0.1,
      rushingTouchdown: 6,
      receivingYard: 0.1,
      receivingTouchdown: 6,
      fieldGoalMade: 3,
      extraPointMade: 1,
      defenseSack: 1,
      defenseInterception: 2,
      defenseFumbleRecovery: 2,
      defenseTouchdown: 6,
      defenseSafety: 2,
    },
    sourcePreferences: {},
    consensusMethod: "weighted-median",
    playerPreferences: {},
    auctionBudget: 200,
    auctionMinimumBid: 1,
    keeperBudgetSpent: 0,
    myKeeperSpend: 0,
    keeperValueRemoved: 0,
  };
}

interface LeagueFormProps {
  league?: League;
  busy: boolean;
  onCancel: () => void;
  onSave: (rules: LeagueRules) => Promise<void>;
}

export function LeagueForm({ league, busy, onCancel, onSave }: LeagueFormProps) {
  const [rules, setRules] = useState<LeagueRules>(() => (league ? leagueToRules(league) : defaultRules()));

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
        <LeagueSettings rules={rules} setRules={setRules} disabled={busy} />
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

      <fieldset disabled={busy}>
        <legend>Scoring values</legend>
        <div className="form-grid scoring-grid">
          {scoringFields.map(([name, label, step]) => (
            <FormField key={name} label={label}>
              <input
                type="number"
                step={step}
                value={rules.scoringRules[name] ?? 0}
                onChange={(event) =>
                  setRules({ ...rules, scoringRules: { ...rules.scoringRules, [name]: Number(event.target.value) } })
                }
              />
            </FormField>
          ))}
        </div>
      </fieldset>

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
