package ktfunc

import (
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
)

// weightingPrec is the big.Float mantissa precision, in bits, used for sqrt
// weights, their sum, and the final probabilities. 128 bits comfortably
// resolves 1-wei differences at wei scale (a raw stake needs ~90 bits;
// float64's 53 would silently collapse nearby whale stakes) while keeping the
// arithmetic cheap.
const weightingPrec uint = 128

// sqrtNormalizeProbabilities applies square-root normalization to the
// probabilities in stakeDataMinsMap: prob_i = sqrt(stake_i) / sum(sqrt(stake_j)).
//
// Larger stakes win proportionally more under sqrt than under log — a wallet
// with 4x the stake gets 2x the weight — while still diluting raw linear
// dominance (1,000,000x the stake buys 1,000x the weight, not 1,000,000x).
//
// Every step is deterministic across operators and runs, which matters
// because the winner draw walks cumulative probabilities and any drift is a
// consensus hazard:
//   - big.Float.Sqrt is correctly rounded pure Go, identical on every platform,
//     and reads the raw stake integer directly (no float64 truncation);
//   - the sum is accumulated in sorted-address order (the same ordering the
//     winner walk uses), so rounding is applied in a fixed sequence instead of
//     Go's randomized map order.
//
// Zero and nil stakes get probability exactly 0 and stay out of the sum.
func sqrtNormalizeProbabilities(stakeDataMinsMap map[common.Address]*UserStakeData) error {
	if stakeDataMinsMap == nil || len(stakeDataMinsMap) == 0 {
		return nil
	}

	addresses := make([]common.Address, 0, len(stakeDataMinsMap))
	for addr := range stakeDataMinsMap {
		addresses = append(addresses, addr)
	}
	sort.Slice(addresses, func(i, j int) bool { return addresses[i].Hex() < addresses[j].Hex() })

	weights := make(map[common.Address]*big.Float, len(addresses))
	sum := new(big.Float).SetPrec(weightingPrec)
	for _, addr := range addresses {
		stakeData := stakeDataMinsMap[addr]
		if stakeData.StakeAmount == nil || stakeData.StakeAmount.Cmp(big.NewInt(0)) <= 0 {
			stakeData.Prob = new(big.Float).SetFloat64(0)
			continue
		}
		w := new(big.Float).SetPrec(weightingPrec).SetInt(stakeData.StakeAmount)
		w.Sqrt(w)
		weights[addr] = w
		sum.Add(sum, w)
	}

	if sum.Sign() == 0 {
		return nil
	}

	for addr, w := range weights {
		stakeDataMinsMap[addr].Prob = new(big.Float).SetPrec(weightingPrec).Quo(w, sum)
	}

	return nil
}
