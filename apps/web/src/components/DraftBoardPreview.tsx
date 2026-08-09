const players = [
  ["1", "Jalen Rivers", "RB", "WES", "7", "1.01"],
  ["2", "Malik Corbin", "WR", "NOR", "9", "1.02"],
  ["3", "Ethan Hale", "RB", "SOU", "5", "1.03"],
  ["4", "Dorian Brooks", "WR", "EAS", "10", "1.04"],
] as const;

export function DraftBoardPreview() {
  return (
    <div
      className="board"
      role="img"
      aria-label="Example DraftMeld board recommending Malik Corbin with clear draft reasons"
    >
      <div aria-hidden="true">
        <div className="board__bar">
          <strong>
            <span className="status-dot" /> Draft
          </strong>
          <span>12-Team PPR</span>
          <span>Pick 1.01</span>
          <span>
            On the clock: <strong>You</strong>
          </span>
        </div>
        <div className="board__tabs" aria-label="Position views">
          <span className="active">Overall</span>
          <span>QB</span>
          <span>RB</span>
          <span>WR</span>
          <span>TE</span>
          <span>FLEX</span>
          <span>K</span>
          <span>DEF</span>
        </div>
        <div className="board__body">
          <div className="board__rankings">
            <div className="player-row player-row--head">
              <span>RK</span>
              <span>PLAYER</span>
              <span>POS</span>
              <span>TEAM</span>
              <span>BYE</span>
              <span>ADP</span>
            </div>
            {players.map((player, index) => (
              <div className={`player-row${index === 1 ? " player-row--selected" : ""}`} key={player[1]}>
                <strong>{player[0]}</strong>
                <span>
                  <span aria-hidden="true">☆ </span>
                  {player[1]}
                </span>
                <span>{player[2]}</span>
                <span>{player[3]}</span>
                <span>{player[4]}</span>
                <span>{player[5]}</span>
              </div>
            ))}
          </div>
          <aside className="recommendation" aria-label="Example recommendation">
            <span className="recommendation__rank">#1</span>
            <strong>Malik Corbin</strong>
            <span>WR · NOR</span>
            <span>Projected 298.7 pts</span>
            <h3>Why this pick</h3>
            <ul>
              <li>Top available at WR</li>
              <li>Addresses roster need</li>
              <li>Strong positional value</li>
            </ul>
            <div className="board__actions">
              <span>Draft Malik</span>
              <span>Mark as Gone</span>
            </div>
          </aside>
        </div>
        <div className="board__keys" aria-hidden="true">
          <span>↑↓ Navigate</span>
          <span>
            <kbd>Enter</kbd> View player
          </span>
          <span>
            <kbd>A</kbd> Add to queue
          </span>
          <span>
            <kbd>/</kbd> Search
          </span>
        </div>
      </div>
    </div>
  );
}
