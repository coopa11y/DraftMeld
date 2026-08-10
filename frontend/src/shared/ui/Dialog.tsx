import { useEffect, useRef, type ReactNode, type RefObject } from "react";
import { classNames } from "./classNames";

interface DialogProps {
  children: ReactNode;
  className?: string;
  labelledBy: string;
  initialFocusRef?: RefObject<HTMLElement | null>;
  onClose: () => void;
  open: boolean;
}

export function Dialog({ children, className, initialFocusRef, labelledBy, onClose, open }: DialogProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (open && !dialog.open) {
      if (typeof dialog.showModal === "function") dialog.showModal();
      else dialog.setAttribute("open", "");
      queueMicrotask(() => initialFocusRef?.current?.focus());
    } else if (!open && dialog.open) {
      if (typeof dialog.close === "function") dialog.close();
      else dialog.removeAttribute("open");
    }
  }, [initialFocusRef, open]);

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
