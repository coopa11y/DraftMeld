import { useRef, useState } from "react";
import { Button } from "../shared/ui/Button";
import { Dialog } from "../shared/ui/Dialog";
import { FormField } from "../shared/ui/FormField";

interface DraftPositionDialogProps {
  busy: boolean;
  initialPosition: number;
  kind: "real" | "mock";
  onClose: () => void;
  onConfirm: (position: number) => Promise<void>;
  teamCount: number;
}

export function DraftPositionDialog(props: DraftPositionDialogProps) {
  const [position, setPosition] = useState(props.initialPosition || 1);
  const heading = useRef<HTMLHeadingElement>(null);

  async function confirm() {
    try {
      await props.onConfirm(position);
    } catch {
      // The parent presents the API error and leaves this dialog open for correction.
    }
  }

  const action = props.kind === "mock" ? "Start mock draft" : "Start draft";
  return (
    <Dialog open labelledBy={`${props.kind}-draft-position-title`} initialFocusRef={heading} onClose={props.onClose}>
      <h2 id={`${props.kind}-draft-position-title`} ref={heading} tabIndex={-1}>
        Choose your draft position
      </h2>
      <p>
        {props.kind === "mock"
          ? "Choose the pick you want to practice. This does not change the position saved for your real league."
          : "Confirm or change your assigned pick before starting. This position will be saved to your league."}
      </p>
      <FormField label="Your draft position">
        <select value={position} disabled={props.busy} onChange={(event) => setPosition(Number(event.target.value))}>
          {Array.from({ length: props.teamCount }, (_, index) => (
            <option key={index + 1} value={index + 1}>
              Pick {index + 1}
            </option>
          ))}
        </select>
      </FormField>
      <div className="dialog-actions">
        <Button variant="primary" disabled={props.busy} onClick={() => void confirm()}>
          {props.busy ? "Starting…" : action}
        </Button>
        <Button disabled={props.busy} onClick={props.onClose}>
          Cancel
        </Button>
      </div>
    </Dialog>
  );
}
