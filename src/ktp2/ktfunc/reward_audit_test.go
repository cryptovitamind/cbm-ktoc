package ktfunc

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"ktp2/src/abis/ktv2"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// rwdTxFromKey signs a real rwd(to, amt) call to the contract from hexKey, so
// the audit can recover both the sender and the _amt argument exactly the way
// it does on-chain.
func rwdTxFromKey(t *testing.T, hexKey string, chainID *big.Int, kt, to common.Address, amt *big.Int) (*types.Transaction, common.Address) {
	t.Helper()
	priv, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		t.Fatalf("HexToECDSA: %v", err)
	}
	parsed, err := ktv2.Ktv2MetaData.GetAbi()
	if err != nil {
		t.Fatalf("GetAbi: %v", err)
	}
	data, err := parsed.Pack("rwd", to, amt)
	if err != nil {
		t.Fatalf("Pack rwd: %v", err)
	}
	tx := types.NewTransaction(0, kt, big.NewInt(0), 100000, big.NewInt(1e9), data)
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), priv)
	if err != nil {
		t.Fatalf("SignTx: %v", err)
	}
	return signed, crypto.PubkeyToAddress(priv.PublicKey)
}

func auditSetup() (*ConnectionProps, *MockEthClient, *MockKtv2) {
	cProps, mockClient, mockKt := feeReserveSetup()
	cProps.ChainID = big.NewInt(1337)
	cProps.ChunkSize = 500
	return cProps, mockClient, mockKt
}

// expectRewardState mocks the balance/tlOcFees reads at block−1 and block.
func expectRewardState(mockClient *MockEthClient, mockKt *MockKtv2, kt common.Address, block uint64, balBefore, tlBefore, balAfter, tlAfter *big.Int) {
	mockClient.On("BalanceAt", mock.Anything, kt, new(big.Int).SetUint64(block-1)).Return(balBefore, nil)
	mockKt.On("TlOcFees", callOptsAtBlock(block-1)).Return(tlBefore, nil)
	mockClient.On("BalanceAt", mock.Anything, kt, new(big.Int).SetUint64(block)).Return(balAfter, nil)
	mockKt.On("TlOcFees", callOptsAtBlock(block)).Return(tlAfter, nil)
}

func rwdEvent(winner common.Address, paid *big.Int, block uint64, tx *types.Transaction) *ktv2.Ktv2Rwd {
	return &ktv2.Ktv2Rwd{Arg0: winner, Arg1: paid, Raw: types.Log{BlockNumber: block, TxHash: tx.Hash()}}
}

var auditWinner = common.HexToAddress("0x4c886b860a54bf938c6c39109215a898759fd57f")

// Two rewards in range: a correct one (X passed balance − tlOcFees) and an
// old-build one (Y passed the whole balance while fees were owed). The
// overpayment is exactly the fees that were owed, and the reserve after Y's
// reward is underfunded.
func TestAuditRewards_AuditsEveryRewardInRange(t *testing.T) {
	cProps, mockClient, mockKt := auditSetup()
	kt := cProps.KtAddr
	txX, addrX := rwdTxFromKey(t, testKeyX, cProps.ChainID, kt, auditWinner, big.NewInt(9e17))
	txY, addrY := rwdTxFromKey(t, testKeyY, cProps.ChainID, kt, auditWinner, big.NewInt(1e18))

	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: []*ktv2.Ktv2Rwd{
		rwdEvent(auditWinner, big.NewInt(882e15), 115, txX), // 0.9 ETH less a 2% rwd fee
		rwdEvent(auditWinner, big.NewInt(98e16), 175, txY),
	}}, nil)
	mockClient.On("TransactionByHash", mock.Anything, txX.Hash()).Return(txX, false, nil)
	mockClient.On("TransactionByHash", mock.Anything, txY.Hash()).Return(txY, false, nil)
	expectRewardState(mockClient, mockKt, kt, 115, big.NewInt(1e18), big.NewInt(1e17), big.NewInt(118e15), big.NewInt(118e15))
	expectRewardState(mockClient, mockKt, kt, 175, big.NewInt(1e18), big.NewInt(1e17), big.NewInt(2e16), big.NewInt(12e16))

	audits, err := AuditRewards(cProps, 100, 200)
	if err != nil {
		t.Fatalf("AuditRewards: %v", err)
	}
	if len(audits) != 2 {
		t.Fatalf("expected 2 audits, got %d", len(audits))
	}

	ok := audits[0]
	assert.Equal(t, uint64(115), ok.Block)
	assert.Equal(t, addrX, ok.Caller)
	assert.Equal(t, auditWinner, ok.Winner)
	assert.Equal(t, RewardVerdictOK, ok.Verdict)
	assert.Equal(t, 0, ok.AmtArg.Cmp(big.NewInt(9e17)))
	assert.Equal(t, 0, ok.Expected.Cmp(big.NewInt(9e17)))
	assert.Equal(t, 0, ok.Overpaid.Sign())
	assert.False(t, ok.FullBalanceSend)
	assert.False(t, ok.After.Underfunded())

	bad := audits[1]
	assert.Equal(t, uint64(175), bad.Block)
	assert.Equal(t, addrY, bad.Caller)
	assert.Equal(t, RewardVerdictOverpaid, bad.Verdict)
	assert.Equal(t, 0, bad.AmtArg.Cmp(big.NewInt(1e18)))
	assert.Equal(t, 0, bad.Paid.Cmp(big.NewInt(98e16)))
	assert.Equal(t, 0, bad.Expected.Cmp(big.NewInt(9e17)))
	assert.Equal(t, 0, bad.Overpaid.Cmp(big.NewInt(1e17)), "overpaid by exactly the fees owed")
	assert.True(t, bad.FullBalanceSend, "_amt == whole balance is the pre-fee-reserve build fingerprint")
	assert.True(t, bad.After.Underfunded())
	assert.Equal(t, 0, bad.After.Deficit().Cmp(big.NewInt(1e17)))
}

