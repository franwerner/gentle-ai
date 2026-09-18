package assets

import (
	"strings"
	"testing"
)

const (
	openRecordDelimOpen  = "<--:openrecord-->"
	openRecordDelimClose = "<--:/openrecord-->"
)

// TestOpenRecordPointerBulletsReachExactlySixPhases pins the spec's
// requirement: explore, propose, spec, design, apply and verify each carry
// exactly one openrecord pointer line, delimiter-bounded; every other phase
// carries none.
func TestOpenRecordPointerBulletsReachExactlySixPhases(t *testing.T) {
	participating := []string{"sdd-explore", "sdd-propose", "sdd-spec", "sdd-design", "sdd-apply", "sdd-verify"}
	nonParticipating := []string{"sdd-init", "sdd-onboard", "sdd-research", "sdd-tasks", "sdd-archive"}

	for _, skillID := range participating {
		t.Run(skillID, func(t *testing.T) {
			content, err := Read("skills/" + skillID + "/SKILL.md")
			if err != nil {
				t.Fatalf("Read() error = %v", err)
			}
			if count := strings.Count(content, "openrecord-convention.md"); count != 1 {
				t.Fatalf("%s references openrecord-convention.md %d times, want exactly 1", skillID, count)
			}
			assertDelimiterPaired(t, content)
		})
	}

	for _, skillID := range nonParticipating {
		t.Run(skillID, func(t *testing.T) {
			content, err := Read("skills/" + skillID + "/SKILL.md")
			if err != nil {
				t.Fatalf("Read() error = %v", err)
			}
			if strings.Contains(content, "openrecord") {
				t.Fatalf("%s carries an openrecord reference; this phase is not supposed to participate", skillID)
			}
		})
	}
}

// TestOpenRecordConventionFileIsDelimited pins that the shared convention
// asset itself is bounded by the same open/close markers, so its content is
// locatable and removable as one unit.
func TestOpenRecordConventionFileIsDelimited(t *testing.T) {
	content, err := Read("skills/_shared/openrecord-convention.md")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	assertDelimiterPaired(t, content)
}

// Exact sentences pinned inside the two shared assets' <--:openrecord-->
// blocks. Checking literal-substring presence of short tokens
// (recordStore.resolved, openrecord, type: architecture) would let a
// garbled rewrite that still contains those tokens pass; these are the
// precise conditional sentences the spec scenarios depend on, matched
// word-for-word via normalizedWords (case/punctuation/line-wrap
// insensitive, word-order sensitive — a dropped "not" or a swapped verb
// still fails).
const (
	// persistenceEnumDropSentence covers the spec scenario "Under
	// openrecord, decision is not an offered save type".
	persistenceEnumDropSentence = "When `recordStore.resolved: openrecord`, the `mem_save` instruction above is conditional: `decision` drops from the `type: \"{decision|bugfix|discovery|pattern}\"` enum at line 121."

	// persistenceUnaffectedTypesSentence covers the spec scenario "The
	// trailing carve-out sentence MUST be pinned": the other three Non-SDD
	// types stay unconditional even though decision drops out.
	persistenceUnaffectedTypesSentence = "`bugfix`, `discovery`, and `pattern` are unaffected and still go to Engram exactly as written above."

	// engramArtifactsUnaffectedSentence covers the spec scenario "A phase
	// artifact save is unaffected under openrecord".
	engramArtifactsUnaffectedSentence = "SDD phase artifacts — every phase from `explore` through `archive-report` — MUST keep saving to Engram with `type: architecture` in every `recordStore.resolved` mode, `openrecord` included."

	// engramHumanDecisionSentence covers the spec scenario "A human
	// decision is not written to Engram under openrecord".
	engramHumanDecisionSentence = "The human architecture- or design-decision sense of that same `type: architecture` MUST NOT be written to Engram when `recordStore.resolved` is `openrecord` — that decision belongs to the record store instead."
)

// TestOpenRecordSharedAssetsCarveOutIsDelimitedAndLiteral pins the
// record-store carve-out blocks added to the two shared SDD persistence
// assets: the <--:openrecord--> pair is balanced in each, and the block's
// prose carries its precise conditional sentence(s) — not merely the short
// literals a sentence happens to contain.
func TestOpenRecordSharedAssetsCarveOutIsDelimitedAndLiteral(t *testing.T) {
	cases := []struct {
		path              string
		requiredSentences []string
	}{
		{
			path:              "skills/_shared/persistence-contract.md",
			requiredSentences: []string{persistenceEnumDropSentence, persistenceUnaffectedTypesSentence},
		},
		{
			path:              "skills/_shared/engram-convention.md",
			requiredSentences: []string{engramArtifactsUnaffectedSentence, engramHumanDecisionSentence},
		},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			content, err := Read(tc.path)
			if err != nil {
				t.Fatalf("Read() error = %v", err)
			}
			_, inside, _ := splitOpenRecordBlock(t, content)
			for _, sentence := range missingCarveOutSentences(inside, tc.requiredSentences) {
				t.Fatalf("%s openrecord block missing required sentence (word-normalized match failed): %q", tc.path, sentence)
			}
		})
	}
}

