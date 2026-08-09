import { links } from "../config";
import { Brand } from "./Brand";

const footerLinks = [
  ["Product", "#product"],
  ["Documentation", links.documentation],
  ["Releases", links.releases],
  ["Contributing", links.contributing],
  ["License", links.license],
  ["GitHub", links.repository],
] as const;

export function Footer() {
  return (
    <footer className="site-footer">
      <div>
        <Brand />
        <p>Draft with every ranking on your side.</p>
      </div>
      <nav aria-label="Footer">
        <ul>
          {footerLinks.map(([label, href]) => (
            <li key={label}>
              <a href={href}>{label}</a>
            </li>
          ))}
        </ul>
      </nav>
      <p className="site-footer__legal">AGPL-3.0 · Made in the open.</p>
    </footer>
  );
}
