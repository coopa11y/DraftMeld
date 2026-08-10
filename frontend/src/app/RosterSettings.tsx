import { useState, type Dispatch, type SetStateAction } from "react";
import type { LeagueRules } from "../shared/api/types";
import { FormField } from "../shared/ui/FormField";
import { AdvancedRosterEditor } from "./AdvancedRosterEditor";
import { playerPositions } from "./leagueDefaults";
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

interface RosterSettingsProps {
  disabled: boolean;
  rules: LeagueRules;
  setRules: Dispatch<SetStateAction<LeagueRules>>;
}

export function RosterSettings({ disabled, rules, setRules }: RosterSettingsProps) {
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const enabled = enabledRosterPositions(rules.rosterSlots);
  const flexIndex = primaryFlexIndex(rules.rosterSlots);
  const flex = flexIndex >= 0 ? rules.rosterSlots[flexIndex] : undefined;
  const updateRoster = (rosterSlots: LeagueRules["rosterSlots"]) =>
    setRules((current) => ({ ...current, rosterSlots }));

  return (
    <fieldset disabled={disabled}>
      <legend>Roster settings</legend>
      <p className="field-help">
        Choose the positions your league uses. Only relevant lineup and scoring options will be shown.
      </p>
      <fieldset className="position-selector">
        <legend>Positions used</legend>
        {playerPositions.map((position) => (
          <label key={position}>
            <input
              type="checkbox"
              checked={enabled.includes(position)}
              disabled={enabled.length === 1 && enabled[0] === position}
              onChange={(event) => updateRoster(setPositionEnabled(rules.rosterSlots, position, event.target.checked))}
            />
            {position}
          </label>
        ))}
      </fieldset>

      <fieldset className="starting-lineup-settings">
        <legend>Starting lineup</legend>
        <div className="form-grid roster-count-grid">
          {enabled.map((position) => (
            <FormField key={position} label={`${position} starters`}>
              <input
                type="number"
                min="1"
                max="10"
                required
                value={starterCount(rules.rosterSlots, position) || 1}
                onChange={(event) =>
                  updateRoster(setStarterCount(rules.rosterSlots, position, Number(event.target.value)))
                }
              />
            </FormField>
          ))}
        </div>
      </fieldset>

      <fieldset className="flex-settings">
        <legend>Flex</legend>
        <div className="form-grid">
          <FormField label="Flex spots" help="Enter 0 if your league does not use a flex slot.">
            <input
              type="number"
              min="0"
              max="10"
              value={flex?.count ?? 0}
              onChange={(event) => updateRoster(setFlexCount(rules.rosterSlots, Number(event.target.value), enabled))}
            />
          </FormField>
        </div>
        {flex ? (
          <fieldset className="position-options">
            <legend>Flex eligible positions</legend>
            {enabled.map((position) => (
              <label key={position}>
                <input
                  type="checkbox"
                  checked={flex.positions.includes(position)}
                  disabled={flex.positions.length === 1 && flex.positions[0] === position}
                  onChange={(event) => updateRoster(setFlexPosition(rules.rosterSlots, position, event.target.checked))}
                />
                {position}
              </label>
            ))}
          </fieldset>
        ) : null}
      </fieldset>

      <div className="form-grid roster-bench-setting">
        <FormField label="Bench spots">
          <input
            type="number"
            min="0"
            max="30"
            value={benchCount(rules.rosterSlots)}
            onChange={(event) => updateRoster(setBenchCount(rules.rosterSlots, Number(event.target.value), enabled))}
          />
        </FormField>
      </div>

      <details className="advanced-roster-disclosure" onToggle={(event) => setAdvancedOpen(event.currentTarget.open)}>
        <summary>Advanced roster slots</summary>
        {advancedOpen ? <AdvancedRosterEditor rules={rules} setRules={setRules} /> : null}
      </details>
    </fieldset>
  );
}
