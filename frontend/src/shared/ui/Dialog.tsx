import { useEffect, useRef, type ReactNode } from "react";
import { classNames } from "./classNames";

interface DialogProps {
  children: ReactNode;
  className?: string;
  labelledBy: string;
  onClose: () => void;
  open: boolean;
}

export function Dialog({ children, className, labelledBy, onClose, open }: DialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (open && !dialog.open) {
      if (typeof dialog.showModal === "function") dialog.showModal();
      else dialog.setAttribute("open", "");
    } else if (!open && dialog.open) {
      if (typeof dialog.close === "function") dialog.close();
      else dialog.removeAttribute("open");
    }
  }, [open]);

  return (
    <dialog
      aria-labelledby={labelledBy}
      className={classNames("app-dialog", className)}
      onCancel={(event) => {
        event.preventDefault();
        onClose();
      }}
      ref={dialogRef}
    >
      {children}
    </dialog>
  );
}
