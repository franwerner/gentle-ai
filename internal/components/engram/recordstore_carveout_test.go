package engram

import (
	"strings"
	"testing"
)

// canonicalRecordStoreCarveOutPassage is the record-store carve-out passage
// design.md's Interfaces/Contracts section pins verbatim (Decision 2):
// authored once and inserted byte-identically into section:full and
// section:slim. "Same decision class" reduces to literal equality of this
// string, not a human reading of two paraphrases.
const canonicalRecordStoreCarveOutPassage = "**Record-store carve-out.** Architecture and design decisions are exempt from the save order above when this workspace's `openspec/config.yaml` (or `.yml`) declares `sdd.record_store: openrecord` — they belong to that record store, not to Engram. Check the declaration yourself; the presence of an `.openrecord/` directory on disk does NOT activate this carve-out, even when the store is fully populated. No other save trigger is affected."

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

// TestRecordStoreCarveOut_SurfaceRegistry is CC4: the surface table has
// exactly 5 entries, ids equal to the five render functions, each carrying
// an explicit wantCarveOut verdict — so a sixth rendered variant added later
// cannot skip CC1-CC3 silently.
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
// MUST carry it (full, slim, codex-instructions).
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
