import type { Dispatch, SetStateAction } from "react";
import type { LeagueRules } from "../shared/api/types";
import { FormField } from "../shared/ui/FormField";
import { ScoringSettings } from "./ScoringSettings";
import { applyReceptionPreset, receptionPreset } from "./scoring";

interface OnboardingScoringStepProps {
  rules: LeagueRules;
  setRules: Dispatch<SetStateAction<LeagueRules>>;
}

export function OnboardingScoringStep({ rules, setRules }: OnboardingScoringStepProps) {
  return (
    <section aria-labelledby="scoring-step-title">
      <h2 id="scoring-step-title" tabIndex={-1}>
        Scoring
      </h2>
      <p>Choose a familiar reception format or adjust any scoring value manually.</p>
      <FormField label="Reception scoring preset" help="This changes only reception and TE-premium values.">
        <select
          value={receptionPreset(rules.scoringRules)}
          onChange={(event) =>
            setRules((current) => ({
              ...current,
              scoringRules: applyReceptionPreset(current.scoringRules, event.target.value),
            }))
          }
        >
          <option value="0">Standard</option>
          <option value="0.5">Half PPR</option>
          <option value="1">PPR</option>
          <option value="te-premium">TE premium PPR (+0.5)</option>
          <option value="custom">Custom</option>
        </select>
      </FormField>

      <ScoringSettings rules={rules} setRules={setRules} disabled={false} />
    </section>
  );
}
