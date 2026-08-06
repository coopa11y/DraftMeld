# Accessibility requirements

DraftMeld treats screen-reader and keyboard access as release requirements, not optional polish.

## Draft-board requirements

- A logical heading and landmark structure identifies the board, recommendations, team, and history.
- Skip links provide direct access to the board, recommendations, and team.
- Player rankings use a semantic table with a caption, column headers, and player row headers.
- Every Draft and Taken control includes the player name and position in its accessible name.
- Draft, Taken, and Undo results are announced through a polite live region.
- When an action removes the focused player, focus moves to the next available player action.
- Undo restores the player and returns focus to that player's Draft control.
- Status and value are never communicated by color alone.
- Controls have visible focus indicators and at least a 44-pixel activation target.
- The interface supports browser zoom, narrow viewports, reduced motion, and forced-colors mode.

## League-setup requirements

- Every league, roster, and scoring input has a persistent visible label.
- Related roster-position controls are grouped with fieldsets and legends.
- Destructive deletion requires a second, clearly labeled confirmation action.
- Save, duplicate, and delete results are announced without moving focus unexpectedly.
- Default presets reduce required input while every stored rule remains editable.

## Verification

Pull requests affecting the draft interface must include keyboard interaction tests. Automated axe scans run with the frontend test suite. Manual checks with NVDA on Windows and VoiceOver on a supported Apple platform should be completed before a stable `1.0.0` release.

## Known limitations

Automated accessibility tests cannot prove screen-reader usability. The current fictional player catalog is development data, and the interface has not yet completed formal usability testing with screen-reader users.
