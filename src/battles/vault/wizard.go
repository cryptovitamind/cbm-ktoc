package vault

// First-run wizard. With no battles.env present the CLI walks the operator
// through endpoint, signing key, and vault (existing address or deploy now),
// then writes the file. The wizard reads from an io.Reader and writes to an
// io.Writer so tests can script it.

import (
	"bufio"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"io"
	"math/big"
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

func (p *prompter) say(format string, args ...interface {
}) {
	fmt.Fprintf(p.out, format+"\n", args...)
}

// ask prints a prompt (with the default shown when there is one) and returns
// the trimmed answer, or the default when the answer is empty.
func (p *prompter) ask(prompt, def string) (string, error) {
	return "", errNotImplemented
}

// RunWizard drives the prompts and returns the config the operator confirmed.
// When they chose to deploy, plan is non-nil and cfg.VaultAddr is unset until
// the caller deploys and fills it in.
func RunWizard(in io.Reader, out io.Writer, opts WizardOptions) (cfg *Config, plan *DeployPlan, err error) {
	return nil, nil, errNotImplemented
}

// maskKey shows enough of a key to recognize it and no more.
func maskKey(k string) string {
	return ""
}

// parseAddressList splits a comma-separated list; bad is the first entry that
// is not an address, empty when all parse.
func parseAddressList(s string) (addrs []common.Address, bad string) {
	return nil, ""
}

// ImportKtocEnv reads ETH_ENDPOINT and MY_PRIVATE_KEY from a ktoc .env file,
// returning empty strings when the file or keys are absent.
func ImportKtocEnv(path string) (endpoint, privateKey string) {
	return "", ""
}
