# KTOC

This app interacts with an Ethereum KT contract to facilitate a decentralized staking and voting system. It allows nodes to vote and reward users who staked tokens over defined epochs, tracked via block ranges, and calculates minimum stakes to determine eligibility. KTOC connects to an Ethereum node and performs admin operations, voting, and rewarding through ABI bindings. At the epoch’s end, it gathers stakes, assigns probabilities based on minimum contributions, and selects a winner probabilistically using a future block hash as a random seed. The system then votes for the winner and, if enough votes meet the consensus threshold, rewards them with the contract’s ETH balance.

# Getting Started...

## Building 

Run build.ps1 (windows) or build.sh (on unix). 
An executable ktoc.exe should be created.

Create a .env file with keys like MY_PUBLIC_KEY and KT_ADDR. 

## Local .env file

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
ETH_ENDPOINT=http://127.0.0.1:8545
KT_START_BLOCK=<Creation block of KT_ADDR - Use -ktBlock to determine block number>
```

## Building the Application
Make sure you have Go installed and properly configured on your system.

### Windows
To build the KTOC application on Windows:
```powershell
.\build.ps1
```

### Linux/macOS
To build the KTOC application on Linux or macOS:
```
./build.sh
```

## Running the Application
After a successful build, you can run the KTOC application:
1. For help and to see available arguments:
```
./ktoc
```
2. To run with the default arguments:
```
./ktoc -run
```

### Testing 

To test locally, download and run geth as follows.
You will need to create the contract factory and its dependencies first.
Sync mode full is required for certain eth client functions that search history.
The default geth dev chain doesn't go back very far.

```
geth --datadir dev-chain --dev --syncmode=full --gcmode=archive --http --http.api admin,web3,eth,net --ws --ws.api admin,web3,eth,net --http.corsdomain "https://remix.ethereum.org,moz-extension://<getyourextensionidfromthebrowser>"
```