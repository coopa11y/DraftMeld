# First-run setup and league rule imports

DraftMeld shows a non-blocking setup prompt until the user completes or dismisses it. The existing demo draft remains usable while setup is unfinished.

The guided flow covers league basics, scoring, and a final review. **Save and finish later** keeps the current step, league draft, import evidence, and pre-import undo snapshot in versioned browser local storage. **Configure manually** opens the complete league editor using the values already entered. **Skip setup guide** dismisses the guide without creating or changing a league.

After a league is created, the guide recommends configuring ranking sources before opening the draft board. Projection imports are optional.

## League rule import

The league-basics step accepts one PDF or CSV file up to 20 MiB. Uploads stay on the DraftMeld host and are discarded after extraction. DraftMeld returns a preview containing the canonical setting, value, confidence, and detected source text. Recognized values are applied to the unsaved league draft and can be edited or undone before creation. Unrecognized settings never erase existing values.

DraftMeld can recognize league name, team count, redraft or dynasty format, snake, linear, or auction draft format, future-pick seasons, rookie rounds, FAAB and auction budgets, budget-trading options, common roster slots, and supported scoring rules. A zero-count roster setting removes that position from the default roster. User-specific draft position, franchise names, full draft order, keeper spend, and ranking-source preferences remain manual because a general league document cannot reliably identify them.

DraftMeld first reads selectable PDF text. When a PDF contains no selectable text, it can recognize up to 25 scanned pages with local English OCR. OCR imports carry a prominent review warning because recognition and provider wording can be imperfect. See [local PDF OCR](ocr.md) for installation, limits, and privacy behavior.

CSV supports a Setting/Value list:

```csv
Setting,Value
League name,Saturday League
Number of teams,10
League format,Dynasty
Draft format,Snake
QB,1
RB,2
WR,3
K,0
Bench,7
```

Scoring rules can use a Statistic/Points list:

```csv
Statistic,Points,Per
Passing yards,1,25
Passing touchdowns,4,1
Interceptions thrown,-2,1
```

```csv
Passing TD,Reception,Field goals made 50+ yards
4,1,5
```

The optional `Per` column converts units such as one point per 25 passing yards into DraftMeld's points-per-yard value. Zero-valued rules remain valid and disable that scoring category.

A two-row wide CSV is also supported. Each first-row header can be a league setting, roster slot, or scoring rule, with its value in the second row.

## Accessibility behavior

- Step changes move focus to the setup heading.
- Progress is an ordered list with `aria-current="step"`.
- Native form labels, fieldsets, summaries, status messages, and tables expose structure without relying on color or pointer input.
- Users can leave, resume, skip, or switch to the full manual editor at any point.
