#!/bin/bash
# Compile src/contracts/*.sol to src/abis/artifacts/artifacts.json and regenerate the Go bindings
# in src/abis. Uses only project-local tooling: solc from tools/node_modules
# (npm install runs on first use) and abigen from the pinned go-ethereum module.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$HERE/../.." && pwd)"
GETH_VERSION="$(cd "$REPO" && go list -m -f '{{.Version}}' github.com/ethereum/go-ethereum)"

cd "$HERE/tools"
[ -d node_modules ] || NODE_OPTIONS= npm install --silent
NODE_OPTIONS= npm run --silent compile

# Bindings for the contracts the Go code talks to. Ktv2 and Ktv2Factory keep
# their existing hand-maintained bindings; only the Vault is generated here.
gen() { # <ContractName> <go package> <out dir>
  local name="$1" pkg="$2" dir="$REPO/src/abis/$3"
  mkdir -p "$dir"
  node -e "const a=require('$REPO/src/abis/artifacts/artifacts.json')['$name'];process.stdout.write(JSON.stringify(a.abi))" > "/tmp/$name.abi"
  node -e "const a=require('$REPO/src/abis/artifacts/artifacts.json')['$name'];process.stdout.write(a.bin.slice(2))" > "/tmp/$name.bin"
  (cd "$REPO" && go run "github.com/ethereum/go-ethereum/cmd/abigen@$GETH_VERSION" \
     --abi "/tmp/$name.abi" --bin "/tmp/$name.bin" --pkg "$pkg" --type "$name" --out "$dir/$name.go")
  echo "  bindings: src/abis/$3/$name.go"
}
NODE_OPTIONS= gen BurnBankVault vault vault
echo "Done."
