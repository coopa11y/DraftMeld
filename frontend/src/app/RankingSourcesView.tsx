import { useState } from "react";
import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";
import { StatusMessage } from "../shared/ui/StatusMessage";
import { AddSourceView, type SourceImportKind } from "./AddSourceView";
import { IdentityReviewQueue } from "./IdentityReviewQueue";
import { RankingEvidencePanels } from "./RankingEvidencePanels";
import { SourceDetailsView } from "./SourceDetailsView";
import { SourceTable } from "./SourceTable";
import type { RankingSourcesController } from "./useRankingSources";

interface RankingSourcesViewProps {
  controller: RankingSourcesController;
}
type SourceView = "list" | "add" | "details" | "identity" | "results";

export function RankingSourcesView({ controller }: RankingSourcesViewProps) {
  const [view, setView] = useState<SourceView>("list");
  const [importKind, setImportKind] = useState<SourceImportKind>();
  const [details, setDetails] = useState<{ id: string; kind: "ranking" | "projection" }>();
  const heading = useViewHeadingFocus<HTMLHeadingElement>(view === "list");
  const { actions, busy, busySourceId, error, identityIssues, message, projectionSources, sources } = controller;

  function openImport(kind?: SourceImportKind) {
    setImportKind(kind);
    setView("add");
  }

  if (view === "add")
    return <AddSourceView controller={controller} initialKind={importKind} onBack={() => setView("list")} />;
  if (view === "details") {
    const source =
      details?.kind === "projection"
        ? projectionSources.find((item) => item.id === details.id)
        : sources.find((item) => item.id === details?.id);
    return <SourceDetailsView source={source} onBack={() => setView("list")} />;
  }
  if (view === "identity") {
    return (
      <main className="focused-page" id="main-content">
        <Button onClick={() => setView("list")}>Back to sources</Button>
        <IdentityReviewQueue
          busy={busy}
          issues={identityIssues}
          status={controller.directoryStatus}
          onBusyChange={actions.setBusy}
          onReviewed={actions.identityReviewed}
          onError={actions.setError}
        />
      </main>
    );
  }
  if (view === "results") {
    return (
      <main className="focused-page" id="main-content">
        <Button onClick={() => setView("list")}>Back to sources</Button>
        <RankingEvidencePanels
          enabledImportedSourceCount={controller.enabledImportedSourceCount}
          leagueName={controller.league.name}
          rankings={controller.rankings}
          watchlist={controller.watchlist}
        />
      </main>
    );
  }

  return (
    <main className="sources-page" id="main-content" aria-busy={busy}>
      <div className="page-title-row">
        <div>
          <p className="eyebrow">Ranking data</p>
          <h1 ref={heading} tabIndex={-1}>
            Sources
          </h1>
          <p>Keep rankings and projections current without leaving this list.</p>
        </div>
        <div className="page-actions">
          <Button variant="primary" onClick={() => openImport()}>
            Add source
          </Button>
          <Button onClick={actions.refresh} disabled={busy || Boolean(busySourceId)}>
            {busy ? "Refreshing…" : "Refresh online sources"}
          </Button>
        </div>
      </div>
      <StatusMessage>{message}</StatusMessage>
      {error ? <StatusMessage tone="error">{error}</StatusMessage> : null}
      <Panel variant="ranking">
        <SourceTable
          busy={busy}
          busySourceId={busySourceId}
          consensusMethod={controller.consensusMethod}
          enabledSourceCount={controller.enabledSourceCount}
          preferences={controller.preferences}
          preferencesChanged={controller.preferencesChanged}
          projectionSources={projectionSources}
          sources={sources}
          onConsensusMethodChange={actions.setConsensusMethod}
          onDetails={(id, kind) => {
            setDetails({ id, kind });
            setView("details");
          }}
          onImport={openImport}
          onPreferencesChange={actions.setPreferences}
          onRefreshSource={actions.refreshSource}
          onReset={actions.resetPreferences}
          onSubmit={actions.saveWeights}
        />
      </Panel>
      <div className="source-followups">
        <p>{identityIssues.filter((issue) => !issue.resolution).length} identity issues need review.</p>
        <Button onClick={() => setView("identity")}>Identity issues</Button>
        <p>
          {controller.rankings.length > 0
            ? "Consensus results are ready."
            : "Import a source to build consensus results."}
        </p>
        <Button onClick={() => setView("results")} disabled={controller.rankings.length === 0}>
          Consensus results
        </Button>
      </div>
    </main>
  );
}
