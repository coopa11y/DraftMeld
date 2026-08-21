import { useState } from "react";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { ProjectionImport } from "./ProjectionImport";
import { RankingCsvImport } from "./RankingCsvImport";
import type { RankingSourcesController } from "./useRankingSources";

export type SourceImportKind = "pdf" | "ranking-csv" | "projection-csv" | "udk";

interface AddSourceViewProps {
  controller: RankingSourcesController;
  initialKind?: SourceImportKind;
  onBack: () => void;
}

export function AddSourceView({ controller, initialKind, onBack }: AddSourceViewProps) {
  const [kind, setKind] = useState<SourceImportKind | undefined>(initialKind);
  const heading = useViewHeadingFocus<HTMLHeadingElement>();
  const { actions, busy, error, message, pdfFile, projectionSources, sources } = controller;

  return (
    <main className="focused-page" id="main-content" aria-busy={busy}>
      <div className="focused-heading">
        <div>
          <p className="eyebrow">Sources</p>
          <h1 ref={heading} tabIndex={-1}>
            Add source
          </h1>
        </div>
        <Button onClick={onBack}>Back to sources</Button>
      </div>
      <StatusMessage>{message}</StatusMessage>
      {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
      {!kind ? (
        <Panel variant="ranking">
          <h2>What are you importing?</h2>
          <div className="source-type-choices">
            <Button className="source-type-choice" onClick={() => setKind("pdf")}>
              <strong>PDF</strong>
              <span>Import supported ranking sheets, including scanned PDFs.</span>
            </Button>
            <Button className="source-type-choice" onClick={() => setKind("ranking-csv")}>
              <strong>CSV</strong>
              <span>Import rankings or projections from any service.</span>
            </Button>
            <Button className="source-type-choice" onClick={() => setKind("udk")}>
              <strong>Fantasy Footballers UDK</strong>
              <span>Import your private subscriber Top 200 export.</span>
            </Button>
          </div>
        </Panel>
      ) : kind === "pdf" ? (
        <Panel variant="ranking">
          <h2>Import a ranking PDF</h2>
          <p>DraftMeld detects supported sheets and uses local OCR for scanned pages. Files are not retained.</p>
          <form className="data-import-form" onSubmit={actions.uploadPDF}>
            <label htmlFor="ranking-pdf">Ranking PDF</label>
            <input
              id="ranking-pdf"
              type="file"
              accept="application/pdf,.pdf"
              onChange={(event) => actions.setPDFFile(event.target.files?.[0] ?? null)}
              disabled={busy}
            />
            <Button variant="primary" type="submit" disabled={busy || !pdfFile}>
              Import PDF
            </Button>
          </form>
        </Panel>
      ) : kind === "ranking-csv" || kind === "udk" ? (
        <Panel variant="ranking">
          <div className="csv-kind-choice">
            <Button aria-pressed="true">Rankings</Button>
            <Button onClick={() => setKind("projection-csv")}>Projections</Button>
          </div>
          <RankingCsvImport
            embedded
            busy={busy}
            provider={kind === "udk" ? "udk" : "generic"}
            sources={sources}
            onImport={actions.uploadRankingCSV}
          />
        </Panel>
      ) : (
        <Panel variant="ranking">
          <div className="csv-kind-choice">
            <Button onClick={() => setKind("ranking-csv")}>Rankings</Button>
            <Button aria-pressed="true">Projections</Button>
          </div>
          <ProjectionImport
            embedded
            busy={busy}
            sources={projectionSources}
            onBusyChange={actions.setBusy}
            onImported={actions.projectionImported}
            onMessage={actions.setMessage}
            onError={actions.setError}
          />
        </Panel>
      )}
    </main>
  );
}
