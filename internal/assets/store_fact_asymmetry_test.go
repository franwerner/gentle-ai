package assets

import (
	"regexp"
	"strings"
	"testing"
)

// #recordstore-fact-injection: the two store facts read the same on the page —
// both name "the store" and both mention branching. Nothing short of a runtime
// check catches the day someone rewords the artifact-store prohibition into a
// sentence that also happens to name recordStore, or reintroduces the
// sdd-status shell-out this change removed from Activation.

type unit struct {
	Line int
	Kind string
	Text string
}

type finding struct {
	Path string
	Line int
	Kind string
	Text string
	Why  string
}

const (
	unitParagraph = "paragraph"
	unitListEntry = "list-entry"
	unitTableRow  = "table-row"
)

var listEntryPattern = regexp.MustCompile(`^([-*+] |\d+\. )`)

const sentenceTerminators = ".!?"
const trailingAllowed = "*`)\"]"

var abbreviations = []string{"e.g", "i.e", "etc", "vs"}

var (
	prohibitionMarkers = []string{
		"Do NOT detect the artifact store",
		"Do NOT determine the artifact store yourself",
		"do NOT branch on it",
		"never branch on the store",
	}
	permissionMarkers = []string{
		"may legitimately behave differently",
		"what is forbidden is determining it yourself",
	}
	scopingMarkers = []string{
		"This prohibition is about the",
		"artifact",
	}
	injectionMarkers = []string{
		"injects",
		"injected",
		"launch prompt",
	}
	resolutionMarkers = []string{
		"already resolved",
		"dispatcher resolved",
	}
	lookupMarkers = []string{
		"gentle-ai sdd-status",
		"sdd-status --json",
		"--json",
		"recordStore.resolved",
	}
)

// segment turns a markdown surface into the units a human reads: a sentence
// inside a paragraph or list entry, or a whole table row. baseLine is the
// 1-indexed line the surface's first physical line occupies in its source
// file; a unit's Line is baseLine plus the zero-based index of its block's
// first physical line within the surface.
func segment(surface string, baseLine int) []unit {
	type block struct {
		kind     string
		startIdx int
		lines    []string
	}

	lines := strings.Split(surface, "\n")
	var blocks []block
	var current *block
	inFence := false

	closeCurrent := func() {
		if current != nil {
			blocks = append(blocks, *current)
			current = nil
		}
	}

	for i, raw := range lines {
		trimmed := strings.TrimSpace(raw)

		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			closeCurrent()
			continue
		}
		if inFence {
			continue
		}
		if trimmed == "" {
			closeCurrent()
			continue
		}
		if strings.HasPrefix(trimmed, "|") {
			closeCurrent()
			blocks = append(blocks, block{kind: unitTableRow, startIdx: i, lines: []string{trimmed}})
			continue
		}
		if listEntryPattern.MatchString(trimmed) {
			closeCurrent()
			current = &block{kind: unitListEntry, startIdx: i, lines: []string{trimmed}}
			continue
		}
		if current != nil && (current.kind == unitParagraph || current.kind == unitListEntry) {
			current.lines = append(current.lines, trimmed)
			continue
		}
		current = &block{kind: unitParagraph, startIdx: i, lines: []string{trimmed}}
	}
	closeCurrent()

	var units []unit
	for _, b := range blocks {
		joined := strings.Join(b.lines, " ")
		line := baseLine + b.startIdx

		if b.kind == unitTableRow {
			units = append(units, unit{Line: line, Kind: unitTableRow, Text: joined})
			continue
		}

		for _, frag := range sentenceCut(joined, maskInlineCode(joined)) {
			frag = strings.TrimSpace(frag)
			if frag == "" {
				continue
			}
			units = append(units, unit{Line: line, Kind: b.kind, Text: frag})
		}
	}
	return units
}

