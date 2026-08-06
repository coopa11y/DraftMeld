import { useEffect, useState } from "react";
import { getHealth, type Health } from "../shared/api/health";

const foundations = [
  ["League profiles", "Custom scoring, roster slots, keepers, and draft formats."],
  ["Source blending", "Weighted rankings and projections with visible disagreement."],
  ["Live draft state", "Persistent picks, roster needs, scarcity, and explainable value."],
] as const;

export function App() {
  const [health, setHealth] = useState<Health | null>(null);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    getHealth()
      .then((result) => {
        setHealth(result);
        setConnected(true);
      })
      .catch(() => setConnected(false));
  }, []);

  return (
    <main>
      <nav aria-label="Primary navigation">
        <a className="brand" href="/">DraftMeld</a>
        <span className={connected ? "status online" : "status"}>
          <span aria-hidden="true" />
          {connected ? `API ${health?.version}` : "Frontend preview"}
        </span>
      </nav>

      <section className="hero" aria-labelledby="page-title">
        <p className="eyebrow">Open-source draft intelligence</p>
        <h1 id="page-title">Every ranking. Every rule. One draft board.</h1>
        <p className="lede">
          DraftMeld is becoming a self-hosted command center for every fantasy
          football league you manage.
        </p>
        <div className="actions">
          <button type="button">Create a league</button>
          <a href="https://github.com/coopa11y/DraftMeld">View the source</a>
        </div>
      </section>

      <section className="foundation" aria-labelledby="foundation-title">
        <div>
          <p className="eyebrow">Foundation milestone</p>
          <h2 id="foundation-title">The architecture is ready for the first feature.</h2>
        </div>
        <div className="cards">
          {foundations.map(([title, description]) => (
            <article key={title}>
              <h3>{title}</h3>
              <p>{description}</p>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}
