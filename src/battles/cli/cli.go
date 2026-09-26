// Package cli is the battles command set. Run takes its inputs, outputs, and
// node connection as parameters so the whole command surface is testable
// against a simulated chain.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"ktp2/src/battles/vault"

	"github.com/common-nighthawk/go-figure"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

// ANSI colors used by the banner and usage; the terminal is wrapped by
// go-colorable in main so these render on Windows too.
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorRed    = "\033[1;31m"
	colorYellow = "\033[1;33m"
	colorCyan   = "\033[1;36m"
	colorDim    = "\033[2m"
)

// Command names.
const (
	CmdInit       = "init"
	CmdDeploy     = "deploy"
	CmdKeygen     = "keygen"
	CmdFund       = "fund"
	CmdStatus     = "status"
	CmdUnlock     = "unlock"
	CmdVerify     = "verify"
	CmdBanks      = "banks"
	CmdAddBank    = "add-bank"
	CmdRemoveBank = "remove-bank"
	CmdChallenges = "challenges"
	CmdVersion    = "version"
	CmdHelp       = "help"

	// Flag names shared by several commands.
	FlagAmount = "amount"
	FlagKey    = "key"
	FlagHash   = "hash"
	FlagBank   = "bank"
	FlagBanks  = "banks"
	FlagTx     = "tx"
	FlagFrom   = "from"

	// KtocEnvFile is the ktoc config the wizard offers to import from.
	KtocEnvFile = ".env"
)

// Deps are the CLI's ties to the outside world; tests substitute all of them.
type Deps struct {
	// Dial opens a connection to an RPC endpoint.
	Dial func(ctx context.Context, endpoint string) (vault.Client, error)
	// ConfigPath is where battles.env lives (default: current directory).
	ConfigPath string
	// ReadSecret reads a private key without echo; nil falls back to stdin.
	ReadSecret func() (string, error)
	// Version is printed by `battles version`.
	Version string
}

// Run executes one command line and returns the process exit code.
func Run(args []string, in io.Reader, out, errOut io.Writer, deps Deps) int {
	if deps.ConfigPath == "" {
		deps.ConfigPath = vault.ConfigFileName
	}
	if len(args) == 0 {
		usage(out, deps.Version)
		return 2
	}
	cmd, rest := args[0], args[1:]
	ctx := context.Background()
	app := &app{in: in, out: out, err: errOut, deps: deps, ctx: ctx}

	var err error
	switch cmd {
	case CmdHelp, "-h", "--help":
		usage(out, deps.Version)
		return 0
	case CmdVersion:
		Banner(out, deps.Version)
		fmt.Fprintf(out, "battles %s\n", deps.Version)
		return 0
	case CmdKeygen:
		err = app.keygen()
	case CmdInit:
		err = app.init(true)
	case CmdDeploy:
		err = app.deploy(rest)
	default:
		err = app.withSession(cmd, rest)
	}
	if err != nil {
		fmt.Fprintf(errOut, "error: %v\n", err)
		return 1
	}
	return 0
}

type app struct {
	in   io.Reader
	out  io.Writer
	err  io.Writer
	deps Deps
	ctx  context.Context
}

// Banner is the fight-card header: two-tone ASCII art, a tagline, the version.
func Banner(w io.Writer, version string) {
	fire := colorRed + "🔥" + colorReset
	rule := colorDim + strings.Repeat("═", 84) + colorReset
	fmt.Fprint(w, "\n", rule, "\n")
	fmt.Fprint(w, figure.NewColorFigure("BURN BANK", "slant", "yellow", true).ColorString())
	fmt.Fprint(w, figure.NewColorFigure("BATTLES", "doom", "red", true).ColorString())
	fmt.Fprintf(w, "   %s  %slock it   ·   match it   ·   unlock it%s  %s        %sthe Vault   %s%s\n",
		fire, colorCyan, colorReset, fire, colorDim, version, colorReset)
	fmt.Fprint(w, rule, "\n")
}