// A give landing between the node's read and the reward being mined makes
// _amt smaller than balance − tlOcFees at block−1. That is harmless and must
// not be reported as an overpayment.
func TestAuditRewards_UnderpayIsNotOverpaid(t *testing.T) {
	cProps, mockClient, mockKt := auditSetup()
	kt := cProps.KtAddr
	tx, _ := rwdTxFromKey(t, testKeyX, cProps.ChainID, kt, auditWinner, big.NewInt(85e16))
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: []*ktv2.Ktv2Rwd{rwdEvent(auditWinner, big.NewInt(833e15), 115, tx)}}, nil)
	mockClient.On("TransactionByHash", mock.Anything, tx.Hash()).Return(tx, false, nil)
	expectRewardState(mockClient, mockKt, kt, 115, big.NewInt(1e18), big.NewInt(1e17), big.NewInt(167e15), big.NewInt(117e15))

	audits, err := AuditRewards(cProps, 100, 200)
	if err != nil || len(audits) != 1 {
		t.Fatalf("AuditRewards: %v (%d audits)", err, len(audits))
	}
	assert.Equal(t, RewardVerdictUnderpaid, audits[0].Verdict)
	assert.Equal(t, 0, audits[0].Overpaid.Sign())
	assert.False(t, audits[0].FullBalanceSend)
}

// The latest SHI reward: the reserve was already underfunded, so a current
// build correctly passed _amt = 0. Expected clamps to zero and that is OK.
func TestAuditRewards_ZeroRewardOnDeficitIsOK(t *testing.T) {
	cProps, mockClient, mockKt := auditSetup()
	kt := cProps.KtAddr
	balance, _ := new(big.Int).SetString("1621000000000000", 10)
	owed, _ := new(big.Int).SetString("4851000000000000", 10)
	tx, addrX := rwdTxFromKey(t, testKeyX, cProps.ChainID, kt, auditWinner, big.NewInt(0))
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: []*ktv2.Ktv2Rwd{rwdEvent(auditWinner, big.NewInt(0), 115, tx)}}, nil)
	mockClient.On("TransactionByHash", mock.Anything, tx.Hash()).Return(tx, false, nil)
	expectRewardState(mockClient, mockKt, kt, 115, balance, owed, balance, owed)

	audits, err := AuditRewards(cProps, 100, 200)
	if err != nil || len(audits) != 1 {
		t.Fatalf("AuditRewards: %v (%d audits)", err, len(audits))
	}
	assert.Equal(t, RewardVerdictOK, audits[0].Verdict)
	assert.Equal(t, addrX, audits[0].Caller)
	assert.Equal(t, 0, audits[0].Expected.Sign())
	assert.False(t, audits[0].FullBalanceSend, "a zero send is not a full-balance send")
	assert.True(t, audits[0].After.Underfunded(), "the pre-existing deficit is still reported")
}

