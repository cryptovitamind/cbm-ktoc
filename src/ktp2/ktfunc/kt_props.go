package ktfunc

import (
	"bufio"
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"

	"ktp2/src/abis" // Your package path
)

// printCurrentEpochInterval logs the current epoch interval details for a KT contract.
func printCurrentEpochInterval(cProps *ConnectionProps, kt *abis.Ktv2) error {
	callOpts := &bind.CallOpts{
		Context: context.Background(),
		Pending: false,
		From:    cProps.MyPubKey,
	}

	currentInterval, err := kt.EpochInterval(callOpts)
	if err != nil {
		return fmt.Errorf("failed to get current epoch interval: %w", err)
	}

	startBlock, err := kt.StartBlock(callOpts)
	if err != nil {
		return fmt.Errorf("failed to get start block: %w", err)
	}

	currentBlock, err := cProps.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	targetBlock := new(big.Int).Add(startBlock, big.NewInt(int64(currentInterval)))

	log.Infof("Current epoch interval: %d blocks", currentInterval)
	log.Infof("Start block: %s", startBlock.String())
	log.Infof("Current block: %d", currentBlock)
	log.Infof("Target block: %s", targetBlock.String())

	return nil
}

// AdjustEpochDuration updates the epoch duration for a KT contract.
func AdjustEpochDuration(cProps *ConnectionProps, newEpochDuration *int64) (common.Address, error) {
	LogOperationStart("Adjusting epoch duration")

	if cProps.Kt == nil {
		return common.Address{}, fmt.Errorf("KT contract instance not initialized")
	}

	newInterval := uint16(*newEpochDuration)
	log.Infof("New epoch duration: %d blocks", newInterval)

	// Print current epoch details before adjustment
	if err := printCurrentEpochInterval(cProps, cProps.Kt); err != nil {
		log.Errorf("Failed to print current epoch: %v", err)
		return common.Address{}, err
	}

	// Prepare transaction options
	auth, err := NewTransactor(cProps)
	if err != nil {
		return common.Address{}, err
	}

	// Set the new epoch interval
	tx, err := cProps.Kt.SetEpochInterval(auth, newInterval)
	if err != nil {
		log.Errorf("Set epoch interval failed: %v", err)
		return common.Address{}, fmt.Errorf("failed to set epoch interval: %w", err)
	}

	log.Infof("Transaction sent: %s", tx.Hash().Hex())

	// Wait for transaction confirmation
	receipt, err := bind.WaitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		log.Errorf("Transaction mining failed: %v", err)
		return common.Address{}, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Info("Transaction successful")
		updatedInterval, err := cProps.Kt.EpochInterval(&bind.CallOpts{Context: context.Background()})
		if err != nil {
			log.Errorf("Failed to get updated interval: %v", err)
			return common.Address{}, fmt.Errorf("failed to get updated interval: %w", err)
		}
		log.Infof("Updated epoch interval: %d blocks", updatedInterval)
	} else {
		log.Warn("Transaction failed: Check receipt for details")
	}

	return receipt.ContractAddress, nil
}