func usage(w io.Writer, version string) {
	Banner(w, version)
	h := func(title string) string { return "\n" + colorYellow + colorBold + title + colorReset + "\n" }
	c := func(cmd, desc string) string {
		return "  " + colorCyan + fmt.Sprintf("%-32s", cmd) + colorReset + desc + "\n"
	}
	fmt.Fprint(w, "\n"+colorBold+"Usage: battles <command> [flags]"+colorReset+"\n")
	fmt.Fprint(w, h("Setup"))
	fmt.Fprint(w, c("init", "Guided setup: writes battles.env (runs automatically when it is missing)"))
	fmt.Fprint(w, c("deploy [-banks a,b]", "Deploy a new Vault contract and record its address in battles.env"))
	fmt.Fprint(w, c("version", "Print the build version"))
	fmt.Fprint(w, h("Challenges"))
	fmt.Fprint(w, c("keygen", "Generate a vault key and its hash (nothing is sent)"))
	fmt.Fprint(w, c("fund -amount <eth> [-key <hex>]", "Lock ETH under a key; prints the SECRET key and the public hash"))
	fmt.Fprint(w, c("status -hash <0x..>", "Show a vault: funder, amount, funded block, spent"))
	fmt.Fprint(w, c("verify -tx <0x..> -hash <0x..>", "Check whether a transaction fulfills the challenge under that hash"))
	fmt.Fprint(w, c("unlock -key <hex> -bank <0x..>", "Spend a vault into a whitelisted Burn Bank's give()"))
	fmt.Fprint(w, c("challenges [-from <block>]", "List every vault funded on this contract and its outcome"))
	fmt.Fprint(w, h("Banks (contract owner only)"))
	fmt.Fprint(w, c("banks", "List whitelisted Burn Banks"))
	fmt.Fprint(w, c("add-bank -bank <0x..>", "Whitelist a Burn Bank"))
	fmt.Fprint(w, c("remove-bank -bank <0x..>", "Remove a Burn Bank from the whitelist"))
	fmt.Fprint(w, h("Configuration"))
	fmt.Fprint(w, "  battles.env in the current directory, or the same names as environment variables\n")
	fmt.Fprint(w, "  (BATTLES_ETH_ENDPOINT, BATTLES_PRIVATE_KEY, BATTLES_VAULT_ADDR, BATTLES_CHAIN_ID,\n")
	fmt.Fprint(w, "  BATTLES_VAULT_BLOCK). A job that sets all of them needs no file.\n\n")
}

// ---------------------------------------------------------------- setup

func (a *app) probe(endpoint string) (*big.Int, error) {
	c, err := a.deps.Dial(a.ctx, endpoint)
	if err != nil {
		return nil, err
	}
	return c.ChainID(a.ctx)
}

// loadOrInit returns the config, running the wizard first when the file is
// missing. It never returns a config without a vault address.
func (a *app) loadOrInit() (*vault.Config, error) {
	cfg, err := vault.LoadConfig(a.deps.ConfigPath)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintf(a.out, "No %s found. Starting setup.\n\n", a.deps.ConfigPath)
		if err := a.init(false); err != nil {
			return nil, err
		}
		cfg, err = vault.LoadConfig(a.deps.ConfigPath)
	}
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", a.deps.ConfigPath, err)
	}
	return cfg, nil
}

// init runs the wizard, deploys if asked, and writes battles.env.
func (a *app) init(explicit bool) error {
	if explicit {
		if _, err := os.Stat(a.deps.ConfigPath); err == nil {
			fmt.Fprintf(a.out, "%s already exists; its values are offered as defaults.\n", a.deps.ConfigPath)
		}
	}
	opts := vault.WizardOptions{Probe: a.probe, ReadSecret: a.deps.ReadSecret}
	if existing, err := vault.LoadConfig(a.deps.ConfigPath); err == nil {
		opts.DefaultEndpoint, opts.DefaultPrivateKey = existing.EthEndpoint, existing.PrivateKey
	} else if ep, key := vault.ImportKtocEnv(filepath.Join(filepath.Dir(a.deps.ConfigPath), KtocEnvFile)); ep != "" || key != "" {
		fmt.Fprintf(a.out, "Found a ktoc %s; its endpoint and key are offered as defaults.\n", KtocEnvFile)
		opts.DefaultEndpoint, opts.DefaultPrivateKey = ep, key
	}

	cfg, plan, err := vault.RunWizard(a.in, a.out, opts)
	if err != nil {
		return err
	}
	if plan != nil {
		addr, err := a.doDeploy(cfg, plan.Banks)
		if err != nil {
			return err
		}
		cfg.VaultAddr = addr
	}
	if err := cfg.Save(a.deps.ConfigPath); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "\nWrote %s. Vault %s on chain %s.\n", a.deps.ConfigPath, cfg.VaultAddr.Hex(), cfg.ChainID)
	return nil
}

