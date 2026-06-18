package ktfunc

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestResolveWaitDuration_DefaultIsOneMinute(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	// No env, flag left at its default sentinel → the loop pauses one minute,
	// not the 5s block-poll interval the help text used to contradict.
	got := ResolveWaitDuration("", DefaultWaitDuration)
	assert.Equal(t, time.Minute, got)
	assert.Equal(t, time.Minute, DefaultWaitDuration, "DefaultWaitDuration must be 1 minute")
}

func TestResolveWaitDuration_FlagWins(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	got := ResolveWaitDuration("30s", 10*time.Second) // flag != default
	assert.Equal(t, 10*time.Second, got, "explicit flag should win over env")
}

func TestResolveWaitDuration_EnvUsedWhenFlagAtDefault(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	got := ResolveWaitDuration("90s", DefaultWaitDuration)
	assert.Equal(t, 90*time.Second, got, "env should apply when flag is left at default")
}

func TestResolveWaitDuration_InvalidEnvFallsBackToDefault(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	got := ResolveWaitDuration("not-a-duration", DefaultWaitDuration)
	assert.Equal(t, DefaultWaitDuration, got)
}

func TestBackoffDuration_NoErrorsReturnsBase(t *testing.T) {
	assert.Equal(t, 30*time.Second, BackoffDuration(30*time.Second, 0))
}

func TestBackoffDuration_GrowsThenCaps(t *testing.T) {
	base := 10 * time.Second
	assert.Equal(t, 20*time.Second, BackoffDuration(base, 1))
	assert.Equal(t, 40*time.Second, BackoffDuration(base, 2))
	assert.Equal(t, 80*time.Second, BackoffDuration(base, 3))

	// Enough doublings to exceed the cap → clamped at MaxBackoff, and it never
	// grows past the cap no matter how many failures accumulate.
	assert.Equal(t, MaxBackoff, BackoffDuration(base, 20))
	assert.Equal(t, MaxBackoff, BackoffDuration(base, 1000))
	assert.LessOrEqual(t, BackoffDuration(time.Minute, 50), MaxBackoff)
}

func TestBackoffDuration_NonPositiveBaseUsesDefault(t *testing.T) {
	assert.Equal(t, DefaultWaitDuration, BackoffDuration(0, 0))
}
