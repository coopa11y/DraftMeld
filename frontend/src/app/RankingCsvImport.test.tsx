import { axe } from "jest-axe";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { RankingCsvImport } from "./RankingCsvImport";

describe("RankingCsvImport", () => {
  afterEach(cleanup);

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

  it("recognizes a UDK Top 200 export and rejects position-relative files", async () => {
    const user = userEvent.setup();
    const onImport = vi.fn().mockResolvedValue(undefined);
    const { container } = render(<RankingCsvImport busy={false} provider="udk" sources={[]} onImport={onImport} />);

    expect(screen.getByRole("link", { name: "UDK Top 200" })).toHaveAttribute(
      "href",
      "https://www.thefantasyfootballers.com/2026-ultimate-draft-kit/udk-top-200-list/",
    );
    expect(screen.getByLabelText("Ranking source name")).toHaveValue("Fantasy Footballers UDK Top 200");

    const fileInput = screen.getByLabelText("Ranking CSV");
    await user.upload(
      fileInput,
      new File(
        ["Rank,Name,Bye,Team,Pos,Andy,Jason,Mike,Markers\n1,Example Runner,7,BUF,RB,1,2,1,Favorite\n"],
        "udk-top-200.csv",
        { type: "text/csv" },
      ),
    );
    expect(await screen.findByText(/UDK Top 200 recognized/)).toBeInTheDocument();
    expect(screen.queryByText("Match required columns")).not.toBeInTheDocument();
    expect((await axe(container)).violations).toHaveLength(0);

    await user.click(screen.getByRole("button", { name: "Import UDK rankings" }));
    await waitFor(() => expect(onImport).toHaveBeenCalledOnce());

    await user.upload(
      fileInput,
      new File(
        [
          "Name,Position,Team,Bye Week,Rank,Points,Risk,Upside,ADP,Tier,Outlook,Dynasty,Markers\nExample Quarterback,QB,BUF,7,1,350,2,9,2.12,1,Private notes,,\n",
        ],
        "udk-qb.csv",
        { type: "text/csv" },
      ),
    );
    expect(await screen.findByText(/This is a position-ranking export/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Import UDK rankings" })).toBeDisabled();
  });
});
