import { useState } from "react";
import {
  createLeague,
  createMockDraft,
  deleteLeague,
  duplicateLeague,
  listLeagues,
  updateLeague,
} from "../shared/api/leagues";
import type { League, LeagueRules, MockDraftSession } from "../shared/api/types";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { Button } from "../shared/ui/Button";
import { Dialog } from "../shared/ui/Dialog";
import { Panel } from "../shared/ui/Panel";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { LeagueDataTools } from "./LeagueDataTools";
import { LeagueCreationForm } from "./LeagueCreationForm";
import { LeagueForm } from "./LeagueForm";
import { formatDraftType, formatScoring, LeagueOverview } from "./LeagueOverview";

interface LeagueManagerProps {
  leagues: League[];
  activeLeagueId: string;
  mode: "list" | "overview";
  onLeaguesChange: (leagues: League[], preferredId?: string) => void;
  onOpenLeague: (id: string) => void;
  onOpenDraft: (id: string) => void;
  onOpenMockDraft: (session: MockDraftSession) => void;
  onOpenSources: (id: string) => void;
  onShowAll: () => void;
}

export function LeagueManager(props: LeagueManagerProps) {
  const [editingId, setEditingId] = useState<string | "new" | null>(null);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const heading = useViewHeadingFocus<HTMLHeadingElement>(props.mode === "list" && editingId === null);
  const editingLeague = props.leagues.find((league) => league.id === editingId);
  const activeLeague = props.leagues.find((league) => league.id === props.activeLeagueId);

  async function refresh(preferredId?: string) {
    const updated = await listLeagues();
    props.onLeaguesChange(updated, preferredId);
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
      props.onOpenLeague(saved.id);
    });
  }

  async function openMockDraft(league: League) {
    await run(async () => {
      const session = await createMockDraft(league.id);
      props.onOpenMockDraft(session);
    });
  }

  if (editingId) {
    return (
      <main className="league-setup-layout">
        {editingId === "new" ? (
          <LeagueCreationForm busy={busy} onCancel={() => setEditingId(null)} onSave={handleSave} />
        ) : (
          <LeagueForm league={editingLeague} busy={busy} onCancel={() => setEditingId(null)} onSave={handleSave} />
        )}
        {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
      </main>
    );
  }

  if (props.mode === "overview" && activeLeague) {
    return (
      <main className="league-setup-layout">
        {message ? <StatusMessage tone="success">{message}</StatusMessage> : null}
        {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
        <LeagueOverview
          league={activeLeague}
          busy={busy}
          onBack={props.onShowAll}
          onEdit={() => setEditingId(activeLeague.id)}
          onOpenDraft={() => props.onOpenDraft(activeLeague.id)}
          onOpenSources={() => props.onOpenSources(activeLeague.id)}
          onMockDraft={() => void openMockDraft(activeLeague)}
        />
        <details className="league-data-disclosure">
          <summary>Import, export, and backups</summary>
          <LeagueDataTools
            activeLeagueId={props.activeLeagueId}
            busy={busy}
            leagues={props.leagues}
            onBusyChange={setBusy}
            onError={setError}
            onImported={async (league) => refresh(league.id)}
            onMessage={setMessage}
          />
        </details>
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
            <p>Create a league or open one to configure its rules, rankings, and drafts.</p>
          </div>
          <div className="league-home-actions">
            <Button variant="primary" onClick={() => setEditingId("new")}>
              Create league
            </Button>
          </div>
        </div>

        {message ? <StatusMessage tone="success">{message}</StatusMessage> : null}
        {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
        {props.leagues.length === 0 ? (
          <div className="empty-leagues">
            <h2>No leagues yet</h2>
            <p>Create your first league to open the draft board.</p>
          </div>
        ) : null}

        <ul className="league-list">
          {props.leagues.map((league) => (
            <li key={league.id} className={league.id === props.activeLeagueId ? "active-league" : ""}>
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
                  {league.id === props.activeLeagueId ? <span className="active-badge">Last opened</span> : null}
                </div>
                <div className="league-actions">
                  <Button
                    variant="primary"
                    aria-label={`Open league: ${league.name}`}
                    onClick={() => props.onOpenLeague(league.id)}
                  >
                    Open
                  </Button>
                  <details className="league-more-actions">
                    <summary>More actions</summary>
                    <div>
                      <Button onClick={() => setEditingId(league.id)}>Edit league settings</Button>
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
                  </details>
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
