import { useViewHeadingFocus } from "../shared/hooks/useViewHeadingFocus";
import type { ProjectionSource, RankingSource } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { Panel } from "../shared/ui/Panel";

interface SourceDetailsViewProps {
  source?: RankingSource | ProjectionSource;
  onBack: () => void;
}

export function SourceDetailsView({ source, onBack }: SourceDetailsViewProps) {
  const heading = useViewHeadingFocus<HTMLHeadingElement>();
  if (!source) return null;
  const ranking = "importMode" in source;
  return (
    <main className="focused-page" id="main-content">
      <div className="focused-heading">
        <div>
          <p className="eyebrow">Source details</p>
          <h1 ref={heading} tabIndex={-1}>
            {source.name}
          </h1>
        </div>
        <Button onClick={onBack}>Back to sources</Button>
      </div>
      <Panel variant="ranking">
        <dl className="source-details-list">
          <div>
            <dt>Type</dt>
            <dd>{ranking ? source.importMode.replace("-", " ") : "CSV projection"}</dd>
          </div>
          <div>
            <dt>Players</dt>
            <dd>{source.recordCount}</dd>
          </div>
          {ranking ? (
            <>
              <div>
                <dt>Description</dt>
                <dd>{source.description}</dd>
              </div>
              <div>
                <dt>Methodology</dt>
                <dd>{source.methodology}</dd>
              </div>
              <div>
                <dt>License</dt>
                <dd>{source.license}</dd>
              </div>
              <div>
                <dt>Published</dt>
                <dd>{source.publishedAt || "Not available"}</dd>
              </div>
              {source.projectUrl ? (
                <div>
                  <dt>Website</dt>
                  <dd>
                    <a href={source.projectUrl} target="_blank" rel="noreferrer">
                      Open source website
                    </a>
                  </dd>
                </div>
              ) : null}
            </>
          ) : (
            <div>
              <dt>Imported</dt>
              <dd>{new Date(source.importedAt).toLocaleString()}</dd>
            </div>
          )}
        </dl>
      </Panel>
    </main>
  );
}
