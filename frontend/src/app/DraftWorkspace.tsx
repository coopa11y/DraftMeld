import { DraftWorkspaceView } from "./DraftWorkspaceView";
import { useDraftWorkspace } from "./useDraftWorkspace";

interface DraftWorkspaceProps {
  leagueId: string;
  autoFocusHeading?: boolean;
  mode?: "board" | "tools";
  onLeagueRulesChanged?: Parameters<typeof useDraftWorkspace>[2];
}

export function DraftWorkspace({
  leagueId,
  autoFocusHeading = true,
  mode = "board",
  onLeagueRulesChanged,
}: DraftWorkspaceProps) {
  return (
    <DraftWorkspaceView controller={useDraftWorkspace(leagueId, autoFocusHeading, onLeagueRulesChanged)} mode={mode} />
  );
}
