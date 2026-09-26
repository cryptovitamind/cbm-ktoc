package integration_test

// End-to-end tests for the BurnBankVault contract and the battles session
// layer, on a simulated go-ethereum backend with real bytecode: a real Ktv2
// Burn Bank receives the unlocked ETH through its give() function.

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"ktp2/src/abis/artifacts"
	"ktp2/src/battles/vault"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

const chainID = 1337

type acct struct {
	key  *ecdsa.PrivateKey
	addr common.Address
}

type world struct {
	t       *testing.T
	backend *simulated.Backend
	client  simulated.Client
	owner   acct           // deploys everything
	donor   acct           // funds vaults
	matcher acct           // fulfills challenges
	bankA   common.Address // real Ktv2
	bankB   common.Address // real Ktv2
	dest    common.Address // Ktv2 donation destination
	stop    func()
	mineMu  sync.Mutex
}

func newAcct(t *testing.T, seed byte) acct {
	t.Helper()
	b := make([]byte, 32)
	b[31] = seed
	k, err := crypto.ToECDSA(b)
	if err != nil {
		t.Fatalf("key: %v", err)
	}
	return acct{key: k, addr: crypto.PubkeyToAddress(k.PublicKey)}
}

func eth(n int64) *big.Int { return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e18)) }

func setup(t *testing.T) *world {
	t.Helper()
	w := &world{t: t, owner: newAcct(t, 1), donor: newAcct(t, 2), matcher: newAcct(t, 3)}
	w.dest = newAcct(t, 9).addr
	alloc := types.GenesisAlloc{}
	for _, a := range []acct{w.owner, w.donor, w.matcher} {
		alloc[a.addr] = types.Account{Balance: eth(100)}
	}
	w.backend = simulated.NewBackend(alloc)
	w.client = w.backend.Client()
	t.Cleanup(func() { _ = w.backend.Close() })

	token := w.deploy(artifacts.MockERC20, w.owner)
	tp := w.deploy(artifacts.StubTokenPrice, w.owner)
	pool := common.HexToAddress("0x00000000000000000000000000000000000000a1")
	w.bankA = w.deploy(artifacts.Ktv2, w.owner, w.owner.addr, token, w.dest, pool, w.owner.addr, tp, false)
	w.bankB = w.deploy(artifacts.Ktv2, w.owner, w.owner.addr, token, w.dest, pool, w.owner.addr, tp, false)

	// Mine continuously so WaitMined in the session layer returns. The
	// simulated beacon's Commit is not safe to call from two goroutines, so
	// every other Commit in the tests goes through advance(), which holds
	// the same lock.
	done := make(chan struct{})
	var once sync.Once
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				w.mineMu.Lock()
				w.backend.Commit()
				w.mineMu.Unlock()
				time.Sleep(3 * time.Millisecond)
			}
		}
	}()
	w.stop = func() { once.Do(func() { close(done) }) }
	t.Cleanup(w.stop)
	return w
}

// advance mines n empty blocks without racing the miner goroutine.
func (w *world) advance(n int) {
	w.mineMu.Lock()
	defer w.mineMu.Unlock()
	for i := 0; i < n; i++ {
		w.backend.Commit()
	}
}

func (w *world) opts(a acct) *bind.TransactOpts {
	o, err := bind.NewKeyedTransactorWithChainID(a.key, big.NewInt(chainID))
	if err != nil {
		w.t.Fatalf("transactor: %v", err)
	}
	return o
}

func (w *world) deploy(name string, by acct, args ...interface{}) common.Address {
	w.t.Helper()
	art, err := artifacts.Load(name)
	if err != nil {
		w.t.Fatalf("artifact %s: %v", name, err)
	}
	addr, tx, _, err := bind.DeployContract(w.opts(by), art.ABI, art.Bin, w.client, args...)
	if err != nil {
		w.t.Fatalf("deploy %s: %v", name, err)
	}
	w.advance(1)
	if _, err := bind.WaitMined(context.Background(), w.client, tx); err != nil {
		w.t.Fatalf("mine %s: %v", name, err)
	}
	return addr
}