// missingCarveOutSentences returns, in declaration order, every sentence in
// required that is absent from inside (the text between the
// <--:openrecord--> pair), matched word-normalized via normalizedWords. An
// empty slice means every pin is satisfied.
func missingCarveOutSentences(inside string, required []string) []string {
	insideWords := normalizedWords(inside)
	var missing []string
	for _, sentence := range required {
		if !strings.Contains(insideWords, normalizedWords(sentence)) {
			missing = append(missing, sentence)
		}
	}
	return missing
}

// TestOpenRecordNonSDDEnumUnchangedOutsideBlock pins the spec's "Outside
// openrecord, the enum is unchanged" scenario: the Non-SDD type: enum in
// persistence-contract.md, read OUTSIDE the <--:openrecord--> block, MUST
// stay byte-identical to what shipped before this change — decision,
// bugfix, discovery and pattern all still offered, exactly once, and never
// folded into the conditional block.
func TestOpenRecordNonSDDEnumUnchangedOutsideBlock(t *testing.T) {
	// Trailing comma disambiguates this from the block's own prose
	// reference to the same enum ("...the `type: \"{decision|bugfix|
	// discovery|pattern}\"` enum at line 121."), which has no comma.
	const enumLiteral = `type: "{decision|bugfix|discovery|pattern}",`

	content, err := Read("skills/_shared/persistence-contract.md")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	before, inside, after := splitOpenRecordBlock(t, content)
	outside := before + after

	if count := strings.Count(outside, enumLiteral); count != 1 {
		t.Fatalf("Non-SDD type: enum literal %q appears %d times outside the openrecord block, want exactly 1", enumLiteral, count)
	}
	if strings.Contains(inside, enumLiteral) {
		t.Fatal("Non-SDD type: enum literal appears inside the openrecord block; the unconditional enum above must not be folded into the conditional prose")
	}

	for _, value := range []string{"decision", "bugfix", "discovery", "pattern"} {
		if !strings.Contains(enumLiteral, value) {
			t.Fatalf("Non-SDD type: enum literal %q is missing %q", enumLiteral, value)
		}
	}
}

// TestCarveOutPinFlagsDeletedSentence pins defect 1: the persistence-contract
// carve-out block's requiredSentences must be able to fail. It first asserts
// both pins are satisfied on the shipped, unmutated block (positive, first),
// then — one subtest per sentence — deletes that sentence via mustMutate and
// asserts missingCarveOutSentences reports exactly that sentence missing.
func TestCarveOutPinFlagsDeletedSentence(t *testing.T) {
	content, err := Read("skills/_shared/persistence-contract.md")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	_, inside, _ := splitOpenRecordBlock(t, content)

	both := []string{persistenceEnumDropSentence, persistenceUnaffectedTypesSentence}
	if missing := missingCarveOutSentences(inside, both); len(missing) != 0 {
		t.Fatalf("missingCarveOutSentences(shipped, both) = %v, want none missing", missing)
	}

	sentences := map[string]string{
		"enum-drop-sentence":        persistenceEnumDropSentence,
		"unaffected-types-sentence": persistenceUnaffectedTypesSentence,
	}
	for name, sentence := range sentences {
		t.Run(name, func(t *testing.T) {
			mutated := mustMutate(t, inside, sentence, "")
			missing := missingCarveOutSentences(mutated, both)
			if len(missing) != 1 {
				t.Fatalf("missingCarveOutSentences(mutated, both) = %v, want exactly [%q]", missing, sentence)
			}
			if missing[0] != sentence {
				t.Fatalf("missingCarveOutSentences(mutated, both) = %v, want [%q]", missing, sentence)
			}
		})
	}
}

func assertDelimiterPaired(t *testing.T, content string) {
	t.Helper()
	opens := strings.Count(content, openRecordDelimOpen)
	closes := strings.Count(content, openRecordDelimClose)
	if opens == 0 {
		t.Fatal("no openrecord delimiter found")
	}
	if opens != closes {
		t.Fatalf("mismatched openrecord delimiters: %d open, %d close", opens, closes)
	}
	if strings.Index(content, openRecordDelimOpen) > strings.LastIndex(content, openRecordDelimClose) {
		t.Fatal("openrecord close delimiter appears before its open delimiter")
	}
}

// splitOpenRecordBlock locates the single <--:openrecord--> /
// <--:/openrecord--> pair in content and splits it into the prose before
// the block, the prose inside it (delimiters excluded), and the prose
// after it — the shape both shared assets carry, exactly one balanced
// pair each.
func splitOpenRecordBlock(t *testing.T, content string) (before, inside, after string) {
	t.Helper()
	assertDelimiterPaired(t, content)

	openIdx := strings.Index(content, openRecordDelimOpen)
	closeIdx := strings.Index(content, openRecordDelimClose)

	before = content[:openIdx]
	inside = content[openIdx+len(openRecordDelimOpen) : closeIdx]
	after = content[closeIdx+len(openRecordDelimClose):]
	return before, inside, after
}
