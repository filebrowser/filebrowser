package cnc

import (
	"runtime"
	"testing"
)

func TestParseKB(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"MemTotal:        4031164 kB", 4031164},
		{"MemAvailable:    1234567 kB", 1234567},
		{"NoColonValue", 0},
		{"OnlyKey:", 0},
		{"BadNumber: abc kB", 0},
	}
	for _, c := range cases {
		got := parseKB(c.in)
		if got != c.want {
			t.Fatalf("parseKB(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestReadHostStats_NoCrashOnNonLinux(t *testing.T) {
	// On non-Linux hosts /proc + /sys don't exist; every field should
	// just zero-out instead of panicking. CI runs Linux but devs may
	// be on macOS; this guard is cheap.
	stats := ReadHostStats()
	if runtime.GOOS == "linux" {
		// On Linux we expect at least one field populated. Don't pin
		// values since CI host differs from a Pi.
		if stats.Load1m < 0 {
			t.Fatalf("load1m negative: %v", stats.Load1m)
		}
		if stats.MemUsedPct < 0 || stats.MemUsedPct > 100 {
			t.Fatalf("mem pct out of [0,100]: %v", stats.MemUsedPct)
		}
		if stats.DiskUsedPct < 0 || stats.DiskUsedPct > 100 {
			t.Fatalf("disk pct out of [0,100]: %v", stats.DiskUsedPct)
		}
	}
}
