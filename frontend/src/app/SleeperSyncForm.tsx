import { useEffect, useState, type FormEvent } from "react";

interface SleeperSyncFormProps {
  busy: boolean;
  onSync: (draftId: string, rosterId: number) => Promise<void>;
}

export function SleeperSyncForm({ busy, onSync }: SleeperSyncFormProps) {
  const [draftId, setDraftId] = useState("");
  const [rosterId, setRosterId] = useState(1);
  const [autoSync, setAutoSync] = useState(false);

  useEffect(() => {
    if (!autoSync || !draftId.trim() || rosterId < 1) return;
    const interval = window.setInterval(() => {
      if (!busy) void onSync(draftId, rosterId);
    }, 15_000);
    return () => window.clearInterval(interval);
  }, [autoSync, busy, draftId, rosterId, onSync]);

  function submit(event: FormEvent) {
    event.preventDefault();
    void onSync(draftId, rosterId);
  }

  return (
    <form className="sleeper-sync-form" onSubmit={submit}>
      <h3>Sleeper live sync</h3>
      <p>Sleeper is treated as the read-only source of truth. Changed or deleted remote picks are reconciled locally; DraftMeld never submits a selection.</p>
      <label>
        Draft ID
        <input required value={draftId} onChange={(event) => setDraftId(event.target.value)} />
      </label>
      <label>
        Your roster ID
        <input required type="number" min="1" value={rosterId} onChange={(event) => setRosterId(Number(event.target.value))} />
      </label>
      <button className="secondary-button" type="submit" disabled={busy || !draftId.trim()}>Sync picks</button>
      <label className="checkbox-label">
        <input type="checkbox" checked={autoSync} disabled={!draftId.trim() || rosterId < 1} onChange={(event) => setAutoSync(event.target.checked)} />
        Automatically check every 15 seconds
      </label>
      {autoSync ? <p className="tool-help" role="status">Automatic read-only synchronization is active while this page remains open.</p> : null}
    </form>
  );
}
