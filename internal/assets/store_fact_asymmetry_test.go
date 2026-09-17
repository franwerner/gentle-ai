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
	// reportMarkers and forwardMarkers are anchored to the backtick-plus-
	// recordStore span so the artifactStore sentence in the same bullet
	// cannot satisfy them (see the design's "framing is checked by surface
	// direction" decision).
	reportMarkers = []string{
		"reports it as `recordStore",
		"reporting it in `recordStore",
	}
	forwardMarkers = []string{
		"into every phase launch",
	}
)

const (
	facingPhase      = "phase"
	facingDispatcher = "dispatcher"
)

// surfaceSpec declares, for one prose surface, which framing checker runs
// (Facing) and a verdict for every condition in the C2/C3a/C3b/C3c family.
// Conditions is a map with an asserted key set, never bare struct bools,
// because a Go zero-value bool cannot be told apart from a field nobody
// filled in — see TestSurfaceRegistryDeclaresAVerdictForEveryCondition.
type surfaceSpec struct {
	ID            string
	Path          string
	Heading       string
	SchemaHeading string // non-empty → C5 runs on this surface
	Facing        string
	Conditions    map[string]bool
}

// surfaceRegistry is the single source of which surfaces the guard covers
// and what applies to each. A false verdict here is a recorded decision
// (see the design's "Architecture Decisions"), not an omission.
var surfaceRegistry = []surfaceSpec{
	{
		ID:      "section-B",
		Path:    "skills/_shared/sdd-phase-common.md",
		Heading: "## B. Artifact Retrieval",
		Facing:  facingPhase,
		Conditions: map[string]bool{
			"C2": false, "C3a": true, "C3b": true, "C3c": true,
		},
	},
	{
		ID:      "activation",
		Path:    "skills/_shared/openrecord-convention.md",
		Heading: "## Activation",
		Facing:  facingPhase,
		Conditions: map[string]bool{
			"C2": true, "C3a": false, "C3b": false, "C3c": true,
		},
	},
	{
		ID:      "claude-workflow",
		Path:    "claude/sdd-orchestrator-workflow.md",
		Heading: "### Native SDD Dispatcher Guard",
		Facing:  facingDispatcher,
		Conditions: map[string]bool{
			"C2": false, "C3a": false, "C3b": true, "C3c": true,
		},
	},
	{
		ID:            "status-contract",
		Path:          "skills/_shared/sdd-status-contract.md",
		Heading:       "## Native Engine",
		SchemaHeading: "## Status Schema",
		Facing:        facingDispatcher,
		Conditions: map[string]bool{
			"C2": false, "C3a": false, "C3b": true, "C3c": true,
		},
	},
}

// validateSurfaceRegistry fails a surface whose Conditions map is missing
// one of the required keys, or whose Facing is not one of the two known
// values — the anti-vacuity check for the registry itself.
func validateSurfaceRegistry(registry []surfaceSpec) []finding {
	requiredConditions := []string{"C2", "C3a", "C3b", "C3c"}
	var findings []finding
	for _, spec := range registry {
		if spec.Facing != facingPhase && spec.Facing != facingDispatcher {
			findings = append(findings, finding{
				Path: spec.ID, Kind: "registry",
				Why: "Facing is not one of the two known values: " + spec.Facing,
			})
		}
		for _, key := range requiredConditions {
			if _, ok := spec.Conditions[key]; !ok {
				findings = append(findings, finding{
					Path: spec.ID, Kind: "registry",
					Why: "Conditions map is missing key " + key,
				})
			}
		}
	}
	return findings
}

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

// c4ForwardFraming fails when a dispatcher-facing surface names no
// recordStore mention at all (the clause has been removed), or names it
// without also stating it is reported and stating it is forwarded. Unlike
// C1, C4 is mandatory: on these two surfaces the mention itself is the
// assertion, so it is never gated behind a conditional presence check.
func c4ForwardFraming(path, text string, baseLine int) []finding {
	if !strings.Contains(text, "recordStore") {
		return []finding{{
			Path: path, Line: baseLine, Kind: "surface",
			Text: truncateRunes(text, 200),
			Why:  "surface does not name recordStore at all — the clause appears to have been removed",
		}}
	}
	var findings []finding
	if !containsAny(text, reportMarkers) {
		findings = append(findings, finding{
			Path: path, Line: baseLine, Kind: "surface",
			Text: truncateRunes(text, 200),
			Why:  "names recordStore with no report-framing marker present in the surface",
		})
	}
	if !containsAny(text, forwardMarkers) {
		findings = append(findings, finding{
			Path: path, Line: baseLine, Kind: "surface",
			Text: truncateRunes(text, 200),
			Why:  "names recordStore with no forward marker present in the surface",
		})
	}
	return findings
}

