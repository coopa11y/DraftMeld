import type { HTMLAttributes, ReactNode } from "react";
import { classNames } from "./classNames";

type PanelVariant = "board" | "side" | "ranking" | "form";

interface PanelProps extends HTMLAttributes<HTMLElement> {
  children: ReactNode;
  variant: PanelVariant;
}

const variantClasses: Record<PanelVariant, string> = {
  board: "board-panel",
  side: "side-panel",
  ranking: "ranking-panel",
  form: "form-panel",
};

export function Panel({ children, className, variant, ...props }: PanelProps) {
  return (
    <section className={classNames(variantClasses[variant], className)} {...props}>
      {children}
    </section>
  );
}
