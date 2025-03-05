package main

import (
	"context"
	"flag"
	"fmt"
	"ktp2/src/abis"
	"ktp2/src/ktp2/ktfunc"
	"ktp2/src/ktp2/tests"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Warnf("Error loading .env file: %v", err)
	}
	log.SetOutput(colorable.NewColorableStdout()) // Ensure colors work on all platforms
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&ktfunc.CustomFormatter{})
}

func LogOperationStart(operation string) {
	timestamp := time.Now().Format("15:04:05")
	log.Info("")
	log.Infof("---- [%s] %s ----", timestamp, strings.ToUpper(operation))
}

type Flags struct {
	continuous       bool
	giveAmount       float64
	stakeAmount      int64
	epochDuration    int64
	moveBlockForward int64
	keys             bool
	initTestWallets  bool
	vote             bool
	run              bool
	findKts          bool
	createKt         bool
	gasLimit         uint64
	blocksToWait     uint64
	ktBlock          bool
	help             bool
	verbose          bool
	ktProps          bool
	queryFees        string
	withdrawFees     string
	currentBlock     bool
	waitDuration     time.Duration
}

func main() {
	mProps := loadMasterProperties()
	displayStartupBanner()
	flags := parseFlags()
	cProps := setupConnectionProps(&mProps, flags)

	if flags.continuous {
		LogOperationStart("Continuous operations")
		testContinuousOperations(cProps)
		return
	}

	handleSingleOperations(cProps, flags)
}

// loadMasterProperties extracts and verifies master properties from environment variables
func loadMasterProperties() ktfunc.MasterProps {
	mProps := ktfunc.MasterProps{
		MyPublicKey:  os.Getenv("MY_PUBLIC_KEY"),
		MyPrivateKey: os.Getenv("MY_PRIVATE_KEY"),
		DeadAddr:     os.Getenv("DEAD_ADDR"),
		TargetAddr:   os.Getenv("TARGET_ADDR"),
		FactoryAddr:  os.Getenv("FACTORY_ADDR"),
		PoolAddr:     os.Getenv("POOL_ADDR"),
		TknAddr:      os.Getenv("TKN_ADDR"),
		TknPrcAddr:   os.Getenv("TKN_PRC_ADDR"),
		KtAddr:       os.Getenv("KT_ADDR"),
		KtStartBlock: os.Getenv("KT_START_BLOCK"),
		EthEndpoint:  os.Getenv("ETH_ENDPOINT"),
		WaitDuration: os.Getenv("WAIT_DURATION"),
	}

	// Verify required properties
	if mProps.EthEndpoint == "" {
		log.Fatal("Required environment variable ETH_ENDPOINT is missing. Please set it in your .env file or environment.")
	}
	if mProps.MyPublicKey == "" {
		log.Fatal("Required environment variable MY_PUBLIC_KEY is missing. Please set it in your .env file or environment.")
	}
	if mProps.MyPrivateKey == "" {
		log.Fatal("Required environment variable MY_PRIVATE_KEY is missing. Please set it in your .env file or environment.")
	}

	return mProps
}

func displayStartupBanner() {
	fmt.Print("\n")
	fmt.Print("──────────────────────────────────────────────────────────────────────────────────────────\n")
	fmt.Print("\n")
	figure.NewColorFigure("KTOC", "larry3d", "red", true).Print()
	figure.NewColorFigure("v0.2-beta", "larry3d", "red", true).Print()
	fmt.Print("\n")
	figure.NewColorFigure("shinatoken", "binary", "red", true).Print()
	fmt.Print("\n")
	fmt.Print("──────────────────────────────────────────────────────────────────────────────────────────\n")
	disclaimer := `
    DISCLAIMER ⚖️
      This is experimental software provided "as is" without any warranties, 
      express or implied, including but not limited to fitness for a particular purpose 
      or merchantability. The developers or contributors are not responsible for 
      any financial losses, damages, bugs, errors, or issues arising from its use. 
      This software is in beta, and unexpected behavior may occur. No technical support 
      or maintenance will be provided. Use at your own risk. By running this program, 
      you acknowledge and accept these terms.
    `
	fmt.Println(disclaimer)
	fmt.Print("Press Enter to continue... ")
	// reader := bufio.NewReader(os.Stdin)
	// _, _ = reader.ReadString('\n') // Wait for Enter key
}

