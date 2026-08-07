import { useState } from "react";
import { reviewIdentity } from "../shared/api/rankings";
import type { IdentityIssue } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { Panel } from "../shared/ui/Panel";

interface IdentityReviewQueueProps {
  busy: boolean;
  issues: IdentityIssue[];
  onBusyChange: (busy: boolean) => void;
  onReviewed: (issueKey: string, resolution: IdentityIssue["resolution"], canonicalPlayerKey?: string) => void;
  onError: (message: string) => void;
}

export function IdentityReviewQueue({ busy, issues, onBusyChange, onReviewed, onError }: IdentityReviewQueueProps) {
  const [canonicalByIssue, setCanonicalByIssue] = useState<Record<string, string>>({});
  const unresolved = issues.filter((issue) => !issue.resolution);
  if (issues.length === 0) return null;

  async function resolve(issue: IdentityIssue, resolution: "confirmed-separate" | "merged") {
    const canonical = resolution === "merged" ? (canonicalByIssue[issue.issueKey] ?? issue.candidates[0]?.playerKey ?? "") : "";
    onBusyChange(true); onError("");
    try {
      await reviewIdentity(issue.issueKey, resolution, canonical);
      onReviewed(issue.issueKey, resolution, canonical);
    } catch (reason) {
      onError(reason instanceof Error ? reason.message : "Unable to save the identity review.");
    } finally {
      onBusyChange(false);
    }
  }

  return (
    <Panel variant="ranking" aria-labelledby="identity-review-heading">
      <div className="section-heading"><div><p className="eyebrow">Data quality</p><h2 id="identity-review-heading">Player identity review</h2><p className="section-description">Keep distinct people separate, or merge aliases that refer to the same player. A merge combines their ranking and projection signals under the selected canonical identity.</p></div><span className="count-badge">{unresolved.length} to review</span></div>
      <ul className="identity-list">
        {issues.map((issue) => <li key={issue.issueKey} className={issue.resolution ? "reviewed" : ""}><div><strong>{issue.reason}</strong><p>{issue.candidates.map((candidate) => `${candidate.name} (${candidate.position}, ${candidate.team})`).join(" / ")}</p>{issue.resolution === "merged" ? <span>Merged as {issue.candidates.find((candidate) => candidate.playerKey === issue.canonicalPlayerKey)?.name ?? issue.canonicalPlayerKey}</span> : null}</div>{issue.resolution ? <span>{issue.resolution === "merged" ? "Merged" : "Kept separate"}</span> : <div className="identity-actions"><FormField label="Canonical player"><select value={canonicalByIssue[issue.issueKey] ?? issue.candidates[0]?.playerKey ?? ""} onChange={(event) => setCanonicalByIssue((current) => ({ ...current, [issue.issueKey]: event.target.value }))} disabled={busy}>{issue.candidates.map((candidate) => <option key={candidate.playerKey} value={candidate.playerKey}>{candidate.name}</option>)}</select></FormField><Button disabled={busy} onClick={() => void resolve(issue, "merged")}>Merge aliases</Button><Button disabled={busy} onClick={() => void resolve(issue, "confirmed-separate")}>Keep separate</Button></div>}</li>)}
      </ul>
    </Panel>
  );
}