// schemaDeclaresRecordStore reports whether the schema region carries a line
// equal to "recordStore:" immediately followed by lines prefixed "  declared:"
// and "  resolved:" — binding the sub-keys to their parent rather than
// matching them bare, so the check stays correct if artifactStore ever
// becomes an object with its own declared/resolved children.
func schemaDeclaresRecordStore(text string) bool {
	lines := strings.Split(text, "\n")
	for i, raw := range lines {
		if strings.TrimRight(raw, "\r") != "recordStore:" {
			continue
		}
		if i+2 >= len(lines) {
			return false
		}
		return strings.HasPrefix(lines[i+1], "  declared:") && strings.HasPrefix(lines[i+2], "  resolved:")
	}
	return false
}

// c5ProseSchemaAgreement fails when the Native Engine prose does not name
// both recordStore.declared and recordStore.resolved, or when the Status
// Schema region does not declare recordStore with both child lines — naming
// which half is missing. It reads two raw regions and never goes through
// segment(), which skips fenced content by design and would see no units in
// the StatusV2Projection yaml block.
func c5ProseSchemaAgreement(path, proseText, schemaText string, proseBase, schemaBase int) []finding {
	var findings []finding
	if !strings.Contains(proseText, "recordStore.declared") || !strings.Contains(proseText, "recordStore.resolved") {
		findings = append(findings, finding{
			Path: path, Line: proseBase, Kind: "surface",
			Text: truncateRunes(proseText, 200),
			Why:  "prose does not name both recordStore.declared and recordStore.resolved",
		})
	}
	if !schemaDeclaresRecordStore(schemaText) {
		findings = append(findings, finding{
			Path: path, Line: schemaBase, Kind: "surface",
			Text: truncateRunes(schemaText, 200),
			Why:  "schema region does not declare recordStore with both a declared: and a resolved: child line",
		})
	}
	return findings
}

func reportFindings(t *testing.T, findings []finding) {
	t.Helper()
	for _, f := range findings {
		t.Errorf("%s:%d [%s] %s: %s", f.Path, f.Line, f.Kind, f.Why, f.Text)
	}
}

// loadedSurface is the segmented, ready-to-check form of one surfaceRegistry
// row, computed once per test run.
type loadedSurface struct {
	text           string
	baseLine       int
	units          []unit
	schemaText     string
	schemaBaseLine int
}