// parseFlags defines and parses command-line flags for the KT application.
// It returns a Flags struct populated with the parsed values.
// Flags are categorized into General Commands (core functionality) and Testing Commands (for development).
func parseFlags() Flags {
	// Define help flag explicitly
	help := flag.Bool("h", false, "Display this help message and exit")
	flag.BoolVar(help, "help", false, "Display this help message and exit") // Supports --help too

	// General Commands (non-testing flags)
	run := flag.Bool("run", false, "Run normal operations in a continuous loop. Monitors the KT contract state until interrupted with CTRL+C. Ideal for ongoing contract management.")
	vote := flag.Bool("vote", false, "Perform a single voting operation on the KT contract. Executes once and exits. Use this to cast a vote + reward without continuous monitoring.")
	findKts := flag.Bool("findKts", false, "List all KT contracts deployed by the factory contract specified by FACTORY_ADDR. Useful for auditing or exploring existing KTs.")
	ktBlock := flag.Bool("ktBlock", false, "Print the block number where the KT contract (specified in KT_ADDR) was created. Helpful for debugging or setting KT_START_BLOCK.")
	createKt := flag.Bool("createKt", false, "Deploy a new KT contract via the factory contract. Requires FACTORY_ADDR and sufficient ETH/gas. Prompts for confirmation.")
	ktProps := flag.Bool("ktProps", false, "Dislpay information about an existing KT contract (specified in KT_ADDR). Useful for debugging or testing KT behavior.")
	gasLimit := flag.Uint64("gasLimit", ktfunc.DefaultGasLimit, fmt.Sprintf("Set the gas limit for transactions (default is %d. Use 3000000 for contract creation).", ktfunc.DefaultGasLimit))
	blocksToWait := flag.Uint64("blocksToWait", ktfunc.DefaultBlocksToWait, fmt.Sprintf("Set the number of blocks to wait for transactions to be mined (default is %d).", ktfunc.DefaultBlocksToWait))
	verbose := flag.Bool("verbose", false, "Display verbose output during operations.")
	queryFees := flag.String("queryFees", "", "Query the current gas fees for a specified block range with syntax <startBlock>:<endBlock>")
	withdrawFees := flag.String("withdrawFees", "", "Withdraw owed fees from kt. syntax: <block1>,<block2>,<blockn>,etc...")
	currentBlock := flag.Bool("currentBlock", false, "Print the current Ethereum block number. Useful for fetching fees owed.")
	waitDuration := flag.Duration("waitDuration", ktfunc.TimeToWaitForBlocks, "Set the duration to wait between operations (ex: 1s, 2m). Default is 1 minute.")

	// Testing Commands (for development and testing)
	continuous := flag.Bool("continuous", false, "TESTING: Run continuous operations in a loop, simulating various actions (e.g., staking, giving ETH). For development use only.")
	giveAmount := flag.Float64("give", 0, "TESTING: Amount of ETH to distribute to test wallets (in ETH, e.g., 0.1). Requires -init or existing wallets.")
	stakeAmount := flag.Int64("stake", 0, "TESTING: Number of tokens to stake from test wallets (in wei, e.g., 1000). Requires -init or configured wallets. Simulates staking behavior.")
	epochDuration := flag.Int64("epochDuration", 0, "TESTING: Set the epoch duration in blocks (e.g., 3600).")
	moveBlockForward := flag.Int64("moveBlockForward", 0, "TESTING: Advance a test Ethereum node by n blocks (e.g., 10). Simulates blockchain progression.")
	keys := flag.Bool("keys", false, "TESTING: Display deterministic private keys for test wallets. For development only—do not use these keys on mainnet!")
	initTestWallets := flag.Bool("init", false, "TESTING: Initialize test wallets with ETH and tokens. Sets up a testing environment. Requires sufficient funds in MY_PUBLIC_KEY.")

	// Custom usage message with guides
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n🌟 Welcome to KT v0.1 (beta) - CLI 🌟\n")
		fmt.Fprintf(os.Stderr, "========================================================\n")
		fmt.Fprintf(os.Stderr, "KT is an experimental blockchain tool. Use these flags to interact with KTv2 contracts.\n")
		fmt.Fprintf(os.Stderr, "Set required environment variables (e.g., MY_PUBLIC_KEY, ETH_ENDPOINT) in a .env file.\n")
		fmt.Fprintf(os.Stderr, "Run with -h or --help to see this guide.\n\n")

		fmt.Fprintf(os.Stderr, "📋 General Commands:\n")
		fmt.Fprintf(os.Stderr, "  -run                %s\n",
			`Run normal vote and reward operations in a continuous loop. Monitors the KT contract state and will perform voting and rewarding automatically as needed. Interrupt with CTRL+C.`)
		fmt.Fprintf(os.Stderr, "  -vote               %s\n", "Perform a single voting operation on the KT contract. Executes once and exits.")
		fmt.Fprintf(os.Stderr, "  -findKts            %s\n", "List all KT contracts deployed by the factory listed in your .env file.")
		fmt.Fprintf(os.Stderr, "  -ktBlock            %s\n", "Print the block number where the KT contract was created. Use this to set KT_START_BLOCK in your .env file.")
		fmt.Fprintf(os.Stderr, "  -createKt           %s\n", "Deploy a new KT contract via the factory contract.")
		fmt.Fprintf(os.Stderr, "  -ktProps            %s\n", "Display information about an existing KT contract.")
		fmt.Fprintf(os.Stderr, "  -gasLimit <limit>   %s\n", fmt.Sprintf("Set gas limit for transactions (default: %d).", ktfunc.DefaultGasLimit))
		fmt.Fprintf(os.Stderr, "  -blocksToWait <n>   %s\n", fmt.Sprintf("Set number of blocks to wait for transactions to be mined (default: %d).", ktfunc.DefaultBlocksToWait))
		fmt.Fprintf(os.Stderr, "  -epochDuration <n>  %s\n", "Set epoch duration in blocks (e.g., -epochDuration 3600). Run with -ktProps to see current value.")
		fmt.Fprintf(os.Stderr, "  -verbose            %s\n", "Display verbose output during operations.")
		fmt.Fprintf(os.Stderr, "  -queryFees <n>:<n>  %s\n", "Query the reward amount owed this node. <startBlock>:<endBlock>.")
		fmt.Fprintf(os.Stderr, "  -withdrawFees <blocks> %s\n", "Withdraw owed fees from kt. <blocks> is a comma-separated list of block numbers.")
		fmt.Fprintf(os.Stderr, "  -currentBlock       %s\n", "Print the current Ethereum block number. Useful for fetching fees owed.")
		fmt.Fprintf(os.Stderr, "  -waitDuration <duration> %s\n", "Set the duration to wait between operations (e.g., 1s, 2m).")

		fmt.Fprintf(os.Stderr, "\n🛠️ Testing Commands (Local Dev Use Only):\n")
		fmt.Fprintf(os.Stderr, "  -continuous         %s\n", "Run continuous operations in a loop for testing.")
		fmt.Fprintf(os.Stderr, "  -give <amount>      %s\n", "Amount of ETH to give to test wallets (e.g., -give 0.1).")
		fmt.Fprintf(os.Stderr, "  -stake <amount>     %s\n", "Number of tokens to stake (e.g., -stake 1000).")
		fmt.Fprintf(os.Stderr, "  -moveBlockForward <n> %s\n", "Advance the test node by n blocks (e.g., -moveBlockForward 10).")
		fmt.Fprintf(os.Stderr, "  -keys               %s\n", "Display deterministic private keys for testing.")
		fmt.Fprintf(os.Stderr, "  -init               %s\n", "Initialize test wallets with ETH and tokens.")

		fmt.Fprintf(os.Stderr, "\n💡 Usage Tips:\n")
		fmt.Fprintf(os.Stderr, "  - Use -gasLimit for operations like -createKt (e.g., 3000000).\n")
		fmt.Fprintf(os.Stderr, "  - Interrupt continuous operations with CTRL+C.\n")
		fmt.Fprintf(os.Stderr, "========================================================\n")
	}

	flag.Parse()

	return Flags{
		help:             *help,
		continuous:       *continuous,
		giveAmount:       *giveAmount,
		stakeAmount:      *stakeAmount,
		epochDuration:    *epochDuration,
		moveBlockForward: *moveBlockForward,
		keys:             *keys,
		initTestWallets:  *initTestWallets,
		vote:             *vote,
		run:              *run,
		findKts:          *findKts,
		createKt:         *createKt,
		gasLimit:         *gasLimit,
		ktBlock:          *ktBlock,
		ktProps:          *ktProps,
		verbose:          *verbose,
		queryFees:        *queryFees,
		withdrawFees:     *withdrawFees,
		currentBlock:     *currentBlock,
		blocksToWait:     *blocksToWait,
		waitDuration:     *waitDuration,
	}
}

