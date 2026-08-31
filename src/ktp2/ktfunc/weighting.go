package ktfunc

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

// WeightingScheme identifies a stake-weighting curve used to turn epoch-minimum
// stakes into winner probabilities.
type WeightingScheme string

const (
	// SchemeSqrt weights each staker by the square root of their epoch-minimum
	// stake: prob_i = sqrt(stake_i) / sum(sqrt(stake_j)). This is the only
	// curve used for live winner selection.
	SchemeSqrt WeightingScheme = "sqrt"

	// SchemeLog weights by log(1+stake). It is no longer used for live
	// selection; it is retained so -verifyLastWinner can replay epochs that
	// were rewarded by builds that selected winners with log weights.
	SchemeLog WeightingScheme = "log"

	// VerifyWeightingFlagName and VerifyWeightingEnvVar name the CLI flag and
	// environment variable that pick which curve the -verifyLastWinner replay
	// uses. They have no effect on live voting.
	VerifyWeightingFlagName = "verifyWeighting"
	VerifyWeightingEnvVar   = "VERIFY_WEIGHTING"

	// DefaultVerifyWeighting matches the live curve, so a plain
	// -verifyLastWinner checks recent epochs against current behavior.
	DefaultVerifyWeighting = SchemeSqrt

	// sqrtWeightingSince names the first release whose live selection uses
	// sqrt weighting. Epochs rewarded by older builds were selected with log
	// weights and only replay correctly with SchemeLog.
	sqrtWeightingSince = "v0.5.0-beta"
)

// normalizeProbabilities fills in Prob for every entry of stakeDataMinsMap
// using the requested weighting curve. An unknown scheme is an error and
// leaves the map untouched: silently falling back to some default curve could
// make a replay disagree with the on-chain winner for the wrong reason.
func normalizeProbabilities(stakeDataMinsMap map[common.Address]*UserStakeData, scheme WeightingScheme) error {
	switch scheme {
	case SchemeSqrt:
		return sqrtNormalizeProbabilities(stakeDataMinsMap)
	case SchemeLog:
		return logNormalizeProbabilities(stakeDataMinsMap)
	default:
		return fmt.Errorf("unknown weighting scheme %q (valid: %q, %q)", scheme, SchemeSqrt, SchemeLog)
	}
}

// calculateProbsForEachWallet assigns a sqrt-normalized probability to each
// wallet in stakeDataMinsMap. Live selection supports exactly this one curve:
// a weighting switch on the live path would be a per-operator choice with no
// on-chain record, so different operators could silently compute different
// winners. The log curve still exists (weighting_log.go) but is reachable only
// from the -verifyLastWinner replay, which reads and never votes.
func calculateProbsForEachWallet(stakeDataMinsMap map[common.Address]*UserStakeData, totalMin *big.Int) bool {
	_ = totalMin // kept in signature for call-site symmetry; the curve doesn't use it.

	if stakeDataMinsMap == nil || len(stakeDataMinsMap) == 0 {
		log.Warn("Stake data map is nil or empty - Cannot calculate probabilities")
		return false
	}

	if err := normalizeProbabilities(stakeDataMinsMap, SchemeSqrt); err != nil {
		log.Errorf("Failed to normalize probabilities: %v", err)
		return false
	}

	foundSomething := false
	for addr, stakeData := range stakeDataMinsMap {
		if stakeData.Prob != nil {
			foundSomething = true
			log.Debugf("Address: %s, Sqrt-normalized Probability: %f\n", addr.Hex(), stakeData.Prob)
		}
	}
	return foundSomething
}

// ResolveVerifyWeighting decides which curve the -verifyLastWinner replay
// uses. Precedence: an explicit -verifyWeighting flag wins, then the
// VERIFY_WEIGHTING env value, otherwise DefaultVerifyWeighting. Invalid values
// from either source are an error, never a silent fallback.
//
// flagVal is compared against DefaultVerifyWeighting to tell "left at the
// default" from "explicitly set": when the flag still holds the default, the
// env value (if any) is allowed to take effect.
func ResolveVerifyWeighting(envVal, flagVal string) (WeightingScheme, error) {
	scheme := DefaultVerifyWeighting

	if envVal != "" {
		parsed, err := parseWeightingScheme(envVal)
		if err != nil {
			return "", fmt.Errorf("invalid %s: %w", VerifyWeightingEnvVar, err)
		}
		scheme = parsed
	}

	if flagVal != string(DefaultVerifyWeighting) {
		parsed, err := parseWeightingScheme(flagVal)
		if err != nil {
			return "", fmt.Errorf("invalid -%s: %w", VerifyWeightingFlagName, err)
		}
		scheme = parsed
	}

	return scheme, nil
}

func parseWeightingScheme(s string) (WeightingScheme, error) {
	switch WeightingScheme(s) {
	case SchemeSqrt:
		return SchemeSqrt, nil
	case SchemeLog:
		return SchemeLog, nil
	default:
		return "", fmt.Errorf("unknown weighting scheme %q (valid: %q, %q)", s, SchemeSqrt, SchemeLog)
	}
}
