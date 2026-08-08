# Draft and dynasty lifecycle

DraftMeld treats a franchise as permanent and its draft position as season-specific. Changing the draft order swaps franchise assignments without renaming teams or changing which franchise belongs to the user.

## Trade rules

- Redraft leagues can trade any unused pick in the active draft.
- Dynasty leagues can also trade rostered players and configured future rookie picks.
- A future pick can carry a plain-language condition. Pending conditional picks are locked until the condition is marked met or not met.
- Auction dollars and FAAB can be included only when that asset is enabled in league settings. Every budget asset identifies its season, and the backend rejects packages that exceed the sending franchise's remaining balance.
- Every accepted trade is recorded in the ledger. The ledger can be filtered by trade season, and a trade can be reversed only while doing so would not invalidate a later transaction.

## Season rollover

Dynasty leagues use the guided rollover in the draft workspace. The review step selects the next season's draft type and assigns every permanent franchise to a draft slot. Closing an incomplete draft requires an explicit acknowledgment.

Rollover advances exactly one year, clears the active draft history, carries rosters and trade ownership forward, and makes that season's rookie picks and budget balances current. Directly changing a dynasty season in league settings is intentionally blocked so those transitions cannot be skipped accidentally.

Redraft leagues do not show future picks, persistent player trades, future budgets, or the rollover workflow.
