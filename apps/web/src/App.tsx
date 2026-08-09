import { ActionLink } from "./components/ActionLink";
import { BuildsSection } from "./components/BuildsSection";
import { DraftBoardPreview } from "./components/DraftBoardPreview";
import { Footer } from "./components/Footer";
import { Header } from "./components/Header";
import { OpenSourceSection } from "./components/OpenSourceSection";
import { WorkflowSection } from "./components/WorkflowSection";
import { links } from "./config";

export function App() {
  return (
    <>
      <Header />
      <main id="main-content">
        <section className="hero field-lines" id="top" aria-labelledby="hero-heading">
          <div className="hero__copy">
            <h1 id="hero-heading">Draft with every ranking on your side.</h1>
            <p>
              Meld trusted rankings with your league rules, roster needs, and live draft state—then make every pick with
              a clear reason.
            </p>
            <div className="actions">
              <ActionLink href="#install" primary icon="docker">
                Run with Docker
              </ActionLink>
              <ActionLink href={links.repository} icon="github" external>
                Explore on GitHub
              </ActionLink>
            </div>
            <p className="hero__note">Free and open source. Your draft data stays yours.</p>
          </div>
          <div className="hero__preview">
            <DraftBoardPreview />
          </div>
        </section>
        <div className="feature-line" aria-label="DraftMeld principles">
          <span>Custom league rules</span>
          <span>Explainable recommendations</span>
          <span>Keyboard and screen-reader friendly</span>
          <span>Open source</span>
        </div>
        <WorkflowSection />
        <OpenSourceSection />
        <BuildsSection />
      </main>
      <Footer />
    </>
  );
}
