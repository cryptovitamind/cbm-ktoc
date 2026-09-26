package vault

import (
	"bytes"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

const (
	bankA  = "0x1111111111111111111111111111111111111111"
	bankB  = "0x2222222222222222222222222222222222222222"
	vaultX = "0x3333333333333333333333333333333333333333"
)

func probeOK(chain int64) func(string) (*big.Int, error) {
	return func(string) (*big.Int, error) { return big.NewInt(chain), nil }
}

func TestRunWizard_ExistingVault(t *testing.T) {
	in := strings.NewReader(strings.Join([]string{
		"https://rpc.example", // endpoint
		testPrivKey,           // private key
		vaultX,                // existing vault
		"Y",                   // confirm
	}, "\n") + "\n")
	var out bytes.Buffer

	cfg, plan, err := RunWizard(in, &out, WizardOptions{Probe: probeOK(1)})
	if err != nil {
		t.Fatalf("RunWizard: %v", err)
	}
	if plan != nil {
		t.Fatal("no deploy plan expected when an address was given")
	}
	if cfg.EthEndpoint != "https://rpc.example" || cfg.PrivateKey != testPrivKey || cfg.VaultAddr != common.HexToAddress(vaultX) || cfg.ChainID.Int64() != 1 {
		t.Fatalf("config mismatch: %+v", cfg)
	}
	for _, p := range []string{PromptEndpoint, PromptPrivateKey, PromptVaultMode, PromptConfirm} {
		if !strings.Contains(out.String(), p) {
			t.Errorf("wizard output lacks prompt %q", p)
		}
	}
	if strings.Contains(out.String(), testPrivKey) {
		t.Error("wizard echoed the private key to output")
	}
}

func TestRunWizard_DeployPath(t *testing.T) {
	in := strings.NewReader(strings.Join([]string{
		"https://rpc.example",
		testPrivKey,
		VaultModeDeploy,
		bankA + ", " + bankB,
		"", // accept default confirm (Y)
	}, "\n") + "\n")
	var out bytes.Buffer

	cfg, plan, err := RunWizard(in, &out, WizardOptions{Probe: probeOK(1337)})
	if err != nil {
		t.Fatalf("RunWizard: %v", err)
	}
	if plan == nil {
		t.Fatal("expected a deploy plan")
	}
	if len(plan.Banks) != 2 || plan.Banks[0] != common.HexToAddress(bankA) || plan.Banks[1] != common.HexToAddress(bankB) {
		t.Fatalf("banks = %v", plan.Banks)
	}
	if cfg.VaultAddr != (common.Address{}) {
		t.Fatal("vault address must stay unset until the caller deploys")
	}
	if !strings.Contains(out.String(), PromptBanks) {
		t.Error("deploy path must ask for banks")
	}
}

func TestRunWizard_DefaultsAreOfferedAndAccepted(t *testing.T) {
	in := strings.NewReader("\n\n" + vaultX + "\n\n") // accept endpoint + key defaults
	var out bytes.Buffer

	cfg, _, err := RunWizard(in, &out, WizardOptions{
		DefaultEndpoint:   "https://from-ktoc.example",
		DefaultPrivateKey: testPrivKey,
		Probe:             probeOK(1),
	})
	if err != nil {
		t.Fatalf("RunWizard: %v", err)
	}
	if cfg.EthEndpoint != "https://from-ktoc.example" || cfg.PrivateKey != testPrivKey {
		t.Fatalf("defaults not applied: %+v", cfg)
	}
	if !strings.Contains(out.String(), "https://from-ktoc.example") {
		t.Error("default endpoint must be shown in the prompt")
	}
	if strings.Contains(out.String(), testPrivKey) {
		t.Error("default private key must not be shown in full")
	}
}

func TestRunWizard_ReprompsOnBadInput(t *testing.T) {
	in := strings.NewReader(strings.Join([]string{
		"",                    // endpoint: empty, must re-ask
		"https://rpc.example", // endpoint
		"not-a-key",           // bad key, must re-ask
		testPrivKey,
		"0xnope", // bad vault address, must re-ask
		vaultX,
		"Y",
	}, "\n") + "\n")
	var out bytes.Buffer

	cfg, _, err := RunWizard(in, &out, WizardOptions{Probe: probeOK(1)})
	if err != nil {
		t.Fatalf("RunWizard: %v", err)
	}
	if cfg.VaultAddr != common.HexToAddress(vaultX) {
		t.Fatalf("vault = %s", cfg.VaultAddr.Hex())
	}
	if strings.Count(out.String(), PromptEndpoint) < 2 || strings.Count(out.String(), PromptPrivateKey) < 2 || strings.Count(out.String(), PromptVaultMode) < 2 {
		t.Errorf("wizard did not re-prompt after bad input:\n%s", out.String())
	}
}

func TestRunWizard_ProbeFailureReprompts(t *testing.T) {
	calls := 0
	probe := func(ep string) (*big.Int, error) {
		calls++
		if ep == "https://bad.example" {
			return nil, errors.New("connection refused")
		}
		return big.NewInt(1), nil
	}
	in := strings.NewReader(strings.Join([]string{"https://bad.example", "https://good.example", testPrivKey, vaultX, "Y"}, "\n") + "\n")
	var out bytes.Buffer

	cfg, _, err := RunWizard(in, &out, WizardOptions{Probe: probe})
	if err != nil {
		t.Fatalf("RunWizard: %v", err)
	}
	if cfg.EthEndpoint != "https://good.example" || calls != 2 {
		t.Fatalf("endpoint %q after %d probes", cfg.EthEndpoint, calls)
	}
	if !strings.Contains(out.String(), "connection refused") {
		t.Error("probe error must be shown to the operator")
	}
}

func TestRunWizard_DeclineConfirmReturnsError(t *testing.T) {
	in := strings.NewReader(strings.Join([]string{"https://rpc.example", testPrivKey, vaultX, "n"}, "\n") + "\n")
	_, _, err := RunWizard(in, &bytes.Buffer{}, WizardOptions{Probe: probeOK(1)})
	if err == nil {
		t.Fatal("declining the confirmation must not produce a config")
	}
}

func TestRunWizard_ReadSecretIsUsedForTheKey(t *testing.T) {
	in := strings.NewReader(strings.Join([]string{"https://rpc.example", vaultX, "Y"}, "\n") + "\n")
	secret := func() (string, error) { return testPrivKey, nil }
	cfg, _, err := RunWizard(in, &bytes.Buffer{}, WizardOptions{Probe: probeOK(1), ReadSecret: secret})
	if err != nil {
		t.Fatalf("RunWizard: %v", err)
	}
	if cfg.PrivateKey != testPrivKey {
		t.Fatalf("key from ReadSecret not used: %q", cfg.PrivateKey)
	}
}

func TestImportKtocEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	os.WriteFile(path, []byte("ETH_ENDPOINT=https://ktoc.example\nMY_PRIVATE_KEY="+testPrivKey+"\nKT_ADDR=0x1\n"), 0o600)

	ep, key := ImportKtocEnv(path)
	if ep != "https://ktoc.example" || key != testPrivKey {
		t.Fatalf("ImportKtocEnv = %q, %q", ep, key)
	}
	if ep, key := ImportKtocEnv(filepath.Join(dir, "missing")); ep != "" || key != "" {
		t.Fatalf("missing file must yield empty strings, got %q %q", ep, key)
	}
}

// On a real terminal the key is read without echo; pressing Enter must accept
// the imported ktoc default, which the prompt shows masked.
func TestRunWizard_ReadSecretEmptyAcceptsDefault(t *testing.T) {
	in := strings.NewReader(strings.Join([]string{"https://rpc.example", vaultX, "Y"}, "\n") + "\n")
	var out bytes.Buffer
	secret := func() (string, error) { return "", nil }
	cfg, _, err := RunWizard(in, &out, WizardOptions{Probe: probeOK(1), ReadSecret: secret, DefaultPrivateKey: testPrivKey})
	if err != nil {
		t.Fatalf("RunWizard: %v", err)
	}
	if cfg.PrivateKey != testPrivKey {
		t.Fatalf("empty secret must accept the default key, got %q", cfg.PrivateKey)
	}
	if !strings.Contains(out.String(), testPrivKey[:4]) || strings.Contains(out.String(), testPrivKey) {
		t.Error("the prompt must show the default masked, never in full")
	}
}