func (w *world) session(a acct, vaultAddr common.Address) *vault.Session {
	w.t.Helper()
	cfg := &vault.Config{EthEndpoint: "sim", PrivateKey: common.Bytes2Hex(crypto.FromECDSA(a.key)), VaultAddr: vaultAddr, ChainID: big.NewInt(chainID)}
	s, err := vault.NewSession(context.Background(), w.client, cfg)
	if err != nil {
		w.t.Fatalf("NewSession: %v", err)
	}
	s.TxTimeout = 20 * time.Second
	return s
}

func (w *world) deployVault(banks ...common.Address) common.Address {
	w.t.Helper()
	addr, rcpt, err := vault.Deploy(context.Background(), w.client, w.owner.key, big.NewInt(chainID), banks)
	if err != nil {
		w.t.Fatalf("Deploy: %v", err)
	}
	if rcpt.Status != types.ReceiptStatusSuccessful {
		w.t.Fatal("vault deploy reverted")
	}
	return addr
}

func (w *world) balance(a common.Address) *big.Int {
	b, err := w.client.BalanceAt(context.Background(), a, nil)
	if err != nil {
		w.t.Fatalf("balance: %v", err)
	}
	return b
}

func (w *world) newKey() vault.Key {
	k, err := vault.GenerateKey()
	if err != nil {
		w.t.Fatalf("GenerateKey: %v", err)
	}
	return k
}

// give sends a give() call from a to bank and returns its hash.
func (w *world) give(a acct, bank common.Address, amount *big.Int) common.Hash {
	w.t.Helper()
	art, _ := artifacts.Load(artifacts.Ktv2)
	c := bind.NewBoundContract(bank, art.ABI, w.client, w.client, w.client)
	o := w.opts(a)
	o.Value = amount
	tx, err := c.Transact(o, "give")
	if err != nil {
		w.t.Fatalf("give: %v", err)
	}
	if _, err := bind.WaitMined(context.Background(), w.client, tx); err != nil {
		w.t.Fatalf("mine give: %v", err)
	}
	return tx.Hash()
}

// cappedClient imitates a public RPC that refuses eth_getLogs over more than
// window blocks, the way free tiers do.
type cappedClient struct {
	simulated.Client
	window uint64
}

func (c *cappedClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if q.FromBlock != nil && q.ToBlock != nil && q.ToBlock.Uint64()-q.FromBlock.Uint64() > c.window {
		return nil, fmt.Errorf("query returned more than 10000 results; log limit exceeded")
	}
	if q.ToBlock == nil {
		return nil, fmt.Errorf("open-ended log query rejected; log limit")
	}
	return c.Client.FilterLogs(ctx, q)
}

func mustRevert(t *testing.T, err error, fragment string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected revert containing %q, got success", fragment)
	}
	if !strings.Contains(err.Error(), fragment) {
		t.Fatalf("expected revert containing %q, got: %v", fragment, err)
	}
}

// ---------------------------------------------------------------- tests

func TestDeploy_WhitelistsInitialBanks(t *testing.T) {
	w := setup(t)
	addr := w.deployVault(w.bankA, w.bankB)
	s := w.session(w.owner, addr)

	banks, err := s.Banks(context.Background())
	if err != nil {
		t.Fatalf("Banks: %v", err)
	}
	if len(banks) != 2 || banks[0] != w.bankA || banks[1] != w.bankB {
		t.Fatalf("banks = %v, want [%s %s]", banks, w.bankA.Hex(), w.bankB.Hex())
	}
	if ok, _ := s.IsBank(context.Background(), w.bankA); !ok {
		t.Fatal("bankA not whitelisted")
	}
	if ok, _ := s.IsBank(context.Background(), w.dest); ok {
		t.Fatal("random address reported as a bank")
	}
}

func TestNewSession_RejectsAddressWithoutCode(t *testing.T) {
	w := setup(t)
	cfg := &vault.Config{EthEndpoint: "sim", PrivateKey: common.Bytes2Hex(crypto.FromECDSA(w.owner.key)), VaultAddr: w.dest, ChainID: big.NewInt(chainID)}
	if _, err := vault.NewSession(context.Background(), w.client, cfg); err == nil {
		t.Fatal("expected an error for an address with no contract")
	}
}

