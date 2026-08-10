import { useState, type Dispatch, type SetStateAction } from "react";
import type { LeagueRules } from "../shared/api/types";
import { FormField } from "../shared/ui/FormField";
import { scoringGroups } from "./scoring";

interface ScoringSettingsProps {
  disabled: boolean;
  rules: LeagueRules;
  setRules: Dispatch<SetStateAction<LeagueRules>>;
  enabledPositions: readonly string[];
}

export function ScoringSettings({ disabled, rules, setRules, enabledPositions }: ScoringSettingsProps) {
  const [openGroups, setOpenGroups] = useState(
    () => new Set(scoringGroups.filter((group) => group.open).map((group) => group.name)),
  );

  function updateScoringRule(name: string, value: number) {
    setRules((current) => ({
      ...current,
      scoringRules: { ...current.scoringRules, [name]: value },
    }));
  }

  return (
    <fieldset disabled={disabled}>
      <legend>Scoring values</legend>
      <p className="field-help">
        Points are awarded per projected event. Enter 0 for categories your league does not use; negative values are
        allowed for turnovers, misses, and high points-allowed tiers.
      </p>
      <div className="scoring-groups">
        {scoringGroups
          .filter((group) => !group.requiresPosition || enabledPositions.includes(group.requiresPosition))
          .map((group) => (
            <details
              key={group.name}
              open={openGroups.has(group.name)}
              onToggle={(event) => {
                const groupName = group.name;
                const isOpen = event.currentTarget.open;
                setOpenGroups((current) => {
                  const next = new Set(current);
                  if (isOpen) next.add(groupName);
                  else next.delete(groupName);
                  return next;
                });
              }}
              className="scoring-group"
            >
              <summary>{group.name}</summary>
              <p className="field-help">{group.description}</p>
              <div className="form-grid scoring-grid">
                {group.fields
                  .filter((field) => !field.requiresPosition || enabledPositions.includes(field.requiresPosition))
                  .map((field) => (
                    <FormField key={field.key} label={`Points per ${field.label.toLowerCase()}`}>
                      <input
                        type="number"
                        step={field.step}
                        value={rules.scoringRules[field.key] ?? 0}
                        onChange={(event) => updateScoringRule(field.key, Number(event.target.value))}
                      />
                    </FormField>
                  ))}
              </div>
            </details>
          ))}
      </div>
    </fieldset>
  );
}