func handleSingleOperations(cProps *ktfunc.ConnectionProps, flags Flags) {
	if flags.help || flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(0)
	}

	if flags.verbose {
		log.SetLevel(log.DebugLevel)
	}

	if flags.currentBlock {
		currentBlock, err := ktfunc.GetCurrentBlock(cProps)
		if err != nil {
			log.Warnf("Error getting current block: %v", err)
		}
		log.Infof("Current Ethereum block number: %d", currentBlock.NumberU64())
	}

	if flags.initTestWallets {
		LogOperationStart("Initializing test wallets")
		initTestWallets(cProps)
	}

	if len(flags.queryFees) > 0 {
		LogOperationStart("Querying fees")
		ktfunc.GetOCFeesOwed(cProps, flags.queryFees)
	}

	if len(flags.withdrawFees) > 0 {
		ktfunc.WithdrawOCFees(cProps, flags.withdrawFees)
	}

	if flags.moveBlockForward != 0 {
		LogOperationStart("Moving blocks forward")

		if flags.moveBlockForward > 0 {
			log.Infof("Target blocks: %d", flags.moveBlockForward)
			err := tests.MoveBlocksForward(cProps, &flags.moveBlockForward, flags.gasLimit)
			if err != nil {
				log.Errorf("Error encountered: %v", err)
			} else {
				log.Info("Blocks moved successfully")
			}
		} else {
			log.Infof("Target blocks: %d", -flags.moveBlockForward)
			tests.KeepMovingBlocks(cProps, flags.gasLimit)
		}
	}

	if flags.ktBlock {
		LogOperationStart("Printing KT contract start block")
		ktfunc.GetContractCreationBlock(cProps)
	}

	if flags.stakeAmount > 0 {
		LogOperationStart("Staking tokens")
		log.Infof("Amount to stake: %d tokens", flags.stakeAmount)
		stakeTokens(cProps, flags.stakeAmount)
	}

	if flags.giveAmount > 0 {
		LogOperationStart("Giving ETH")
		log.Infof("Amount to give: %.4f ETH", flags.giveAmount)
		giveETH(cProps, flags.giveAmount)
	}

	if flags.epochDuration > 0 {
		LogOperationStart("Adjusting epoch duration")
		log.Infof("New duration: %d seconds", flags.epochDuration)
		ktfunc.AdjustEpochDuration(cProps, &flags.epochDuration)
		ktfunc.PrintKtContractVariables(cProps)
	}

	if flags.keys {
		LogOperationStart("Printing deterministic private keys - FOR TESTING ONLY!")
		printDeterministicKeys()
	}

	if flags.vote {
		LogOperationStart("Finding receiver for voting")
		ktfunc.VoteAndReward(cProps)
	}

	if flags.run {
		LogOperationStart("Starting normal operations... Press CTRL+C to stop")
		ktfunc.PrintKtContractVariables(cProps)
		KeepRunning(cProps)
	}

	if flags.findKts {
		LogOperationStart("Listing all KTs in factory contract")
		ktfunc.PrintKtFactContracts(cProps)
	}

	if flags.createKt {
		LogOperationStart("Creating a new KT")
		ktfunc.CreateKtFromFact(cProps)
	}

	if flags.ktProps {
		LogOperationStart("Displaying KT contract properties")
		ktfunc.PrintKtContractVariables(cProps)
	}
}

