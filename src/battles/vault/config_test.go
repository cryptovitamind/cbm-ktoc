package vault

import (
	"errors"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const testPrivKey = "fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d4977e62bc6535e9a"

func sampleConfig() *Config {
	return &Config{
		EthEndpoint: "http://127.0.0.1:8545",
		PrivateKey:  testPrivKey,
		VaultAddr:   common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ChainID:     big.NewInt(1337),
		VaultBlock:  4242,
	}
}

func TestConfig_SaveThenLoadRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), ConfigFileName)
	want := sampleConfig()
	if err := want.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.EthEndpoint != want.EthEndpoint || got.PrivateKey != want.PrivateKey || got.VaultAddr != want.VaultAddr || got.ChainID.Cmp(want.ChainID) != 0 || got.VaultBlock != want.VaultBlock {
		t.Fatalf("round trip mismatch: got %+v want %+v", got, want)
	}
}

// A CI job sets the BATTLES_* variables and has no file at all; that must
// load as a config, not trigger the wizard.
func TestLoadConfig_EnvOnlyWithoutFile(t *testing.T) {
	t.Setenv(EnvEndpoint, "https://ci.example")
	t.Setenv(EnvPrivateKey, testPrivKey)
	t.Setenv(EnvVaultAddr, "0x000000000000000000000000000000000000dEaD")
	t.Setenv(EnvChainID, "1")
	got, err := LoadConfig(filepath.Join(t.TempDir(), "absent.env"))
	if err != nil {
		t.Fatalf("LoadConfig with env only: %v", err)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("env-only config invalid: %v", err)
	}
	if got.EthEndpoint != "https://ci.example" || got.ChainID.Int64() != 1 {
		t.Fatalf("config = %+v", got)
	}
}

// The file holds a private key: it must never be group- or world-readable.
func TestConfig_SaveWritesOwnerOnlyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ConfigFileName)
	if err := sampleConfig().Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != configFileMode {
		t.Fatalf("file mode = %o, want %o", perm, configFileMode)
	}
}

func TestConfig_SaveUsesTheDocumentedEnvNames(t *testing.T) {
	path := filepath.Join(t.TempDir(), ConfigFileName)
	if err := sampleConfig().Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	raw, _ := os.ReadFile(path)
	for _, name := range []string{EnvEndpoint, EnvPrivateKey, EnvVaultAddr, EnvChainID, EnvVaultBlock} {
		if !containsLine(string(raw), name+"=") {
			t.Errorf("battles.env lacks a %s= line:\n%s", name, raw)
		}
	}
}

func TestLoadConfig_MissingFileIsNotExist(t *testing.T) {
	_, err := LoadConfig(filepath.Join(t.TempDir(), "nope.env"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected fs.ErrNotExist, got %v", err)
	}
}

// Environment variables override the file so CI can run without one.
func TestLoadConfig_EnvOverridesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ConfigFileName)
	if err := sampleConfig().Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	t.Setenv(EnvEndpoint, "https://override.example")
	t.Setenv(EnvVaultAddr, "0x000000000000000000000000000000000000dEaD")
	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got.EthEndpoint != "https://override.example" {
		t.Errorf("endpoint not overridden: %q", got.EthEndpoint)
	}
	if got.VaultAddr != common.HexToAddress("0x000000000000000000000000000000000000dEaD") {
		t.Errorf("vault addr not overridden: %s", got.VaultAddr.Hex())
	}
	if got.PrivateKey != testPrivKey {
		t.Errorf("unset env must not clobber the file value; got key %q", got.PrivateKey)
	}
}

func TestConfig_Validate(t *testing.T) {
	ok := sampleConfig()
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	cases := map[string]func(c *Config){
		"no endpoint":   func(c *Config) { c.EthEndpoint = "" },
		"no key":        func(c *Config) { c.PrivateKey = "" },
		"bad key":       func(c *Config) { c.PrivateKey = "zz" },
		"no vault":      func(c *Config) { c.VaultAddr = common.Address{} },
		"no chain":      func(c *Config) { c.ChainID = nil },
		"0x-prefix key": func(c *Config) { c.PrivateKey = "0x" + testPrivKey },
	}
	for name, mutate := range cases {
		c := sampleConfig()
		mutate(c)
		err := c.Validate()
		if name == "0x-prefix key" {
			if err != nil {
				t.Errorf("%s: a 0x-prefixed key must be accepted, got %v", name, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
}

func TestConfig_SignerDerivesAddress(t *testing.T) {
	c := sampleConfig()
	key, addr, err := c.Signer()
	if err != nil {
		t.Fatalf("Signer: %v", err)
	}
	want := crypto.PubkeyToAddress(key.PublicKey)
	if addr != want {
		t.Fatalf("Signer address %s != derived %s", addr.Hex(), want.Hex())
	}
	c.PrivateKey = "0x" + testPrivKey
	if _, addr2, err := c.Signer(); err != nil || addr2 != want {
		t.Fatalf("Signer with 0x prefix: %s, %v", addr2.Hex(), err)
	}
}

func containsLine(s, prefix string) bool {
	for _, line := range splitLines(s) {
		if len(line) >= len(prefix) && line[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
