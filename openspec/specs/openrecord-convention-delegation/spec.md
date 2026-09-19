# openrecord-convention-delegation Specification

## Purpose

Fix the observable content of the shared convention asset
`internal/assets/skills/_shared/openrecord-convention.md`: it delegates retrieval and write
**mechanics** to the skills the openrecord binary emits and keeps no private copy of either; it
declares what an agent that never receives those skills must do; and it fixes the boundary of the
delegation — the two guarantees that stay local precisely because no emitted skill carries them (the
ban on editing a record file with a file-editing tool, and the `body-hash-mismatch` half of
`sdd-verify`'s validate duty).

## Requirements

### Requirement: Retrieval mechanics are a redirect, never a local copy

The convention MUST NOT document any retrieval command, current or withdrawn. The heading opening
`## The walk` MUST survive with a redirect-only body naming `skills/openrecord-consult/SKILL.md` as
where the retrieval moves are documented. The section `## Say what you found` (L117-139) MUST be gone
entirely.

#### Scenario: The walk heading survives carrying only a redirect

- GIVEN the convention file after the change
- WHEN the section opening `## The walk` is read
- THEN its body MUST name `skills/openrecord-consult/SKILL.md` as where the retrieval moves are documented
- AND it MUST contain no command invocation, no collection-naming rule, and no result-coordinate translation rule

#### Scenario: The report-format section is gone

- GIVEN the convention file after the change
- WHEN its headings are enumerated
- THEN no heading `## Say what you found` MUST appear
- AND no `Governs this work:` / `Looked at, does not apply` report template MUST remain anywhere in the file

#### Scenario: No withdrawn command shape survives anywhere in the file

- GIVEN the convention file after the change
- WHEN it is searched for the shapes the private copy had drifted to
- THEN no `qmd query` reference MUST remain
- AND no instruction to invoke the `qmd` binary directly MUST remain
- AND no claim that `openrecord qmd` lacks a query subcommand MUST remain
- AND no `--section "<heading>"` flag form MUST remain

#### Scenario: The live retrieval command is not re-copied under a new name

- GIVEN the live command is `openrecord search TERM --for COORDINATE [--omit PATH]`, running a literal pass and — when qmd is registered — a semantic one in a single call, reporting which ran via a `semantic` field (`used` / `lexical-only` / `unavailable`)
- WHEN the convention file is read after the change
- THEN it MUST NOT restate that command, its flags, or its `semantic` field values
- AND accuracy MUST rest on the emitted skill alone, so no local copy exists that can drift again

### Requirement: Write mechanics split — the command shapes go, the local guarantees stay

`## Writing records — apply is the sole writer` MUST survive carrying, in this order: (1) the
sole-writer paragraph, (2) a redirect naming `skills/openrecord-capture/SKILL.md`, (3) the
gentle-ai-owned materialization guidance, (4) the file-editing-ban paragraph, (5) the `sdd-verify` /
`body-hash-mismatch` paragraph. The heading and blocks 1, 4 and 5 MUST be byte-for-byte identical to
the pre-change file. The two usage bullets (L183-188 of the file before the delegation change) MUST
remain deleted.

The split is three-way, not two. (a) **openrecord command shapes go** — they are named only by the
redirect. (b) **The local guarantees stay** — the file-editing ban and the `body-hash-mismatch` duty,
which no emitted skill carries. (c) **gentle-ai-owned mapping knowledge stays local** — guidance whose
left-hand side is a gentle-ai asset that openrecord has never heard of, and which therefore has no
upstream to drift from. The classification test MUST be decidable per passage: content whose left-hand
side is a gentle-ai asset lives in the convention; content whose left-hand side is an openrecord
command lives in the emitted skill. Block (3) MUST contain no openrecord flag name and no command
invocation, so category (c) cannot re-copy what category (a) removed.

(Previously: the section carried four blocks and the split was stated as two-way — command shapes out,
local guarantees in — with no category for gentle-ai-owned mapping knowledge and no test for
classifying a passage.)

#### Scenario: The three kept blocks survive byte-for-byte

- GIVEN the pre-change file's heading `## Writing records — apply is the sole writer` (L178), its sole-writer paragraph opening "Only `sdd-apply` writes into the record store" (L180-181), its paragraph opening "**Never touch a record file with a file-editing tool.**" through "…carrying this guard is this convention's job, not the tool's." (L190-198), and its paragraph opening "`sdd-verify` runs `openrecord validate` for structural shape" through "…`sdd-archive` writes nothing into the record store, ever." (L200-207)
- WHEN each is compared against the same block in the post-change file
- THEN all four MUST be byte-for-byte identical
- AND deleting, shortening or rewording any of them MUST be treated as a failure, not a simplification

#### Scenario: The two usage bullets are gone

- GIVEN the two bullets opening "**A record that does not exist yet** is created with `openrecord record write`" and "**A record that already exists** is updated with `openrecord record edit --section \"<heading>\"`"
- WHEN the post-change file is read
- THEN neither bullet MUST remain
- AND no passage MUST instruct which verb to pick for a new versus an existing record

#### Scenario: The redirect sits between the sole-writer paragraph and the ban paragraph

- GIVEN the ban paragraph opens by referring to "the `record write` / `record edit` path"
- WHEN the post-change section is read top to bottom
- THEN the redirect MUST appear after the sole-writer paragraph and before the ban paragraph
- AND the materialization guidance MUST sit after the redirect and before the ban paragraph, so the ban paragraph's reference still has a named documentation source already in view

#### Scenario: The bare verb names legitimately survive inside the kept paragraphs

- GIVEN the strings `openrecord record write` and `openrecord record edit` appear inside the kept L190-198 and L200-207 paragraphs
- WHEN a check asserts the usage bullets are gone
- THEN those occurrences MUST NOT be treated as drift or removed
- AND only the `--section "<heading>"` flag form and the verb-choice instruction count as the shapes withdrawn

#### Scenario: The redirect names the skill as the source without instructing usage

- GIVEN the redirect replacing the bullets
- WHEN it is read
- THEN it MUST name `skills/openrecord-capture/SKILL.md` as where the write commands — `level add`, `record write --body-file`, `record edit --section --body-file`, and which verb applies when — are documented
- AND it MUST NOT itself instruct how to invoke any of them or which one to pick

#### Scenario: The redirect carries no degradation line

- GIVEN the preamble declares the skill-less-agent degradation once for the whole file
- WHEN the write-section redirect is read
- THEN it MUST NOT restate that degradation

#### Scenario: A passage is classified by its left-hand side

- GIVEN a candidate passage for this section
- WHEN it is classified under the three-way split
- THEN a passage whose left-hand side is a gentle-ai asset — the design artifact's `## Architecture Decisions` template — MUST stay in the convention
- AND a passage whose left-hand side is an openrecord command MUST be left to the emitted skill

#### Scenario: The local mapping block carries no command shape

- GIVEN block (3), the gentle-ai-owned materialization guidance
- WHEN it is searched for openrecord flag forms and command invocations
- THEN none MUST appear, so the block cannot reopen the drift the command shapes were removed to close
- AND the redirect at block (2) MUST remain the only passage in the section that names a flag form

### Requirement: The preamble declares the dependency and the degradation

The preamble (L5-6) MUST NOT claim the file is self-contained. It MUST state that the convention
depends on the skills openrecord emits, and MUST declare what an agent that never receives them does:
say so and stop, never improvise the moves from memory. This declaration MUST appear once, for the
whole file.

#### Scenario: The self-containment claim is gone

- GIVEN the pre-change preamble claims a phase agent that has read only this file can run every retrieval move without loading any emitted skill
- WHEN the post-change preamble is read
- THEN no such claim MUST remain
- AND it MUST state instead that the convention depends on the emitted skills

#### Scenario: The degradation is declared for an agent with no skills directory

- GIVEN an agent whose skills directory resolves to none, which the fan-out silently skips
- WHEN that agent reads the convention
- THEN the preamble MUST tell it to say it cannot read the emitted skills and stop
- AND it MUST NOT be told, anywhere in the file, to improvise the retrieval or write moves from memory

#### Scenario: The degradation is stated once, restated only for retrieval

- GIVEN the preamble carries the declaration
- WHEN the rest of the file is read
- THEN only the `## The walk` redirect MAY restate it, for retrieval
- AND no third statement of it MUST appear

### Requirement: The responsibility map keeps no dangling reference

`## Per-phase responsibility map` (L23-35) MUST keep its shape — same six phase rows, same columns —
and MUST contain no reference to a section this change deletes. Exactly two cells change: the
`sdd-explore` row and the `sdd-apply` row.

#### Scenario: The two stale cells name their emitted skill

- GIVEN the `sdd-explore` cell says "Runs the full walk below and reports how each record surfaced" and the `sdd-apply` cell says "see \"Writing records\" below"
- WHEN both are read after the change
- THEN neither MUST point at deleted content
- AND each MUST name the emitted skill that now carries the mechanics it describes

#### Scenario: The other four rows are unchanged

- GIVEN the `sdd-propose`, `sdd-spec`, `sdd-design` and `sdd-verify` rows, plus the two closing lines (L34-35)
- WHEN they are compared against the pre-change file
- THEN each MUST be byte-for-byte identical
- AND the `sdd-propose` / `sdd-spec` "check below" wording MUST stay valid because the contradiction contract stays in place

### Requirement: Guard surfaces and the delimiter pair are untouched

`## Activation — declared, never presence-detected` (L8-21) and `## When your work contradicts an
accepted record` (L141-176) MUST survive byte-for-byte. The file MUST keep exactly one balanced
`<--:openrecord-->` / `<--:/openrecord-->` pair, open before close.

#### Scenario: The registered guard surface is unchanged

- GIVEN `store_fact_asymmetry_test.go` registers the `## Activation` section as the `activation` surface, extracted by heading and cut at the next `## `
- WHEN the post-change file is extracted the same way
- THEN the extracted text MUST be byte-for-byte identical to the pre-change extraction
- AND its C1 / C2 / C3c checks MUST pass with no edit to that test file

#### Scenario: The contradiction contract is unchanged

- GIVEN the section `## When your work contradicts an accepted record`, including the `STOPPED:` template and the "Never resolve it on your own" paragraph
- WHEN it is compared against the pre-change file
- THEN it MUST be byte-for-byte identical

#### Scenario: Both test files pass with zero edits

- GIVEN `internal/assets/openrecord_pointer_test.go` and `internal/assets/store_fact_asymmetry_test.go` as they stand on `feat/recordstore-test-defect-closure`
- WHEN `go test ./internal/assets/...` runs after the change
- THEN it MUST pass
- AND neither test file MUST have been edited

### Requirement: The pointer paths resolve to what the fan-out actually writes

Both redirects MUST address their target as `skills/openrecord-consult/SKILL.md` and
`skills/openrecord-capture/SKILL.md` — the emitted skills' locations relative to the skills root,
written in the same form the six phase SKILL.md files already use for
`skills/_shared/openrecord-convention.md`.

#### Scenario: The named paths match the emitted locations

- GIVEN the fan-out writes `openrecord-consult/SKILL.md` and `openrecord-capture/SKILL.md` under each selected agent's skills directory
- WHEN the two paths named in the convention are compared against those locations
- THEN each MUST match its emitted location exactly, below the shared skills root
- AND both MUST be expressed relative to that root, as the existing `skills/_shared/…` references are

#### Scenario: No third skill is named

- GIVEN the fan-out also emits skills this flow's phase map does not use
- WHEN the convention is read
- THEN it MUST name only `openrecord-consult` and `openrecord-capture`
