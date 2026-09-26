package integration_test

// The battles command surface, driven end to end on the simulated chain:
// first-run wizard that deploys, then fund / status / verify / unlock /
// challenges / banks as an operator would type them.

import (
	"bytes"
	"context"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"ktp2/src/battles/cli"
	"ktp2/src/battles/vault"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type cliWorld struct {
	*world
	cfgPath string
	deps    cli.Deps
}

func setupCLI(t *testing.T) *cliWorld {
	t.Helper()
	w := setup(t)
	dir := t.TempDir()
	return &cliWorld{
		world:   w,
		cfgPath: filepath.Join(dir, vault.ConfigFileName),
		deps: cli.Deps{
			Dial:       func(context.Context, string) (vault.Client, error) { return w.client, nil },
			ConfigPath: filepath.Join(dir, vault.ConfigFileName),
			Version:    "test",
		},
	}
}

// run executes one command with scripted stdin and returns exit code + output.
func (c *cliWorld) run(stdin string, args ...string) (int, string) {
	var out, errOut bytes.Buffer
	code := cli.Run(args, strings.NewReader(stdin), &out, &errOut, c.deps)
	return code, out.String() + errOut.String()
}

func privHex(a acct) string { return common.Bytes2Hex(crypto.FromECDSA(a.key)) }

var (
	keyRe  = regexp.MustCompile(`SECRET key : (0x[0-9a-f]{64})`)
	hashRe = regexp.MustCompile(`Public hash: (0x[0-9a-f]{64})`)
)

func TestCLI_FirstRunWizardDeploysAndWritesConfig(t *testing.T) {
	c := setupCLI(t)
	// Any command with no battles.env triggers the wizard. The operator
	// chooses "deploy" with bankA whitelisted.
	script := strings.Join([]string{"sim://node", privHex(c.owner), vault.VaultModeDeploy, c.bankA.Hex(), "Y"}, "\n") + "\n"
	code, out := c.run(script, cli.CmdBanks)
	if code != 0 {
		t.Fatalf("exit %d:\n%s", code, out)
	}
	if !strings.Contains(out, "Vault deployed at 0x") || !strings.Contains(out, "Whitelisted Burn Banks") || !strings.Contains(out, c.bankA.Hex()) {
		t.Fatalf("unexpected output:\n%s", out)
	}
	if strings.Contains(out, privHex(c.owner)) {
		t.Fatal("wizard output contains the private key")
	}
	cfg, err := vault.LoadConfig(c.cfgPath)
	if err != nil {
		t.Fatalf("battles.env not written: %v", err)
	}
	if cfg.VaultAddr == (common.Address{}) || cfg.ChainID.Int64() != chainID || cfg.PrivateKey != privHex(c.owner) || cfg.VaultBlock == 0 {
		t.Fatalf("config = %+v (VaultBlock must be recorded from the deploy receipt)", cfg)
	}
	info, _ := os.Stat(c.cfgPath)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("battles.env mode %o", info.Mode().Perm())
	}

	// Second run: no wizard, straight to the command.
	code, out = c.run("", cli.CmdBanks)
	if code != 0 || strings.Contains(out, vault.PromptEndpoint) {
		t.Fatalf("second run re-ran the wizard or failed (exit %d):\n%s", code, out)
	}
}

func TestCLI_ChallengeLifecycle(t *testing.T) {
	c := setupCLI(t)
	vaultAddr := c.deployVault(c.bankA, c.bankB)
	cfg := &vault.Config{EthEndpoint: "sim://node", PrivateKey: privHex(c.donor), VaultAddr: vaultAddr, ChainID: big.NewInt(chainID)}
	if err := cfg.Save(c.cfgPath); err != nil {
		t.Fatal(err)
	}

	// keygen sends nothing.
	code, out := c.run("", cli.CmdKeygen)
	if code != 0 || !keyRe.MatchString(out) || !hashRe.MatchString(out) {
		t.Fatalf("keygen (exit %d):\n%s", code, out)
	}

	// fund generates a key, prints it once, and locks the ETH.
	code, out = c.run("", cli.CmdFund, "-"+cli.FlagAmount, "0.01")
	if code != 0 {
		t.Fatalf("fund exit %d:\n%s", code, out)
	}
	key := keyRe.FindStringSubmatch(out)[1]
	hash := hashRe.FindStringSubmatch(out)[1]
	k, _ := vault.ParseKey(key)
	if k.Hash().Hex() != hash {
		t.Fatalf("printed hash %s does not match printed key", hash)
	}
	if got := c.balance(vaultAddr); got.Cmp(big.NewInt(1e16)) != 0 {
		t.Fatalf("vault holds %s after fund", got)
	}

	code, out = c.run("", cli.CmdStatus, "-"+cli.FlagHash, hash)
	if code != 0 || !strings.Contains(out, "unlockable") || !strings.Contains(out, "0.010000 ETH") {
		t.Fatalf("status (exit %d):\n%s", code, out)
	}

	// A participant matches; verify accepts it, and rejects a too-small one.
	good := c.give(c.matcher, c.bankA, big.NewInt(1e16))
	small := c.give(c.matcher, c.bankB, big.NewInt(1e15))
	code, out = c.run("", cli.CmdVerify, "-"+cli.FlagTx, good.Hex(), "-"+cli.FlagHash, hash)
	if code != 0 || !strings.Contains(out, "FULFILLED") || !strings.Contains(out, c.matcher.addr.Hex()) {
		t.Fatalf("verify good (exit %d):\n%s", code, out)
	}
	code, out = c.run("", cli.CmdVerify, "-"+cli.FlagTx, small.Hex(), "-"+cli.FlagHash, hash)
	if code != 0 || !strings.Contains(out, "NOT FULFILLED") || !strings.Contains(out, vault.MatchTooSmall) {
		t.Fatalf("verify small (exit %d):\n%s", code, out)
	}

	// Unlock refuses a non-whitelisted bank before sending anything, then
	// pays a whitelisted one.
	code, out = c.run("", cli.CmdUnlock, "-"+cli.FlagKey, key, "-"+cli.FlagBank, c.dest.Hex())
	if code == 0 || !strings.Contains(out, "not a whitelisted Burn Bank") {
		t.Fatalf("unlock to non-bank (exit %d):\n%s", code, out)
	}
	before := c.balance(c.bankB)
	code, out = c.run("", cli.CmdUnlock, "-"+cli.FlagKey, key, "-"+cli.FlagBank, c.bankB.Hex())
	if code != 0 || !strings.Contains(out, "Unlocked in block") {
		t.Fatalf("unlock (exit %d):\n%s", code, out)
	}
	if got := new(big.Int).Sub(c.balance(c.bankB), before); got.Cmp(big.NewInt(5e15)) != 0 {
		t.Fatalf("bankB gained %s, want half of 0.01 ETH", got)
	}

	// Second unlock is refused locally, before a transaction is built.
	code, out = c.run("", cli.CmdUnlock, "-"+cli.FlagKey, key, "-"+cli.FlagBank, c.bankB.Hex())
	if code == 0 || !strings.Contains(out, "already unlocked") {
		t.Fatalf("replay (exit %d):\n%s", code, out)
	}

	code, out = c.run("", cli.CmdChallenges)
	if code != 0 || !strings.Contains(out, hash) || !strings.Contains(out, "unlocked into "+c.bankB.Hex()) {
		t.Fatalf("challenges (exit %d):\n%s", code, out)
	}
}