// initTestWallets initializes test wallets by sending ETH and tokens.
func initTestWallets(cProps *ktfunc.ConnectionProps) {
	keyPairs := tests.DeterministicPrivateKeys(10)
	for _, kp := range keyPairs {
		log.Printf("Initializing wallet: %s", kp.Address)
		tests.SendSomeEth(cProps, kp.PrivateKey, 0.5)
		tests.SendSomeTestTokens(cProps, kp.PrivateKey, 100)
		tests.PrintBalances(cProps, ktfunc.GetPublicAddress(kp.PrivateKey))
	}
}

// stakeTokens stakes the specified amount of tokens for each test wallet.
func stakeTokens(cProps *ktfunc.ConnectionProps, stakeAmount int64) {
	keyPairs := tests.DeterministicPrivateKeys(10)
	for _, kp := range keyPairs {
		log.Printf("Staking for address: %s", kp.Address)
		tests.StakeTokensToKt(cProps, kp.PrivateKey, big.NewInt(stakeAmount))
	}
}

// giveETH gives the specified amount of ETH from each test wallet.
func giveETH(cProps *ktfunc.ConnectionProps, giveAmount float64) {
	giveAmountWei := big.NewInt(int64(giveAmount * 1e18))
	keyPairs := tests.DeterministicPrivateKeys(10)
	for _, kp := range keyPairs {
		log.Printf("Giving from address: %s", kp.Address)
		ktfunc.Give(cProps, kp.PrivateKey, giveAmountWei)
	}
}

