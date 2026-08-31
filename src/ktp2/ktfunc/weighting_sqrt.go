package ktfunc

import (
	"github.com/ethereum/go-ethereum/common"
)

// sqrtNormalizeProbabilities is not implemented yet. It temporarily delegates
// to the log curve so the package compiles while the sqrt tests
// (weighting_sqrt_test.go) pin the intended behavior.
func sqrtNormalizeProbabilities(stakeDataMinsMap map[common.Address]*UserStakeData) error {
	return logNormalizeProbabilities(stakeDataMinsMap)
}
