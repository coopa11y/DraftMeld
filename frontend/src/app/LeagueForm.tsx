import { useState, type FormEvent } from "react";
import { leagueToRules } from "../shared/api/leagues";
import type { League, LeagueRules, RosterSlot } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";

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

function AuctionSettings({ rules, onChange }: { rules: LeagueRules; onChange: (update: Partial<LeagueRules>) => void }) {
  return <>
    <FormField label="Team auction budget"><input type="number" min="1" step="1" required value={rules.auctionBudget} onChange={(event) => onChange({ auctionBudget: Number(event.target.value) })} /></FormField>
    <FormField label="Minimum bid"><input type="number" min="1" max={rules.auctionBudget} step="1" required value={rules.auctionMinimumBid} onChange={(event) => onChange({ auctionMinimumBid: Number(event.target.value) })} /></FormField>
    <FormField label="My keeper spend" help="Dollars already committed by your team."><input type="number" min="0" max={rules.auctionBudget} step="1" value={rules.myKeeperSpend} onChange={(event) => onChange({ myKeeperSpend: Number(event.target.value) })} /></FormField>
    <FormField label="League-wide keeper spend" help="Total dollars already committed to keepers across every team."><input type="number" min={rules.myKeeperSpend} max={rules.auctionBudget * rules.teamCount} step="1" value={rules.keeperBudgetSpent} onChange={(event) => onChange({ keeperBudgetSpent: Number(event.target.value) })} /></FormField>
    <FormField label="Keeper value removed" help="DraftMeld baseline dollar value of players already kept."><input type="number" min="0" step="1" value={rules.keeperValueRemoved} onChange={(event) => onChange({ keeperValueRemoved: Number(event.target.value) })} /></FormField>
  </>;
}

export function LeagueForm({ league, busy, onCancel, onSave }: LeagueFormProps) {
  const [rules, setRules] = useState<LeagueRules>(() => league ? leagueToRules(league) : defaultRules());

  function updateSlot(index: number, update: Partial<RosterSlot>) {
    setRules((current) => ({
      ...current,
      rosterSlots: current.rosterSlots.map((slot, slotIndex) => slotIndex === index ? { ...slot, ...update } : slot),
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
        <Button onClick={onCancel} disabled={busy}>Cancel</Button>
      </div>

      <fieldset disabled={busy}>
        <legend>Basic information</legend>
        <div className="form-grid">
          <FormField label="League name"><input required maxLength={80} value={rules.name} onChange={(event) => setRules({ ...rules, name: event.target.value })} /></FormField>
          <FormField label="Draft format"><select value={rules.draftType} onChange={(event) => setRules({ ...rules, draftType: event.target.value as LeagueRules["draftType"] })}>
            <option value="snake">Snake</option><option value="linear">Linear</option><option value="auction">Auction</option>
          </select></FormField>
          <FormField label="Number of teams"><input type="number" min="2" max="32" required value={rules.teamCount} onChange={(event) => {
            const teamCount = Number(event.target.value);
            setRules({ ...rules, teamCount, draftPosition: Math.min(rules.draftPosition, teamCount) });
          }} /></FormField>
          <FormField label="Your draft position"><input type="number" min="1" max={rules.teamCount} required value={rules.draftPosition} onChange={(event) => setRules({ ...rules, draftPosition: Number(event.target.value) })} /></FormField>
          <FormField label="Scoring preset"><select value={scoringPreset(rules.scoringRules.reception)} onChange={(event) => {
            if (event.target.value === "custom") return;
            setRules({ ...rules, scoringRules: { ...rules.scoringRules, reception: Number(event.target.value) } });
          }}>
            <option value="0">Standard</option><option value="0.5">Half PPR</option><option value="1">PPR</option><option value="custom">Custom</option>
          </select></FormField>
          <FormField label="Consensus method"><select value={rules.consensusMethod} onChange={(event) => setRules({ ...rules, consensusMethod: event.target.value as LeagueRules["consensusMethod"] })}>
            <option value="weighted-median">Weighted median</option><option value="trimmed-mean">Trimmed mean</option><option value="weighted-average">Weighted average</option>
          </select></FormField>
          {rules.draftType === "auction" ? <AuctionSettings rules={rules} onChange={(update) => setRules((current) => ({ ...current, ...update }))} /> : null}
        </div>
      </fieldset>

      <fieldset disabled={busy}>
        <legend>Roster slots</legend>
        <p className="field-help">Choose how many players can fill each slot and which positions are eligible.</p>
        <div className="roster-editor">
          {rules.rosterSlots.map((slot, index) => (
            <fieldset className="roster-slot" key={index}>
              <legend>Roster slot {index + 1}</legend>
              <FormField label="Slot name"><input required value={slot.name} onChange={(event) => updateSlot(index, { name: event.target.value })} /></FormField>
              <FormField label="Count"><input type="number" min="1" max="30" required value={slot.count} onChange={(event) => updateSlot(index, { count: Number(event.target.value) })} /></FormField>
              <fieldset className="position-options">
                <legend>Eligible positions</legend>
                {playerPositions.map((position) => <label key={position}><input type="checkbox" checked={slot.positions.includes(position)} onChange={() => togglePosition(index, position)} />{position}</label>)}
              </fieldset>
              <label className="checkbox-label"><input type="checkbox" checked={slot.isStarting} onChange={(event) => updateSlot(index, { isStarting: event.target.checked })} />Starting lineup slot</label>
              <Button variant="dangerText" onClick={() => setRules({ ...rules, rosterSlots: rules.rosterSlots.filter((_, slotIndex) => slotIndex !== index) })} disabled={rules.rosterSlots.length === 1}>Remove slot</Button>
            </fieldset>
          ))}
        </div>
        <Button onClick={() => setRules({ ...rules, rosterSlots: [...rules.rosterSlots, { name: "FLEX", count: 1, positions: ["RB", "WR", "TE"], isStarting: true }] })}>Add roster slot</Button>
      </fieldset>

      <fieldset disabled={busy}>
        <legend>Scoring values</legend>
        <div className="form-grid scoring-grid">
          {scoringFields.map(([name, label, step]) => <FormField key={name} label={label}><input type="number" step={step} value={rules.scoringRules[name] ?? 0} onChange={(event) => setRules({ ...rules, scoringRules: { ...rules.scoringRules, [name]: Number(event.target.value) } })} /></FormField>)}
        </div>
      </fieldset>

      <div className="form-actions">
        <Button type="submit" variant="primary" disabled={busy}>{busy ? "Saving…" : "Save league"}</Button>
        <Button onClick={onCancel} disabled={busy}>Cancel</Button>
      </div>
    </form>
  );
}

function scoringPreset(receptions: number): string {
  return receptions === 0 || receptions === 0.5 || receptions === 1 ? String(receptions) : "custom";
}
