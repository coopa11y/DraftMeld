import { useState } from "react";
import { links } from "../config";
import { ActionLink } from "./ActionLink";
import { Brand } from "./Brand";
import { Icon } from "./Icon";

const navItems = [
  ["Product", "#product"],
  ["Open source", "#open-source"],
  ["Official builds", "#official-builds"],
  ["Contribute", "#contribute"],
] as const;

export function Header() {
  const [open, setOpen] = useState(false);

  return (
    <header className="site-header">
      <a className="brand-link" href="#top" aria-label="DraftMeld home">
        <Brand />
      </a>
      <button
        className="menu-button"
        type="button"
        aria-expanded={open}
        aria-controls="site-navigation"
        onClick={() => setOpen(!open)}
      >
        <span>Menu</span>
        <Icon name="menu" />
      </button>
      <nav id="site-navigation" className="site-nav" aria-label="Primary" data-open={open || undefined}>
        <ul>
          {navItems.map(([label, href]) => (
            <li key={href}>
              <a href={href} onClick={() => setOpen(false)}>
                {label}
              </a>
            </li>
          ))}
        </ul>
        <a className="github-link" href={links.repository}>
          GitHub <Icon name="external" />
        </a>
        <ActionLink href="#open-source" primary>
          Run DraftMeld
        </ActionLink>
      </nav>
    </header>
  );
}
