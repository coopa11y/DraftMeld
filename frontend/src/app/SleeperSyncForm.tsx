import { useEffect, useState, type FormEvent } from "react";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { StatusMessage } from "../shared/ui/StatusMessage";

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
      <p>
        Sleeper is treated as the read-only source of truth. Changed or deleted remote picks are reconciled locally;
        DraftMeld never submits a selection.
      </p>
      <FormField label="Draft ID">
        <input
          required
          value={draftId}
          onChange={(event) => {
            const value = event.target.value;
            setDraftId(value);
            if (!value.trim()) setAutoSync(false);
          }}
        />
      </FormField>
      <FormField label="Your roster ID">
        <input
          required
          type="number"
          min="1"
          value={rosterId}
          onChange={(event) => {
            const value = Number(event.target.value);
            setRosterId(value);
            if (value < 1) setAutoSync(false);
          }}
        />
      </FormField>
      <Button type="submit" disabled={busy || !draftId.trim()}>
        Sync picks
      </Button>
      {draftId.trim() && rosterId >= 1 ? (
        <label className="checkbox-label">
          <input type="checkbox" checked={autoSync} onChange={(event) => setAutoSync(event.target.checked)} />
          Automatically check every 15 seconds
        </label>
      ) : null}
      {autoSync ? (
        <StatusMessage className="tool-help">
          Automatic read-only synchronization is active while this page remains open.
        </StatusMessage>
      ) : null}
    </form>
  );
}
