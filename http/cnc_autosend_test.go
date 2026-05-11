package fbhttp

import (
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/cnc"
)

func TestAutoSendBlockReasonClear(t *testing.T) {
	pf := &cnc.Preflight{
		Summary: cnc.PreflightSummary{OK: 4},
	}
	if got := autoSendBlockReason(pf); got != "" {
		t.Fatalf("expected clear path, got %q", got)
	}
}

func TestAutoSendBlockReasonNilPreflight(t *testing.T) {
	if got := autoSendBlockReason(nil); got == "" {
		t.Fatal("expected nil preflight to block")
	}
}

func TestAutoSendBlockReasonTableMissing(t *testing.T) {
	pf := &cnc.Preflight{TableMissing: true}
	got := autoSendBlockReason(pf)
	if !strings.Contains(got, "tool-table") {
		t.Fatalf("expected table-missing reason, got %q", got)
	}
}

func TestAutoSendBlockReasonMissingTrumpsWarn(t *testing.T) {
	pf := &cnc.Preflight{
		Summary: cnc.PreflightSummary{Missing: 1, Warn: 2},
	}
	got := autoSendBlockReason(pf)
	if !strings.Contains(got, "missing") {
		t.Fatalf("expected missing-first reason, got %q", got)
	}
}

func TestAutoSendBlockReasonSpindleSwap(t *testing.T) {
	pf := &cnc.Preflight{
		Summary:     cnc.PreflightSummary{OK: 3},
		SpindleSwap: true,
	}
	got := autoSendBlockReason(pf)
	if !strings.Contains(got, "spindle swap") {
		t.Fatalf("expected spindle-swap reason, got %q", got)
	}
}

func TestAutoSendBlockReasonAllSummaryStates(t *testing.T) {
	cases := []struct {
		name string
		pf   *cnc.Preflight
		want string
	}{
		{"empty", &cnc.Preflight{Summary: cnc.PreflightSummary{Empty: 1}}, "empty"},
		{"offline", &cnc.Preflight{Summary: cnc.PreflightSummary{Offline: 1}}, "errored"},
		{"warn", &cnc.Preflight{Summary: cnc.PreflightSummary{Warn: 1}}, "warn"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := autoSendBlockReason(c.pf)
			if !strings.Contains(got, c.want) {
				t.Fatalf("expected %q in %q", c.want, got)
			}
		})
	}
}
