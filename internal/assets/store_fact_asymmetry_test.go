package assets

import (
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"
)

// #openrecord-unconditional: openrecord used to be gated on a declared
// `sdd.record_store` axis, and the shipped markdown carried that gating three
// ways — a `<--:openrecord-->` wrapper around every participating region, a
// "when status reports `recordStore.resolved: openrecord`" clause on every
// pointer, and prose describing what a phase does when the store is
// undeclared. The axis is gone: openrecord is installed unconditionally and
// nothing declares or resolves a record store any more.
//
// This guard is the reconverted form of the old store-fact asymmetry checker.
// Its subject changed, not its discipline: it pins that the conditionality
// stays gone (U1, U2) while the substance it used to wrap stays present (U3),
// that the one genuinely surviving `recordStore` surface — the frozen v2
// status projection — keeps its shape pinned to the constant (U5), and that
// the artifact-store facts the old checker also carried (C3a, C3b, C3c) are
// untouched. Nothing short of a runtime check catches the day someone
// reintroduces a wrapper, a gating clause, or an instruction to read the axis.

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

const (
	openRecordDelimOpen  = "<--:openrecord-->"
	openRecordDelimClose = "<--:/openrecord-->"
)

var (
	prohibitionMarkers = []string{
		"Do NOT detect the artifact store",
		"Do NOT determine the artifact store yourself",
		"do NOT branch on it",
		"never branch on the store",
	}
	scopingMarkers = []string{
		"This prohibition is about the",
		"artifact",
	}
	// gatingMarkers are the literal shapes a record-store condition took
	// before the axis was deleted. Every one of them names the axis itself
	// rather than merely naming openrecord, so unconditional prose that
	// simply says "the openrecord record store" never matches.
	gatingMarkers = []string{
		"recordStore.resolved",
		"recordStore.declared",
		"sdd.record_store",
		"record_store",
		"the record store the workspace declares",
		"declares no record store",
		"when the injected",
		"is empty (undeclared)",
	}
)

// surfaceSpec declares, for one prose surface, the passages that MUST survive
// on it and a verdict for every condition in the C3a/C3b family. Conditions is
// a map with an asserted key set, never bare struct bools, because a Go
// zero-value bool cannot be told apart from a field nobody filled in — see
// TestSurfaceRegistryDeclaresAVerdictForEveryCondition.
type surfaceSpec struct {
	ID            string
	Path          string
	Heading       string
	SchemaHeading string // non-empty → U5 runs on this surface
	Passages      []string
	Conditions    map[string]bool
}

// surfaceRegistry is the single source of which surfaces the guard covers and
// what applies to each. A false verdict here is a recorded decision, not an
// omission, and an empty Passages list is rejected outright by
// TestSurfaceRegistryDeclaresAVerdictForEveryCondition — a surface nobody
// pinned content on is exactly the vacuous checker this program hit twice.
var surfaceRegistry = []surfaceSpec{
	{
		ID:      "section-B",
		Path:    "skills/_shared/sdd-phase-common.md",
		Heading: "## B. Artifact Retrieval",
		Passages: []string{
			"This prohibition is about the *artifact* store specifically.",
			"read `skills/_shared/openrecord-convention.md` and follow it",
		},
		Conditions: map[string]bool{"C3a": true, "C3b": true, "C3c": true},
	},
	{
		ID:      "activation",
		Path:    "skills/_shared/openrecord-convention.md",
		Heading: "## Activation",
		Passages: []string{
			"This convention always applies.",
			"openrecord is installed unconditionally",
		},
		Conditions: map[string]bool{"C3a": false, "C3b": false, "C3c": true},
	},
	{
		ID:      "writer-contract",
		Path:    "skills/_shared/openrecord-convention.md",
		Heading: "## Writing records",
		Passages: []string{
			"Only `sdd-apply` writes into the record store, in the same step that implements the governing code",
			"**A record store that is not on disk.** Write nothing and create nothing",
		},
		Conditions: map[string]bool{"C3a": false, "C3b": false, "C3c": true},
	},
	{
		ID:      "claude-workflow",
		Path:    "claude/sdd-orchestrator-workflow.md",
		Heading: "### Native SDD Dispatcher Guard",
		Passages: []string{
			"Do NOT determine the artifact store yourself, and do NOT branch on it.",
		},
		Conditions: map[string]bool{"C3a": false, "C3b": true, "C3c": true},
	},
	{
		ID:            "status-contract",
		Path:          "skills/_shared/sdd-status-contract.md",
		Heading:       "## Native Engine",
		SchemaHeading: "## Status Schema",
		Passages: []string{
			"never branch on the store",
		},
		Conditions: map[string]bool{"C3a": false, "C3b": true, "C3c": true},
	},
	{
		ID:      "apply-step-5a",
		Path:    "skills/sdd-apply/SKILL.md",
		Heading: "#### Step 5a: Materialize Records, Then Prune What Moved",
		Passages: []string{
			"When this batch leaves no pending task in the tasks artifact, write the change's durable records into the record store before you return",
			"When the batch still leaves pending tasks, write nothing, prune nothing and report nothing",
		},
		Conditions: map[string]bool{"C3a": false, "C3b": false, "C3c": true},
	},
	{
		ID:      "spec-step-4a",
		Path:    "skills/sdd-spec/SKILL.md",
		Heading: "#### Step 4a: Name the Type and Compose Whole Sections",
		Passages: []string{
			"Each capability this change touches also becomes a durable record",
			"`sdd-apply` is the store's sole writer",
		},
		Conditions: map[string]bool{"C3a": false, "C3b": false, "C3c": true},
	},
}

