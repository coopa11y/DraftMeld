import type { DraftSnapshot } from "../shared/api/types";

export function DraftTeamSelect({
  label,
  value,
  snapshot,
  onChange,
}: {
  label: string;
  value: number;
  snapshot: DraftSnapshot;
  onChange: (team: number) => void;
}) {
  return (
    <label>
      <span>{label}</span>
      <select value={value} onChange={(event) => onChange(Number(event.target.value))}>
        {snapshot.teams.map((team) => (
          <option key={team.number} value={team.number}>
            {team.name}
            {team.isUser ? " (you)" : ""}
          </option>
        ))}
      </select>
    </label>
  );
}

export function DraftTradeSummary({
  firstName,
  secondName,
  firstAssets,
  secondAssets,
}: {
  firstName: string;
  secondName: string;
  firstAssets: string[];
  secondAssets: string[];
}) {
  return (
    <dl className="trade-summary">
      <div>
        <dt>{firstName} receives</dt>
        <dd>{firstAssets.join("; ") || "External consideration not tracked in DraftMeld"}</dd>
      </div>
      <div>
        <dt>{secondName} receives</dt>
        <dd>{secondAssets.join("; ") || "External consideration not tracked in DraftMeld"}</dd>
      </div>
    </dl>
  );
}