func (a *app) doDeploy(cfg *vault.Config, banks []common.Address) (common.Address, error) {
	addr, block, err := a.doDeployAt(cfg, banks)
	if err == nil {
		cfg.VaultBlock = block
	}
	return addr, err
}

func (a *app) doDeployAt(cfg *vault.Config, banks []common.Address) (common.Address, uint64, error) {
	key, from, err := cfg.Signer()
	if err != nil {
		return common.Address{}, 0, err
	}
	client, err := a.deps.Dial(a.ctx, cfg.EthEndpoint)
	if err != nil {
		return common.Address{}, 0, fmt.Errorf("failed to connect to %s: %w", cfg.EthEndpoint, err)
	}
	if cfg.ChainID == nil {
		if cfg.ChainID, err = client.ChainID(a.ctx); err != nil {
			return common.Address{}, 0, fmt.Errorf("failed to read chain ID: %w", err)
		}
	}
	fmt.Fprintf(a.out, "Deploying Vault from %s with %d bank(s)...\n", from.Hex(), len(banks))
	addr, rcpt, err := vault.Deploy(a.ctx, client, key, cfg.ChainID, banks)
	if err != nil {
		return common.Address{}, 0, err
	}
	fmt.Fprintf(a.out, "Vault deployed at %s (block %s, tx %s).\n", addr.Hex(), rcpt.BlockNumber, rcpt.TxHash.Hex())
	fmt.Fprintf(a.out, "You (%s) are its owner: only you can add or remove banks.\n", from.Hex())
	return addr, rcpt.BlockNumber.Uint64(), nil
}

// deploy: `battles deploy [-banks a,b]` on an existing config (or after init).
func (a *app) deploy(args []string) error {
	flags := flag.NewFlagSet(CmdDeploy, flag.ContinueOnError)
	banksArg := flags.String(FlagBanks, "", "comma-separated Burn Bank addresses to whitelist")
	if err := flags.Parse(args); err != nil {
		return err
	}
	banks, err := parseBanks(*banksArg)
	if err != nil {
		return err
	}
	cfg, err := vault.LoadConfig(a.deps.ConfigPath)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("no %s yet; run `battles %s` first (it can deploy for you)", a.deps.ConfigPath, CmdInit)
	}
	if err != nil {
		return err
	}
	if cfg.VaultAddr != (common.Address{}) {
		fmt.Fprintf(a.out, "Note: %s already points at %s; the new address will replace it.\n", a.deps.ConfigPath, cfg.VaultAddr.Hex())
	}
	addr, err := a.doDeploy(cfg, banks)
	if err != nil {
		return err
	}
	cfg.VaultAddr = addr
	if err := cfg.Save(a.deps.ConfigPath); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "Updated %s.\n", a.deps.ConfigPath)
	return nil
}

// --------------------------------------------------------------- session

func (a *app) withSession(cmd string, args []string) error {
	cfg, err := a.loadOrInit()
	if err != nil {
		return err
	}
	client, err := a.deps.Dial(a.ctx, cfg.EthEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", cfg.EthEndpoint, err)
	}
	s, err := vault.NewSession(a.ctx, client, cfg)
	if err != nil {
		return err
	}
	if cfg.VaultBlock == 0 && s.DeployBlock != 0 {
		// Found by search; remember it so later runs skip the lookup.
		cfg.VaultBlock = s.DeployBlock
		if _, statErr := os.Stat(a.deps.ConfigPath); statErr == nil {
			_ = cfg.Save(a.deps.ConfigPath)
		}
	}
	switch cmd {
	case CmdFund:
		return a.fund(s, args)
	case CmdStatus:
		return a.status(s, args)
	case CmdUnlock:
		return a.unlock(s, args)
	case CmdVerify:
		return a.verify(s, args)
	case CmdBanks:
		return a.banks(s)
	case CmdAddBank:
		return a.bankOp(s, CmdAddBank, args)
	case CmdRemoveBank:
		return a.bankOp(s, CmdRemoveBank, args)
	case CmdChallenges:
		return a.challenges(s, args)
	}
	usage(a.err, a.deps.Version)
	return fmt.Errorf("unknown command %q", cmd)
}

func (a *app) keygen() error {
	k, err := vault.GenerateKey()
	if err != nil {
		return err
	}
	a.printKey(k)
	return nil
}