func TestFundThenUnlock_PaysTheBankThroughGive(t *testing.T) {
	w := setup(t)
	addr := w.deployVault(w.bankA, w.bankB)
	donor := w.session(w.donor, addr)
	anyone := w.session(w.matcher, addr)
	ctx := context.Background()
	key := w.newKey()
	amount := big.NewInt(1e16)

	if _, err := donor.Fund(ctx, key.Hash(), amount); err != nil {
		t.Fatalf("Fund: %v", err)
	}
	st, err := donor.Status(ctx, key.Hash())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.Funder != w.donor.addr || st.Amount.Cmp(amount) != 0 || st.Spent || !st.Unlockable || st.FundedBlock == 0 {
		t.Fatalf("status after fund: %+v", st)
	}
	if got := w.balance(addr); got.Cmp(amount) != 0 {
		t.Fatalf("vault balance %s, want %s", got, amount)
	}

	bankBefore, destBefore := w.balance(w.bankB), w.balance(w.dest)
	rcpt, err := anyone.Unlock(ctx, key, w.bankB)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if rcpt.Status != types.ReceiptStatusSuccessful {
		t.Fatal("unlock reverted")
	}
	// Ktv2.give() donates donationPrc (50%) to dest and keeps the rest.
	half := new(big.Int).Div(amount, big.NewInt(2))
	if got := new(big.Int).Sub(w.balance(w.bankB), bankBefore); got.Cmp(half) != 0 {
		t.Fatalf("bank retained %s, want %s", got, half)
	}
	if got := new(big.Int).Sub(w.balance(w.dest), destBefore); got.Cmp(half) != 0 {
		t.Fatalf("dest received %s, want %s", got, half)
	}
	if got := w.balance(addr); got.Sign() != 0 {
		t.Fatalf("vault still holds %s", got)
	}

	st, _ = donor.Status(ctx, key.Hash())
	if !st.Spent || st.Unlockable {
		t.Fatalf("status after unlock: %+v", st)
	}
	ch, err := donor.Challenges(ctx, 0)
	if err != nil {
		t.Fatalf("Challenges: %v", err)
	}
	if len(ch) != 1 || ch[0].KeyHash != key.Hash() || !ch[0].Spent || ch[0].Bank != w.bankB || ch[0].Unlocker != w.matcher.addr || ch[0].UnlockBlock == 0 {
		t.Fatalf("challenges = %+v", ch)
	}
}

func TestFund_SameHashTwiceReverts(t *testing.T) {
	w := setup(t)
	addr := w.deployVault(w.bankA)
	donor := w.session(w.donor, addr)
	key := w.newKey()
	if _, err := donor.Fund(context.Background(), key.Hash(), big.NewInt(1e16)); err != nil {
		t.Fatalf("first Fund: %v", err)
	}
	_, err := donor.Fund(context.Background(), key.Hash(), big.NewInt(1e18))
	mustRevert(t, err, "Hash already funded")
	if got := w.balance(addr); got.Cmp(big.NewInt(1e16)) != 0 {
		t.Fatalf("top-up changed the vault balance to %s", got)
	}
}

func TestFund_ZeroValueReverts(t *testing.T) {
	w := setup(t)
	donor := w.session(w.donor, w.deployVault(w.bankA))
	_, err := donor.Fund(context.Background(), w.newKey().Hash(), big.NewInt(0))
	mustRevert(t, err, "No ETH")
}

func TestUnlock_Rejections(t *testing.T) {
	w := setup(t)
	addr := w.deployVault(w.bankA)
	donor := w.session(w.donor, addr)
	ctx := context.Background()
	key := w.newKey()
	if _, err := donor.Fund(ctx, key.Hash(), big.NewInt(1e16)); err != nil {
		t.Fatalf("Fund: %v", err)
	}

	_, err := donor.Unlock(ctx, w.newKey(), w.bankA)
	mustRevert(t, err, "Unknown key")

	_, err = donor.Unlock(ctx, key, w.bankB) // bankB not whitelisted on this vault
	mustRevert(t, err, "Not a bank")

	if _, err := donor.Unlock(ctx, key, w.bankA); err != nil {
		t.Fatalf("legit Unlock: %v", err)
	}
	_, err = donor.Unlock(ctx, key, w.bankA)
	mustRevert(t, err, "Already unlocked")
}

