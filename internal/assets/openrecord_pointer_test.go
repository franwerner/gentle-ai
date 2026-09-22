package assets

import (
	"strings"
	"testing"
)

// TestOpenRecordPointerBulletsReachExactlySixPhases pins the spec's
// requirement: explore, propose, spec, design, apply and verify each carry
// exactly one openrecord pointer line; every other phase carries none.
//
// The pointer used to be delimiter-bounded and conditional. Both went with the
// record_store axis, and the delimiter half of this test went with them — a
// pairing assertion over markers that must never reappear is the inverse of
// what is now true, and store_fact_asymmetry_test.go's U1 asserts their
// absence over the whole asset tree. The citation count kept its subject
// intact: one citation per participating phase is what stops a phase from
// reconstructing the convention inline or quietly losing its pointer, and it
// is checked here independently of the pointer's wording, which U4 pins.
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

// Exact sentences pinned in the two shared assets' record-store carve-outs.
// Checking literal-substring presence of short tokens ("openrecord",
// "type: architecture") would let a garbled rewrite that still contains those
// tokens pass; these are the precise sentences the spec scenarios depend on,
// matched word-for-word via normalizedWords (case/punctuation/line-wrap
// insensitive, word-order sensitive — a dropped "not" or a swapped verb still
// fails).
//
// They used to be located by splitting the file on its <--:openrecord-->
// pair and searching only inside it. With the wrapper gone the search runs
// over the whole file: locating the passage was never the assertion, the
// passage's exact words were.
const (
	// persistenceEnumDropSentence covers the spec scenario "Under
	// openrecord, decision is not an offered save type".
	persistenceEnumDropSentence = "The `mem_save` instruction above carries one carve-out: `decision` drops from the `type: \"{decision|bugfix|discovery|pattern}\"` enum at line 121."

	// persistenceUnaffectedTypesSentence covers the spec scenario "The
	// trailing carve-out sentence MUST be pinned": the other three Non-SDD
	// types stay unconditional even though decision drops out.
	persistenceUnaffectedTypesSentence = "`bugfix`, `discovery`, and `pattern` are unaffected and still go to Engram exactly as written above."

	// engramArtifactsUnaffectedSentence covers the spec scenario "A phase
	// artifact save is unaffected by the carve-out".
	engramArtifactsUnaffectedSentence = "SDD phase artifacts — every phase from `explore` through `archive-report` — MUST keep saving to Engram with `type: architecture`."

	// engramHumanDecisionSentence covers the spec scenario "A human
	// decision is not written to Engram".
	engramHumanDecisionSentence = "The human architecture- or design-decision sense of that same `type: architecture` MUST NOT be written to Engram — that decision belongs to the openrecord record store instead."
)

// TestOpenRecordSharedAssetsCarveOutIsLiteral pins the record-store carve-out
// prose in the two shared SDD persistence assets: each carries its precise
// sentence(s) — not merely the short literals a sentence happens to contain.
func TestOpenRecordSharedAssetsCarveOutIsLiteral(t *testing.T) {
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
			for _, sentence := range missingCarveOutSentences(content, tc.requiredSentences) {
				t.Fatalf("%s carve-out missing required sentence (word-normalized match failed): %q", tc.path, sentence)
			}
		})
	}
}

// missingCarveOutSentences returns, in declaration order, every sentence in
// required that is absent from content, matched word-normalized via
// normalizedWords. An empty slice means every pin is satisfied.
func missingCarveOutSentences(content string, required []string) []string {
	contentWords := normalizedWords(content)
	var missing []string
	for _, sentence := range required {
		if !strings.Contains(contentWords, normalizedWords(sentence)) {
			missing = append(missing, sentence)
		}
	}
	return missing
}

// TestOpenRecordNonSDDEnumStillOffersEveryType pins the spec's "the Non-SDD
// enum is unchanged" scenario: the Non-SDD type: enum in
// persistence-contract.md MUST stay byte-identical to what shipped —
// decision, bugfix, discovery and pattern all still offered, exactly once.
//
// The old form of this test read the enum "outside the <--:openrecord-->
// block" and separately asserted it had not been folded into the conditional
// prose. There is no conditional prose to fold it into any more, so the
// containment half went; the enum's own shape is the part that had to
// survive, and it is asserted here unchanged.
func TestOpenRecordNonSDDEnumStillOffersEveryType(t *testing.T) {
	// Trailing comma disambiguates this from the carve-out's own prose
	// reference to the same enum ("...the `type: \"{decision|bugfix|
	// discovery|pattern}\"` enum at line 121."), which has no comma.
	const enumLiteral = `type: "{decision|bugfix|discovery|pattern}",`

	content, err := Read("skills/_shared/persistence-contract.md")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if count := strings.Count(content, enumLiteral); count != 1 {
		t.Fatalf("Non-SDD type: enum literal %q appears %d times, want exactly 1", enumLiteral, count)
	}

	for _, value := range []string{"decision", "bugfix", "discovery", "pattern"} {
		if !strings.Contains(enumLiteral, value) {
			t.Fatalf("Non-SDD type: enum literal %q is missing %q", enumLiteral, value)
		}
	}
}

// TestCarveOutPinFlagsDeletedSentence is the pins' anti-vacuity harness: the
// persistence-contract carve-out's requiredSentences must be able to fail. It
// first asserts both pins are satisfied on the shipped asset (positive,
// first), then — one subtest per sentence — deletes that sentence via
// mustMutate and asserts missingCarveOutSentences reports exactly that
// sentence missing. A checker that passes when the thing it checks is absent
// is the failure mode this harness exists to prevent.
func TestCarveOutPinFlagsDeletedSentence(t *testing.T) {
	content, err := Read("skills/_shared/persistence-contract.md")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	both := []string{persistenceEnumDropSentence, persistenceUnaffectedTypesSentence}
	if missing := missingCarveOutSentences(content, both); len(missing) != 0 {
		t.Fatalf("missingCarveOutSentences(shipped, both) = %v, want none missing", missing)
	}

	sentences := map[string]string{
		"enum-drop-sentence":        persistenceEnumDropSentence,
		"unaffected-types-sentence": persistenceUnaffectedTypesSentence,
	}
	for name, sentence := range sentences {
		t.Run(name, func(t *testing.T) {
			mutated := mustMutate(t, content, sentence, "")
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