// TestStoreFactsAreNotConflated pins the spec's "The two store facts are not
// conflated" scenario over the four shipped surfaces the registry declares.
// One t.Run per condition, each iterating the registry filtered by its own
// flag — an empty markdownSection result is always t.Fatalf, never a skip,
// so a renamed heading breaks the guard loudly instead of passing with
// nothing to check.
func TestStoreFactsAreNotConflated(t *testing.T) {
	surfaces := make(map[string]loadedSurface, len(surfaceRegistry))
	for _, spec := range surfaceRegistry {
		content := MustRead(spec.Path)
		text := markdownSection(content, spec.Heading)
		if text == "" {
			t.Fatalf("%s: heading %q not found", spec.Path, spec.Heading)
		}
		baseLine := 1 + strings.Count(content[:strings.Index(content, spec.Heading)], "\n")
		loaded := loadedSurface{text: text, baseLine: baseLine, units: segment(text, baseLine)}

		if spec.SchemaHeading != "" {
			schemaText := markdownSection(content, spec.SchemaHeading)
			if schemaText == "" {
				t.Fatalf("%s: schema heading %q not found", spec.Path, spec.SchemaHeading)
			}
			loaded.schemaText = schemaText
			loaded.schemaBaseLine = 1 + strings.Count(content[:strings.Index(content, spec.SchemaHeading)], "\n")
		}
		surfaces[spec.ID] = loaded
	}

	t.Run("C1", func(t *testing.T) {
		for _, spec := range surfaceRegistry {
			if spec.Facing != facingPhase {
				continue
			}
			s := surfaces[spec.ID]
			reportFindings(t, c1InjectionFraming(spec.Path, s.text, s.baseLine))
		}
	})

	t.Run("C2", func(t *testing.T) {
		for _, spec := range surfaceRegistry {
			if !spec.Conditions["C2"] {
				continue
			}
			s := surfaces[spec.ID]
			reportFindings(t, c2NoStatusLookup(spec.Path, s.text, s.baseLine))
		}
	})

	t.Run("C3a", func(t *testing.T) {
		for _, spec := range surfaceRegistry {
			if !spec.Conditions["C3a"] {
				continue
			}
			s := surfaces[spec.ID]
			reportFindings(t, c3aScopingPresent(spec.Path, s.units))
		}
	})

	t.Run("C3b", func(t *testing.T) {
		for _, spec := range surfaceRegistry {
			if !spec.Conditions["C3b"] {
				continue
			}
			s := surfaces[spec.ID]
			reportFindings(t, c3bMarkersStillMatch(spec.Path, s.units))
		}
	})

	t.Run("C3c", func(t *testing.T) {
		for _, spec := range surfaceRegistry {
			if !spec.Conditions["C3c"] {
				continue
			}
			s := surfaces[spec.ID]
			reportFindings(t, c3cNoConflation(spec.Path, s.units))
		}
	})

	t.Run("C4", func(t *testing.T) {
		for _, spec := range surfaceRegistry {
			if spec.Facing != facingDispatcher {
				continue
			}
			s := surfaces[spec.ID]
			reportFindings(t, c4ForwardFraming(spec.Path, s.text, s.baseLine))
		}
	})

	t.Run("C5", func(t *testing.T) {
		for _, spec := range surfaceRegistry {
			if spec.SchemaHeading == "" {
				continue
			}
			s := surfaces[spec.ID]
			reportFindings(t, c5ProseSchemaAgreement(spec.Path, s.text, s.schemaText, s.baseLine, s.schemaBaseLine))
		}
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

// TestC4FlagsDispatcherSurfaceWithNoRecordStoreMention pins the CRITICAL this
// checker exists to close: a guard-shaped dispatcher surface that names only
// the artifactStore sentence and bullets, with no recordStore mention at
// all, must fail naming the missing mention — without this fixture, deleting
// the clause would be invisible.
func TestC4FlagsDispatcherSurfaceWithNoRecordStoreMention(t *testing.T) {
	text := "It resolves the artifact store the workspace declares and reports it in `artifactStore`.\n\n" +
		"- Do NOT determine the artifact store yourself, and do NOT branch on it. The dispatcher already did.\n" +
		"- Use the dispatcher for every store and treat its JSON as authoritative over prompt inference."
	findings := c4ForwardFraming("synthetic-dispatcher", text, 1)
	if len(findings) < 1 {
		t.Fatalf("c4ForwardFraming() = %d findings, want at least 1", len(findings))
	}
	found := false
	for _, f := range findings {
		if strings.Contains(f.Why, "does not name recordStore") {
			found = true
		}
	}
	if !found {
		t.Fatalf("findings = %+v, want one naming the missing recordStore mention", findings)
	}
}

// TestC4FlagsRecordStoreMentionedButNotForwarded proves C4 is more than
// strings.Contains("recordStore"): a surface that reports the fact but never
// states it forwards it must fail on the forward half alone.
func TestC4FlagsRecordStoreMentionedButNotForwarded(t *testing.T) {
	text := "It also resolves the record store the workspace declares, reporting it in `recordStore.resolved`."
	findings := c4ForwardFraming("synthetic-dispatcher", text, 1)
	if len(findings) != 1 {
		t.Fatalf("c4ForwardFraming() = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "forward") {
		t.Fatalf("finding.Why = %q, want it to name the missing forward marker", findings[0].Why)
	}
}

// TestC4AcceptsBothDispatcherRegisters proves the report/forward families are
// not so narrow that they match one file only: one synthetic surface mirrors
// claude-workflow's terse one-line bullet register, the other mirrors
// status-contract's long-bullet register — both must pass with zero findings.
func TestC4AcceptsBothDispatcherRegisters(t *testing.T) {
	registers := map[string]string{
		"terse register (claude-workflow-shaped)": "- Forward `recordStore.resolved` into every phase launch alongside " +
			"`artifactStore` and `artifactPaths`; do NOT resolve the record store yourself — the dispatcher already " +
			"did, reporting it in `recordStore.resolved`.",
		"long-bullet register (status-contract-shaped)": "The native dispatcher also resolves the record store the " +
			"workspace declares in `sdd.record_store` and reports it as `recordStore.declared` (the verbatim config " +
			"value) and `recordStore.resolved` (`openrecord`, or empty when the workspace declares no record " +
			"store). Forward `recordStore` into every phase launch alongside `artifactStore` and `artifactPaths`; " +
			"never resolve it yourself.",
	}
	for name, text := range registers {
		findings := c4ForwardFraming("synthetic-"+name, text, 1)
		if len(findings) != 0 {
			t.Fatalf("c4ForwardFraming(%s) = %d findings, want 0: %+v", name, len(findings), findings)
		}
	}
}

// TestC5FlagsProseWithoutSchemaDeclaration proves C5 asserts the schema half:
// prose names both JSON fields, but the schema region drops the recordStore
// block entirely.
func TestC5FlagsProseWithoutSchemaDeclaration(t *testing.T) {
	prose := "reports it as `recordStore.declared` (the verbatim config value) and `recordStore.resolved`."
	schema := "```yaml\nschemaName: gentle-ai.sdd-status\nschemaVersion: 2\n" +
		"artifactStore: openspec | engram | hybrid | none\n```"
	findings := c5ProseSchemaAgreement("synthetic-status-contract", prose, schema, 1, 10)
	if len(findings) != 1 {
		t.Fatalf("c5ProseSchemaAgreement() = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "schema") {
		t.Fatalf("finding.Why = %q, want it to name the schema half", findings[0].Why)
	}
}

// TestC5FlagsSchemaDeclarationWithoutProse proves C5 asserts the prose half
// too — without this fixture, a one-sided C5 checking only the schema region
// would still pass TestC5FlagsProseWithoutSchemaDeclaration.
func TestC5FlagsSchemaDeclarationWithoutProse(t *testing.T) {
	prose := "The native dispatcher resolves the artifact store the workspace declares and reports it in `artifactStore`."
	schema := "```yaml\nartifactStore: openspec | engram | hybrid | none\nrecordStore:\n" +
		"  declared: <verbatim sdd.record_store value from openspec/config.yaml, empty when absent>\n" +
		"  resolved: openrecord | <empty>\n```"
	findings := c5ProseSchemaAgreement("synthetic-status-contract", prose, schema, 1, 10)
	if len(findings) != 1 {
		t.Fatalf("c5ProseSchemaAgreement() = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "prose") {
		t.Fatalf("finding.Why = %q, want it to name the prose half", findings[0].Why)
	}
}

// TestC3cDetectsConflationInBulletRegister proves C3c reaches the new
// surfaces' bullet register, not only the paragraph shape the shipped
// fixture uses.
func TestC3cDetectsConflationInBulletRegister(t *testing.T) {
	text := "- Do NOT determine the artifact store yourself, and do NOT branch on it, including `recordStore`."
	units := segment(text, 1)
	findings := c3cNoConflation("synthetic", units)
	if len(findings) != 1 {
		t.Fatalf("c3cNoConflation() = %d findings, want 1: %+v", len(findings), findings)
	}
	if findings[0].Kind != unitListEntry {
		t.Fatalf("finding.Kind = %q, want %q", findings[0].Kind, unitListEntry)
	}
}

// TestC3bFlagsWhenStatusContractProhibitionVanishes proves the anti-vacuity
// self-check extends to status-contract's own prohibition literal
// ("never branch on the store"), which section-B's shipped fixture never
// exercises.
func TestC3bFlagsWhenStatusContractProhibitionVanishes(t *testing.T) {
	text := "The native dispatcher resolves the artifact store the workspace declares and reports it in " +
		"`artifactStore`. Never re-resolve artifact status yourself, and treat the dispatcher's locators as authoritative."
	units := segment(text, 1)
	findings := c3bMarkersStillMatch("synthetic-status-contract", units)
	if len(findings) != 1 {
		t.Fatalf("c3bMarkersStillMatch() = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "must be updated, not deleted") {
		t.Fatalf("finding.Why = %q, want it to state the markers must be updated, not deleted", findings[0].Why)
	}
}

// TestSurfaceRegistryDeclaresAVerdictForEveryCondition is what makes "scope
// is declared, never omitted" enforceable rather than aspirational: the real
// shipped registry passes; a copy missing a Conditions key or carrying an
// unknown Facing value fails, naming the offending surface and key/value.
func TestSurfaceRegistryDeclaresAVerdictForEveryCondition(t *testing.T) {
	if findings := validateSurfaceRegistry(surfaceRegistry); len(findings) != 0 {
		t.Fatalf("validateSurfaceRegistry(surfaceRegistry) = %d findings, want 0: %+v", len(findings), findings)
	}

	missingKey := []surfaceSpec{{
		ID: "broken-missing-key", Path: "x", Heading: "x", Facing: facingPhase,
		Conditions: map[string]bool{"C2": false, "C3a": true, "C3b": true},
	}}
	findings := validateSurfaceRegistry(missingKey)
	if len(findings) != 1 {
		t.Fatalf("validateSurfaceRegistry(missing key) = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "C3c") {
		t.Fatalf("finding.Why = %q, want it to name the missing key C3c", findings[0].Why)
	}

	unknownFacing := []surfaceSpec{{
		ID: "broken-facing", Path: "x", Heading: "x", Facing: "neither",
		Conditions: map[string]bool{"C2": false, "C3a": false, "C3b": false, "C3c": false},
	}}
	findings = validateSurfaceRegistry(unknownFacing)
	if len(findings) != 1 {
		t.Fatalf("validateSurfaceRegistry(unknown Facing) = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "neither") {
		t.Fatalf("finding.Why = %q, want it to name the offending Facing value", findings[0].Why)
	}
}