// maskInlineCode marks every byte between an opening backtick and its
// closing backtick, so a sentence terminator inside `sdd.record_store` or
// `config.yaml` is never mistaken for the end of a sentence.
func maskInlineCode(text string) []bool {
	mask := make([]bool, len(text))
	open := false
	for i := 0; i < len(text); i++ {
		if text[i] == '`' {
			mask[i] = true
			open = !open
			continue
		}
		if open {
			mask[i] = true
		}
	}
	return mask
}

func endsWithAbbreviation(before string) bool {
	for _, abbr := range abbreviations {
		if strings.HasSuffix(before, abbr) {
			return true
		}
	}
	return false
}

// sentenceCut cuts a joined block into sentences after an unmasked
// terminator followed by zero or more trailing closers and then a space or
// the end of the text, skipping abbreviations.
func sentenceCut(text string, mask []bool) []string {
	var fragments []string
	start := 0

	for i := 0; i < len(text); i++ {
		if mask[i] || !strings.ContainsRune(sentenceTerminators, rune(text[i])) {
			continue
		}
		if endsWithAbbreviation(text[:i]) {
			continue
		}
		j := i + 1
		for j < len(text) && strings.ContainsRune(trailingAllowed, rune(text[j])) {
			j++
		}
		if j != len(text) && text[j] != ' ' {
			continue
		}
		fragments = append(fragments, text[start:j])
		start = j
		if start < len(text) && text[start] == ' ' {
			start++
		}
		i = start - 1
	}
	if start < len(text) {
		fragments = append(fragments, text[start:])
	}
	return fragments
}

func containsAny(text string, markers []string) bool {
	for _, m := range markers {
		if strings.Contains(text, m) {
			return true
		}
	}
	return false
}

func containsAll(text string, markers []string) bool {
	for _, m := range markers {
		if !strings.Contains(text, m) {
			return false
		}
	}
	return true
}

