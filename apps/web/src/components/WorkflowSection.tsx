import { DraftBoardPreview } from "./DraftBoardPreview";

const steps = [
  {
    title: "Bring your rules",
    text: "Start simple or configure scoring, rosters, keepers, dynasty, auction, and custom rules. Import a PDF or CSV whenever that is faster.",
  },
  {
    title: "Meld the rankings",
    text: "Weight every source, disable one without losing its signal, and reconcile player identities when teams change.",
  },
  {
    title: "Draft with reasons",
    text: "See live recommendations, every position, Draft and Gone actions, undo, and pick trades as the room changes.",
  },
] as const;

export function WorkflowSection() {
  return (
    <section className="section field-lines" id="product" aria-labelledby="workflow-heading">
      <div className="section__inner">
        <h2 id="workflow-heading">Your league. Your board. Your call.</h2>
        <ol className="workflow">
          {steps.map((step, index) => (
            <li key={step.title}>
              <span className="workflow__number">{index + 1}</span>
              <div>
                <h3>{step.title}</h3>
                <p>{step.text}</p>
              </div>
            </li>
          ))}
        </ol>
        <DraftBoardPreview />
      </div>
    </section>
  );
}
