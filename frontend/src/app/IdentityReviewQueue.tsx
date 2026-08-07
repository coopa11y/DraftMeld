import { reviewIdentity } from "../shared/api/rankings";
import type { IdentityIssue } from "../shared/api/types";

interface IdentityReviewQueueProps {
  busy: boolean;
  issues: IdentityIssue[];
  onBusyChange: (busy: boolean) => void;
  onReviewed: (issueKey: string, resolution: IdentityIssue["resolution"]) => void;
  onError: (message: string) => void;
}

export function IdentityReviewQueue({ busy, issues, onBusyChange, onReviewed, onError }: IdentityReviewQueueProps) {
  const unresolved = issues.filter((issue) => !issue.resolution);
  if (issues.length === 0) return null;

  async function resolve(issueKey: string, resolution: "confirmed-separate" | "acknowledged") {
    onBusyChange(true);
    onError("");
    try {
      await reviewIdentity(issueKey, resolution);
      onReviewed(issueKey, resolution);
    } catch (reason) {
      onError(reason instanceof Error ? reason.message : "Unable to save the identity review.");
    } finally {
      onBusyChange(false);
    }
  }

  return (
    <section className="ranking-panel" aria-labelledby="identity-review-heading">
      <div className="section-heading"><div><p className="eyebrow">Data quality</p><h2 id="identity-review-heading">Player identity review</h2><p className="section-description">DraftMeld found similar names on the same NFL team and position. Review these exceptions so a risky match never happens silently.</p></div><span className="count-badge">{unresolved.length} to review</span></div>
      <ul className="identity-list">
        {issues.map((issue) => <li key={issue.issueKey} className={issue.resolution ? "reviewed" : ""}><div><strong>{issue.reason}</strong><p>{issue.candidates.map((candidate) => `${candidate.name} (${candidate.position}, ${candidate.team})`).join(" / ")}</p></div>{issue.resolution ? <span>Reviewed</span> : <button className="secondary-button" type="button" disabled={busy} onClick={() => void resolve(issue.issueKey, "confirmed-separate")}>Keep as separate players</button>}</li>)}
      </ul>
    </section>
  );
}
