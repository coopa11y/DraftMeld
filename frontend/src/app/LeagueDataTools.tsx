import { useState, type FormEvent } from "react";
import { downloadLeagueExport, importLeagueBackup, type LeagueExportKind } from "../shared/api/leagues";
import type { League, LeagueBackup } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { FormField } from "../shared/ui/FormField";
import { Panel } from "../shared/ui/Panel";

const maximumBackupBytes = 256 * 1024;

interface LeagueDataToolsProps {
  activeLeagueId: string;
  busy: boolean;
  leagues: League[];
  onBusyChange: (busy: boolean) => void;
  onError: (message: string) => void;
  onImported: (league: League) => Promise<void>;
  onMessage: (message: string) => void;
}

const exports: Array<{ kind: LeagueExportKind; label: string }> = [
  { kind: "backup", label: "League backup (JSON)" },
  { kind: "rankings.csv", label: "Consensus rankings (CSV)" },
  { kind: "draft.csv", label: "Draft results (CSV)" },
  { kind: "draft.json", label: "Draft results (JSON)" },
];

export function LeagueDataTools({
  activeLeagueId,
  busy,
  leagues,
  onBusyChange,
  onError,
  onImported,
  onMessage,
}: LeagueDataToolsProps) {
  const [selectedLeagueId, setSelectedLeagueId] = useState(activeLeagueId || leagues[0]?.id || "");
  const [backupFile, setBackupFile] = useState<File | null>(null);
  const exportLeagueId = leagues.some((league) => league.id === selectedLeagueId)
    ? selectedLeagueId
    : activeLeagueId || leagues[0]?.id || "";

  async function download(kind: LeagueExportKind) {
    if (!exportLeagueId || busy) return;
    onBusyChange(true);
    onError("");
    try {
      const filename = await downloadLeagueExport(exportLeagueId, kind);
      onMessage(`${filename} was downloaded.`);
    } catch (reason) {
      onError(reason instanceof Error ? reason.message : "Unable to download that export.");
    } finally {
      onBusyChange(false);
    }
  }

  async function restore(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!backupFile || busy) return;
    const form = event.currentTarget;
    onBusyChange(true);
    onError("");
    try {
      if (backupFile.size > maximumBackupBytes) throw new Error("League backups must be 256 KiB or smaller.");
      const backup = JSON.parse(await backupFile.text()) as LeagueBackup;
      const imported = await importLeagueBackup(backup);
      await onImported(imported);
      setBackupFile(null);
      form.reset();
      onMessage(`${imported.name} was restored as a new league.`);
    } catch (reason) {
      const message =
        reason instanceof SyntaxError
          ? "That file is not a valid DraftMeld JSON backup."
          : reason instanceof Error
            ? reason.message
            : "Unable to restore that league backup.";
      onError(message);
    } finally {
      onBusyChange(false);
    }
  }

  return (
    <Panel variant="form" className="league-data-tools" aria-labelledby="league-data-tools-title">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Data portability</p>
          <h2 id="league-data-tools-title">Export and backup center</h2>
          <p className="section-description">
            Keep a portable league backup, share an inspectable consensus board, or save your current draft results.
            Restoring a backup always creates a new league and never overwrites existing data.
          </p>
        </div>
      </div>

      <div className="data-tools-grid">
        <div>
          <h3>Download league data</h3>
          {leagues.length > 0 ? (
            <>
              <FormField label="League to export">
                <select
                  value={exportLeagueId}
                  onChange={(event) => setSelectedLeagueId(event.target.value)}
                  disabled={busy}
                >
                  {leagues.map((league) => (
                    <option key={league.id} value={league.id}>
                      {league.name}
                    </option>
                  ))}
                </select>
              </FormField>
              <div className="export-actions">
                {exports.map((item) => (
                  <Button key={item.kind} disabled={busy} onClick={() => void download(item.kind)}>
                    {item.label}
                  </Button>
                ))}
              </div>
            </>
          ) : (
            <p className="empty-state">Create or restore a league before exporting data.</p>
          )}
        </div>

        <form onSubmit={restore}>
          <h3>Restore a league backup</h3>
          <FormField label="DraftMeld backup file" help="Choose a versioned JSON backup no larger than 256 KiB.">
            <input
              type="file"
              accept="application/json,.json"
              required
              disabled={busy}
              onChange={(event) => setBackupFile(event.target.files?.[0] ?? null)}
            />
          </FormField>
          <Button type="submit" variant="primary" disabled={busy || !backupFile}>
            Restore as new league
          </Button>
        </form>
      </div>
    </Panel>
  );
}
