package engram

import (
	"strings"
	"testing"
)

// canonicalRecordStoreCarveOutPassage is the record-store carve-out passage
// authored once and inserted byte-identically into section:full and
// section:slim. "Same decision class" reduces to literal equality of this
// string, not a human reading of two paraphrases.
//
// The passage used to gate itself on a declared `sdd.record_store: openrecord`
// config key, and carried a paragraph explaining that an `.openrecord/`
// directory on disk did NOT activate it. openrecord is installed
// unconditionally now, so both clauses went; the exemption they wrapped did
// not. CC5 below is what keeps them from coming back.
const canonicalRecordStoreCarveOutPassage = "**Record-store carve-out.** Architecture and design decisions are exempt from the save order above — they belong to the openrecord record store, not to Engram. No other save trigger is affected."

// recordStoreAxisMarkers are the literal shapes the deleted record-store axis
// took in this passage and in the prose around it: a config-key lookup, a
// declaration check, or the conditional wrapper the shipped skills used to
// carry. Each names the axis itself rather than merely naming openrecord, so
// the unconditional passage above matches none of them.
var recordStoreAxisMarkers = []string{
	"sdd.record_store",
	"record_store",
	"recordstore",
	"declares",
	"declaration",
	"<--:openrecord-->",
	"<--:/openrecord-->",
}

// recordStoreCarveOutSurface is one entry of the CC4 surface registry: a
// rendered protocol variant paired with an explicit verdict on whether it
// MUST carry the carve-out.
type recordStoreCarveOutSurface struct {
	name         string
	content      string
	wantCarveOut bool
}

// recordStoreCarveOutSurfaces is the CC4 registry: exactly the five render
// functions protocol.go exposes, each with an explicit wantCarveOut verdict.
func recordStoreCarveOutSurfaces() []recordStoreCarveOutSurface {
	return []recordStoreCarveOutSurface{
		{name: "full", content: protocolFull(), wantCarveOut: true},
		{name: "slim", content: protocolSlim(), wantCarveOut: true},
		{name: "passive-capture", content: protocolPassiveCapture(), wantCarveOut: false},
		{name: "compact", content: codexCompact(), wantCarveOut: false},
		{name: "codex-instructions", content: codexInstructions(), wantCarveOut: true},
	}
}

// countCanonicalPassage is the CC1 assertion body: how many times the
// normalized canonical carve-out passage appears in normalized content.
// Shared with the mutation fixtures so they exercise the exact check they
// claim to flip, rather than a parallel reimplementation of it.
func countCanonicalPassage(content string) int {
	return strings.Count(normalizeForSemanticMatch(content), normalizeForSemanticMatch(canonicalRecordStoreCarveOutPassage))
}

// countOpenrecord is the CC2/CC3 assertion body: how many times the
// substring "openrecord" appears in normalized content.
func countOpenrecord(content string) int {
	return strings.Count(normalizeForSemanticMatch(content), "openrecord")
}

// axisMarkersPresent is the CC5 assertion body: every deleted-axis marker the
// normalized content still carries, in declaration order. An empty slice means
// the surface states the carve-out unconditionally.
func axisMarkersPresent(content string) []string {
	normalized := normalizeForSemanticMatch(content)
	var found []string
	for _, marker := range recordStoreAxisMarkers {
		if strings.Contains(normalized, strings.ToLower(marker)) {
			found = append(found, marker)
		}
	}
	return found
}

// TestRecordStoreCarveOut_SurfaceRegistry is CC4: the surface table has
// exactly 5 entries, ids equal to the five render functions, each carrying
// an explicit wantCarveOut verdict — so a sixth rendered variant added later
// cannot skip CC1-CC3 and CC5 silently.
func TestRecordStoreCarveOut_SurfaceRegistry(t *testing.T) {
	surfaces := recordStoreCarveOutSurfaces()
	if len(surfaces) != 5 {
		t.Fatalf("surface registry has %d entries, want exactly 5", len(surfaces))
	}

	wantVerdicts := map[string]bool{
		"full":               true,
		"slim":               true,
		"passive-capture":    false,
		"compact":            false,
		"codex-instructions": true,
	}

	seen := make(map[string]bool, len(surfaces))
	for _, s := range surfaces {
		want, known := wantVerdicts[s.name]
		if !known {
			t.Errorf("surface registry carries unexpected id %q", s.name)
			continue
		}
		if s.wantCarveOut != want {
			t.Errorf("surface %q: wantCarveOut = %v, want %v", s.name, s.wantCarveOut, want)
		}
		seen[s.name] = true
	}
	for name := range wantVerdicts {
		if !seen[name] {
			t.Errorf("surface registry is missing required id %q", name)
		}
	}
}