func TestCLI_BankManagementRequiresOwner(t *testing.T) {
	c := setupCLI(t)
	vaultAddr := c.deployVault(c.bankA)
	owner := &vault.Config{EthEndpoint: "sim://node", PrivateKey: privHex(c.owner), VaultAddr: vaultAddr, ChainID: big.NewInt(chainID)}
	donor := &vault.Config{EthEndpoint: "sim://node", PrivateKey: privHex(c.donor), VaultAddr: vaultAddr, ChainID: big.NewInt(chainID)}

	donor.Save(c.cfgPath)
	code, out := c.run("", cli.CmdAddBank, "-"+cli.FlagBank, c.bankB.Hex())
	if code == 0 || !strings.Contains(out, "not the owner") {
		t.Fatalf("non-owner add-bank (exit %d):\n%s", code, out)
	}

	owner.Save(c.cfgPath)
	if code, out = c.run("", cli.CmdAddBank, "-"+cli.FlagBank, c.bankB.Hex()); code != 0 {
		t.Fatalf("add-bank (exit %d):\n%s", code, out)
	}
	if code, out = c.run("", cli.CmdRemoveBank, "-"+cli.FlagBank, c.bankA.Hex()); code != 0 {
		t.Fatalf("remove-bank (exit %d):\n%s", code, out)
	}
	code, out = c.run("", cli.CmdBanks)
	if code != 0 || strings.Contains(out, c.bankA.Hex()) || !strings.Contains(out, c.bankB.Hex()) {
		t.Fatalf("banks (exit %d):\n%s", code, out)
	}
}

func TestCLI_DeployCommandReplacesVaultInConfig(t *testing.T) {
	c := setupCLI(t)
	first := c.deployVault(c.bankA)
	cfg := &vault.Config{EthEndpoint: "sim://node", PrivateKey: privHex(c.owner), VaultAddr: first, ChainID: big.NewInt(chainID)}
	cfg.Save(c.cfgPath)

	code, out := c.run("", cli.CmdDeploy, "-"+cli.FlagBanks, c.bankA.Hex()+","+c.bankB.Hex())
	if code != 0 || !strings.Contains(out, "Vault deployed at") {
		t.Fatalf("deploy (exit %d):\n%s", code, out)
	}
	after, _ := vault.LoadConfig(c.cfgPath)
	if after.VaultAddr == first || after.VaultAddr == (common.Address{}) {
		t.Fatalf("config vault not replaced: %s", after.VaultAddr.Hex())
	}
	code, out = c.run("", cli.CmdBanks)
	if code != 0 || !strings.Contains(out, c.bankB.Hex()) {
		t.Fatalf("banks on new vault (exit %d):\n%s", code, out)
	}
}

func TestCLI_UsageAndBadInput(t *testing.T) {
	c := setupCLI(t)
	if code, out := c.run(""); code != 2 || !strings.Contains(out, "Usage: battles") {
		t.Fatalf("no args (exit %d):\n%s", code, out)
	}
	if code, out := c.run("", cli.CmdVersion); code != 0 || !strings.Contains(out, "battles test") {
		t.Fatalf("version (exit %d):\n%s", code, out)
	}
	cfg := &vault.Config{EthEndpoint: "sim://node", PrivateKey: privHex(c.owner), VaultAddr: c.deployVault(c.bankA), ChainID: big.NewInt(chainID)}
	cfg.Save(c.cfgPath)
	for _, args := range [][]string{
		{cli.CmdFund},
		{cli.CmdFund, "-" + cli.FlagAmount, "-1"},
		{cli.CmdStatus, "-" + cli.FlagHash, "nope"},
		{cli.CmdUnlock, "-" + cli.FlagKey, "abc", "-" + cli.FlagBank, c.bankA.Hex()},
		{"frobnicate"},
	} {
		if code, out := c.run("", args...); code == 0 {
			t.Errorf("%v succeeded:\n%s", args, out)
		}
	}
}
