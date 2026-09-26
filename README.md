# KTOC

KTOC runs an operator node for a KTv2 staking lottery on Ethereum. Each epoch it
gathers stake events from the contract, weights every staker by the square root
of the minimum balance they held across the epoch, and picks a winner using a
fixed future block hash as the random seed. The seed block is the same for every operator, so
all nodes compute the same winner independently. The node then votes on-chain;
once votes reach the consensus threshold, the winner receives the contract's ETH
balance.

## Building

You need Go 1.23 or newer on your PATH. Nothing else.

### Linux / macOS

```
./build.sh
```

### Windows

```powershell
.\build.ps1
```

Both scripts download dependencies, build both binaries, run the test suite
with coverage, and write a `build.log`. A successful run leaves `ktoc` and
`battles` at the repo root (`.exe` on Windows).

To build a binary alone, skipping the tests:

```
go build -o ktoc ./src/ktp2/cmd
go build -o battles ./src/battles/cmd
```

## Configuration

KTOC reads its settings from a `.env` file in the working directory. Use
`-ktBlock` to find the creation block for `KT_START_BLOCK`.

```
MY_PUBLIC_KEY=
MY_PRIVATE_KEY=
DEAD_ADDR=
TARGET_ADDR=
FACTORY_ADDR=
POOL_ADDR=
TKN_ADDR=
TKN_PRC_ADDR=
KT_ADDR=
QUERY_DELAY=
ETH_ENDPOINT=http://127.0.0.1:8545
KT_START_BLOCK=<creation block of KT_ADDR, from -ktBlock>
```

## Running

Run with no flags to print the full list of commands:

```
./ktoc
```

Run a node in its normal vote-and-reward loop:

```
./ktoc -run
```

A few flags worth knowing beyond the help text:

- `-showVotes` prints the current epoch's reward votes: the tally per candidate
  and which OC voted for which address. Start here when an epoch looks stuck.
- `-voteFor <address>` and `-resetLotteryVote <address>` recover a wedged epoch.
  Reset undoes this node's vote; voteFor forces a vote on an agreed address so
  the operators can converge.
- `-ktProps` now includes the fee reserve: contract balance minus the OC fees
  owed. That difference is the next pot; when it is negative the contract
  cannot pay every operator and winners get nothing until income refills it.
- `-auditRewards all` (or `<startBlock>:<endBlock>`) replays every past reward:
  which OC sent it, the amount it passed to the contract against
  balance-minus-fees at that block, and an OK/OVERPAID verdict. Use it when the
  reserve is underfunded to see which node drained it. Judging old blocks needs
  an archive-capable RPC.
- `-confirmationDepth <n>` sets how many blocks a node waits past the seed block
  before submitting, for reorg safety. It does not change which block seeds the
  lottery, so operators can set it independently.
- `-logDir <dir>` chooses where logs are written (default `logs`). `-zipLogs`
  bundles recent logs into a zip for a bug report, then exits.

## battles: the Vault tool

`battles` is a second executable in this repo, built by the same scripts. It
drives the `BurnBankVault` contract (`src/contracts/BurnBankVault.sol`): an
escrow where a donor locks ETH under the hash of a secret key, and anyone who
learns the key can send that ETH, once, into the `give()` of a whitelisted
Burn Bank. Nobody can withdraw it, the owner included.

Run any command with no `battles.env` present and a guided setup starts. It
asks for an RPC endpoint (probed for its chain ID), the signing key (offered
from a ktoc `.env` if one is beside it, never echoed), and either an existing
Vault address or `deploy` to publish a new one. It then writes `battles.env`
with owner-only permissions. Every value can also be set as an environment
variable of the same name (`BATTLES_ETH_ENDPOINT`, `BATTLES_PRIVATE_KEY`,
`BATTLES_VAULT_ADDR`, `BATTLES_CHAIN_ID`, `BATTLES_VAULT_BLOCK`), and a job
that sets all of them needs no file. `BATTLES_VAULT_BLOCK` is the deploy
block, where event scans start; the tool records it on deploy and finds it by
search otherwise.

```
./battles init                          # guided setup (also runs on first use)
./battles deploy -banks 0xA,0xB         # deploy a Vault whitelisting those Burn Banks
./battles fund -amount 0.01             # lock ETH; prints the SECRET key and the public hash
./battles status -hash 0x..             # funder, amount, funded block, unlockable or spent
./battles verify -tx 0x.. -hash 0x..    # did this transaction fulfill the challenge?
./battles unlock -key 0x.. -bank 0x..   # send the vault into that Burn Bank's give()
./battles challenges                    # every vault on the contract and its outcome
./battles banks | add-bank | remove-bank
```

Challenge flow: fund, post the hash and the funding transaction, keep the key
secret. When someone matches with a `give()` to a whitelisted bank, `verify`
confirms it was successful, at least the vault amount, and mined after the
funding block. Publish the key; anyone can then `unlock`.

Design choices worth knowing: one deposit per key hash (a top-up needs a new
key, so the posted amount is exactly what the key unlocks); plain ETH
transfers to the Vault revert; a lost key means the ETH is stuck for good; and
once a key is public the unlock is a race, so a mempool watcher can pick a
different bank than the matcher intended.

To recompile the contracts and regenerate the Go bindings after editing
anything in `src/contracts`, run `./src/contracts/compile.sh`. It uses a
project-local solc (installed under `src/contracts/tools` on first use) and
the pinned go-ethereum's abigen; nothing global.

## Local testing

To test against a local chain, run geth as a dev node. Full sync mode and an
archive gcmode are required so the eth client can search far enough back in
history; the default dev chain doesn't retain enough. Create the factory
contract and its dependencies first.

```
geth --datadir dev-chain --dev --syncmode=full --gcmode=archive \
  --http --http.api admin,web3,eth,net \
  --ws --ws.api admin,web3,eth,net \
  --http.corsdomain "https://remix.ethereum.org,moz-extension://<your-extension-id>"
```
