package version

import "testing"

func TestStringDev(t *testing.T) {
	want := "pcm dev (commit none, built unknown)"
	if got := String("pcm"); got != want {
		t.Errorf("String = %q, want %q", got, want)
	}
}

func TestStringSet(t *testing.T) {
	oldV, oldC, oldD := Version, Commit, Date
	t.Cleanup(func() { Version, Commit, Date = oldV, oldC, oldD })
	Version, Commit, Date = "0.1.0", "abc1234", "2026-10-29"

	want := "pcm-server 0.1.0 (commit abc1234, built 2026-10-29)"
	if got := String("pcm-server"); got != want {
		t.Errorf("String = %q, want %q", got, want)
	}
}
