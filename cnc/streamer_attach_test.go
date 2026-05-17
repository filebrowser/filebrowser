package cnc

import (
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
)

type fakeSettings struct {
	s *settings.Settings
}

func (f *fakeSettings) Get() (*settings.Settings, error) { return f.s, nil }

func newTestStreamer() *Streamer {
	return New(&fakeSettings{s: &settings.Settings{}}, "m1")
}

func TestAttachAuto_RejectsEmpty(t *testing.T) {
	s := newTestStreamer()
	if s.AttachAuto("") {
		t.Fatal("AttachAuto should reject empty file_path")
	}
}

func TestAttachAuto_HappyPath(t *testing.T) {
	s := newTestStreamer()
	if !s.AttachAuto("/op1.nc") {
		t.Fatal("first AttachAuto should succeed")
	}
	st := s.Status()
	if st.AttachedFile != "/op1.nc" || st.AttachedSource != "auto" {
		t.Fatalf("status mismatch: %+v", st)
	}
}

func TestAttachAuto_NoopWhenSameAuto(t *testing.T) {
	s := newTestStreamer()
	s.AttachAuto("/op1.nc")
	if s.AttachAuto("/op1.nc") {
		t.Fatal("AttachAuto should be a no-op when the same file is already auto-attached")
	}
}

func TestAttachAuto_SwapsAutoToAuto(t *testing.T) {
	s := newTestStreamer()
	s.AttachAuto("/op1.nc")
	if !s.AttachAuto("/op2.nc") {
		t.Fatal("AttachAuto should swap from one auto file to another")
	}
	st := s.Status()
	if st.AttachedFile != "/op2.nc" {
		t.Fatalf("want /op2.nc, got %q", st.AttachedFile)
	}
}

func TestAttachAuto_RefusesOverManual(t *testing.T) {
	s := newTestStreamer()
	if err := s.Attach("/manual.nc", "manual"); err != nil {
		t.Fatalf("manual attach: %v", err)
	}
	if s.AttachAuto("/auto.nc") {
		t.Fatal("AttachAuto should refuse to overwrite a manual attach")
	}
	st := s.Status()
	if st.AttachedFile != "/manual.nc" || st.AttachedSource != "manual" {
		t.Fatalf("manual attach got clobbered: %+v", st)
	}
}

func TestDetachAuto_LeavesManualAlone(t *testing.T) {
	s := newTestStreamer()
	_ = s.Attach("/manual.nc", "manual")
	if s.DetachAuto() {
		t.Fatal("DetachAuto should be a no-op on a manual attach")
	}
	if s.Status().AttachedFile != "/manual.nc" {
		t.Fatal("manual attach got cleared by DetachAuto")
	}
}

func TestDetachAuto_ClearsAuto(t *testing.T) {
	s := newTestStreamer()
	s.AttachAuto("/op1.nc")
	if !s.DetachAuto() {
		t.Fatal("DetachAuto should clear an auto attach")
	}
	if s.Status().AttachedFile != "" {
		t.Fatal("AttachedFile should be empty after DetachAuto")
	}
}
