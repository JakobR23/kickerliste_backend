package entity

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestScoreHistoryEntryDiscriminatedJSON verifies the discriminated-union wire
// shape: a fixture entry carries fixtureId and omits reason; an adjustment
// entry carries reason and omits fixtureId.
func TestScoreHistoryEntryDiscriminatedJSON(t *testing.T) {
	at := time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	fid := 12
	reason := "Korrektur: Fehleingabe"

	fixture := ScoreHistoryEntry{At: at, Delta: 3, RunningTotal: 3, Type: ScoreHistoryTypeFixture, FixtureId: &fid}
	adjustment := ScoreHistoryEntry{At: at, Delta: -2, RunningTotal: 1, Type: ScoreHistoryTypeAdjustment, Reason: &reason}

	fs := mustMarshal(t, fixture)
	if !strings.Contains(fs, `"type":"fixture"`) || !strings.Contains(fs, `"fixtureId":12`) {
		t.Errorf("fixture entry missing expected fields: %s", fs)
	}
	if strings.Contains(fs, "reason") {
		t.Errorf("fixture entry must omit reason: %s", fs)
	}

	as := mustMarshal(t, adjustment)
	if !strings.Contains(as, `"type":"adjustment"`) || !strings.Contains(as, `"reason":"Korrektur: Fehleingabe"`) {
		t.Errorf("adjustment entry missing expected fields: %s", as)
	}
	if strings.Contains(as, "fixtureId") {
		t.Errorf("adjustment entry must omit fixtureId: %s", as)
	}
}

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(b)
}
