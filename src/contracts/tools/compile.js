// Compiles every contract under src/contracts with the project-local solc
// 0.8.16 (OpenZeppelin imports resolved from ./node_modules) and writes
// abi + creation bytecode per contract to src/abis/artifacts/artifacts.json. That file is
// checked in: the Go tests embed it to deploy real bytecode on a simulated
// chain, and abigen reads it to produce src/abis bindings.
// Run with: NODE_OPTIONS= npm run compile   (from src/contracts/tools)
const fs = require('fs');
const path = require('path');
const solc = require('solc');

const TOOLS = __dirname;
const CONTRACTS = path.resolve(TOOLS, '..');
const OUT = path.resolve(CONTRACTS, '..', 'abis', 'artifacts', 'artifacts.json');

// Only 0.8.16 contracts; TokenPrice.sol is 0.7.6 and built elsewhere.
const files = ['Ktv2.sol', 'Ktv2Factory.sol', 'BurnBankVault.sol', 'test/MockERC20.sol', 'test/StubTokenPrice.sol', 'test/ReentrantBank.sol'];
const sources = {};
for (const f of files) {
  const p = path.join(CONTRACTS, f);
  if (fs.existsSync(p)) sources[f] = { content: fs.readFileSync(p, 'utf8') };
}

function findImports(importPath) {
  for (const base of [path.join(TOOLS, 'node_modules'), CONTRACTS]) {
    const full = path.join(base, importPath);
    if (fs.existsSync(full)) return { contents: fs.readFileSync(full, 'utf8') };
  }
  return { error: 'File not found: ' + importPath };
}

const input = {
  language: 'Solidity',
  sources,
  settings: {
    optimizer: { enabled: true, runs: 200 },
    outputSelection: { '*': { '*': ['abi', 'evm.bytecode.object'] } },
  },
};

const out = JSON.parse(solc.compile(JSON.stringify(input), { import: findImports }));
if (out.errors) {
  const fatal = out.errors.filter((e) => e.severity === 'error');
  for (const e of out.errors) console.error(e.formattedMessage);
  if (fatal.length) { console.error(`Compilation failed with ${fatal.length} error(s).`); process.exit(1); }
}

const artifacts = {};
for (const [file, contracts] of Object.entries(out.contracts)) {
  if (!sources[file]) continue; // skip imported library contracts
  for (const [name, c] of Object.entries(contracts)) {
    if (!c.evm.bytecode.object) continue; // interfaces / abstract
    artifacts[name] = { abi: c.abi, bin: '0x' + c.evm.bytecode.object };
  }
}
fs.writeFileSync(OUT, JSON.stringify(artifacts, null, 2) + '\n');
console.log('Wrote', OUT);
for (const [name, a] of Object.entries(artifacts)) console.log(`  ${name}: ${a.abi.length} abi entries, ${(a.bin.length - 2) / 2} bytes`);