// TestRecordStoreCarveOut_PresenceAndSingularity is CC1: the normalized
// canonical carve-out passage appears exactly once in every surface that
// MUST carry it (full, slim, codex-instructions). This is the half that keeps
// CC5 honest — without it, deleting the passage outright would satisfy the
// no-axis check.
func TestRecordStoreCarveOut_PresenceAndSingularity(t *testing.T) {
	for _, s := range recordStoreCarveOutSurfaces() {
		if !s.wantCarveOut {
			continue
		}
		t.Run(s.name, func(t *testing.T) {
			if got := countCanonicalPassage(s.content); got != 1 {
				t.Errorf("surface %q contains the canonical carve-out passage %d times, want exactly 1", s.name, got)
			}
		})
	}
}

// TestRecordStoreCarveOut_NoBroadening is CC2: the "openrecord" count in
// each carve-out-carrying surface equals the canonical passage's own count,
// so a second, wider exemption anywhere in the variant fails.
func TestRecordStoreCarveOut_NoBroadening(t *testing.T) {
	canonicalCount := countOpenrecord(canonicalRecordStoreCarveOutPassage)
	if canonicalCount == 0 {
		t.Fatalf("canonical carve-out passage contains zero occurrences of %q — the CC2 baseline is degenerate", "openrecord")
	}
	for _, s := range recordStoreCarveOutSurfaces() {
		if !s.wantCarveOut {
			continue
		}
		t.Run(s.name, func(t *testing.T) {
			if got := countOpenrecord(s.content); got != canonicalCount {
				t.Errorf("surface %q contains %d occurrences of %q, want exactly %d (the canonical passage's own count)", s.name, got, "openrecord", canonicalCount)
			}
		})
	}
}

// TestRecordStoreCarveOut_Containment is CC3: surfaces that MUST stay
// byte-identical (passive-capture, compact) carry zero occurrences of
// "openrecord" — the carve-out MUST NOT leak into them.
func TestRecordStoreCarveOut_Containment(t *testing.T) {
	for _, s := range recordStoreCarveOutSurfaces() {
		if s.wantCarveOut {
			continue
		}
		t.Run(s.name, func(t *testing.T) {
			if got := countOpenrecord(s.content); got != 0 {
				t.Errorf("surface %q contains %d occurrences of %q, want 0 — this surface MUST NOT carry the carve-out", s.name, got, "openrecord")
			}
		})
	}
}

// TestRecordStoreCarveOut_NoAxisLookup is CC5: no rendered surface may gate
// the carve-out on the deleted record-store axis, name the config key that
// used to declare it, instruct anyone to check a declaration, or re-wrap the
// passage in the conditional delimiters the shipped skills used to carry.
// It runs over every surface — including the two that must not carry the
// carve-out at all — because an axis lookup smuggled into a variant the
// exemption never reached is still an axis lookup.
func TestRecordStoreCarveOut_NoAxisLookup(t *testing.T) {
	if found := axisMarkersPresent(canonicalRecordStoreCarveOutPassage); len(found) != 0 {
		t.Fatalf("the canonical passage itself carries deleted-axis markers %v — the CC5 baseline is self-contradictory", found)
	}
	for _, s := range recordStoreCarveOutSurfaces() {
		t.Run(s.name, func(t *testing.T) {
			if found := axisMarkersPresent(s.content); len(found) != 0 {
				t.Errorf("surface %q gates the record store on the deleted axis via %v; openrecord is unconditional and nothing declares or resolves a record store", s.name, found)
			}
		})
	}
}

// TestRecordStoreCarveOut_MutationBroadensClassInSlim is M1: broadening
// slim's exempted class to "all decisions are exempt" must make CC1 fail.
// The positive text is asserted found first, so a stale fixture turns this
// guard red instead of vacuously green.
func TestRecordStoreCarveOut_MutationBroadensClassInSlim(t *testing.T) {
	original := protocolSlim()
	const positive = "Architecture and design decisions are exempt"
	const broadened = "All decisions are exempt"

	if !strings.Contains(original, positive) {
		t.Fatalf("fixture is stale: slim no longer contains %q", positive)
	}

	mutated := strings.Replace(original, positive, broadened, 1)
	if mutated == original {
		t.Fatalf("mutation did not change slim's content")
	}
	if got := countCanonicalPassage(mutated); got == 1 {
		t.Fatalf("CC1 did not fail after broadening slim's exempted class: the passage still matched exactly once")
	}
}

// TestRecordStoreCarveOut_MutationDeletesPassageFromFull is M2: deleting the
// passage from a copy of full must make CC1 fail.
func TestRecordStoreCarveOut_MutationDeletesPassageFromFull(t *testing.T) {
	original := protocolFull()

	if !strings.Contains(original, canonicalRecordStoreCarveOutPassage) {
		t.Fatalf("fixture is stale: full no longer contains the canonical carve-out passage verbatim")
	}

	mutated := strings.Replace(original, canonicalRecordStoreCarveOutPassage, "", 1)
	if mutated == original {
		t.Fatalf("mutation did not change full's content")
	}
	if got := countCanonicalPassage(mutated); got == 1 {
		t.Fatalf("CC1 did not fail after deleting the passage from full: the passage still matched exactly once")
	}
}

