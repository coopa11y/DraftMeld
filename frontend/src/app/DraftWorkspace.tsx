import { DraftWorkspaceView } from "./DraftWorkspaceView";
import { useDraftWorkspace } from "./useDraftWorkspace";

interface DraftWorkspaceProps {
  leagueId: string;
  autoFocusHeading?: boolean;
  mode?: "board" | "tools";
}

export function DraftWorkspace({ leagueId, autoFocusHeading = true, mode = "board" }: DraftWorkspaceProps) {
  return <DraftWorkspaceView controller={useDraftWorkspace(leagueId, autoFocusHeading)} mode={mode} />;
}
