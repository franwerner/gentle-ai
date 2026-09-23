package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/franwerner/gentle-ai/v3/internal/catalog"
)

// TestBuildEventNeverLeaksIdentifyingData builds an event from a fixture
// whose home directory, hostname, and username are realistic-looking and
// asserts none of those bytes ever reach the marshaled payload. This is the
// behavior the contract cares about most: the event carries only fixed
// enums, a random id, and counts.
func TestBuildEventNeverLeaksIdentifyingData(t *testing.T) {
	realHostname := "alans-macbook-pro.local"
	realUsername := "alanbuscaglia"
	// The fixture home directory embeds a realistic-looking username and
	// path shape (/Users/<name>/...), but lives under t.TempDir() so the
	// test never touches the real filesystem outside its own sandbox.
	realHomeDir := filepath.Join(t.TempDir(), "Users", realUsername, "gentle-ai-real-home")

	if err := os.MkdirAll(realHomeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	persisted, err := NewState()
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(realHomeDir, persisted); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(realHomeDir)
	if err != nil {
		t.Fatal(err)
	}

	ev := Build(BuildInput{
		Kind:       EventHeartbeat,
		InstallID:  loaded.InstallID,
		Now:        time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC),
		Version:    "2.2.0",
		Agents:     []string{"claude-code", "opencode"},
		Components: []string{"sdd", "engram"},
		RDDEnabled: true,
		Counters:   Counters{Syncs: 3},
	})
	payload, err := Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}

	haystack := string(payload)
	for _, forbidden := range []string{realHostname, realUsername, realHomeDir, os.TempDir()} {
		if strings.Contains(haystack, forbidden) {
			t.Fatalf("payload leaks %q:\n%s", forbidden, haystack)
		}
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"path", "cwd", "home", "hostname", "username", "prompt", "diff", "ip"} {
		if _, present := decoded[key]; present {
			t.Fatalf("payload carries unexpected field %q", key)
		}
	}
	if decoded["schema"] != EventSchema {
		t.Fatalf("schema = %v, want %v", decoded["schema"], EventSchema)
	}
	if len(payload) > MaxPayloadBytes {
		t.Fatalf("payload is %d bytes, want <= %d", len(payload), MaxPayloadBytes)
	}
}

func TestBuildInstallEventCarriesNoCounters(t *testing.T) {
	ev := Build(BuildInput{Kind: EventInstall, InstallID: "x", Now: time.Now(), Counters: Counters{Syncs: 5}})
	payload, err := Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "counters") {
		t.Fatalf("install event must never carry counters: %s", payload)
	}
}

func TestBuildHeartbeatEventCarriesCountersEvenWhenZero(t *testing.T) {
	ev := Build(BuildInput{Kind: EventHeartbeat, InstallID: "x", Now: time.Now()})
	payload, err := Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"counters"`) {
		t.Fatalf("heartbeat event must always carry counters: %s", payload)
	}
}

// componentEnumFromSchema reads a telemetry event schema and returns its
// closed `components` enum.
func componentEnumFromSchema(t *testing.T, path string) []string {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Properties struct {
			Components struct {
				Items struct {
					Enum []string `json:"enum"`
				} `json:"items"`
				MaxItems *int `json:"maxItems"`
			} `json:"components"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	enum := doc.Properties.Components.Items.Enum
	if len(enum) == 0 {
		t.Fatalf("%s: no components enum", path)
	}
	if max := doc.Properties.Components.MaxItems; max != nil && *max < len(enum) {
		t.Fatalf("%s: components maxItems = %d, below the %d values its enum allows", path, *max, len(enum))
	}
	return enum
}

// TestComponentAllowListsAgreeWithCatalog pins the three closed lists a
// component must appear in — knownComponents here, the collector's embedded
// schema, and the published contract copy — against catalog.MVPComponents().
// A component added to the catalog and forgotten in any of them is either
// silently stripped from every event (filterKnown) or gets the whole event
// rejected by the collector, and nothing else fails.
func TestComponentAllowListsAgreeWithCatalog(t *testing.T) {
	want := make(map[string]bool)
	for _, component := range catalog.MVPComponents() {
		want[string(component.ID)] = true
	}

	lists := map[string][]string{
		"../telemetrycollector/schema/event.schema.json":         componentEnumFromSchema(t, filepath.Join("..", "telemetrycollector", "schema", "event.schema.json")),
		"../../contracts/telemetry/v1/schemas/event.schema.json": componentEnumFromSchema(t, filepath.Join("..", "..", "contracts", "telemetry", "v1", "schemas", "event.schema.json")),
	}
	knownList := make([]string, 0, len(knownComponents))
	for id := range knownComponents {
		knownList = append(knownList, id)
	}
	lists["internal/telemetry.knownComponents"] = knownList

	for name, got := range lists {
		have := make(map[string]bool, len(got))
		for _, id := range got {
			if have[id] {
				t.Fatalf("%s: duplicate component %q", name, id)
			}
			have[id] = true
			if !want[id] {
				t.Errorf("%s: has %q, which catalog.MVPComponents() does not", name, id)
			}
		}
		for id := range want {
			if !have[id] {
				t.Errorf("%s: missing %q, which catalog.MVPComponents() declares", name, id)
			}
		}
	}
}
