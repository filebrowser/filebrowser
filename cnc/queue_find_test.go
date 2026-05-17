package cnc

import (
	"testing"
)

func TestFindByONumber(t *testing.T) {
	qs, err := NewQueueStore(t.TempDir(), nil)
	if err != nil {
		t.Fatalf("new queue store: %v", err)
	}
	mid := "m1"
	qs.queues[mid] = []QueueItem{
		{ID: "a", FilePath: "/op1.nc", OnumberHint: "O00057", State: QueueStateQueued},
		{ID: "b", FilePath: "/op2.nc", OnumberHint: "O00123", State: QueueStateQueued},
	}

	cases := []struct {
		name string
		in   string
		want string // expected FilePath; "" means no match
	}{
		{"exact match", "O00057", "/op1.nc"},
		{"short form normalized", "O57", "/op1.nc"},
		{"lowercase normalized", "o123", "/op2.nc"},
		{"miss", "O99999", ""},
		{"empty", "", ""},
		{"unknown machine", "O00057", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mid := mid
			if c.name == "unknown machine" {
				mid = "other"
			}
			got := qs.FindByONumber(mid, c.in)
			if c.want == "" {
				if got != nil {
					t.Fatalf("want nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("want %q, got nil", c.want)
			}
			if got.FilePath != c.want {
				t.Fatalf("want %q, got %q", c.want, got.FilePath)
			}
		})
	}
}

func TestFindByONumberReturnsCopy(t *testing.T) {
	qs, err := NewQueueStore(t.TempDir(), nil)
	if err != nil {
		t.Fatalf("new queue store: %v", err)
	}
	mid := "m1"
	qs.queues[mid] = []QueueItem{
		{ID: "a", FilePath: "/op1.nc", OnumberHint: "O00057", State: QueueStateQueued},
	}
	got := qs.FindByONumber(mid, "O00057")
	if got == nil {
		t.Fatal("expected match")
	}
	got.FilePath = "/mutated"
	again := qs.FindByONumber(mid, "O00057")
	if again.FilePath != "/op1.nc" {
		t.Fatalf("store mutated through returned pointer: %q", again.FilePath)
	}
}