// Without archive state the audit cannot know what should have been paid, but
// it must still name the caller and the amount passed, and say why it can't
// judge, rather than fail the whole run.
func TestAuditRewards_UnknownWhenHistoricStateUnavailable(t *testing.T) {
	cProps, mockClient, mockKt := auditSetup()
	kt := cProps.KtAddr
	tx, addrX := rwdTxFromKey(t, testKeyX, cProps.ChainID, kt, auditWinner, big.NewInt(1e18))
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: []*ktv2.Ktv2Rwd{rwdEvent(auditWinner, big.NewInt(98e16), 115, tx)}}, nil)
	mockClient.On("TransactionByHash", mock.Anything, tx.Hash()).Return(tx, false, nil)
	mockClient.On("BalanceAt", mock.Anything, kt, big.NewInt(114)).Return((*big.Int)(nil), errors.New("missing trie node"))

	audits, err := AuditRewards(cProps, 100, 200)
	if err != nil || len(audits) != 1 {
		t.Fatalf("AuditRewards: %v (%d audits)", err, len(audits))
	}
	a := audits[0]
	assert.Equal(t, RewardVerdictUnknown, a.Verdict)
	assert.Nil(t, a.Before)
	assert.Equal(t, addrX, a.Caller)
	assert.Equal(t, 0, a.AmtArg.Cmp(big.NewInt(1e18)))
	assert.Contains(t, a.Note, "missing trie node")
}

// A Rwd event whose transaction is not a plain rwd call (e.g. a contract
// wallet forwarding) cannot have its _amt decoded. Report it, don't crash.
func TestAuditRewards_UndecodableCalldataIsUnknown(t *testing.T) {
	cProps, mockClient, mockKt := auditSetup()
	tx, addrX := signedTxFromKey(t, testKeyX, cProps.ChainID) // no calldata
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{events: []*ktv2.Ktv2Rwd{rwdEvent(auditWinner, big.NewInt(1), 115, tx)}}, nil)
	mockClient.On("TransactionByHash", mock.Anything, tx.Hash()).Return(tx, false, nil)

	audits, err := AuditRewards(cProps, 100, 200)
	if err != nil || len(audits) != 1 {
		t.Fatalf("AuditRewards: %v (%d audits)", err, len(audits))
	}
	assert.Equal(t, RewardVerdictUnknown, audits[0].Verdict)
	assert.Equal(t, addrX, audits[0].Caller)
	assert.Contains(t, strings.ToLower(audits[0].Note), "calldata")
}

func TestResolveAuditRange_AllSpansCreationToHead(t *testing.T) {
	cProps, mockClient, _ := auditSetup()
	cProps.KtBlock = big.NewInt(24179279)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(25944796), nil)

	from, to, err := ResolveAuditRange(cProps, AuditRangeAll)
	if err != nil {
		t.Fatalf("ResolveAuditRange: %v", err)
	}
	assert.Equal(t, uint64(24179279), from)
	assert.Equal(t, uint64(25944796), to)
}

func TestResolveAuditRange_ParsesExplicitRange(t *testing.T) {
	cProps, _, _ := auditSetup()
	from, to, err := ResolveAuditRange(cProps, "100:200")
	if err != nil {
		t.Fatalf("ResolveAuditRange: %v", err)
	}
	assert.Equal(t, uint64(100), from)
	assert.Equal(t, uint64(200), to)
}

func TestResolveAuditRange_RejectsGarbage(t *testing.T) {
	cProps, _, _ := auditSetup()
	if _, _, err := ResolveAuditRange(cProps, "banana"); err == nil {
		t.Error("expected an error for an unparseable range")
	}
}

// Run-loop epoch-advance check.

func captureLogs(t *testing.T) *logtest.Hook {
	t.Helper()
	prev := logrus.GetLevel()
	logrus.SetLevel(logrus.InfoLevel)
	hook := logtest.NewGlobal()
	t.Cleanup(func() {
		hook.Reset()
		logrus.SetLevel(prev)
	})
	return hook
}

