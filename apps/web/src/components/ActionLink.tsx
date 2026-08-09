import type { ReactNode } from "react";
import { Icon } from "./Icon";

type ActionLinkProps = {
  children: ReactNode;
  href: string;
  icon?: "docker" | "github";
  primary?: boolean;
  external?: boolean;
};

export function ActionLink({ children, href, icon, primary = false, external = false }: ActionLinkProps) {
  return (
    <a className={`action-link${primary ? " action-link--primary" : ""}`} href={href}>
      {icon && <Icon name={icon} />}
      <span>{children}</span>
      <Icon name={external ? "external" : "arrow"} />
    </a>
  );
}
