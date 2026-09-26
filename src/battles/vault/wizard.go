package vault

// First-run wizard. With no battles.env present the CLI walks the operator
// through endpoint, signing key, and vault (existing address or deploy now),
// then writes the file. The wizard reads from an io.Reader and writes to an
// io.Writer so tests can script it.

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
)

// Wizard prompts, exported so tests and docs quote the same words.
const (
	PromptEndpoint   = "Ethereum RPC endpoint"
	PromptPrivateKey = "Private key of the wallet that will sign transactions (hex)"
	PromptVaultMode  = "Vault: enter an existing address, or type 'deploy' to deploy a new one"
	PromptBanks      = "Burn Bank addresses to whitelist, comma-separated (empty for none)"
	PromptConfirm    = "Write battles.env with these settings? [Y/n]"

	VaultModeDeploy = "deploy"

	// ktoc .env keys the wizard can import defaults from.
	ktocEnvEndpoint   = "ETH_ENDPOINT"
	ktocEnvPrivateKey = "MY_PRIVATE_KEY"
)

// WizardOptions supplies defaults and the probes the wizard needs.
type WizardOptions struct {
	// DefaultEndpoint and DefaultPrivateKey pre-fill the prompts, typically
	// imported from a ktoc .env in the same directory.
	DefaultEndpoint   string
	DefaultPrivateKey string
	// Probe connects to an endpoint and returns its chain ID; nil skips probing.
	Probe func(endpoint string) (*big.Int, error)
	// ReadSecret reads the private key without echo; nil reads a line from in.
	ReadSecret func() (string, error)
}

// DeployPlan is returned when the operator chose to deploy a new Vault.
type DeployPlan struct {
	Banks []common.Address
}

type prompter struct {
	in  *bufio.Scanner
	out io.Writer
}

func (p *prompter) say(format string, args ...interface{}) {
	fmt.Fprintf(p.out, format+"\n", args...)
}

// ask prints a prompt (with the default shown when there is one) and returns
// the trimmed answer, or the default when the answer is empty.
func (p *prompter) ask(prompt, def string) (string, error) {
	if def != "" {
		fmt.Fprintf(p.out, "%s [%s]: ", prompt, def)
	} else {
		fmt.Fprintf(p.out, "%s: ", prompt)
	}
	if !p.in.Scan() {
		if err := p.in.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("input ended before the wizard finished")
	}
	answer := strings.TrimSpace(p.in.Text())
	if answer == "" {
		return def, nil
	}
	return answer, nil
}

// RunWizard drives the prompts and returns the config the operator confirmed.
// When they chose to deploy, plan is non-nil and cfg.VaultAddr is unset until
// the caller deploys and fills it in.
func RunWizard(in io.Reader, out io.Writer, opts WizardOptions) (cfg *Config, plan *DeployPlan, err error) {
	p := &prompter{in: bufio.NewScanner(in), out: out}
	cfg = &Config{}

	p.say("battles setup: answer a few questions and a %s file will be written here.", ConfigFileName)
	p.say("")

	// 1. Endpoint, probed for its chain ID.
	for {
		ep, err := p.ask(PromptEndpoint, opts.DefaultEndpoint)
		if err != nil {
			return nil, nil, err
		}
		if ep == "" {
			p.say("  An endpoint is required (e.g. https://mainnet.infura.io/v3/<key> or http://127.0.0.1:8545).")
			continue
		}
		if opts.Probe != nil {
			id, err := opts.Probe(ep)
			if err != nil {
				p.say("  Could not reach %s: %v", ep, err)
				continue
			}
			cfg.ChainID = id
			p.say("  Connected. Chain ID %s.", id)
		}
		cfg.EthEndpoint = ep
		break
	}

	// 2. Private key, never echoed back in full.
	for {
		var raw string
		if opts.ReadSecret != nil {
			if opts.DefaultPrivateKey != "" {
				fmt.Fprintf(p.out, "%s [%s, Enter to keep]: ", PromptPrivateKey, maskKey(opts.DefaultPrivateKey))
			} else {
				fmt.Fprintf(p.out, "%s: ", PromptPrivateKey)
			}
			raw, err = opts.ReadSecret()
			if err != nil {
				return nil, nil, err
			}
			p.say("")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				raw = opts.DefaultPrivateKey
			}
		} else {
			def := ""
			if opts.DefaultPrivateKey != "" {
				def = maskKey(opts.DefaultPrivateKey)
			}
			raw, err = p.ask(PromptPrivateKey, def)
			if err != nil {
				return nil, nil, err
			}
			if raw == def && def != "" {
				raw = opts.DefaultPrivateKey
			}
		}
		probe := &Config{PrivateKey: raw}
		if _, addr, err := probe.Signer(); err != nil {
			p.say("  That is not a valid private key: 64 hex characters, 0x optional.")
			continue
		} else {
			p.say("  Signing as %s.", addr.Hex())
		}
		cfg.PrivateKey = strings.TrimPrefix(strings.TrimSpace(raw), "0x")
		break
	}

	// 3. Vault: existing address or deploy.
	for {
		answer, err := p.ask(PromptVaultMode, "")
		if err != nil {
			return nil, nil, err
		}
		if strings.EqualFold(answer, VaultModeDeploy) {
			banksAnswer, err := p.ask(PromptBanks, "")
			if err != nil {
				return nil, nil, err
			}
			banks, bad := parseAddressList(banksAnswer)
			if bad != "" {
				p.say("  %q is not an address; try again.", bad)
				continue
			}
			plan = &DeployPlan{Banks: banks}
			break
		}
		if !common.IsHexAddress(answer) {
			p.say("  Enter a 0x address, or the word %q.", VaultModeDeploy)
			continue
		}
		cfg.VaultAddr = common.HexToAddress(answer)
		break
	}

	// 4. Confirm.
	p.say("")
	p.say("  Endpoint : %s", cfg.EthEndpoint)
	if cfg.ChainID != nil {
		p.say("  Chain ID : %s", cfg.ChainID)
	}
	p.say("  Key      : %s", maskKey(cfg.PrivateKey))
	if plan != nil {
		p.say("  Vault    : deploy a new one with %d bank(s)", len(plan.Banks))
	} else {
		p.say("  Vault    : %s", cfg.VaultAddr.Hex())
	}
	answer, err := p.ask(PromptConfirm, "Y")
	if err != nil {
		return nil, nil, err
	}
	if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
		return nil, nil, fmt.Errorf("setup cancelled; nothing was written")
	}
	return cfg, plan, nil
}

// maskKey shows enough of a key to recognize it and no more.
func maskKey(k string) string {
	k = strings.TrimPrefix(k, "0x")
	if len(k) < 8 {
		return "****"
	}
	return k[:4] + "…" + k[len(k)-4:]
}

// parseAddressList splits a comma-separated list; bad is the first entry that
// is not an address, empty when all parse.
func parseAddressList(s string) (addrs []common.Address, bad string) {
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !common.IsHexAddress(part) {
			return nil, part
		}
		addrs = append(addrs, common.HexToAddress(part))
	}
	return addrs, ""
}

// ImportKtocEnv reads ETH_ENDPOINT and MY_PRIVATE_KEY from a ktoc .env file,
// returning empty strings when the file or keys are absent.
func ImportKtocEnv(path string) (endpoint, privateKey string) {
	values, err := godotenv.Read(path)
	if err != nil {
		return "", ""
	}
	return values[ktocEnvEndpoint], values[ktocEnvPrivateKey]
}
