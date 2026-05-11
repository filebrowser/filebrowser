package cnc

import (
	"reflect"
	"testing"
)

// parseFloatTail is the workhorse that turns a Q600 MACRO frame into a
// float. The frame format varies slightly across Haas firmware (with /
// without spaces around the commas) and the macro var itself may carry
// negatives or integers. We need all of these to round-trip cleanly.
func TestParseFloatTailFormats(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  float64
		ok    bool
	}{
		{"haas-with-spaces", "MACRO, 2001, 3.5400", 3.54, true},
		{"haas-no-spaces", "MACRO,2001,3.5400", 3.54, true},
		{"zero", "MACRO, 5021, 0.0000", 0, true},
		{"integer-tail", "MACRO, 3022, 42", 42, true},
		{"negative", "MACRO, 5041, -2.1234", -2.1234, true},
		{"long-decimal", "MACRO, 5023, 12.123456", 12.123456, true},
		{"trailing-spaces", "MACRO, 3122, 17  ", 17, true},
		{"empty", "", 0, false},
		{"only-prefix", "MACRO", 0, false},
		{"missing-trailing-num", "MACRO, 5021,", 0, false},
		{"non-numeric-tail", "MACRO, 5021, NA", 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v, ok := parseFloatTail(c.value)
			if ok != c.ok {
				t.Fatalf("ok mismatch: got %v want %v (value=%q)", ok, c.ok, c.value)
			}
			if c.ok && v != c.want {
				t.Fatalf("value mismatch: got %v want %v (value=%q)", v, c.want, c.value)
			}
		})
	}
}

func TestLifeSampleRankOrdering(t *testing.T) {
	nonZero := 3.14
	zero := 0.0
	cases := []struct {
		name string
		s    ToolLifeProbeSample
		want int
	}{
		{"non-zero", ToolLifeProbeSample{Number: &nonZero}, 0},
		{"error", ToolLifeProbeSample{Error: "timeout"}, 1},
		{"zero", ToolLifeProbeSample{Number: &zero}, 2},
		{"unparsed", ToolLifeProbeSample{Value: "MACRO, 5021, NA"}, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := lifeSampleRank(c.s); got != c.want {
				t.Fatalf("rank %d want %d", got, c.want)
			}
		})
	}
}

func TestClassifyLifeProbeEmpty(t *testing.T) {
	rep := &ToolLifeProbeReport{}
	v, _ := classifyLifeProbe(rep)
	if v != "empty-range" {
		t.Fatalf("expected empty-range, got %q", v)
	}
}

func TestClassifyLifeProbeAllErrors(t *testing.T) {
	rep := &ToolLifeProbeReport{Probed: 5, Errors: 5}
	v, _ := classifyLifeProbe(rep)
	if v != "no-bridge" {
		t.Fatalf("expected no-bridge, got %q", v)
	}
}

func TestClassifyLifeProbeAllZero(t *testing.T) {
	rep := &ToolLifeProbeReport{Probed: 10, OK: 10, Empty: 10}
	v, _ := classifyLifeProbe(rep)
	if v != "all-zero" {
		t.Fatalf("expected all-zero, got %q", v)
	}
}

func TestClassifyLifeProbeCandidates(t *testing.T) {
	n := 17.0
	rep := &ToolLifeProbeReport{
		Probed:  3,
		OK:      3,
		NonZero: 1,
		Empty:   2,
		Samples: []ToolLifeProbeSample{
			{Macro: 3122, Number: &n},
		},
	}
	v, msg := classifyLifeProbe(rep)
	if v != "candidates-found" {
		t.Fatalf("expected candidates-found, got %q", v)
	}
	if msg == "" {
		t.Fatal("recommendation should not be empty for candidates")
	}
}

// FindNonZeroClusters detects contiguous runs of macros that carry
// non-zero values. Operators eyeballing the report want "macros
// 3122..3141 are populated" rather than 20 individual lines — a tool
// table laid out per-slot will show up as exactly that kind of run.
func TestFindNonZeroClusters(t *testing.T) {
	mk := func(n float64) *float64 { v := n; return &v }
	samples := []ToolLifeProbeSample{
		{Macro: 3122, Number: mk(2)},
		{Macro: 3123, Number: mk(0)},
		{Macro: 3124, Number: mk(5)},
		{Macro: 3125, Number: mk(8)},
		{Macro: 3126, Number: mk(0)},
		{Macro: 3200, Number: mk(1)},
	}
	got := FindNonZeroClusters(samples, 1)
	want := []MacroCluster{
		{Start: 3122, End: 3122, Count: 1},
		{Start: 3124, End: 3125, Count: 2},
		{Start: 3200, End: 3200, Count: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("clusters mismatch: got %+v want %+v", got, want)
	}
}

func TestFindNonZeroClustersHonorsStep(t *testing.T) {
	mk := func(n float64) *float64 { v := n; return &v }
	samples := []ToolLifeProbeSample{
		{Macro: 3000, Number: mk(1)},
		{Macro: 3002, Number: mk(2)},
		{Macro: 3004, Number: mk(3)},
		{Macro: 3010, Number: mk(4)}, // gap > step → new cluster
	}
	got := FindNonZeroClusters(samples, 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 clusters with step=2, got %d (%+v)", len(got), got)
	}
	if got[0].Start != 3000 || got[0].End != 3004 || got[0].Count != 3 {
		t.Fatalf("first cluster wrong: %+v", got[0])
	}
}
