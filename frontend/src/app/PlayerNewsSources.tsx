import { useEffect, useRef, useState, type FormEvent } from "react";
import {
  addPlayerNewsSource,
  deletePlayerNewsSource,
  listPlayerNewsSources,
  refreshPlayerNews,
  setPlayerNewsSourceEnabled,
} from "../shared/api/news";
import type { PlayerNewsSource } from "../shared/api/types";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { Button } from "../shared/ui/Button";
import { Dialog } from "../shared/ui/Dialog";
import { FormField } from "../shared/ui/FormField";
import { Panel } from "../shared/ui/Panel";
import { StatusMessage } from "../shared/ui/StatusMessage";

interface PlayerNewsSourcesProps {
  onBack: () => void;
}

export function PlayerNewsSources({ onBack }: PlayerNewsSourcesProps) {
  const [sources, setSources] = useState<PlayerNewsSource[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [addOpen, setAddOpen] = useState(false);
  const [deleteSource, setDeleteSource] = useState<PlayerNewsSource | null>(null);
  const heading = useViewHeadingFocus<HTMLHeadingElement>();
  const nameInput = useRef<HTMLInputElement>(null);

  useEffect(() => {
    let active = true;
    listPlayerNewsSources()
      .then((loaded) => {
        if (active) setSources(loaded);
      })
      .catch((reason: Error) => {
        if (active) setError(reason.message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  async function run(action: () => Promise<void>) {
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await action();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "The player-news action failed.");
    } finally {
      setBusy(false);
    }
  }

  async function reload() {
    setSources(await listPlayerNewsSources());
  }

  async function addSource(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const fields = new FormData(form);
    await run(async () => {
      const created = await addPlayerNewsSource({
        name: String(fields.get("name") ?? ""),
        url: String(fields.get("url") ?? ""),
        attribution: String(fields.get("attribution") ?? ""),
        refreshMinutes: Number(fields.get("refreshMinutes") ?? 15),
      });
      await reload();
      setAddOpen(false);
      form.reset();
      setMessage(`${created.name} was added.`);
    });
  }

  return (
    <main className="league-setup-layout">
      <Panel variant="form" aria-labelledby="player-news-sources-title">
        <Button onClick={onBack}>Back to league</Button>
        <div className="form-heading">
          <div>
            <p className="eyebrow">League configuration</p>
            <h1 id="player-news-sources-title" ref={heading} tabIndex={-1}>
              Player news sources
            </h1>
            <p>Track injuries and availability without adding another permanent section to the draft board.</p>
          </div>
          <div className="league-home-actions">
            <Button variant="primary" onClick={() => setAddOpen(true)}>
              Add RSS source
            </Button>
            <Button
              disabled={busy}
              onClick={() =>
                void run(async () => {
                  const result = await refreshPlayerNews();
                  await reload();
                  if (result.errors.length) setError(result.errors.join(" "));
                  setMessage(`${result.refreshed} sources refreshed; ${result.skipped} skipped.`);
                })
              }
            >
              Refresh news now
            </Button>
          </div>
        </div>

        {message ? <StatusMessage tone="success">{message}</StatusMessage> : null}
        {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
        {loading ? (
          <StatusMessage>Loading player news sources…</StatusMessage>
        ) : (
          <div className="table-scroll" role="region" aria-label="Player news source settings" tabIndex={0}>
            <table className="source-settings-table">
              <caption>Enabled sources are refreshed automatically at the displayed interval.</caption>
              <thead>
                <tr>
                  <th scope="col">Source</th>
                  <th scope="col">Type</th>
                  <th scope="col">Refresh</th>
                  <th scope="col">Enabled</th>
                  <th scope="col">Status</th>
                  <th scope="col">Actions</th>
                </tr>
              </thead>
              <tbody>
                {sources.map((source) => (
                  <tr key={source.id}>
                    <th scope="row">
                      <a href={source.url} target="_blank" rel="noreferrer">
                        {source.name}
                      </a>
                      <span className="player-meta">Content provided by {source.attribution}</span>
                    </th>
                    <td>{source.kind === "rss" ? "RSS/Atom" : "Structured status"}</td>
                    <td>{formatInterval(source.refreshMinutes)}</td>
                    <td>
                      <input
                        type="checkbox"
                        checked={source.enabled}
                        disabled={busy}
                        aria-label={`Enable ${source.name}`}
                        onChange={(event) => {
                          const enabled = event.target.checked;
                          void run(async () => {
                            await setPlayerNewsSourceEnabled(source.id, enabled);
                            setSources((current) =>
                              current.map((item) => (item.id === source.id ? { ...item, enabled } : item)),
                            );
                            setMessage(`${source.name} ${enabled ? "enabled" : "disabled"}.`);
                          });
                        }}
                      />
                    </td>
                    <td>{source.lastError || formatLastRefresh(source.lastRefreshedAt)}</td>
                    <td>
                      {source.builtIn ? (
                        <span>Built in</span>
                      ) : (
                        <Button
                          variant="dangerText"
                          aria-label={`Remove ${source.name}`}
                          onClick={() => setDeleteSource(source)}
                        >
                          Remove
                        </Button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Panel>

      <Dialog
        open={addOpen}
        labelledBy="add-news-source-title"
        initialFocusRef={nameInput}
        onClose={() => setAddOpen(false)}
      >
        <form onSubmit={addSource}>
          <h2 id="add-news-source-title">Add RSS source</h2>
          <p>Use a public HTTPS RSS or Atom feed. DraftMeld stores the provided headline and links to the publisher.</p>
          <FormField label="Source name">
            <input ref={nameInput} name="name" required />
          </FormField>
          <FormField label="RSS or Atom URL" help="Private network and non-HTTPS addresses are blocked.">
            <input name="url" type="url" inputMode="url" required />
          </FormField>
          <FormField label="Publisher attribution" help="The publisher name shown beside imported stories.">
            <input name="attribution" required />
          </FormField>
          <FormField label="Refresh interval">
            <select name="refreshMinutes" defaultValue="15">
              <option value="5">Every 5 minutes</option>
              <option value="10">Every 10 minutes</option>
              <option value="15">Every 15 minutes</option>
              <option value="30">Every 30 minutes</option>
              <option value="60">Every hour</option>
            </select>
          </FormField>
          <div className="dialog-actions">
            <Button type="submit" variant="primary" disabled={busy}>
              Add source
            </Button>
            <Button onClick={() => setAddOpen(false)}>Cancel</Button>
          </div>
        </form>
      </Dialog>

      <Dialog open={deleteSource !== null} labelledBy="remove-news-source-title" onClose={() => setDeleteSource(null)}>
        <h2 id="remove-news-source-title">Remove {deleteSource?.name}?</h2>
        <p>Its imported stories will also be removed. This does not affect rankings or league settings.</p>
        <div className="dialog-actions">
          <Button
            variant="danger"
            disabled={busy}
            onClick={() =>
              void run(async () => {
                if (!deleteSource) return;
                const name = deleteSource.name;
                await deletePlayerNewsSource(deleteSource.id);
                await reload();
                setDeleteSource(null);
                setMessage(`${name} was removed.`);
              })
            }
          >
            Remove source
          </Button>
          <Button onClick={() => setDeleteSource(null)}>Cancel</Button>
        </div>
      </Dialog>
    </main>
  );
}

function formatInterval(minutes: number): string {
  if (minutes === 1440) return "Daily";
  if (minutes === 60) return "Hourly";
  return `Every ${minutes} minutes`;
}

function formatLastRefresh(value?: string): string {
  if (!value) return "Not refreshed yet";
  return `Updated ${new Intl.DateTimeFormat(undefined, { dateStyle: "short", timeStyle: "short" }).format(new Date(value))}`;
}