// printDeterministicKeys prints deterministic private keys for testing.
func printDeterministicKeys() {
	fmt.Println("Deterministic Private Keys (please don't use these in the real world):")
	keyPairs := tests.DeterministicPrivateKeys(10)
	for i, kp := range keyPairs {
		fmt.Printf("Key Pair %d:\n", i+1)
		fmt.Printf("  Private Key: %s\n", kp.PrivateKey)
		fmt.Printf("  Public Addr: %s\n\n", kp.Address)
	}
}

// continuousOperations runs continuous operations for testing.
func testContinuousOperations(cProps *ktfunc.ConnectionProps) {
	keyPairs := tests.DeterministicPrivateKeys(10)

	for {
		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
		LogOperationStart("Continuous operation iteration")
		moveBlockForward := rand.Int63n(81) + 2
		stakeAmount := rand.Int63n(9001) + 1000
		giveAmount := rand.Float64()*0.1 + 0.04

		log.Infof("Moving blocks: %d", moveBlockForward)
		log.Infof("Staking: %d tokens", stakeAmount)
		log.Infof("Giving: %.4f ETH", giveAmount)

		tests.MoveBlocksForward(cProps, &moveBlockForward, cProps.GasLimit)

		for _, kp := range keyPairs {
			log.Infof("Staking for: %s", kp.Address)
			tests.StakeTokensToKt(cProps, kp.PrivateKey, big.NewInt(stakeAmount))
		}

		giveAmountWei := big.NewInt(int64(giveAmount * 1e18))
		for _, kp := range keyPairs {
			ethBalance, err := cProps.Client.BalanceAt(context.Background(), *kp.Address, nil)
			if err != nil {
				log.Errorf("Balance check failed for: %s - %v", kp.Address, err)
				continue
			}

			float64Eth, _ := ethBalance.Float64()
			ethBalancePretty := fmt.Sprintf("%.4f ETH", float64Eth/1e18)

			if ethBalance.Cmp(giveAmountWei) > 0 {
				log.Infof("Balance: %s", ethBalancePretty)
				log.Infof("Giving from: %s - %.4f ETH", kp.Address, giveAmount)
				ktfunc.Give(cProps, kp.PrivateKey, giveAmountWei)
			} else {
				log.Warnf("Insufficient balance for: %s - %s", kp.Address, ethBalancePretty)
			}
		}

		log.Printf("Iteration complete: Sleeping for %d seconds", int(cProps.WaitDuration.Seconds()))
		time.Sleep(time.Duration(cProps.WaitDuration))
	}
}