func TestVault_RefusesPlainTransfers(t *testing.T) {
	w := setup(t)
	addr := w.deployVault(w.bankA)
	ctx := context.Background()
	nonce, _ := w.client.PendingNonceAt(ctx, w.donor.addr)
	gp, _ := w.client.SuggestGasPrice(ctx)
	tx := types.NewTransaction(nonce, addr, big.NewInt(1e16), 100000, gp, nil)
	signed, _ := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(chainID)), w.donor.key)
	if err := w.client.SendTransaction(ctx, signed); err != nil {
		// Simulated backend may reject at gas estimation; that is also a refusal.
		mustRevert(t, err, "Use fund")
		return
	}
	rcpt, err := bind.WaitMined(ctx, w.client, signed)
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if rcpt.Status == types.ReceiptStatusSuccessful {
		t.Fatal("plain transfer to the vault must revert")
	}
	if got := w.balance(addr); got.Sign() != 0 {
		t.Fatalf("vault accepted stray ETH: %s", got)
	}
}

func TestBanks_OwnerAddsAndRemoves_NonOwnerCannot(t *testing.T) {
	w := setup(t)
	addr := w.deployVault(w.bankA)
	owner := w.session(w.owner, addr)
	donor := w.session(w.donor, addr)
	ctx := context.Background()

	if _, err := owner.AddBank(ctx, w.bankB); err != nil {
		t.Fatalf("AddBank: %v", err)
	}
	banks, _ := owner.Banks(ctx)
	if len(banks) != 2 {
		t.Fatalf("banks after add = %v", banks)
	}
	if _, err := owner.RemoveBank(ctx, w.bankA); err != nil {
		t.Fatalf("RemoveBank: %v", err)
	}
	banks, _ = owner.Banks(ctx)
	if len(banks) != 1 || banks[0] != w.bankB {
		t.Fatalf("banks after remove = %v", banks)
	}
	_, err := donor.AddBank(ctx, w.dest)
	mustRevert(t, err, "Ownable: caller is not the owner")
	_, err = owner.AddBank(ctx, w.bankB)
	mustRevert(t, err, "Already a bank")
}

func TestRemovedBank_CannotReceiveUnlock(t *testing.T) {
	w := setup(t)
	addr := w.deployVault(w.bankA, w.bankB)
	owner := w.session(w.owner, addr)
	ctx := context.Background()
	key := w.newKey()
	if _, err := owner.Fund(ctx, key.Hash(), big.NewInt(1e16)); err != nil {
		t.Fatalf("Fund: %v", err)
	}
	if _, err := owner.RemoveBank(ctx, w.bankA); err != nil {
		t.Fatalf("RemoveBank: %v", err)
	}
	_, err := owner.Unlock(ctx, key, w.bankA)
	mustRevert(t, err, "Not a bank")
	if _, err := owner.Unlock(ctx, key, w.bankB); err != nil {
		t.Fatalf("unlock to the remaining bank: %v", err)
	}
}

func TestUnlock_ReentrantBankCannotDoubleSpend(t *testing.T) {
	w := setup(t)
	evil := w.deploy("ReentrantBank", w.owner)
	addr := w.deployVault(evil)
	owner := w.session(w.owner, addr)
	ctx := context.Background()
	key := w.newKey()
	amount := big.NewInt(1e16)
	if _, err := owner.Fund(ctx, key.Hash(), amount); err != nil {
		t.Fatalf("Fund: %v", err)
	}

	art, _ := artifacts.Load("ReentrantBank")
	c := bind.NewBoundContract(evil, art.ABI, w.client, w.client, w.client)
	tx, err := c.Transact(w.opts(w.owner), "arm", addr, [32]byte(key))
	if err != nil {
		t.Fatalf("arm: %v", err)
	}
	bind.WaitMined(ctx, w.client, tx)

	if _, err := owner.Unlock(ctx, key, evil); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if got := w.balance(evil); got.Cmp(amount) != 0 {
		t.Fatalf("evil bank holds %s, want exactly one payout of %s", got, amount)
	}
	var out []interface{}
	if err := c.Call(nil, &out, "reentryReverted"); err != nil {
		t.Fatalf("reentryReverted: %v", err)
	}
	if reverted, _ := out[0].(bool); !reverted {
		t.Fatal("nested unlock from inside give() did not revert")
	}
}

