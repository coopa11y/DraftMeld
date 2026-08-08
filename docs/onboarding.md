# First-run setup and league rule imports

DraftMeld shows a non-blocking setup prompt until the user completes or dismisses it. The existing demo draft remains usable while setup is unfinished.

The guided flow covers league basics, scoring, and a final review. **Save and finish later** keeps the current step and league draft in versioned browser local storage. **Configure manually** opens the complete league editor using the values already entered. **Skip setup guide** dismisses the guide without creating or changing a league.

After a league is created, the guide recommends configuring ranking sources before opening the draft board. Projection imports are optional.

## Scoring rule import

The scoring step accepts one PDF or CSV file up to 20 MiB. Uploads are processed in memory and discarded after extraction. DraftMeld returns a preview containing the canonical rule, value, confidence, and detected source text. Recognized values are applied to the unsaved league draft and can be edited or undone before creation. Unrecognized settings never erase existing values.

PDF files must contain selectable text. Scanned documents need OCR before upload. Because provider layouts and wording vary, every imported value must be compared with the original league settings.

CSV supports either of these common shapes:

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

## Accessibility behavior

- Step changes move focus to the setup heading.
- Progress is an ordered list with `aria-current="step"`.
- Native form labels, fieldsets, summaries, status messages, and tables expose structure without relying on color or pointer input.
- Users can leave, resume, skip, or switch to the full manual editor at any point.
