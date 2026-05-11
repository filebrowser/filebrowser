package cnc

import (
	"reflect"
	"testing"
)

// captureSink builds a closure that captures DPRNT emissions for assertion.
func captureSink() (func(string), func(level, msg string), *[]string, *[]string) {
	emitted := &[]string{}
	logged := &[]string{}
	emit := func(s string) { *emitted = append(*emitted, s) }
	dbg := func(_, msg string) { *logged = append(*logged, msg) }
	return emit, dbg, emitted, logged
}

func TestDprntDrainSimpleLine(t *testing.T) {
	d := &dprntBuffer{}
	d.buf.WriteString("X1.234Y5.678\r\n")
	emit, dbg, out, _ := captureSink()
	d.drain(emit, dbg)
	if !reflect.DeepEqual(*out, []string{"X1.234Y5.678"}) {
		t.Fatalf("expected one line emitted, got %v", *out)
	}
	if d.buf.Len() != 0 {
		t.Fatalf("expected buffer drained, %d bytes left", d.buf.Len())
	}
}

func TestDprntDrainMultiLine(t *testing.T) {
	d := &dprntBuffer{}
	d.buf.WriteString("A1\nA2\r\nA3\r")
	emit, dbg, out, _ := captureSink()
	d.drain(emit, dbg)
	if !reflect.DeepEqual(*out, []string{"A1", "A2", "A3"}) {
		t.Fatalf("expected three lines, got %v", *out)
	}
}

func TestDprntDrainPartialLine(t *testing.T) {
	d := &dprntBuffer{}
	d.buf.WriteString("partial")
	emit, dbg, out, _ := captureSink()
	d.drain(emit, dbg)
	if len(*out) != 0 {
		t.Fatalf("expected no emit on partial, got %v", *out)
	}
	if d.buf.Len() != len("partial") {
		t.Fatalf("expected buffer kept intact, %d bytes left", d.buf.Len())
	}
	d.buf.WriteString("-line\n")
	d.drain(emit, dbg)
	if !reflect.DeepEqual(*out, []string{"partial-line"}) {
		t.Fatalf("expected joined line, got %v", *out)
	}
}

func TestDprntDrainDiscardsQCodeFrame(t *testing.T) {
	d := &dprntBuffer{}
	d.buf.WriteByte(stxByte)
	d.buf.WriteString("MACRO, 5021, 0.0000")
	d.buf.WriteByte(etbByte)
	d.buf.WriteString("\r\nDPRNT-output\n")
	emit, dbg, out, logged := captureSink()
	d.drain(emit, dbg)
	if !reflect.DeepEqual(*out, []string{"DPRNT-output"}) {
		t.Fatalf("expected only DPRNT line emitted, got %v", *out)
	}
	if len(*logged) == 0 {
		t.Fatal("expected debug log for discarded Q-code frame")
	}
}

func TestDprntSidecarPath(t *testing.T) {
	cases := []struct {
		absPath string
		jobID   string
		want    string
	}{
		{"/srv/share/nc/part.nc", "abc123", "/srv/share/nc/part.nc.abc123.dprnt.log"},
		{"/tmp/x.txt", "j-1", "/tmp/x.txt.j-1.dprnt.log"},
	}
	for _, c := range cases {
		got := dprntSidecarPath(c.absPath, c.jobID)
		if got != c.want {
			t.Errorf("sidecarPath(%q, %q) = %q, want %q", c.absPath, c.jobID, got, c.want)
		}
	}
}

func TestDprntDrainOversizedBufferFlushed(t *testing.T) {
	d := &dprntBuffer{}
	huge := make([]byte, dprntMaxLineBytes+10)
	for i := range huge {
		huge[i] = 'x'
	}
	d.buf.Write(huge)
	emit, dbg, out, logged := captureSink()
	d.drain(emit, dbg)
	if len(*out) != 0 {
		t.Fatalf("oversized buffer with no newline should not emit, got %v", *out)
	}
	if d.buf.Len() != 0 {
		t.Fatalf("expected buffer flushed, %d bytes left", d.buf.Len())
	}
	if len(*logged) == 0 {
		t.Fatal("expected debug log for flushed oversized buffer")
	}
}
