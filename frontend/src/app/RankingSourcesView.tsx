import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { IdentityReviewQueue } from "./IdentityReviewQueue";
import { ProjectionImport } from "./ProjectionImport";
import { RankingCsvImport } from "./RankingCsvImport";
import { RankingEvidencePanels } from "./RankingEvidencePanels";
import { RankingSourcePreferences } from "./RankingSourcePreferences";
import type { RankingSourcesController } from "./useRankingSources";

interface RankingSourcesViewProps {
  controller: RankingSourcesController;
}

export function RankingSourcesView({ controller }: RankingSourcesViewProps) {
  const heading = useViewHeadingFocus<HTMLHeadingElement>();
  const {
    busy,
    consensusMethod,
    directoryStatus,
    enabledImportedSourceCount,
    enabledSourceCount,
    error,
    identityIssues,
    league,
    message,
    pdfFile,
    preferences,
    preferencesChanged,
    projectionSources,
    rankings,
    sources,
    watchlist,
    actions,
  } = controller;

  return (
    <main className="ranking-page" id="main-content" aria-busy={busy}>
      <Panel variant="ranking" aria-labelledby="ranking-sources-heading">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Ranking data</p>
            <h1 id="ranking-sources-heading" ref={heading} tabIndex={-1}>
              Ranking sources
            </h1>
            <p className="section-description">
              Choose which sources shape {league.name}, then tune their relative influence. Equal weights have equal
              pull.
            </p>
          </div>
          <Button variant="primary" onClick={actions.refresh} disabled={busy}>
            {busy ? "Working..." : "Refresh all sources"}
          </Button>
        </div>
        <StatusMessage>{message}</StatusMessage>
        {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
        <form className="pdf-import-form" onSubmit={actions.uploadPDF}>
          <div>
            <label htmlFor="ranking-pdf">
              <strong>Import a ranking PDF</strong>
            </label>
            <p id="ranking-pdf-help">
              DraftMeld detects supported ESPN PPR Top 300 and Dynasty cheat sheets. Scanned pages use local OCR when
              available. Files are processed locally and are not retained.
            </p>
          </div>
          <input
            id="ranking-pdf"
            name="file"
            type="file"
            accept="application/pdf,.pdf"
            aria-describedby="ranking-pdf-help"
            onChange={(event) => actions.setPDFFile(event.target.files?.[0] ?? null)}
            disabled={busy}
          />
          <Button type="submit" disabled={busy || !pdfFile}>
            Import PDF
          </Button>
        </form>
        <RankingSourcePreferences
          busy={busy}
          consensusMethod={consensusMethod}
          enabledSourceCount={enabledSourceCount}
          leagueName={league.name}
          preferences={preferences}
          preferencesChanged={preferencesChanged}
          sources={sources}
          onConsensusMethodChange={actions.setConsensusMethod}
          onPreferencesChange={actions.setPreferences}
          onReset={actions.resetPreferences}
          onSubmit={actions.saveWeights}
        />
      </Panel>

      <RankingCsvImport busy={busy} sources={sources} onImport={actions.uploadRankingCSV} />
      <ProjectionImport
        busy={busy}
        sources={projectionSources}
        onBusyChange={actions.setBusy}
        onImported={actions.projectionImported}
        onMessage={actions.setMessage}
        onError={actions.setError}
      />
      <IdentityReviewQueue
        busy={busy}
        issues={identityIssues}
        status={directoryStatus}
        onBusyChange={actions.setBusy}
        onReviewed={actions.identityReviewed}
        onError={actions.setError}
      />
      <RankingEvidencePanels
        enabledImportedSourceCount={enabledImportedSourceCount}
        leagueName={league.name}
        rankings={rankings}
        watchlist={watchlist}
      />
    </main>
  );
}
