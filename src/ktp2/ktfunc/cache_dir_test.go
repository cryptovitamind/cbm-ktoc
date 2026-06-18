package ktfunc

import "testing"

func TestResolvedCacheDir_DefaultsAndOverride(t *testing.T) {
	if got := (&ConnectionProps{}).ResolvedCacheDir(); got != "cache" {
		t.Errorf("default cache dir = %q, want %q", got, "cache")
	}
	if got := (&ConnectionProps{CacheDir: "/tmp/node1"}).ResolvedCacheDir(); got != "/tmp/node1" {
		t.Errorf("override cache dir = %q, want %q", got, "/tmp/node1")
	}
}
