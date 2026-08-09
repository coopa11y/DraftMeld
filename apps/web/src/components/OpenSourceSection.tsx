import { links } from "../config";
import { ActionLink } from "./ActionLink";
import { Icon } from "./Icon";

const benefits = [
  ["Open source", "AGPL-3.0 licensed, auditable, extensible, and built in public."],
  ["Run anywhere", "Use the published Docker image on any machine that runs Docker."],
  ["Your data stays yours", "Self-hosted by default, with no account or vendor lock-in."],
  ["Accessible by design", "Keyboard and screen-reader workflows are release requirements."],
] as const;

export function OpenSourceSection() {
  return (
    <section className="section section--open" id="open-source" aria-labelledby="open-heading">
      <div className="section__inner open-layout">
        <div>
          <h2 id="open-heading">Free to run. Built in the open.</h2>
          <p className="section__lead">
            Every DraftMeld feature stays in the community project. Run it, inspect it, change it, and keep control of
            your league data.
          </p>
        </div>
        <div className="terminal" aria-label="Docker Compose command">
          <span aria-hidden="true">$</span>
          <code>docker compose up -d</code>
          <Icon name="code" />
        </div>
        <div className="benefits">
          {benefits.map(([title, text]) => (
            <div className="benefit" key={title}>
              <Icon name="check" />
              <div>
                <h3>{title}</h3>
                <p>{text}</p>
              </div>
            </div>
          ))}
        </div>
        <div className="actions">
          <ActionLink href="#install" primary icon="docker">
            Run with Docker
          </ActionLink>
          <ActionLink href={links.repository} icon="github" external>
            Explore on GitHub
          </ActionLink>
        </div>
        <div className="install" id="install">
          <h3>Start with Compose</h3>
          <pre>
            <code>{`curl -O https://raw.githubusercontent.com/coopa11y/DraftMeld/main/deployments/compose.yaml
docker compose up -d`}</code>
          </pre>
          <p>
            Then open <code>http://localhost:8080</code>. Pin <code>DRAFTMELD_VERSION</code> in your environment for
            repeatable upgrades.
          </p>
        </div>
      </div>
    </section>
  );
}
