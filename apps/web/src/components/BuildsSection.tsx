import { links } from "../config";
import { ActionLink } from "./ActionLink";
import { Icon } from "./Icon";

const community = [
  "Source code and Docker are free",
  "Every product feature included",
  "Self-host and customize",
  "Community support on GitHub",
];
const official = [
  "Signed Windows and Linux packages",
  "Fastest setup",
  "Verified updates",
  "Direct installation support",
];

function BuildOption({
  title,
  points,
  officialBuild = false,
}: {
  title: string;
  points: string[];
  officialBuild?: boolean;
}) {
  return (
    <article className={`build-option${officialBuild ? " build-option--official" : ""}`}>
      <h3>{title}</h3>
      <ul>
        {points.map((point) => (
          <li key={point}>
            <Icon name="check" />
            {point}
          </li>
        ))}
      </ul>
      {officialBuild && <p className="coming-soon">Coming with the first public release</p>}
    </article>
  );
}

export function BuildsSection() {
  return (
    <section className="section field-lines" id="official-builds" aria-labelledby="builds-heading">
      <div className="section__inner">
        <h2 id="builds-heading">Skip the setup, not the transparency.</h2>
        <p className="section__lead section__lead--center">
          Choose the path that fits you. Same product. Same features. Different delivery.
        </p>
        <div className="build-grid">
          <BuildOption title="Community" points={community} />
          <BuildOption title="Official supported builds" points={official} officialBuild />
        </div>
        <p className="plain-language">
          <strong>Official packaging and support are paid convenience, not feature gating.</strong> AGPL source and
          Docker remain open.
        </p>
        <div className="contribute" id="contribute">
          <div>
            <h2>Build it with us.</h2>
            <p>
              Meaningful contributions keep DraftMeld open and independent. Code, documentation, translations, testing,
              ranking adapters, and accessibility work can earn access to official builds.
            </p>
          </div>
          <div className="actions actions--stack">
            <ActionLink href={links.buildPolicy} primary>
              Read the build policy
            </ActionLink>
            <ActionLink href={links.contributing} icon="github" external>
              Contribute on GitHub
            </ActionLink>
          </div>
        </div>
        <div className="version-strip">
          <p>
            <strong>Semantic versioning</strong>
            <span>0.4.0 features · 0.4.1 fixes · 0.5.0 breaking changes</span>
          </p>
          <p>
            <strong>Docker tags</strong>
            <span>latest · 0.4 · 0.4.0 · edge</span>
          </p>
        </div>
      </div>
    </section>
  );
}