// setupConnectionProps initializes Ethereum connection properties.
func setupConnectionProps(mstProps *ktfunc.MasterProps, flags Flags) *ktfunc.ConnectionProps {
	fmt.Println("")
	log.Println("Setting up connection properties...")
	cProps := &ktfunc.ConnectionProps{}

	// Use the defined gas limit.
	cProps.GasLimit = flags.gasLimit
	cProps.BlocksToWait = flags.blocksToWait

	duration := ktfunc.TimeToWaitForBlocks
	if mstProps.WaitDuration != "" {
		if parsed, err := time.ParseDuration(mstProps.WaitDuration); err != nil {
			log.Warnf("Invalid wait duration '%s': %v. Using default: %v",
				mstProps.WaitDuration, err, ktfunc.TimeToWaitForBlocks)
		} else {
			duration = parsed
		}
	}

	// Override with flags if different from default
	if flags.waitDuration != ktfunc.TimeToWaitForBlocks {
		log.Infof("Overriding wait duration with command line flag: %v", flags.waitDuration)
		duration = flags.waitDuration
	}

	cProps.WaitDuration = duration

	// Connect to Ethereum node.
	client, err := ethclient.Dial(mstProps.EthEndpoint)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	cProps.Client = client
	cProps.Backend = client
	log.Println("Connected to Ethereum node successfully")

	// Set public key.
	cProps.MyPubKey = ktfunc.ToAddr(mstProps.MyPublicKey)
	ktfunc.PrintBalanceOfAddr(cProps)

	privateKey, err := crypto.HexToECDSA(mstProps.MyPrivateKey)
	if err != nil {
		log.Fatalf("Invalid private key: %v", err)
	}
	cProps.MyPrivateKey = privateKey

	// Get chain ID.
	cProps.ChainID, err = ktfunc.GetChainId(client)
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}
	cProps.Addresses = mstProps

	// Initialize KT contract instance if address is provided.
	if mstProps.KtAddr != "" {
		cProps.KtAddr = ktfunc.ToAddr(mstProps.KtAddr)
		cProps.Kt, err = getKtInstance(cProps.Backend, cProps.KtAddr)
		if err != nil {
			log.Fatalf("Failed to initialize KT contract: %v", err)
		}
	}

	if mstProps.KtStartBlock == "" {
		startBlock, err := ktfunc.GetContractCreationBlock(cProps)
		if err != nil {
			log.Warnf("Failed to get KT contract creation block. Is KT_ADDR defined in your .env? - %v", err)
		}

		if startBlock > 0 {
			log.Printf("KT contract created at block: %d", startBlock)
			log.Printf("Place this in your .env file: KT_START_BLOCK=%d", startBlock)
			log.Printf("If you change your kt contract, you'll need to update this in your .env file again.")
		}
	} else {
		block, err := strconv.ParseInt(mstProps.KtStartBlock, 10, 64)
		if err != nil {
			log.Fatalf("Invalid KT_START_BLOCK: %v", err)
		}
		cProps.KtBlock = big.NewInt(block)
	}

	return cProps
}

// getKtInstance initializes and returns a KT contract instance for the given address.
// Returns the instance or nil if instantiation fails.
func getKtInstance(client bind.ContractBackend, ktAddr common.Address) (*abis.Ktv2, error) {
	// Validate inputs
	if client == nil {
		log.Errorf("Invalid ConnectionProps - Client: %v", client)
		return nil, fmt.Errorf("invalid ConnectionProps: client is nil")
	}

	// Instantiate the KT contract
	instance, err := abis.NewKtv2(ktAddr, client)
	if err != nil {
		log.Errorf("Failed to instantiate KT contract at %s: %v", ktAddr.Hex(), err)
		return nil, fmt.Errorf("failed to instantiate KT contract: %w", err)
	}

	log.Debugf("KT instance created for contract: %s", ktAddr.Hex())
	return instance, nil
}

func KeepRunning(cProps *ktfunc.ConnectionProps) {
	for {
		err := ktfunc.VoteAndReward(cProps)
		if err != nil {
			log.Printf("Error in VoteAndReward: %v", err)
		}

		log.Printf("Iteration complete. Sleeping for %d seconds", int(cProps.WaitDuration.Seconds()))
		time.Sleep(time.Duration(cProps.WaitDuration))
	}
}
