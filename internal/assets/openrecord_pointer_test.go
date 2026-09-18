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

// TestOpenRecordSharedAssetsCarveOutIsDelimitedAndLiteral pins the
// record-store carve-out blocks added to the two shared SDD persistence
// assets: the <--:openrecord--> pair is balanced in each, and each carries
// its required literal — recordStore.resolved plus openrecord on both,
// type: architecture on engram-convention.md.
func TestOpenRecordSharedAssetsCarveOutIsDelimitedAndLiteral(t *testing.T) {
	cases := []struct {
		path             string
		requiredLiterals []string
	}{
		{
			path:             "skills/_shared/persistence-contract.md",
			requiredLiterals: []string{"recordStore.resolved", "openrecord"},
		},
		{
			path:             "skills/_shared/engram-convention.md",
			requiredLiterals: []string{"recordStore.resolved", "openrecord", "type: architecture"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			content, err := Read(tc.path)
			if err != nil {
				t.Fatalf("Read() error = %v", err)
			}
			assertDelimiterPaired(t, content)
			for _, literal := range tc.requiredLiterals {
				if !strings.Contains(content, literal) {
					t.Fatalf("%s missing required literal %q", tc.path, literal)
				}
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