func hasEntry(hook *logtest.Hook, level logrus.Level, substrs ...string) bool {
	for _, e := range hook.AllEntries() {
		if e.Level != level {
			continue
		}
		match := true
		for _, s := range substrs {
			if !strings.Contains(e.Message, s) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// When startBlock has moved since the last cycle, the reward that moved it is
// audited and an overpayment is logged as an error naming the OC.
func TestCheckEpochAdvance_LogsErrorNamingOverpayingOC(t *testing.T) {
	hook := captureLogs(t)
	cProps, mockClient, mockKt := auditSetup()
	kt := cProps.KtAddr
	cProps.lastEpochStart = big.NewInt(50) // previous cycle saw epoch 50..110
	txY, addrY := rwdTxFromKey(t, testKeyY, cProps.ChainID, kt, auditWinner, big.NewInt(1e18))

	// The audit window is the old epoch end through the chain head, re-read
	// so a reward mined right after the loop's header read is not missed.
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(131), nil)
	mockKt.On("FilterRwd", mock.MatchedBy(func(o *bind.FilterOpts) bool {
		return o.Start == 110 && o.End != nil && *o.End == 131
	})).Return(&mockRwdIter{events: []*ktv2.Ktv2Rwd{rwdEvent(auditWinner, big.NewInt(98e16), 115, txY)}}, nil)
	mockClient.On("TransactionByHash", mock.Anything, txY.Hash()).Return(txY, false, nil)
	expectRewardState(mockClient, mockKt, kt, 115, big.NewInt(1e18), big.NewInt(1e17), big.NewInt(2e16), big.NewInt(12e16))
	mockClient.On("BalanceAt", mock.Anything, kt, (*big.Int)(nil)).Return(big.NewInt(2e16), nil).Maybe()
	mockKt.On("TlOcFees", callOptsLatest()).Return(big.NewInt(12e16), nil).Maybe()

	checkEpochAdvance(cProps, big.NewInt(110), 60, 130)

	assert.True(t, hasEntry(hook, logrus.ErrorLevel, addrY.Hex(), RewardVerdictOverpaid),
		"expected an ERROR naming the overpaying OC; got: %v", messages(hook))
	assert.True(t, hasEntry(hook, logrus.ErrorLevel, "0.100000"), "the overpaid amount is logged in ETH")
	assert.Equal(t, int64(110), cProps.lastEpochStart.Int64(), "the new epoch start is recorded")
}

func TestCheckEpochAdvance_FirstCycleOnlyRecordsEpoch(t *testing.T) {
	cProps, _, mockKt := auditSetup()

	checkEpochAdvance(cProps, big.NewInt(110), 60, 130)

	mockKt.AssertNotCalled(t, "FilterRwd", mock.Anything)
	if cProps.lastEpochStart == nil || cProps.lastEpochStart.Int64() != 110 {
		t.Fatalf("expected lastEpochStart recorded as 110, got %v", cProps.lastEpochStart)
	}
}

func TestCheckEpochAdvance_SameEpochDoesNotAudit(t *testing.T) {
	cProps, _, mockKt := auditSetup()
	cProps.lastEpochStart = big.NewInt(110)

	checkEpochAdvance(cProps, big.NewInt(110), 60, 130)

	mockKt.AssertNotCalled(t, "FilterRwd", mock.Anything)
}

// The advance check must run before VoteAndReward's "not time to vote yet"
// early return, or a healthy node idling in the new epoch never reports the
// reward that opened it.
func TestVoteAndReward_EpochAdvanceCheckRunsBeforeEpochEndWait(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	cProps, mockClient, mockKt := auditSetup()
	cProps.lastEpochStart = big.NewInt(50)

	mockClient.On("HeaderByNumber", mock.Anything, (*big.Int)(nil)).Return(&types.Header{Number: big.NewInt(120)}, nil)
	mockKt.On("StartBlock", mock.Anything).Return(big.NewInt(110), nil) // advanced; new epoch ends at 170
	mockKt.On("EpochInterval", mock.Anything).Return(uint16(60), nil)
	mockKt.On("FilterRwd", mock.Anything).Return(&mockRwdIter{}, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(120), nil).Maybe()
	mockClient.On("BalanceAt", mock.Anything, cProps.KtAddr, (*big.Int)(nil)).Return(big.NewInt(1e18), nil).Maybe()
	mockKt.On("TlOcFees", mock.Anything).Return(big.NewInt(0), nil).Maybe()

	err := VoteAndReward(cProps)

	assert.NoError(t, err, "not time to vote is not an error")
	mockKt.AssertCalled(t, "FilterRwd", mock.Anything)
}

func messages(hook *logtest.Hook) []string {
	out := make([]string, 0, len(hook.AllEntries()))
	for _, e := range hook.AllEntries() {
		out = append(out, e.Level.String()+": "+e.Message)
	}
	return out
}
