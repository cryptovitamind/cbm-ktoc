# Release Notes

## v0.4.8-beta

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
