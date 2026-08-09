import { axe } from "jest-axe";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { RankingCsvImport } from "./RankingCsvImport";

describe("RankingCsvImport", () => {
  it("auto-maps a private ranking CSV and submits the confirmed columns accessibly", async () => {
    const user = userEvent.setup();
    const onImport = vi.fn().mockResolvedValue(undefined);
    const { container } = render(<RankingCsvImport busy={false} sources={[]} onImport={onImport} />);

    await user.type(screen.getByLabelText("Ranking source name"), "Marcus board");
    await user.upload(
      screen.getByLabelText("Ranking CSV"),
      new File(
        ["Overall Rank,Player Name,Pos,Tm,ADP,Tier,Player ID\n1,Example Runner,RB,ATL,4.2,1,runner-1\n"],
        "board.csv",
        {
          type: "text/csv",
        },
      ),
    );

    expect(await screen.findByLabelText("Overall rank")).toHaveValue("Overall Rank");
    expect(screen.getByLabelText("Player name")).toHaveValue("Player Name");
    expect(screen.getByLabelText("Position")).toHaveValue("Pos");
    await user.click(screen.getByText("Map optional ranking details"));
    expect(screen.getByLabelText("NFL team")).toHaveValue("Tm");
    expect(screen.getByLabelText("Average draft position")).toHaveValue("ADP");
    expect(screen.getByLabelText("Tier")).toHaveValue("Tier");
    expect(screen.getByLabelText("Source player ID")).toHaveValue("Player ID");
    expect((await axe(container)).violations).toHaveLength(0);

    await user.click(screen.getByRole("button", { name: "Import rankings" }));
    await waitFor(() =>
      expect(onImport).toHaveBeenCalledWith(
        "Marcus board",
        expect.objectContaining({ name: "board.csv" }),
        expect.objectContaining({
          rank: "Overall Rank",
          name: "Player Name",
          position: "Pos",
          team: "Tm",
          adp: "ADP",
          tier: "Tier",
          providerId: "Player ID",
        }),
      ),
    );
    expect(screen.getByLabelText("Ranking source name")).toHaveValue("");
  });
});
