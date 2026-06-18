package ktfunc

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// MaxBackoff caps the loop pause when vote cycles fail repeatedly, so a node
// that keeps erroring (RPC outage, rate limit, stalled chain) backs off but
// never sleeps so long it stops trying.
const MaxBackoff = 5 * time.Minute

// BackoffDuration returns how long to pause after a vote cycle, given how many
// times in a row the cycle has failed. With zero consecutive errors it returns
// base (the normal loop interval); each additional failure doubles the wait,
// capped at MaxBackoff. This keeps a healthy node on its normal cadence while
// preventing a failing node from hammering the RPC every base interval.
func BackoffDuration(base time.Duration, consecutiveErrors int) time.Duration {
	if base <= 0 {
		base = DefaultWaitDuration
	}
	d := base
	for i := 0; i < consecutiveErrors; i++ {
		if d >= MaxBackoff/2 {
			return MaxBackoff
		}
		d *= 2
	}
	if d > MaxBackoff {
		return MaxBackoff
	}
	return d
}

// ResolveWaitDuration decides how long the continuous loop pauses between
// vote-cycle iterations. Precedence: an explicit -waitDuration flag wins, then
// the WAIT_DURATION env value, otherwise DefaultWaitDuration.
//
// flagWaitDuration is compared against DefaultWaitDuration to tell "left at the
// default" from "explicitly set": when the flag still holds the default, the
// env value (if any) is allowed to take effect.
func ResolveWaitDuration(envWaitDuration string, flagWaitDuration time.Duration) time.Duration {
	duration := DefaultWaitDuration

	if envWaitDuration != "" {
		if parsed, err := time.ParseDuration(envWaitDuration); err != nil {
			log.Warnf("Invalid WAIT_DURATION '%s': %v. Using default: %v", envWaitDuration, err, DefaultWaitDuration)
		} else {
			duration = parsed
		}
	}

	if flagWaitDuration != DefaultWaitDuration {
		duration = flagWaitDuration
	}

	return duration
}
