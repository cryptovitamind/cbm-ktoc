# Release Notes

## v0.5.0-beta

### Changed: winner weighting is now square root

The lottery now weights every staker by the square root of the minimum balance
they held across the epoch. Before this release it used the logarithm, and the
community noticed what the math confirms: log made the odds nearly flat. A
wallet holding a billion times more tokens only won 1.5x more often, and
splitting one stake across many wallets multiplied its total odds by hundreds.

Square root keeps a real edge for larger stakes without handing everything to
the largest one. A staker with 4x the tokens now gets 2x the weight. A billion
times the tokens gets about 31,623x the weight, not 1.5x (log) and not a
billion (linear). Splitting a stake across N wallets still helps, but only by
sqrt(N) instead of the runaway gain log allowed.

The new weighting also computes probabilities with 128-bit arithmetic in a
fixed order, so every operator derives bit-identical odds from the same chain
state. Stakes too close together for the old floating-point math to tell apart
now weigh in correctly.

### Upgrade notes: coordinate this one

**Every operator must upgrade before the next epoch ends.** A node on this
release and a node on an older release will pick different winners from the
same chain state and vote against each other. Funds stay safe either way,
because a reward needs the consensus threshold, but a split fleet can leave an
epoch without a rewarded winner until enough nodes agree. If that happens,
upgrade the stragglers, then use `-showVotes` to see the tallies and `-voteFor`
to converge on a winner.

### New: pick the curve when verifying old epochs

`-verifyLastWinner` replays with square-root weighting by default, which means
it will report a mismatch for epochs rewarded before this release. That
mismatch is expected: those winners were selected under log. To check one of
those epochs, re-run with the old curve:

- Flag: `-verifyWeighting log`
- Env var: `VERIFY_WEIGHTING=log`

The flag wins over the env var. This setting only affects the verification
replay. Live voting has no weighting switch, deliberately, so no operator can
run a different curve from the rest of the fleet.



### Fixed: node could hang after "Winner selected"

Some operators saw their node stop making progress. The last log line was
`Winner selected: 0x...`, and a restart cleared it every time. The node was not
crashing. It was stuck.

The cause was the wait for a transaction to be mined. After picking a winner the
node submits a vote, then waits for that transaction's receipt. The wait had no
time limit, so if the transaction never landed (dropped from the mempool,
underpriced, or stuck behind an earlier pending transaction) the node waited
forever and never moved on to the next cycle. A restart fixed it because the
node re-read chain state and resubmitted with a fresh nonce.

Every transaction wait in the node is now bounded. If a transaction does not
mine within the timeout, the node logs the reason, returns, and retries on the
next cycle instead of parking. This covers the vote and reward steps of the main
loop as well as the manual operator commands (add or remove an OC, reset a vote,
give, withdraw fees, set the OC fee).

### New: configurable mining timeout

Two ways to set how long the node waits for a transaction before giving up and
retrying:

- Flag: `-txMineTimeout 2m`
- Env var: `TX_MINE_TIMEOUT=2m`

The flag wins over the env var, which wins over the default of 5 minutes. Five
minutes is about 25 blocks on a 12-second chain, well beyond what a properly
priced transaction needs, while still freeing a stuck node promptly. Most
operators do not need to change it.

### Upgrade notes

Recommended for every operator. The node now recovers on its own from a stuck
transaction, so the "restart to unstick it" workaround is no longer needed. No
configuration changes are required to get the fix; the bounded wait is on by
default.