func isPermissionExempt(text string) bool {
	return containsAny(text, permissionMarkers)
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// c1InjectionFraming fails when a surface names recordStore without also
// framing it, within the same surface, as an injected and already-resolved
// value.
func c1InjectionFraming(path, text string, baseLine int) []finding {
	if !strings.Contains(text, "recordStore") {
		return nil
	}
	var findings []finding
	if !containsAny(text, injectionMarkers) {
		findings = append(findings, finding{
			Path: path, Line: baseLine, Kind: "surface",
			Text: truncateRunes(text, 200),
			Why:  "names recordStore with no injection-framing marker present in the surface",
		})
	}
	if !containsAny(text, resolutionMarkers) {
		findings = append(findings, finding{
			Path: path, Line: baseLine, Kind: "surface",
			Text: truncateRunes(text, 200),
			Why:  "names recordStore with no resolution marker present in the surface",
		})
	}
	return findings
}

// c2NoStatusLookup fails when the Activation surface still carries a
// status-projection invocation.
func c2NoStatusLookup(path, text string, baseLine int) []finding {
	var found []string
	for _, literal := range lookupMarkers {
		if strings.Contains(text, literal) {
			found = append(found, literal)
		}
	}
	if len(found) == 0 {
		return nil
	}
	return []finding{{
		Path: path, Line: baseLine, Kind: "surface",
		Text: strings.Join(found, ", "),
		Why:  "activation still invokes a status lookup instead of reading the injected value",
	}}
}

// c3aScopingPresent fails when no unit of Section B scopes the branching
// prohibition to the artifact store.
func c3aScopingPresent(path string, units []unit) []finding {
	for _, u := range units {
		if containsAll(u.Text, scopingMarkers) {
			return nil
		}
	}
	return []finding{{
		Path: path, Kind: "section",
		Why: "no unit scopes the branching prohibition to the artifact store",
	}}
}

// c3bMarkersStillMatch fails when none of Section B's non-exempt units
// carries a prohibition marker any more — a sign the markers went stale.
func c3bMarkersStillMatch(path string, units []unit) []finding {
	for _, u := range units {
		if isPermissionExempt(u.Text) {
			continue
		}
		if containsAny(u.Text, prohibitionMarkers) {
			return nil
		}
	}
	return []finding{{
		Path: path, Kind: "section",
		Why: "no unit carries a prohibition marker; the markers no longer match the shipped prose and must be updated, not deleted",
	}}
}

// c3cNoConflation fails when a non-exempt unit carries both an
// artifact-store branching prohibition and recordStore.
func c3cNoConflation(path string, units []unit) []finding {
	var findings []finding
	for _, u := range units {
		if isPermissionExempt(u.Text) {
			continue
		}
		if containsAny(u.Text, prohibitionMarkers) && strings.Contains(u.Text, "recordStore") {
			findings = append(findings, finding{
				Path: path, Line: u.Line, Kind: u.Kind,
				Text: truncateRunes(u.Text, 200),
				Why:  "unit carries both an artifact-store branching prohibition and recordStore",
			})
		}
	}
	return findings
}

func reportFindings(t *testing.T, findings []finding) {
	t.Helper()
	for _, f := range findings {
		t.Errorf("%s:%d [%s] %s: %s", f.Path, f.Line, f.Kind, f.Why, f.Text)
	}
}

// TestStoreFactsAreNotConflated pins the spec's "The two store facts are not
// conflated" scenario over the shipped assets.
func TestStoreFactsAreNotConflated(t *testing.T) {
	const sectionBPath = "skills/_shared/sdd-phase-common.md"
	const activationPath = "skills/_shared/openrecord-convention.md"
	const sectionBHeading = "## B. Artifact Retrieval"
	const activationHeading = "## Activation"

	sectionBContent := MustRead(sectionBPath)
	activationContent := MustRead(activationPath)

	sectionBText := markdownSection(sectionBContent, sectionBHeading)
	if sectionBText == "" {
		t.Fatalf("%s: heading %q not found", sectionBPath, sectionBHeading)
	}
	activationText := markdownSection(activationContent, activationHeading)
	if activationText == "" {
		t.Fatalf("%s: heading %q not found", activationPath, activationHeading)
	}

	sectionBBase := 1 + strings.Count(sectionBContent[:strings.Index(sectionBContent, sectionBHeading)], "\n")
	activationBase := 1 + strings.Count(activationContent[:strings.Index(activationContent, activationHeading)], "\n")

	sectionBUnits := segment(sectionBText, sectionBBase)
	activationUnits := segment(activationText, activationBase)

	t.Run("C1", func(t *testing.T) {
		reportFindings(t, c1InjectionFraming(sectionBPath, sectionBText, sectionBBase))
		reportFindings(t, c1InjectionFraming(activationPath, activationText, activationBase))
	})

	t.Run("C2", func(t *testing.T) {
		reportFindings(t, c2NoStatusLookup(activationPath, activationText, activationBase))
	})

	t.Run("C3a", func(t *testing.T) {
		reportFindings(t, c3aScopingPresent(sectionBPath, sectionBUnits))
	})

	t.Run("C3b", func(t *testing.T) {
		reportFindings(t, c3bMarkersStillMatch(sectionBPath, sectionBUnits))
	})

	t.Run("C3c", func(t *testing.T) {
		reportFindings(t, c3cNoConflation(sectionBPath, sectionBUnits))
		reportFindings(t, c3cNoConflation(activationPath, activationUnits))
	})
}

// TestSegmentMasksInlineCodeSpans proves step 3 of segment(): a period
// inside an inline-code span is never a sentence terminator.
func TestSegmentMasksInlineCodeSpans(t *testing.T) {
	text := "The value is `sdd.record_store` and it stays inside one sentence."
	units := segment(text, 1)
	if len(units) != 1 {
		t.Fatalf("segment() returned %d units, want 1 (inline code must not fracture the sentence): %+v", len(units), units)
	}
}

// TestSegmentJoinsWrappedLinesIntoOneUnit proves step 2 of segment(): a
// sentence hard-wrapped across two physical lines still joins into one unit,
// anchored to the block's first physical line.
func TestSegmentJoinsWrappedLinesIntoOneUnit(t *testing.T) {
	const baseLine = 5
	text := "Do NOT detect the artifact store, and do NOT branch on it,\nincluding `recordStore`."
	units := segment(text, baseLine)
	if len(units) != 1 {
		t.Fatalf("segment() returned %d units for a hard-wrapped sentence, want 1: %+v", len(units), units)
	}
	if units[0].Line != baseLine {
		t.Fatalf("unit.Line = %d, want %d (the block's first physical line)", units[0].Line, baseLine)
	}
}

// TestC2FlagsStatusLookupInActivation proves C2 fires, with exactly one
// finding, when Activation prose still names a status-projection call.
func TestC2FlagsStatusLookupInActivation(t *testing.T) {
	text := "Before this rewrite, a phase would run `gentle-ai sdd-status --json` to learn the value."
	findings := c2NoStatusLookup("synthetic-activation", text, 1)
	if len(findings) != 1 {
		t.Fatalf("c2NoStatusLookup() = %d findings, want 1: %+v", len(findings), findings)
	}
}

// TestC3cDetectsConflationFixture pins the spec's first failure scenario:
// widening the artifact-store prohibition to also name recordStore.
func TestC3cDetectsConflationFixture(t *testing.T) {
	text := "Do NOT detect the artifact store, and do NOT branch on it, including `recordStore`."
	units := segment(text, 1)
	findings := c3cNoConflation("synthetic", units)
	if len(findings) != 1 {
		t.Fatalf("c3cNoConflation() = %d findings, want 1: %+v", len(findings), findings)
	}
}

// TestC3cAllowsProhibitionAndRecordStoreInSeparateSentences proves step 4 of
// segment(): the prohibition and recordStore in separate sentences of the
// same paragraph do not conflate.
func TestC3cAllowsProhibitionAndRecordStoreInSeparateSentences(t *testing.T) {
	text := "Do NOT detect the artifact store, and do NOT branch on it. `recordStore` is different."
	units := segment(text, 1)
	if len(units) != 2 {
		t.Fatalf("segment() returned %d units, want 2 (proves the sentence cut): %+v", len(units), units)
	}
	findings := c3cNoConflation("synthetic", units)
	if len(findings) != 0 {
		t.Fatalf("c3cNoConflation() = %d findings, want 0: %+v", len(findings), findings)
	}
}

// TestC3aFlagsMissingScopingSentence pins the spec's second failure
// scenario: removing Section B's scoping sentence.
func TestC3aFlagsMissingScopingSentence(t *testing.T) {
	text := "The orchestrator injects the artifact store and the locators native status already resolved. Read what you are given.\n\n" +
		"**Do NOT detect the artifact store, and do NOT branch on it.** The dispatcher resolved it from the store the workspace DECLARES."
	units := segment(text, 1)
	findings := c3aScopingPresent("synthetic-section-b", units)
	if len(findings) != 1 {
		t.Fatalf("c3aScopingPresent() = %d findings, want 1: %+v", len(findings), findings)
	}
}

// TestC3bFlagsWhenProhibitionMarkersVanish is the guard's own anti-vacuity
// self-check: when every prohibition literal disappears from Section B, the
// guard must say so rather than pass with nothing left to compare.
func TestC3bFlagsWhenProhibitionMarkersVanish(t *testing.T) {
	text := "The orchestrator injects the store you were given. This prohibition is about the artifact " +
		"store specifically; recordStore may legitimately behave differently."
	units := segment(text, 1)
	findings := c3bMarkersStillMatch("synthetic-section-b", units)
	if len(findings) != 1 {
		t.Fatalf("c3bMarkersStillMatch() = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "must be updated, not deleted") {
		t.Fatalf("finding.Why = %q, want it to state the markers must be updated, not deleted", findings[0].Why)
	}
}
