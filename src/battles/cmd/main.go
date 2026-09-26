// battles: command-line tool for the Burn Bank Battles Vault contract.
// All behavior lives in ktp2/src/battles/cli; this file only wires the
// terminal, the RPC dialer, and the version.
package main

import (
	"context"
	"fmt"
	"os"

	"ktp2/src/battles/cli"
	"ktp2/src/battles/vault"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mattn/go-colorable"
	"golang.org/x/term"
)

// bannerVersion identifies this build in `battles version`.
const bannerVersion = "v0.1.0-beta"

func main() {
	deps := cli.Deps{
		Dial: func(ctx context.Context, endpoint string) (vault.Client, error) {
			return ethclient.DialContext(ctx, endpoint)
		},
		Version: bannerVersion,
	}
	// Read the private key without echo when stdin is a terminal; a piped
	// stdin (scripts, tests) falls back to a plain line read.
	if term.IsTerminal(int(os.Stdin.Fd())) {
		deps.ReadSecret = func() (string, error) {
			b, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return "", fmt.Errorf("failed to read the key: %w", err)
			}
			return string(b), nil
		}
	}
	// colorable makes the ANSI colors in the banner and usage render on Windows.
	os.Exit(cli.Run(os.Args[1:], os.Stdin, colorable.NewColorableStdout(), colorable.NewColorableStderr(), deps))
}
