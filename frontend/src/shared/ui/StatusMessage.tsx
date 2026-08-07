import { forwardRef, type HTMLAttributes } from "react";
import { classNames } from "./classNames";

type StatusTone = "info" | "success" | "error";

interface StatusMessageProps extends HTMLAttributes<HTMLDivElement> {
  tone?: StatusTone;
  visuallyHidden?: boolean;
}

export const StatusMessage = forwardRef<HTMLDivElement, StatusMessageProps>(function StatusMessage(
  { children, className, tone = "info", visuallyHidden = false, ...props },
  ref,
) {
  return (
    <div
      className={classNames("status-message", `status-message-${tone}`, visuallyHidden && "sr-only", className)}
      ref={ref}
      role={tone === "error" ? "alert" : "status"}
      {...props}
    >
      {children}
    </div>
  );
});