func (a *app) printKey(k vault.Key) {
	fmt.Fprintf(a.out, "%sSECRET key : %s   <- keep this private until the challenge is fulfilled%s\n", colorRed, k.Hex(), colorReset)
	fmt.Fprintf(a.out, "%sPublic hash: %s   <- this identifies the challenge%s\n", colorCyan, k.Hash().Hex(), colorReset)
}

func (a *app) fund(s *vault.Session, args []string) error {
	fs := flag.NewFlagSet(CmdFund, flag.ContinueOnError)
	amountArg := fs.String(FlagAmount, "", "amount in ETH, e.g. 0.01")
	keyArg := fs.String(FlagKey, "", "vault key (hex); generated when omitted")
	if err := fs.Parse(args); err != nil {
		return err
	}
	amount, err := parseEth(*amountArg)
	if err != nil {
		return err
	}
	var k vault.Key
	if *keyArg == "" {
		if k, err = vault.GenerateKey(); err != nil {
			return err
		}
	} else if k, err = vault.ParseKey(*keyArg); err != nil {
		return err
	}
	a.printKey(k)
	fmt.Fprintf(a.out, "Funding %s ETH from %s...\n", fmtEth(amount), s.From.Hex())
	rcpt, err := s.Fund(a.ctx, k.Hash(), amount)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.out, "Funded in block %s (tx %s).\n", rcpt.BlockNumber, rcpt.TxHash.Hex())
	fmt.Fprintf(a.out, "Post the challenge with the hash and this tx; reveal the key only once it is fulfilled.\n")
	return nil
}

func (a *app) status(s *vault.Session, args []string) error {
	fs := flag.NewFlagSet(CmdStatus, flag.ContinueOnError)
	hashArg := fs.String(FlagHash, "", "vault key hash (0x..)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	h, err := vault.ParseKeyHash(*hashArg)
	if err != nil {
		return err
	}
	st, err := s.Status(a.ctx, h)
	if err != nil {
		return err
	}
	if st.Amount.Sign() == 0 {
		fmt.Fprintf(a.out, "No vault is funded under %s.\n", h.Hex())
		return nil
	}
	state := "unlockable"
	if st.Spent {
		state = "spent"
	}
	fmt.Fprintf(a.out, "Vault %s\n  funder : %s\n  amount : %s ETH\n  funded : block %d\n  state  : %s\n", h.Hex(), st.Funder.Hex(), fmtEth(st.Amount), st.FundedBlock, state)
	return nil
}

func (a *app) unlock(s *vault.Session, args []string) error {
	fs := flag.NewFlagSet(CmdUnlock, flag.ContinueOnError)
	keyArg := fs.String(FlagKey, "", "vault key (hex)")
	bankArg := fs.String(FlagBank, "", "Burn Bank address to receive the ETH")
	if err := fs.Parse(args); err != nil {
		return err
	}
	k, err := vault.ParseKey(*keyArg)
	if err != nil {
		return err
	}
	bank, err := parseAddress(*bankArg, FlagBank)
	if err != nil {
		return err
	}
	st, err := s.Status(a.ctx, k.Hash())
	if err != nil {
		return err
	}
	if st.Amount.Sign() == 0 {
		return fmt.Errorf("no vault is funded under this key (hash %s)", k.Hash().Hex())
	}
	if st.Spent {
		return fmt.Errorf("this vault was already unlocked")
	}
	if ok, err := s.IsBank(a.ctx, bank); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("%s is not a whitelisted Burn Bank (see `battles %s`)", bank.Hex(), CmdBanks)
	}
	fmt.Fprintf(a.out, "Unlocking %s ETH into %s...\n", fmtEth(st.Amount), bank.Hex())
	rcpt, err := s.Unlock(a.ctx, k, bank)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.out, "Unlocked in block %s (tx %s). The key is now spent.\n", rcpt.BlockNumber, rcpt.TxHash.Hex())
	return nil
}

