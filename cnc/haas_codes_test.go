package cnc

import "testing"

func TestLookupCodeKnownSettings(t *testing.T) {
	e, ok := LookupCode(CodeKindSetting, 414)
	if !ok {
		t.Fatal("Setting 414 must be in the catalog")
	}
	if e.Number != 414 || e.Kind != CodeKindSetting {
		t.Fatalf("entry shape wrong: %+v", e)
	}
	if e.Title == "" || e.Summary == "" {
		t.Fatal("Setting 414 must have title + summary")
	}
}

func TestLookupCodeUnknown(t *testing.T) {
	if _, ok := LookupCode(CodeKindSetting, 99999); ok {
		t.Fatal("unknown setting should return ok=false")
	}
	if _, ok := LookupCode(CodeKindAlarm, 99999); ok {
		t.Fatal("unknown alarm should return ok=false")
	}
}

func TestSearchCodesCaseInsensitive(t *testing.T) {
	results := SearchCodes(CodeKindSetting, "PROBE", 10)
	if len(results) == 0 {
		t.Fatal("uppercase probe search should hit Setting 414")
	}
	found414 := false
	for _, e := range results {
		if e.Number == 414 {
			found414 = true
			break
		}
	}
	if !found414 {
		t.Fatal("Setting 414 should appear in probe search")
	}
}

func TestSearchCodesKindFilter(t *testing.T) {
	results := SearchCodes(CodeKindAlarm, "", 200)
	for _, e := range results {
		if e.Kind != CodeKindAlarm {
			t.Fatalf("kind filter leaked: %+v", e)
		}
	}
	if len(results) == 0 {
		t.Fatal("alarm catalog should have entries")
	}
}

func TestNormalizeKindFallback(t *testing.T) {
	if NormalizeKind("garbage") != CodeKindSetting {
		t.Fatal("unknown kind should fall back to setting")
	}
	if NormalizeKind("ALARM") != CodeKindAlarm {
		t.Fatal("uppercase alarm should normalize")
	}
}
