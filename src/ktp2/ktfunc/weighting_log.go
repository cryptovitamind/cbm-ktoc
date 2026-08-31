package ktfunc

// The log weighting curve. No longer used for live winner selection (that is
// sqrt, see weighting_sqrt.go); retained so -verifyLastWinner can replay
// epochs that were rewarded by builds that selected winners with log weights.
// The arithmetic below must stay byte-for-byte what those builds shipped, or
// replays of old epochs stop reproducing their winners.

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// logNormalizeProbabilities applies log-scale normalization to the
// probabilities in stakeDataMinsMap: prob_i = log(1+stake_i) / sum(log(1+stake_j)).
//
// Larger stakes still win more often, but the log compresses disparities so
// a wallet with 1000x more stake is only modestly more likely to win, not
// 1000x. This keeps whales from drowning out smaller stakers.
//
// We use log1p (log(1+x)) instead of log(x) so a 1-wei stake gets a small
// but non-zero probability instead of being silently excluded (log(1)=0).
func logNormalizeProbabilities(stakeDataMinsMap map[common.Address]*UserStakeData) error {
	if stakeDataMinsMap == nil || len(stakeDataMinsMap) == 0 {
		return nil
	}

	sumLog := new(big.Float)
	validCount := 0

	// First pass: compute sum of log(1+stake) for valid stakes
	for _, stakeData := range stakeDataMinsMap {
		if stakeData.StakeAmount == nil || stakeData.StakeAmount.Cmp(big.NewInt(0)) <= 0 {
			stakeData.Prob = new(big.Float).SetFloat64(0)
			continue
		}
		stakeFloat := new(big.Float).SetInt(stakeData.StakeAmount)
		stakeF64, _ := stakeFloat.Float64()
		logStake := math.Log1p(stakeF64)
		logStakeBig := new(big.Float).SetFloat64(logStake)
		sumLog.Add(sumLog, logStakeBig)
		validCount++
	}

	if validCount == 0 || sumLog.Cmp(new(big.Float).SetFloat64(0)) == 0 {
		return nil
	}

	// Second pass: set probabilities
	for _, stakeData := range stakeDataMinsMap {
		if stakeData.StakeAmount == nil || stakeData.StakeAmount.Cmp(big.NewInt(0)) <= 0 {
			continue
		}
		stakeFloat := new(big.Float).SetInt(stakeData.StakeAmount)
		stakeF64, _ := stakeFloat.Float64()
		logStake := math.Log1p(stakeF64)
		logStakeBig := new(big.Float).SetFloat64(logStake)
		stakeData.Prob = new(big.Float).Quo(logStakeBig, sumLog)
	}

	return nil
}