// participatingPhasePointers is the pointer bullet each of the six
// participating phases carries, unwrapped and unconditional. The bullet is the
// smallest thing that could quietly regain a condition, so it is pinned in
// full rather than by its citation alone.
var participatingPhasePointers = map[string]string{
	"sdd-explore": "- **record store**: read and follow `skills/_shared/openrecord-convention.md`.",
	"sdd-propose": "- **record store**: read and follow `skills/_shared/openrecord-convention.md`.",
	"sdd-spec":    "- **record store**: read and follow `skills/_shared/openrecord-convention.md`.",
	"sdd-design":  "- **record store**: read and follow `skills/_shared/openrecord-convention.md`.",
	"sdd-apply":   "- **record store**: read and follow `skills/_shared/openrecord-convention.md`.",
	"sdd-verify":  "- **record store**: read and follow `skills/_shared/openrecord-convention.md`. Run `openrecord validate` for structural shape and separately check that each record's prose still describes the system as actually built.",
}

// validateSurfaceRegistry fails a surface whose Conditions map is missing one
// of the required keys, or which pins no passage at all — the anti-vacuity
// check for the registry itself.
func validateSurfaceRegistry(registry []surfaceSpec) []finding {
	requiredConditions := []string{"C3a", "C3b", "C3c"}
	var findings []finding
	for _, spec := range registry {
		if len(spec.Passages) == 0 {
			findings = append(findings, finding{
				Path: spec.ID, Kind: "registry",
				Why: "surface pins no passage; a registered surface with nothing to preserve checks nothing",
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
// first physical line within the surface. Fenced content is skipped by design,
// which is what keeps the frozen yaml projection out of the prose checkers.
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

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// u1NoConditionalWrapper fails when content carries either half of the
// `<--:openrecord-->` pair. The wrapper was never processed by anything — no
// renderer strips it, so it shipped verbatim into every installed skill — which
// is exactly why its reappearance has to be caught by a test rather than by
// a rendering step noticing it.
func u1NoConditionalWrapper(assetPath, content string) []finding {
	var findings []finding
	for _, delim := range []string{openRecordDelimOpen, openRecordDelimClose} {
		if idx := strings.Index(content, delim); idx >= 0 {
			findings = append(findings, finding{
				Path: assetPath, Line: 1 + strings.Count(content[:idx], "\n"), Kind: "delimiter",
				Text: delim,
				Why:  "asset carries an openrecord conditional wrapper; openrecord is unconditional and nothing wraps it",
			})
		}
	}
	return findings
}

// u2NoAxisGating fails when a prose unit names the deleted record-store axis:
// a `recordStore.resolved`/`recordStore.declared` gate, an `sdd.record_store`
// config lookup, or the undeclared-store prose that described a state which
// can no longer occur. Unconditional prose naming "the openrecord record
// store" matches none of these.
func u2NoAxisGating(assetPath string, units []unit) []finding {
	var findings []finding
	for _, u := range units {
		for _, marker := range gatingMarkers {
			if !strings.Contains(u.Text, marker) {
				continue
			}
			findings = append(findings, finding{
				Path: assetPath, Line: u.Line, Kind: u.Kind,
				Text: truncateRunes(u.Text, 200),
				Why:  "unit gates behaviour on the deleted record-store axis via " + marker,
			})
		}
	}
	return findings
}

// u3PassagesSurvive fails when a registered surface has lost a passage the
// registry says it must keep. This is the half that makes U1 and U2 safe to
// run: without it, deleting a section outright would satisfy both.
func u3PassagesSurvive(assetPath, text string, baseLine int, passages []string) []finding {
	var findings []finding
	for _, passage := range passages {
		if !strings.Contains(text, passage) {
			findings = append(findings, finding{
				Path: assetPath, Line: baseLine, Kind: "surface",
				Text: truncateRunes(passage, 200),
				Why:  "surface no longer carries a passage the registry requires; the content survives this change, only its conditionality was removed",
			})
		}
	}
	return findings
}

// c3aScopingPresent fails when no unit of Section B scopes the branching
// prohibition to the artifact store.
func c3aScopingPresent(assetPath string, units []unit) []finding {
	for _, u := range units {
		if containsAll(u.Text, scopingMarkers) {
			return nil
		}
	}
	return []finding{{
		Path: assetPath, Kind: "section",
		Why: "no unit scopes the branching prohibition to the artifact store",
	}}
}

// c3bMarkersStillMatch fails when none of a surface's units carries a
// prohibition marker any more — a sign the markers went stale.
func c3bMarkersStillMatch(assetPath string, units []unit) []finding {
	for _, u := range units {
		if containsAny(u.Text, prohibitionMarkers) {
			return nil
		}
	}
	return []finding{{
		Path: assetPath, Kind: "section",
		Why: "no unit carries a prohibition marker; the markers no longer match the shipped prose and must be updated, not deleted",
	}}
}

// c3cNoConflation fails when a unit carries both an artifact-store branching
// prohibition and the recordStore token. Before the axis was deleted this
// caught the two store facts being merged into one sentence; it now catches
// the axis being reintroduced through the prohibition that outlived it.
func c3cNoConflation(assetPath string, units []unit) []finding {
	var findings []finding
	for _, u := range units {
		if containsAny(u.Text, prohibitionMarkers) && strings.Contains(u.Text, "recordStore") {
			findings = append(findings, finding{
				Path: assetPath, Line: u.Line, Kind: u.Kind,
				Text: truncateRunes(u.Text, 200),
				Why:  "unit carries both an artifact-store branching prohibition and recordStore",
			})
		}
	}
	return findings
}

// schemaRecordStoreLines returns the two child lines of the frozen projection's
// `recordStore:` key, or ok=false when the key is absent or not followed by a
// `declared:` and a `resolved:` line — binding the sub-keys to their parent
// rather than matching them bare.
func schemaRecordStoreLines(text string) (declared, resolved string, ok bool) {
	lines := strings.Split(text, "\n")
	for i, raw := range lines {
		if strings.TrimRight(raw, "\r") != "recordStore:" {
			continue
		}
		if i+2 >= len(lines) {
			return "", "", false
		}
		declared, okDeclared := strings.CutPrefix(strings.TrimRight(lines[i+1], "\r"), "  declared:")
		resolved, okResolved := strings.CutPrefix(strings.TrimRight(lines[i+2], "\r"), "  resolved:")
		if !okDeclared || !okResolved {
			return "", "", false
		}
		return strings.TrimSpace(declared), strings.TrimSpace(resolved), true
	}
	return "", "", false
}

// u5SchemaPinsRecordStoreToTheConstant fails when the frozen v2 status schema
// drops `recordStore`, or describes it as anything but the constant the
// projection actually emits. The field survived the axis deletion: v2 changes
// additively, so dropping a key released in v2.8.2 needs a contract version
// bump, not an edit to the asset. The documented value therefore has to track
// StatusV2Projection's constant, not the config key that no longer exists.
func u5SchemaPinsRecordStoreToTheConstant(assetPath, schemaText string, schemaBase int) []finding {
	declared, resolved, ok := schemaRecordStoreLines(schemaText)
	if !ok {
		return []finding{{
			Path: assetPath, Line: schemaBase, Kind: "surface",
			Text: truncateRunes(schemaText, 200),
			Why:  "frozen schema region does not declare recordStore with both a declared: and a resolved: child line",
		}}
	}
	var findings []finding
	for _, pair := range []struct{ key, value string }{{"declared", declared}, {"resolved", resolved}} {
		if pair.value != openRecordProjectionConstant {
			findings = append(findings, finding{
				Path: assetPath, Line: schemaBase, Kind: "surface",
				Text: pair.key + ": " + pair.value,
				Why:  "frozen schema documents recordStore." + pair.key + " as something other than the constant the projection emits (" + openRecordProjectionConstant + ")",
			})
		}
	}
	return findings
}

// openRecordProjectionConstant mirrors internal/sddstatus's openRecordStoreV2,
// which is unexported. Both halves are pinned in lockstep by
// internal/sddstatus/status_v2_clean_break_test.go on the Go side and by U5 on
// the asset side.
const openRecordProjectionConstant = "openrecord"

func reportFindings(t *testing.T, findings []finding) {
	t.Helper()
	for _, f := range findings {
		t.Errorf("%s:%d [%s] %s: %s", f.Path, f.Line, f.Kind, f.Why, f.Text)
	}
}

// mustMutate returns src with the first occurrence of old replaced by new,
// and fails the test when the replacement changed nothing — a no-op mutation
// means the fixture's needle has drifted out of the source it was written
// against, and an unchanged input would pass every assertion below it.
func mustMutate(t *testing.T, src, old, new string) string {
	t.Helper()
	mutated := strings.Replace(src, old, new, 1)
	if mutated == src {
		t.Fatalf("mustMutate: needle %q not found in source: %q", old, truncateRunes(src, 200))
	}
	return mutated
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

func loadSurfaces(t *testing.T) map[string]loadedSurface {
	t.Helper()
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
	return surfaces
}

// markdownAssetPaths walks the whole embedded asset tree and returns every
// markdown file in it. U1 runs over this rather than over surfaceRegistry: a
// wrapper reintroduced in an asset nobody registered is exactly the shape a
// registry-scoped check cannot see.
func markdownAssetPaths(t *testing.T) []string {
	t.Helper()
	var paths []string
	err := fs.WalkDir(FS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && path.Ext(p) == ".md" {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(assets) error = %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("asset walk found no markdown files; U1 would pass over nothing")
	}
	return paths
}

// TestOpenRecordBehaviourIsUnconditional is the guard proper: U1 over the whole
// embedded tree, U2/U3 and the surviving artifact-store conditions over the
// registered surfaces, U5 over the frozen projection. One t.Run per condition,
// each iterating its own scope — an empty markdownSection result is always
// t.Fatalf inside loadSurfaces, never a skip, so a renamed heading breaks the
// guard loudly instead of passing with nothing to check.
func TestOpenRecordBehaviourIsUnconditional(t *testing.T) {
	surfaces := loadSurfaces(t)

	t.Run("U1-no-conditional-wrapper", func(t *testing.T) {
		for _, assetPath := range markdownAssetPaths(t) {
			reportFindings(t, u1NoConditionalWrapper(assetPath, MustRead(assetPath)))
		}
	})

	t.Run("U2-no-axis-gating", func(t *testing.T) {
		for _, spec := range surfaceRegistry {
			s := surfaces[spec.ID]
			reportFindings(t, u2NoAxisGating(spec.Path, s.units))
		}
	})

	t.Run("U3-passages-survive", func(t *testing.T) {
		for _, spec := range surfaceRegistry {
			s := surfaces[spec.ID]
			reportFindings(t, u3PassagesSurvive(spec.Path, s.text, s.baseLine, spec.Passages))
		}
	})

	t.Run("U4-phase-pointers-unconditional", func(t *testing.T) {
		if len(participatingPhasePointers) != 6 {
			t.Fatalf("participatingPhasePointers has %d entries, want the six participating phases", len(participatingPhasePointers))
		}
		for skillID, bullet := range participatingPhasePointers {
			content := MustRead("skills/" + skillID + "/SKILL.md")
			if count := strings.Count(content, bullet); count != 1 {
				t.Errorf("%s carries the unconditional record-store pointer %d times, want exactly 1: %q", skillID, count, bullet)
			}
		}
	})

	t.Run("U5-frozen-schema-pins-the-constant", func(t *testing.T) {
		for _, spec := range surfaceRegistry {
			if spec.SchemaHeading == "" {
				continue
			}
			s := surfaces[spec.ID]
			reportFindings(t, u5SchemaPinsRecordStoreToTheConstant(spec.Path, s.schemaText, s.schemaBaseLine))
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

// TestU1FlagsReintroducedWrapperInAShippedAsset is U1's anti-vacuity fixture,
// run against the real asset rather than a synthetic string: the shipped file
// passes first, then wrapping one of its sections back up — via mustMutate,
// so a drifted needle turns this red instead of vacuously green — must be
// flagged, once per delimiter half.
func TestU1FlagsReintroducedWrapperInAShippedAsset(t *testing.T) {
	const assetPath = "skills/_shared/openrecord-convention.md"
	content := MustRead(assetPath)

	if findings := u1NoConditionalWrapper(assetPath, content); len(findings) != 0 {
		t.Fatalf("u1NoConditionalWrapper(shipped) = %d findings, want 0: %+v", len(findings), findings)
	}

	const heading = "## Activation"
	mutated := mustMutate(t, content, heading, openRecordDelimOpen+"\n"+heading)
	mutated = mustMutate(t, mutated, "## Per-phase responsibility map", openRecordDelimClose+"\n\n## Per-phase responsibility map")

	findings := u1NoConditionalWrapper(assetPath, mutated)
	if len(findings) != 2 {
		t.Fatalf("u1NoConditionalWrapper(rewrapped) = %d findings, want 2 (one per delimiter half): %+v", len(findings), findings)
	}
	for _, f := range findings {
		if !strings.Contains(f.Why, "conditional wrapper") {
			t.Fatalf("finding.Why = %q, want it to name the conditional wrapper", f.Why)
		}
	}
}

// TestU1ReachesUnregisteredAssets proves U1's scope is the whole embedded
// tree, not surfaceRegistry: the walk must reach an asset no registry row
// names, so a wrapper smuggled into one is still caught.
func TestU1ReachesUnregisteredAssets(t *testing.T) {
	registered := map[string]bool{}
	for _, spec := range surfaceRegistry {
		registered[spec.Path] = true
	}
	for _, assetPath := range markdownAssetPaths(t) {
		if !registered[assetPath] {
			return
		}
	}
	t.Fatal("every markdown asset is registered; U1's whole-tree scope proves nothing beyond the registry")
}

// TestU2FlagsReintroducedGatingClause pins the gating shape this change
// removed: the shipped pointer passes, and putting the "when status reports
// recordStore.resolved" condition back on it must fire.
func TestU2FlagsReintroducedGatingClause(t *testing.T) {
	const assetPath = "skills/sdd-apply/SKILL.md"
	content := MustRead(assetPath)
	section := markdownSection(content, "#### Step 5a: Materialize Records, Then Prune What Moved")
	if section == "" {
		t.Fatal("heading \"#### Step 5a: Materialize Records, Then Prune What Moved\" not found")
	}

	if findings := u2NoAxisGating(assetPath, segment(section, 1)); len(findings) != 0 {
		t.Fatalf("u2NoAxisGating(shipped) = %d findings, want 0: %+v", len(findings), findings)
	}

	mutated := mustMutate(t, section,
		"When this batch leaves no pending task",
		"When status reports `recordStore.resolved: openrecord` AND this batch leaves no pending task")

	findings := u2NoAxisGating(assetPath, segment(mutated, 1))
	if len(findings) != 1 {
		t.Fatalf("u2NoAxisGating(gated) = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "recordStore.resolved") {
		t.Fatalf("finding.Why = %q, want it to name the gating marker", findings[0].Why)
	}
}

// TestU2FlagsReintroducedDeclarationLookup covers the other gating shape: an
// instruction to read the deleted `sdd.record_store` config key.
func TestU2FlagsReintroducedDeclarationLookup(t *testing.T) {
	text := "Read the value the workspace's `openspec/config.yaml` declares in `sdd.record_store` before acting."
	findings := u2NoAxisGating("synthetic-activation", segment(text, 1))
	if len(findings) != 2 {
		t.Fatalf("u2NoAxisGating() = %d findings, want 2 (sdd.record_store and its record_store substring): %+v", len(findings), findings)
	}
}

// TestU2IgnoresUnconditionalRecordStoreProse proves U2 is narrower than
// strings.Contains("record store"): prose that names the store without gating
// on an axis must not fire, or the guard would forbid the content this change
// deliberately kept.
func TestU2IgnoresUnconditionalRecordStoreProse(t *testing.T) {
	text := "Write the change's durable records into the openrecord record store before you return. " +
		"Only `sdd-apply` writes into the record store."
	if findings := u2NoAxisGating("synthetic", segment(text, 1)); len(findings) != 0 {
		t.Fatalf("u2NoAxisGating(unconditional prose) = %d findings, want 0: %+v", len(findings), findings)
	}
}

// TestU3FlagsDeletedPassage is the "content survives" half's own anti-vacuity
// fixture: the shipped surface satisfies its pins, then deleting one — via
// mustMutate — must report exactly that passage.
func TestU3FlagsDeletedPassage(t *testing.T) {
	const assetPath = "skills/_shared/openrecord-convention.md"
	content := MustRead(assetPath)
	section := markdownSection(content, "## Writing records")
	if section == "" {
		t.Fatal("heading \"## Writing records\" not found")
	}

	var spec surfaceSpec
	for _, candidate := range surfaceRegistry {
		if candidate.ID == "writer-contract" {
			spec = candidate
		}
	}
	if len(spec.Passages) == 0 {
		t.Fatal("surfaceRegistry no longer carries the writer-contract row")
	}

	if findings := u3PassagesSurvive(assetPath, section, 1, spec.Passages); len(findings) != 0 {
		t.Fatalf("u3PassagesSurvive(shipped) = %d findings, want 0: %+v", len(findings), findings)
	}

	for _, passage := range spec.Passages {
		t.Run(truncateRunes(passage, 40), func(t *testing.T) {
			mutated := mustMutate(t, section, passage, "")
			findings := u3PassagesSurvive(assetPath, mutated, 1, spec.Passages)
			if len(findings) != 1 {
				t.Fatalf("u3PassagesSurvive(mutated) = %d findings, want exactly 1: %+v", len(findings), findings)
			}
			if findings[0].Text != passage {
				t.Fatalf("finding.Text = %q, want %q", findings[0].Text, passage)
			}
		})
	}
}

// TestU4FlagsAReconditionedPhasePointer proves U4 pins the bullet's wording
// rather than its mere presence: re-adding the condition changes the bullet, so
// the exact-match count drops to zero.
func TestU4FlagsAReconditionedPhasePointer(t *testing.T) {
	const skillID = "sdd-design"
	content := MustRead("skills/" + skillID + "/SKILL.md")
	bullet := participatingPhasePointers[skillID]

	if count := strings.Count(content, bullet); count != 1 {
		t.Fatalf("shipped %s carries the pointer %d times, want 1", skillID, count)
	}

	mutated := mustMutate(t, content, bullet,
		"- **record store**: when status reports `recordStore.resolved: openrecord`, read and follow `skills/_shared/openrecord-convention.md`.")
	if count := strings.Count(mutated, bullet); count != 0 {
		t.Fatalf("reconditioned %s still carries the unconditional pointer %d times, want 0", skillID, count)
	}
}

// TestU5FlagsSchemaDroppingRecordStore proves U5's first branch: the frozen
// projection losing its recordStore block must fail. The field is not a
// leftover — v2 changes additively, so it stays until a contract bump.
func TestU5FlagsSchemaDroppingRecordStore(t *testing.T) {
	schema := "```yaml\nschemaName: gentle-ai.sdd-status\nschemaVersion: 2\n" +
		"artifactStore: openspec | engram | hybrid | none\n```"
	findings := u5SchemaPinsRecordStoreToTheConstant("synthetic-status-contract", schema, 10)
	if len(findings) != 1 {
		t.Fatalf("u5SchemaPinsRecordStoreToTheConstant() = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "does not declare recordStore") {
		t.Fatalf("finding.Why = %q, want it to name the missing recordStore block", findings[0].Why)
	}
}

// TestU5FlagsSchemaDescribingTheDeletedAxis proves U5's second branch: keeping
// the block but documenting the old config-derived values — the exact drift
// this change repaired — must fail on both children.
func TestU5FlagsSchemaDescribingTheDeletedAxis(t *testing.T) {
	schema := "```yaml\nartifactStore: openspec | engram | hybrid | none\nrecordStore:\n" +
		"  declared: <verbatim sdd.record_store value from openspec/config.yaml, empty when absent>\n" +
		"  resolved: openrecord | <empty>\n```"
	findings := u5SchemaPinsRecordStoreToTheConstant("synthetic-status-contract", schema, 10)
	if len(findings) != 2 {
		t.Fatalf("u5SchemaPinsRecordStoreToTheConstant() = %d findings, want 2 (declared and resolved): %+v", len(findings), findings)
	}
	for _, f := range findings {
		if !strings.Contains(f.Why, openRecordProjectionConstant) {
			t.Fatalf("finding.Why = %q, want it to name the projection constant", f.Why)
		}
	}
}

// TestC3cDetectsConflationFixture pins the conflation shape: widening the
// artifact-store prohibition to also name recordStore.
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

// TestC3aFlagsMissingScopingSentence pins the scoping failure scenario:
// removing Section B's scoping sentence.
func TestC3aFlagsMissingScopingSentence(t *testing.T) {
	text := "The orchestrator injects the artifact store and the locators native status already resolved. Read what you are given.\n\n" +
		"**Do NOT detect the artifact store, and do NOT branch on it.** The dispatcher resolved it from the store the workspace DECLARES."
	units := segment(text, 1)
	findings := c3aScopingPresent("synthetic-section-b", units)
	if len(findings) != 1 {
		t.Fatalf("c3aScopingPresent() = %d findings, want 1: %+v", len(findings), findings)
	}
}

// TestC3bFlagsWhenProhibitionMarkersVanish is the artifact-store half's own
// anti-vacuity self-check: when every prohibition literal disappears from a
// surface, the guard must say so rather than pass with nothing left to
// compare.
func TestC3bFlagsWhenProhibitionMarkersVanish(t *testing.T) {
	text := "The orchestrator injects the store you were given. This prohibition is about the artifact " +
		"store specifically; the record store is not an axis at all."
	units := segment(text, 1)
	findings := c3bMarkersStillMatch("synthetic-section-b", units)
	if len(findings) != 1 {
		t.Fatalf("c3bMarkersStillMatch() = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "must be updated, not deleted") {
		t.Fatalf("finding.Why = %q, want it to state the markers must be updated, not deleted", findings[0].Why)
	}
}

// TestC3cDetectsConflationInBulletRegister proves C3c reaches the bullet
// register, not only the paragraph shape the shipped fixture uses.
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

// TestSurfaceRegistryDeclaresAVerdictForEveryCondition is what makes "scope is
// declared, never omitted" enforceable rather than aspirational: the real
// shipped registry passes; a copy missing a Conditions key, or pinning no
// passage at all, fails naming the offending surface.
func TestSurfaceRegistryDeclaresAVerdictForEveryCondition(t *testing.T) {
	if findings := validateSurfaceRegistry(surfaceRegistry); len(findings) != 0 {
		t.Fatalf("validateSurfaceRegistry(surfaceRegistry) = %d findings, want 0: %+v", len(findings), findings)
	}

	missingKey := []surfaceSpec{{
		ID: "broken-missing-key", Path: "x", Heading: "x",
		Passages:   []string{"something"},
		Conditions: map[string]bool{"C3a": true, "C3b": true},
	}}
	findings := validateSurfaceRegistry(missingKey)
	if len(findings) != 1 {
		t.Fatalf("validateSurfaceRegistry(missing key) = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "C3c") {
		t.Fatalf("finding.Why = %q, want it to name the missing key C3c", findings[0].Why)
	}

	noPassages := []surfaceSpec{{
		ID: "broken-no-passages", Path: "x", Heading: "x",
		Conditions: map[string]bool{"C3a": false, "C3b": false, "C3c": false},
	}}
	findings = validateSurfaceRegistry(noPassages)
	if len(findings) != 1 {
		t.Fatalf("validateSurfaceRegistry(no passages) = %d findings, want 1: %+v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Why, "pins no passage") {
		t.Fatalf("finding.Why = %q, want it to name the empty passage list", findings[0].Why)
	}
}
