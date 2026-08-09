import { DraftWorkspaceView } from "./DraftWorkspaceView";
import { useDraftWorkspace } from "./useDraftWorkspace";

interface DraftWorkspaceProps {
  leagueId: string;
}

export function DraftWorkspace({ leagueId }: DraftWorkspaceProps) {
  return <DraftWorkspaceView controller={useDraftWorkspace(leagueId)} />;
}
