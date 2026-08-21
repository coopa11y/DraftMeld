import { cleanup, render, screen } from "@testing-library/react";
import { axe } from "jest-axe";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ProjectionSource, RankingSource } from "../shared/api/types";
import { SourceTable } from "./SourceTable";

afterEach(cleanup);

const source = (id: string, name: string, profile: string): RankingSource => ({
  id,
  name,
  profile,
  variantGroup: "draft-sharks",
  description: "Tiered rankings and projections.",
  methodology: "3D Value",
  license: "Proprietary",
  projectUrl: "https://www.draftsharks.com/rankings",
  dataUrl: "https://www.draftsharks.com/rankings/load-table",
  defaultWeight: 0.9,
  defaultEnabled: false,
  importMode: "download",
  role: "ranking",
  isCustom: false,
  recordCount: 0,
});

describe("SourceTable", () => {
  it("shows only the Draft Sharks variant matched to the league", async () => {
    const matched = source("draft-sharks-ppr-1qb", "Draft Sharks - PPR, 1QB", "PPR, 1QB");
    const other = source(
      "draft-sharks-standard-superflex",
      "Draft Sharks - Standard, Superflex",
      "Standard, Superflex",
    );
    const { container } = render(
      <SourceTable
        busy={false}
        busySourceId=""
        consensusMethod="weighted-median"
        enabledSourceCount={1}
        preferences={{ [matched.id]: { weight: 0.9, enabled: true }, [other.id]: { weight: 0.9, enabled: false } }}
        preferencesChanged={false}
        projectionSources={[]}
        recommendations={{
          profile: "redraft, 1-QB, 1 PPR",
          sources: [
            {
              sourceId: matched.id,
              fit: "recommended",
              reason: "Matches this league.",
              preference: { weight: 0.9, enabled: true },
            },
            {
              sourceId: other.id,
              fit: "not recommended",
              reason: "Different preset.",
              preference: { weight: 0.9, enabled: false },
            },
          ],
        }}
        sources={[matched, other]}
        onConsensusMethodChange={vi.fn()}
        onDetails={vi.fn()}
        onImport={vi.fn()}
        onPreferencesChange={vi.fn()}
        onRefreshSource={vi.fn()}
        onRefreshProjectionSource={vi.fn()}
        onReset={vi.fn()}
        onApplyRecommendations={vi.fn()}
        onSubmit={vi.fn()}
      />,
    );
    expect(screen.getByRole("rowheader", { name: matched.name })).toBeVisible();
    expect(screen.queryByRole("rowheader", { name: other.name })).not.toBeInTheDocument();
    expect((await axe(container)).violations).toHaveLength(0);
  });

  it("offers a direct update action for Sleeper projections", async () => {
    const refresh = vi.fn();
    const projection: ProjectionSource = {
      id: "sleeper-projections",
      name: "Sleeper statistical projections",
      description: "Raw offensive projections.",
      methodology: "Raw passing, rushing, and receiving statistics.",
      license: "Noncommercial",
      projectUrl: "https://sleeper.com/fantasy-football",
      dataUrl: "https://api.sleeper.com/projections/nfl/2026",
      importMode: "download",
      publishedAt: "",
      recordCount: 0,
      importedAt: "0001-01-01T00:00:00Z",
    };
    render(
      <SourceTable
        busy={false}
        busySourceId=""
        consensusMethod="weighted-median"
        enabledSourceCount={1}
        preferences={{}}
        preferencesChanged={false}
        projectionSources={[projection]}
        recommendations={{ profile: "redraft, 1-QB, 1 PPR", sources: [] }}
        sources={[]}
        onConsensusMethodChange={vi.fn()}
        onDetails={vi.fn()}
        onImport={vi.fn()}
        onPreferencesChange={vi.fn()}
        onRefreshSource={vi.fn()}
        onRefreshProjectionSource={refresh}
        onReset={vi.fn()}
        onApplyRecommendations={vi.fn()}
        onSubmit={vi.fn()}
      />,
    );
    screen.getByRole("button", { name: "Update Sleeper statistical projections now" }).click();
    expect(refresh).toHaveBeenCalledWith("sleeper-projections");
    expect(screen.getByText("Not imported")).toBeVisible();
  });
});