func (a *app) verify(s *vault.Session, args []string) error {
	fs := flag.NewFlagSet(CmdVerify, flag.ContinueOnError)
	txArg := fs.String(FlagTx, "", "transaction hash the participant submitted")
	hashArg := fs.String(FlagHash, "", "vault key hash of the challenge")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !strings.HasPrefix(*txArg, "0x") || len(*txArg) != 66 {
		return fmt.Errorf("-%s must be a 0x-prefixed 32-byte transaction hash", FlagTx)
	}
	h, err := vault.ParseKeyHash(*hashArg)
	if err != nil {
		return err
	}
	r, err := s.VerifyMatch(a.ctx, common.HexToHash(*txArg), h)
	if err != nil {
		return err
	}
	if r.OK {
		fmt.Fprintf(a.out, "FULFILLED: %s gave %s ETH to %s in block %d.\n", r.Sender.Hex(), fmtEth(r.Amount), r.Bank.Hex(), r.Block)
		fmt.Fprintf(a.out, "The key for %s can now be published.\n", h.Hex())
		return nil
	}
	fmt.Fprintf(a.out, "NOT FULFILLED: %s.\n", r.Reason)
	return nil
}

func (a *app) banks(s *vault.Session) error {
	banks, err := s.Banks(a.ctx)
	if err != nil {
		return err
	}
	if len(banks) == 0 {
		fmt.Fprintf(a.out, "No Burn Banks are whitelisted on %s.\n", s.VaultAddr.Hex())
		return nil
	}
	fmt.Fprintf(a.out, "Whitelisted Burn Banks on %s:\n", s.VaultAddr.Hex())
	for _, b := range banks {
		fmt.Fprintf(a.out, "  %s\n", b.Hex())
	}
	return nil
}

func (a *app) bankOp(s *vault.Session, op string, args []string) error {
	fs := flag.NewFlagSet(op, flag.ContinueOnError)
	bankArg := fs.String(FlagBank, "", "Burn Bank address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	bank, err := parseAddress(*bankArg, FlagBank)
	if err != nil {
		return err
	}
	var rcpt interface{ String() string }
	if op == CmdAddBank {
		r, err := s.AddBank(a.ctx, bank)
		if err != nil {
			return err
		}
		rcpt = r.BlockNumber
		fmt.Fprintf(a.out, "Added %s (block %s).\n", bank.Hex(), rcpt)
	} else {
		r, err := s.RemoveBank(a.ctx, bank)
		if err != nil {
			return err
		}
		rcpt = r.BlockNumber
		fmt.Fprintf(a.out, "Removed %s (block %s).\n", bank.Hex(), rcpt)
	}
	return nil
}

func (a *app) challenges(s *vault.Session, args []string) error {
	fs := flag.NewFlagSet(CmdChallenges, flag.ContinueOnError)
	from := fs.Uint64(FlagFrom, 0, "first block to scan")
	if err := fs.Parse(args); err != nil {
		return err
	}
	list, err := s.Challenges(a.ctx, *from)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Fprintf(a.out, "No vaults funded on %s.\n", s.VaultAddr.Hex())
		return nil
	}
	for _, c := range list {
		outcome := "unlockable"
		if c.Spent {
			outcome = fmt.Sprintf("unlocked into %s by %s at block %d", c.Bank.Hex(), c.Unlocker.Hex(), c.UnlockBlock)
		}
		fmt.Fprintf(a.out, "%s  %s ETH  funded by %s at block %d  %s\n", c.KeyHash.Hex(), fmtEth(c.Amount), c.Funder.Hex(), c.FundedBlock, outcome)
	}
	return nil
}

// --------------------------------------------------------------- helpers

func parseBanks(s string) ([]common.Address, error) {
	var out []common.Address
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		addr, err := parseAddress(part, FlagBanks)
		if err != nil {
			return nil, err
		}
		out = append(out, addr)
	}
	return out, nil
}

func parseAddress(s, flagName string) (common.Address, error) {
	if !common.IsHexAddress(s) {
		return common.Address{}, fmt.Errorf("-%s must be a 0x address, got %q", flagName, s)
	}
	return common.HexToAddress(s), nil
}

// parseEth turns "0.01" into wei.
func parseEth(s string) (*big.Int, error) {
	if s == "" {
		return nil, fmt.Errorf("-%s is required (in ETH, e.g. 0.01)", FlagAmount)
	}
	f, ok := new(big.Float).SetPrec(256).SetString(s)
	if !ok || f.Sign() <= 0 {
		return nil, fmt.Errorf("-%s must be a positive number of ETH, got %q", FlagAmount, s)
	}
	wei, _ := new(big.Float).Mul(f, big.NewFloat(params.Ether)).Int(nil)
	if wei.Sign() <= 0 {
		return nil, fmt.Errorf("-%s is too small to represent in wei", FlagAmount)
	}
	return wei, nil
}

func fmtEth(wei *big.Int) string {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether)).Text('f', 6)
}
