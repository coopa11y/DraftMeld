import { useState, type FormEvent } from "react";
import { importProjectionCSV } from "../shared/api/rankings";
import type { ProjectionSource } from "../shared/api/types";

const projectionTemplate = "data:text/csv;charset=utf-8,name%2Cposition%2Cteam%2Cadp%2CbyeWeek%2Creception%2CpassingYard%2CpassingTouchdown%2Cinterception%2CrushingYard%2CrushingTouchdown%2CreceivingYard%2CreceivingTouchdown%2CfieldGoalMade%2CextraPointMade%2CdefenseSack%2CdefenseInterception%2CdefenseFumbleRecovery%2CdefenseTouchdown%2CdefenseSafety%0A";

interface ProjectionImportProps {
  busy: boolean;
  sources: ProjectionSource[];
  onBusyChange: (busy: boolean) => void;
  onImported: (source: ProjectionSource) => void;
  onMessage: (message: string) => void;
  onError: (message: string) => void;
}

export function ProjectionImport({ busy, sources, onBusyChange, onImported, onMessage, onError }: ProjectionImportProps) {
  const [name, setName] = useState("");
  const [file, setFile] = useState<File | null>(null);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!name.trim() || !file) return;
    onBusyChange(true);
    onError("");
    try {
      const imported = await importProjectionCSV(name, file);
      onImported(imported);
      onMessage(`${imported.name} imported with ${imported.recordCount} granular player projections.`);
      setName("");
      setFile(null);
      event.currentTarget.reset();
    } catch (reason) {
      onError(reason instanceof Error ? reason.message : "Unable to import projection CSV.");
    } finally {
      onBusyChange(false);
    }
  }

  return (
    <section className="ranking-panel" aria-labelledby="projection-import-heading">
      <div className="section-heading"><div><p className="eyebrow">League scoring</p><h2 id="projection-import-heading">Projection sources</h2><p className="section-description">Import granular statistics so DraftMeld can calculate league-specific points, replacement value, tiers, and auction values. Required columns are name, position, and team; optional statistic columns use the scoring-rule names shown in league setup.</p><a className="inline-action-link" href={projectionTemplate} download="draftmeld-projection-template.csv">Download projection CSV template</a></div></div>
      <form className="data-import-form" onSubmit={submit}>
        <label>Projection source name<input required value={name} onChange={(event) => setName(event.target.value)} placeholder="My projection model" disabled={busy} /></label>
        <label>Projection CSV<input type="file" accept="text/csv,.csv" required onChange={(event) => setFile(event.target.files?.[0] ?? null)} disabled={busy} /></label>
        <button className="secondary-button" type="submit" disabled={busy || !name.trim() || !file}>Import projections</button>
      </form>
      {sources.length > 0 ? <ul className="compact-data-list">{sources.map((source) => <li key={source.id}><strong>{source.name}</strong><span>{source.recordCount} players</span></li>)}</ul> : <p className="empty-state panel-empty-state">No projections imported. Consensus rankings still work, but projected points and VOR remain unavailable.</p>}
    </section>
  );
}
