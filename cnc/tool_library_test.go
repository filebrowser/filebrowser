package cnc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Minimal subset of a real Fusion 360 export. Two cutting tools
// (different pockets), one holder-only entry, one tool with pocket 0
// (in-library but not loaded).
const fixtureFusionLibrary = `{
  "data": [
    {
      "BMC": "carbide",
      "description": "4mm ball aluminum",
      "geometry": { "DC": 0.15748, "NOF": 4, "OAL": 1.5 },
      "guid": "00929141",
      "holder": {
        "description": "CAT40 ER25",
        "gaugeLength": 2.7083,
        "segments": [
          { "height": 0.0098, "lower-diameter": 1.158, "upper-diameter": 1.2178 }
        ],
        "type": "holder"
      },
      "post-process": { "number": 1 },
      "type": "ball end mill",
      "unit": "inches"
    },
    {
      "description": "HELICAL - 81608 - 45 DEG HELIX CORNER RADIUS END MILL",
      "geometry": { "DC": 0.375, "NOF": 3, "OAL": 3, "RE": 0.06 },
      "guid": "55c86795",
      "holder": { "description": "CAT40 ER16" },
      "post-process": { "number": 4 },
      "product-id": "81608",
      "product-link": "https://www.helicaltool.com/products/tool-details-81608",
      "type": "bull nose end mill",
      "vendor": "HELICAL SOLUTIONS"
    },
    {
      "description": "unloaded backup",
      "geometry": { "DC": 0.5 },
      "post-process": { "number": 0 },
      "type": "flat end mill"
    },
    {
      "description": "bare CAT40 holder",
      "type": "holder",
      "segments": []
    }
  ],
  "version": 36
}`

func TestNewToolLibrary_IndexesByPocket(t *testing.T) {
	var fl FusionLibrary
	if err := json.Unmarshal([]byte(fixtureFusionLibrary), &fl); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	lib := NewToolLibrary(fl)

	// Lookup hits.
	if got, ok := lib.Lookup(1); !ok || got.Type != "ball end mill" {
		t.Fatalf("Lookup(1) want ball end mill, got %+v / ok=%v", got, ok)
	}
	if got, ok := lib.Lookup(4); !ok || got.ProductID != "81608" {
		t.Fatalf("Lookup(4) want 81608, got %+v / ok=%v", got, ok)
	}

	// Number 0 is "not loaded" — must NOT appear.
	if _, ok := lib.Lookup(0); ok {
		t.Fatal("Lookup(0) should miss — unloaded tools aren't indexed")
	}

	// Holder-only entry must not show under a pocket number.
	for _, n := range lib.AssignedSlots() {
		entry, _ := lib.Lookup(n)
		if entry.IsHolderOnly() {
			t.Fatalf("holder-only entry leaked into pocket %d", n)
		}
	}

	slots := lib.AssignedSlots()
	if len(slots) != 2 || slots[0] != 1 || slots[1] != 4 {
		t.Fatalf("AssignedSlots want [1,4], got %v", slots)
	}
}

func TestLibraryStore_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.json")
	store, err := NewLibraryStore(path)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if store.Library() != nil {
		t.Fatal("fresh store should be empty")
	}

	lib, err := store.Replace([]byte(fixtureFusionLibrary))
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if _, ok := lib.Lookup(1); !ok {
		t.Fatal("after replace, pocket 1 should be findable")
	}

	// Reload from disk — second store should see the file we just wrote.
	store2, err := NewLibraryStore(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if store2.Library() == nil {
		t.Fatal("reopen should rehydrate the library from disk")
	}
	if _, ok := store2.Library().Lookup(4); !ok {
		t.Fatal("pocket 4 missing after reload")
	}

	// Persisted file should have uploaded_at populated.
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(buf), "uploaded_at") {
		t.Fatal("persisted file should include uploaded_at timestamp")
	}

	// Clear removes the file.
	if err := store.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("Clear should remove the file")
	}
}

func TestLibraryStore_RejectsEmpty(t *testing.T) {
	store, _ := NewLibraryStore(filepath.Join(t.TempDir(), "library.json"))
	if _, err := store.Replace([]byte(`{"data":[]}`)); err == nil {
		t.Fatal("empty data should be rejected")
	}
	if _, err := store.Replace([]byte(`not json`)); err == nil {
		t.Fatal("malformed JSON should be rejected")
	}
}
