# KTOC

KTOC runs an operator node for a KTv2 staking lottery on Ethereum. Each epoch it
gathers stake events from the contract, weights every staker by the minimum
balance they held across the epoch, and picks a winner using a fixed future
block hash as the random seed. The seed block is the same for every operator, so
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

Both scripts download dependencies, build the binary, run the test suite with
coverage, and write a `build.log`. A successful run leaves the executable at the
repo root: `ktoc` on Linux/macOS, `ktoc.exe` on Windows.

To build the binary alone, skipping the tests:

```
go build -o ktoc ./src/ktp2/cmd
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
- `-confirmationDepth <n>` sets how many blocks a node waits past the seed block
  before submitting, for reorg safety. It does not change which block seeds the
  lottery, so operators can set it independently.
- `-logDir <dir>` chooses where logs are written (default `logs`). `-zipLogs`
  bundles recent logs into a zip for a bug report, then exits.

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