// CreateKtFromFact creates a new KT instance using a factory contract.
// It submits a transaction to the Ethereum blockchain and waits for confirmation.
// The gas limit must be set in cProps.GasLimit, typically via the -gasLimit flag.
func CreateKtFromFact(cProps *ConnectionProps) (common.Address, error) {
	LogOperationStart("Creating a new KT")

	// Convert factory address from string to Ethereum address type
	factoryAddr := ToAddr(cProps.Addresses.FactoryAddr)
	log.Infof("Using factory contract at address: %s", factoryAddr.Hex())

	// Instantiate the factory contract with the Ethereum client
	instance, err := abis.NewKtv2fact(factoryAddr, cProps.Backend)
	if err != nil {
		log.Errorf("Failed to instantiate factory contract: %v", err)
		return common.Address{}, fmt.Errorf("failed to instantiate factory: %w", err)
	}

	// Fetch the suggested gas price from the network
	gasPrice, err := cProps.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Errorf("Failed to retrieve suggested gas price: %v", err)
		return common.Address{}, fmt.Errorf("failed to get gas price: %w", err)
	}
	log.Infof("Suggested gas price: %s wei", gasPrice.String())

	// Set up transaction options with the sender's private key and chain ID
	auth, err := NewTransactor(cProps)
	if err != nil {
		log.Errorf("Failed to create transactor: %v", err)
		return common.Address{}, fmt.Errorf("failed to create transactor: %w", err)
	}

	// Get the next nonce for the sender’s account to ensure transaction ordering
	nextNonce, err := getNextNonce(cProps.Client, cProps.MyPubKey)
	if err != nil {
		log.Errorf("Failed to retrieve next nonce: %v", err)
		return common.Address{}, fmt.Errorf("failed to get next nonce: %w", err)
	}

	// Configure transaction parameters
	auth.Nonce = big.NewInt(int64(nextNonce))
	auth.Value = big.NewInt(0)      // No ETH sent with this transaction
	auth.GasLimit = cProps.GasLimit // Gas limit set via command line or default
	auth.GasPrice = gasPrice        // Suggested gas price from the network
	log.Debugf("Transaction config - Nonce: %d, Gas Limit: %d, Gas Price: %s wei",
		auth.Nonce, auth.GasLimit, auth.GasPrice.String())

	// Prepare arguments for the Create function
	args := struct {
		BurnDest   common.Address
		Token      common.Address
		Dest       common.Address
		Pool       common.Address
		OCPrice    common.Address
		TokenPrice common.Address
	}{
		BurnDest:   ToAddr(cProps.Addresses.DeadAddr),
		Token:      ToAddr(cProps.Addresses.TknAddr),
		Dest:       ToAddr(cProps.Addresses.TargetAddr),
		Pool:       ToAddr(cProps.Addresses.PoolAddr),
		OCPrice:    ToAddr(cProps.Addresses.MyPublicKey),
		TokenPrice: ToAddr(cProps.Addresses.TknPrcAddr),
	}

	// Prompt user for confirmation
	log.Info("KT creation arguments:")
	log.Infof("  Burn Destination: %s", args.BurnDest.Hex())
	log.Infof("  Token Address: %s", args.Token.Hex())
	log.Infof("  Target Address: %s", args.Dest.Hex())
	log.Infof("  Pool Address: %s", args.Pool.Hex())
	log.Infof("  OC Price Address: %s", args.OCPrice.Hex())
	log.Infof("  Token Price Address: %s", args.TokenPrice.Hex())
	log.Infof("  Gas Limit: %d", auth.GasLimit)

	fmt.Print("Are you sure you want to create this contract with these arguments? (y/N): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if response != "y" && response != "yes" {
		log.Info("KT creation aborted by user")
		return common.Address{}, fmt.Errorf("user aborted KT creation")
	}

	// Execute the Create function on the factory contract to deploy a new KT
	tx, err := instance.Create(auth,
		args.BurnDest,   // Burn destination address
		args.Token,      // Token contract address
		args.Dest,       // Target address
		args.Pool,       // Pool address
		args.OCPrice,    // OC price address (using sender’s public key)
		args.TokenPrice) // Token price address
	if err != nil {
		log.Errorf("Failed to submit KT creation transaction: %v", err)
		log.Warn("This might be due to insufficient funds, low gas limit, or invalid contract addresses.")
		return common.Address{}, fmt.Errorf("failed to create KT: %w", err)
	}

	log.Infof("Transaction submitted: %s (Nonce: %d)", tx.Hash().Hex(), tx.Nonce())

	// Wait for the transaction to be mined and get the receipt
	receipt, err := bind.WaitMined(context.Background(), cProps.Client, tx)
	if err != nil {
		log.Errorf("Failed to confirm transaction mining: %v (Tx Hash: %s)", err, tx.Hash().Hex())
		return common.Address{}, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	// Check transaction status and provide detailed feedback
	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Info("KT creation successful")
		log.Debugf("Confirmed in block: %d, Gas used: %d",
			receipt.BlockNumber.Uint64(), receipt.GasUsed)
	} else {
		log.Errorf("KT creation failed in block %d (Tx Hash: %s)",
			receipt.BlockNumber.Uint64(), tx.Hash().Hex())
		log.Warnf("Gas used: %d out of %d (%.2f%%)",
			receipt.GasUsed, auth.GasLimit, float64(receipt.GasUsed)/float64(auth.GasLimit)*100)
		log.Warn("Possible reasons: insufficient funds, gas limit too low, or contract logic reverted.")
		log.Warn("Try increasing -gasLimit (e.g., 4000000) or checking account balance and contract addresses.")
		return common.Address{}, fmt.Errorf("KT creation failed: transaction reverted")
	}

	// Retrieve and return the created KT contract address
	newKtAddr, _ := PrintKtFactContracts(cProps)
	log.Infof("New KT created at: %s", newKtAddr.Hex())
	return newKtAddr, nil
}

// printKtFactContracts lists all KT contracts created by the factory and returns the first address found.
func PrintKtFactContracts(cProps *ConnectionProps) (common.Address, error) {
	factoryAddr := ToAddr(cProps.Addresses.FactoryAddr)
	log.Infof("Factory address: %s", cProps.Addresses.FactoryAddr)

	// Instantiate the factory contract
	instance, err := abis.NewKtv2fact(factoryAddr, cProps.Backend)
	if err != nil {
		log.Errorf("Factory instantiation failed: %v", err)
		return common.Address{}, fmt.Errorf("failed to instantiate factory: %w", err)
	}

	var lastAddressFound common.Address
	i := int64(0)

	for {
		address, err := instance.Created(nil, big.NewInt(i))
		if err != nil {
			log.Warnf("Failed to fetch KT at index %d: %v", i, err)
			break // No more KTs found in the factory's created list
		}

		lastAddressFound = address
		log.Infof("KT at index %d: %s", i, address.Hex())
		i++
	}

	if i == 0 {
		log.Warn("No KTs found in factory")
	} else {
		log.Infof("Total KTs found: %d", i)
	}

	return lastAddressFound, nil
}

