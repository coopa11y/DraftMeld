import { DraftWorkspaceView } from "./DraftWorkspaceView";
import { useDraftWorkspace } from "./useDraftWorkspace";

interface DraftWorkspaceProps {
  leagueId: string;
  autoFocusHeading?: boolean;
}

export function DraftWorkspace({ leagueId, autoFocusHeading = true }: DraftWorkspaceProps) {
  return <DraftWorkspaceView controller={useDraftWorkspace(leagueId, autoFocusHeading)} />;
}
