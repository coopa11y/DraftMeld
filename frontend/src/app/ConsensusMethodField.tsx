import type { LeagueRules } from "../shared/api/types";
import { FormField } from "../shared/ui/FormField";

type ConsensusMethod = LeagueRules["consensusMethod"];

const consensusMethods: Array<{ value: ConsensusMethod; label: string; description: string }> = [
  {
    value: "weighted-median",
    label: "Weighted median",
    description: "Favors the middle weighted rank and limits the effect of one unusually high or low source.",
  },
  {
    value: "trimmed-mean",
    label: "Trimmed mean",
    description: "Removes the highest and lowest ranks before averaging when enough sources are available.",
  },
  {
    value: "weighted-average",
    label: "Weighted average",
    description: "Averages every enabled source, so each source influences the result according to its weight.",
  },
];

interface ConsensusMethodFieldProps {
  disabled?: boolean;
  onChange: (method: ConsensusMethod) => void;
  value: ConsensusMethod;
}

export function ConsensusMethodField({ disabled, onChange, value }: ConsensusMethodFieldProps) {
  const selected = consensusMethods.find((method) => method.value === value) ?? consensusMethods[0];

  return (
    <FormField label="Consensus method" help={selected.description}>
      <select value={value} disabled={disabled} onChange={(event) => onChange(event.target.value as ConsensusMethod)}>
        {consensusMethods.map((method) => (
          <option key={method.value} value={method.value}>
            {method.label} — {method.description}
          </option>
        ))}
      </select>
    </FormField>
  );
}