// TestRecordStoreCarveOut_MutationAppendsSecondOpenrecordSentenceToSlim is
// M3: appending a second openrecord sentence to a copy of slim must make CC2
// fail — the count no longer equals the canonical passage's own count.
func TestRecordStoreCarveOut_MutationAppendsSecondOpenrecordSentenceToSlim(t *testing.T) {
	original := protocolSlim()
	canonicalCount := countOpenrecord(canonicalRecordStoreCarveOutPassage)

	if baseline := countOpenrecord(original); baseline == 0 {
		t.Fatalf("fixture is stale: slim contains zero occurrences of %q before mutation", "openrecord")
	}

	const secondSentence = "\n- A second, unrelated exemption also applies whenever openrecord is present."
	mutated := original + secondSentence

	if got := countOpenrecord(mutated); got == canonicalCount {
		t.Fatalf("CC2 did not fail after appending a second openrecord sentence: the count still equals the canonical baseline (%d)", canonicalCount)
	}
}

// TestRecordStoreCarveOut_MutationNarrowsClassInFull is M4: narrowing the
// exempted class to "tool or library choices are exempt" must make CC1
// fail — proving the guard checks for THIS class, not merely for the
// presence of some carve-out.
func TestRecordStoreCarveOut_MutationNarrowsClassInFull(t *testing.T) {
	original := protocolFull()
	const positive = "Architecture and design decisions are exempt"
	const narrowed = "Tool or library choices are exempt"

	if !strings.Contains(original, positive) {
		t.Fatalf("fixture is stale: full no longer contains %q", positive)
	}

	mutated := strings.Replace(original, positive, narrowed, 1)
	if mutated == original {
		t.Fatalf("mutation did not change full's content")
	}
	if got := countCanonicalPassage(mutated); got == 1 {
		t.Fatalf("CC1 did not fail after narrowing full's exempted class: the passage still matched exactly once")
	}
}

// TestRecordStoreCarveOut_MutationRegatesFullOnTheDeletedAxis is M5, CC5's own
// anti-vacuity fixture: putting the declaration condition back onto full — the
// exact wording this change removed — must make CC5 fail, naming the config
// key it reintroduced.
func TestRecordStoreCarveOut_MutationRegatesFullOnTheDeletedAxis(t *testing.T) {
	original := protocolFull()

	if found := axisMarkersPresent(original); len(found) != 0 {
		t.Fatalf("fixture is stale: full already carries deleted-axis markers %v", found)
	}

	const positive = "exempt from the save order above — they belong to the openrecord record store"
	const regated = "exempt from the save order above when this workspace's `openspec/config.yaml` declares `sdd.record_store: openrecord` — they belong to that record store"

	if !strings.Contains(original, positive) {
		t.Fatalf("fixture is stale: full no longer contains %q", positive)
	}
	mutated := strings.Replace(original, positive, regated, 1)
	if mutated == original {
		t.Fatalf("mutation did not change full's content")
	}

	found := axisMarkersPresent(mutated)
	if len(found) == 0 {
		t.Fatal("CC5 did not fail after re-gating full on the declared record_store axis")
	}
	if !strings.Contains(strings.Join(found, " "), "sdd.record_store") {
		t.Fatalf("axisMarkersPresent(regated) = %v, want it to name the reintroduced config key", found)
	}
}

// TestRecordStoreCarveOut_MutationRewrapsSlimInTheConditionalDelimiters is M6:
// the other half of CC5's subject — the `<--:openrecord-->` wrapper. Nothing
// ever processed those markers (no renderer strips them; they shipped verbatim
// into installed skills), so only a test catches their return.
func TestRecordStoreCarveOut_MutationRewrapsSlimInTheConditionalDelimiters(t *testing.T) {
	original := protocolSlim()

	if found := axisMarkersPresent(original); len(found) != 0 {
		t.Fatalf("fixture is stale: slim already carries deleted-axis markers %v", found)
	}

	if !strings.Contains(original, canonicalRecordStoreCarveOutPassage) &&
		countCanonicalPassage(original) != 1 {
		t.Fatal("fixture is stale: slim no longer carries the canonical carve-out passage")
	}
	mutated := "<--:openrecord-->\n" + original + "\n<--:/openrecord-->"

	found := axisMarkersPresent(mutated)
	if len(found) != 2 {
		t.Fatalf("axisMarkersPresent(rewrapped) = %v, want both delimiter halves", found)
	}
}