func TestVerifyMatch_EndToEnd(t *testing.T) {
	w := setup(t)
	addr := w.deployVault(w.bankA)
	site := w.session(w.owner, addr)
	ctx := context.Background()
	min := big.NewInt(1e16)

	// A give() that landed BEFORE the vault was funded must not count.
	early := w.give(w.matcher, w.bankA, min)

	key := w.newKey()
	if _, err := site.Fund(ctx, key.Hash(), min); err != nil {
		t.Fatalf("Fund: %v", err)
	}

	good := w.give(w.matcher, w.bankA, min)
	small := w.give(w.matcher, w.bankA, big.NewInt(1e15))

	r, err := site.VerifyMatch(ctx, good, key.Hash())
	if err != nil {
		t.Fatalf("VerifyMatch: %v", err)
	}
	if !r.OK || r.Sender != w.matcher.addr || r.Bank != w.bankA || r.Amount.Cmp(min) != 0 {
		t.Fatalf("good match rejected: %+v", r)
	}
	if r, _ := site.VerifyMatch(ctx, early, key.Hash()); r.OK || r.Reason != vault.MatchTooEarly {
		t.Fatalf("early give accepted: %+v", r)
	}
	if r, _ := site.VerifyMatch(ctx, small, key.Hash()); r.OK || r.Reason != vault.MatchTooSmall {
		t.Fatalf("small give accepted: %+v", r)
	}
	if _, err := site.VerifyMatch(ctx, common.HexToHash("0xdead"), key.Hash()); err == nil {
		t.Fatal("unknown tx hash must error")
	}
	if _, err := site.VerifyMatch(ctx, good, w.newKey().Hash()); err == nil {
		t.Fatal("unknown challenge must error")
	}
}

// Public RPCs cap eth_getLogs ranges. The session must scan from the vault's
// deploy block in chunks and shrink the chunk when the provider refuses it,
// so banks and challenges keep working on a free-tier endpoint months later.
func TestScans_SurviveProviderLogWindowLimit(t *testing.T) {
	w := setup(t)
	addr := w.deployVault(w.bankA, w.bankB)
	// Put some history behind the vault, then more after funding.
	w.advance(60)
	owner := w.session(w.owner, addr)
	ctx := context.Background()
	key := w.newKey()
	if _, err := owner.Fund(ctx, key.Hash(), big.NewInt(1e16)); err != nil {
		t.Fatalf("Fund: %v", err)
	}
	w.advance(60)

	capped := &cappedClient{Client: w.client, window: 25}
	cfg := &vault.Config{EthEndpoint: "sim", PrivateKey: common.Bytes2Hex(crypto.FromECDSA(w.owner.key)), VaultAddr: addr, ChainID: big.NewInt(chainID), VaultBlock: owner.DeployBlock}
	s, err := vault.NewSession(ctx, capped, cfg)
	if err != nil {
		t.Fatalf("NewSession over capped client: %v", err)
	}
	s.ScanChunk = 100 // larger than the window: the halving must kick in

	banks, err := s.Banks(ctx)
	if err != nil {
		t.Fatalf("Banks over capped client: %v", err)
	}
	if len(banks) != 2 {
		t.Fatalf("banks = %v", banks)
	}
	ch, err := s.Challenges(ctx, 0)
	if err != nil {
		t.Fatalf("Challenges over capped client: %v", err)
	}
	if len(ch) != 1 || ch[0].KeyHash != key.Hash() {
		t.Fatalf("challenges = %+v", ch)
	}
}

// The deploy block is what scans start from; NewSession must find it when
// the config lacks it, and a session must expose it so the CLI can save it.
func TestNewSession_FindsDeployBlockWhenUnset(t *testing.T) {
	w := setup(t)
	addr, rcpt, err := vault.Deploy(context.Background(), w.client, w.owner.key, big.NewInt(chainID), []common.Address{w.bankA})
	if err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	w.advance(30)
	s := w.session(w.owner, addr) // config has VaultBlock 0
	if s.DeployBlock != rcpt.BlockNumber.Uint64() {
		t.Fatalf("DeployBlock = %d, want %d", s.DeployBlock, rcpt.BlockNumber.Uint64())
	}
}

func TestNewSession_RejectsWrongChainID(t *testing.T) {
	w := setup(t)
	addr := w.deployVault(w.bankA)
	cfg := &vault.Config{EthEndpoint: "sim", PrivateKey: common.Bytes2Hex(crypto.FromECDSA(w.owner.key)), VaultAddr: addr, ChainID: big.NewInt(1)}
	_, err := vault.NewSession(context.Background(), w.client, cfg)
	if err == nil || !strings.Contains(err.Error(), "chain") {
		t.Fatalf("expected a chain ID mismatch error, got %v", err)
	}
}
