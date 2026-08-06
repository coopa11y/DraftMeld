import { useRef, useState } from "react";
import { createLeague, deleteLeague, duplicateLeague, listLeagues, updateLeague } from "../shared/api/leagues";
import type { League, LeagueRules } from "../shared/api/types";
import { LeagueForm } from "./LeagueForm";

interface LeagueManagerProps {
  leagues: League[];
  activeLeagueId: string;
  onLeaguesChange: (leagues: League[], preferredId?: string) => void;
  onOpenDraft: (id: string) => void;
}

export function LeagueManager({ leagues, activeLeagueId, onLeaguesChange, onOpenDraft }: LeagueManagerProps) {
  const [editingId, setEditingId] = useState<string | "new" | null>(null);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const heading = useRef<HTMLHeadingElement>(null);
  const editingLeague = leagues.find((league) => league.id === editingId);

  async function refresh(preferredId?: string) {
    const updated = await listLeagues();
    onLeaguesChange(updated, preferredId);
  }

  async function run(action: () => Promise<void>) {
    setBusy(true);
    setError("");
    try {
      await action();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "The league action failed.");
    } finally {
      setBusy(false);
    }
  }

  async function handleSave(rules: LeagueRules) {
    await run(async () => {
      const saved = editingLeague ? await updateLeague(editingLeague.id, rules) : await createLeague(rules);
      await refresh(saved.id);
      setEditingId(null);
      setMessage(`${saved.name} was saved.`);
      heading.current?.focus();
    });
  }

  if (editingId) {
    return (
      <main className="league-setup-layout">
        <LeagueForm league={editingLeague} busy={busy} onCancel={() => setEditingId(null)} onSave={handleSave} />
        {error ? <div className="error-banner" role="alert">{error}</div> : null}
      </main>
    );
  }

  return (
    <main className="league-setup-layout">
      <section className="league-manager" aria-labelledby="league-manager-title">
        <div className="form-heading">
          <div>
            <p className="eyebrow">DraftMeld settings</p>
            <h1 id="league-manager-title" ref={heading} tabIndex={-1}>Your leagues</h1>
            <p>Create a league from familiar defaults, then customize every rule that matters.</p>
          </div>
          <button type="button" className="primary-button" onClick={() => setEditingId("new")}>Create league</button>
        </div>

        <div className="sr-only" role="status" aria-live="polite">{message}</div>
        {error ? <div className="error-banner" role="alert">{error}</div> : null}
        {leagues.length === 0 ? <div className="empty-leagues"><h2>No leagues yet</h2><p>Create your first league to open the draft board.</p></div> : null}

        <ul className="league-list">
          {leagues.map((league) => (
            <li key={league.id} className={league.id === activeLeagueId ? "active-league" : ""}>
              <article>
                <div>
                  <h2>{league.name}</h2>
                  <p>{league.teamCount} teams · {formatDraftType(league.draftType)} · Draft position {league.draftPosition}</p>
                  <p>{formatScoring(league.scoringRules.reception)} · {league.rosterSlots.reduce((total, slot) => total + slot.count, 0)} roster spots</p>
                  {league.id === activeLeagueId ? <span className="active-badge">Active league</span> : null}
                </div>
                <div className="league-actions">
                  <button type="button" className="primary-button" onClick={() => onOpenDraft(league.id)}>Open draft</button>
                  <button type="button" className="secondary-button" onClick={() => setEditingId(league.id)}>Edit</button>
                  <button type="button" className="secondary-button" disabled={busy} onClick={() => run(async () => {
                    const copy = await duplicateLeague(league.id);
                    await refresh(copy.id);
                    setMessage(`${copy.name} was created.`);
                  })}>Duplicate</button>
                  <button type="button" className="danger-text-button" onClick={() => setDeleteId(league.id)}>Delete</button>
                </div>
                {deleteId === league.id ? (
                  <div className="delete-confirmation" role="group" aria-label={`Confirm deletion of ${league.name}`}>
                    <p><strong>Delete {league.name}?</strong> Its draft history will also be permanently deleted.</p>
                    <button type="button" className="danger-button" disabled={busy} onClick={() => run(async () => {
                      await deleteLeague(league.id);
                      await refresh();
                      setDeleteId(null);
                      setMessage(`${league.name} was deleted.`);
                      heading.current?.focus();
                    })}>Yes, delete league</button>
                    <button type="button" className="secondary-button" onClick={() => setDeleteId(null)}>Cancel</button>
                  </div>
                ) : null}
              </article>
            </li>
          ))}
        </ul>
      </section>
    </main>
  );
}

function formatDraftType(type: League["draftType"]): string {
  return `${type.charAt(0).toUpperCase()}${type.slice(1)} draft`;
}

function formatScoring(receptions: number): string {
  if (receptions === 1) return "PPR scoring";
  if (receptions === 0.5) return "Half-PPR scoring";
  if (receptions === 0) return "Standard scoring";
  return `${receptions} points per reception`;
}
