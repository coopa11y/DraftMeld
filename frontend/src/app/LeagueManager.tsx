import { useState } from "react";
import { createLeague, deleteLeague, duplicateLeague, listLeagues, updateLeague } from "../shared/api/leagues";
import type { League, LeagueRules } from "../shared/api/types";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { Button } from "../shared/ui/Button";
import { Dialog } from "../shared/ui/Dialog";
import { Panel } from "../shared/ui/Panel";
import { StatusMessage } from "../shared/ui/StatusMessage";
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
  const heading = useViewHeadingFocus<HTMLHeadingElement>(editingId === null);
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
        {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
      </main>
    );
  }

  return (
    <main className="league-setup-layout">
      <Panel variant="form" className="league-manager" aria-labelledby="league-manager-title">
        <div className="form-heading">
          <div>
            <p className="eyebrow">DraftMeld settings</p>
            <h1 id="league-manager-title" ref={heading} tabIndex={-1}>
              Your leagues
            </h1>
            <p>Create a league from familiar defaults, then customize every rule that matters.</p>
          </div>
          <Button variant="primary" onClick={() => setEditingId("new")}>
            Create league
          </Button>
        </div>

        <StatusMessage visuallyHidden>{message}</StatusMessage>
        {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
        {leagues.length === 0 ? (
          <div className="empty-leagues">
            <h2>No leagues yet</h2>
            <p>Create your first league to open the draft board.</p>
          </div>
        ) : null}

        <ul className="league-list">
          {leagues.map((league) => (
            <li key={league.id} className={league.id === activeLeagueId ? "active-league" : ""}>
              <article>
                <div>
                  <h2>{league.name}</h2>
                  <p>
                    {league.teamCount} teams · {formatDraftType(league.draftType)} · Draft position{" "}
                    {league.draftPosition}
                  </p>
                  <p>
                    {formatScoring(league.scoringRules.reception)} ·{" "}
                    {league.rosterSlots.reduce((total, slot) => total + slot.count, 0)} roster spots
                  </p>
                  {league.id === activeLeagueId ? <span className="active-badge">Active league</span> : null}
                </div>
                <div className="league-actions">
                  <Button variant="primary" onClick={() => onOpenDraft(league.id)}>
                    Open draft
                  </Button>
                  <Button onClick={() => setEditingId(league.id)}>Edit</Button>
                  <Button
                    disabled={busy}
                    onClick={() =>
                      run(async () => {
                        const copy = await duplicateLeague(league.id);
                        await refresh(copy.id);
                        setMessage(`${copy.name} was created.`);
                      })
                    }
                  >
                    Duplicate
                  </Button>
                  <Button variant="dangerText" onClick={() => setDeleteId(league.id)}>
                    Delete
                  </Button>
                </div>
                <Dialog
                  open={deleteId === league.id}
                  labelledBy={`delete-league-${league.id}`}
                  onClose={() => setDeleteId(null)}
                >
                  <h2 id={`delete-league-${league.id}`}>Delete {league.name}?</h2>
                  <p>Its draft history will also be permanently deleted.</p>
                  <div className="dialog-actions">
                    <Button
                      variant="danger"
                      disabled={busy}
                      onClick={() =>
                        run(async () => {
                          await deleteLeague(league.id);
                          await refresh();
                          setDeleteId(null);
                          setMessage(`${league.name} was deleted.`);
                          heading.current?.focus();
                        })
                      }
                    >
                      Yes, delete league
                    </Button>
                    <Button onClick={() => setDeleteId(null)}>Cancel</Button>
                  </div>
                </Dialog>
              </article>
            </li>
          ))}
        </ul>
      </Panel>
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