func PrintKtContractVariables(cProps *ConnectionProps) {
	callOpts := &bind.CallOpts{
		Context: context.Background(),
		Pending: false,
		From:    cProps.MyPubKey,
	}

	totalStk, err := cProps.Kt.TotalStk(callOpts)
	if err != nil {
		log.Printf("Error fetching TotalStk: %v", err)
		return
	}
	totalGvn, err := cProps.Kt.TotalGvn(callOpts)
	if err != nil {
		log.Printf("Error fetching TotalGvn: %v", err)
		return
	}
	totalBurned, err := cProps.Kt.TotalBurned(callOpts)
	if err != nil {
		log.Printf("Error fetching TotalBurned: %v", err)
		return
	}
	maxBrnPrc, err := cProps.Kt.MaxBrnPrc(callOpts)
	if err != nil {
		log.Printf("Error fetching MaxBrnPrc: %v", err)
		return
	}
	donationPrc, err := cProps.Kt.DonationPrc(callOpts)
	if err != nil {
		log.Printf("Error fetching DonationPrc: %v", err)
		return
	}
	burnFactor, err := cProps.Kt.BurnFactor(callOpts)
	if err != nil {
		log.Printf("Error fetching BurnFactor: %v", err)
		return
	}
	startBlock, err := cProps.Kt.StartBlock(callOpts)
	if err != nil {
		log.Printf("Error fetching StartBlock: %v", err)
		return
	}
	epochInterval, err := cProps.Kt.EpochInterval(callOpts)
	if err != nil {
		log.Printf("Error fetching EpochInterval: %v", err)
		return
	}
	consensusReq, err := cProps.Kt.ConsensusReq(callOpts)
	if err != nil {
		log.Printf("Error fetching ConsensusReq: %v", err)
		return
	}
	ocFee, err := cProps.Kt.OcFee(callOpts)
	if err != nil {
		log.Printf("Error fetching OcFee: %v", err)
		return
	}
	tlOcFees, err := cProps.Kt.TlOcFees(callOpts)
	if err != nil {
		log.Printf("Error fetching TlOcFees: %v", err)
		return
	}
	tokenAddr, err := cProps.Kt.TokenAddr(callOpts)
	if err != nil {
		log.Printf("Error fetching TokenAddr: %v", err)
		return
	}
	dest, err := cProps.Kt.Dest(callOpts)
	if err != nil {
		log.Printf("Error fetching Dest: %v", err)
		return
	}
	burnDest, err := cProps.Kt.BurnDest(callOpts)
	if err != nil {
		log.Printf("Error fetching BurnDest: %v", err)
		return
	}
	pool, err := cProps.Kt.Pool(callOpts)
	if err != nil {
		log.Printf("Error fetching Pool: %v", err)
		return
	}

	log.Printf("KT Contract Variables:")

	// Convert totalGvn from Wei to ETH
	totalGvnEth := new(big.Float).Quo(new(big.Float).SetInt(totalGvn), big.NewFloat(1e18))

	log.Printf("Total Staked: %s", totalStk.String())
	log.Printf("Total Given: %s", totalGvnEth.String())
	log.Printf("Total Burned: %s", totalBurned.String())
	log.Printf("Max Burn Percentage: %d", maxBrnPrc)
	log.Printf("Donation Percentage: %d", donationPrc)
	log.Printf("Burn Factor: %d", burnFactor)
	log.Printf("Start Block: %s", startBlock.String())
	log.Printf("Epoch Interval: %d", epochInterval)
	log.Printf("Consensus Requirement: %d", consensusReq)
	log.Printf("OC Fee: %d", ocFee)
	log.Printf("Total OC Fees: %s", tlOcFees.String())

	log.Print("Addresses:")
	log.Printf(" Token Address: %s", tokenAddr.Hex())
	log.Printf(" Destination: %s", dest.Hex())
	log.Printf(" Burn Destination: %s", burnDest.Hex())
	log.Printf(" Pool: %s", pool.Hex())

	// Convert Wei to ETH
	PrintKtBalance(cProps)
	log.Println("Note: Mappings (blockRwd, ocRwdrVote, ocRwdrs, ocFees) will not be printed as they require specific keys.")
	log.Println()
}

func PrintKtBalance(cProps *ConnectionProps) {
	balance, err := cProps.Client.BalanceAt(context.Background(), cProps.KtAddr, nil)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}

	ethBalance := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))

	log.Printf("Balance: %f ETH", ethBalance)
}
