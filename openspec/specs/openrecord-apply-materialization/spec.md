# openrecord-apply-materialization Specification

## Purpose

Fix the observable behaviour of the decision-materialization step that `sdd-apply` performs when a
project's record store resolves to openrecord: when the step runs, what it writes, how the record's
path is derived, what happens when the declared store is absent, and the authorship contract that binds
the record body sections no design field feeds.

## Requirements

### Requirement: Materialization runs inside the `ready` run, in the batch that clears the last task

The convention MUST place the write inside an `applyState: ready` run — after tasks are marked complete
(Step 5), before the return — and MUST gate it on **this batch leaving no pending task in the tasks
artifact**. It MUST NOT place the write under `applyState: all_done`: that is a non-edit entry
condition supplied by the orchestrator, and the shipped gate answers it with "do not edit. Return
`success`". Under a multi-batch apply, the batch that clears the last pending task materializes; every
earlier batch writes nothing and reports nothing about it.

#### Scenario: A single batch clears every task

- GIVEN a change whose design carries two `## Architecture Decisions` entries
- WHEN an `applyState: ready` run completes every pending task in the tasks artifact
- THEN both decisions MUST be written as records before the run returns
- AND the write MUST happen after the tasks are marked complete, not during the per-task loop

#### Scenario: An earlier batch leaves tasks pending

- GIVEN a multi-batch apply where this batch finishes its assigned tasks but the tasks artifact still lists pending ones
- WHEN the run reaches the materialization step
- THEN it MUST write no record
- AND it MUST report nothing about the deferral — a later batch owning the write is the normal case, not an issue

#### Scenario: The run is entered with `all_done`

- GIVEN the orchestrator supplies `applyState: all_done`
- WHEN the run starts
- THEN it MUST return without editing, exactly as the shipped gate already requires
- AND no record MUST be written on that path

#### Scenario: Nothing about the record store gates the step

- GIVEN openrecord is installed unconditionally, so no project can report a different record store
- WHEN apply reaches the step
- THEN the only condition governing whether it runs MUST be whether the batch leaves a pending task
- AND the step MUST NOT consult a config key, a status field, or an injected value to decide

### Requirement: The design-to-record mapping is fixed, and stated in record-field terms only

The convention MUST fix how one design decision becomes one record, naming **record fields and body
sections** and never an openrecord flag or command invocation. Which flag carries each field stays with
the emitted capture skill, which MUST remain the only place that says it.

| Design field (`## Architecture Decisions`) | Record input |
|---|---|
| `### Decision: {Title}` | the record's title |
| `**Choice**` | the record's one-line description — compressed to one line when the design's is longer |
| `**Alternatives considered**` | the body's `## Alternatives` section |
| `**Rationale**` | the body's `## Decision` section, developed |
| *(no source field)* | the body's `## Context` — authored under the pointing test below |
| *(no source field)* | the body's `## Consequences` — authored under the pointing test below |
| *(no source field)* | status: always `accepted` |

Status is not a judgement call here: apply writes after the governing code has landed, so the
proposal-stage value cannot arise on this path.

#### Scenario: A design decision becomes a record

- GIVEN a design decision with a title, a `**Choice**`, `**Alternatives considered**` and a `**Rationale**`
- WHEN apply materializes it
- THEN the record's title MUST come from the decision title, its one-line description from `**Choice**`, its `## Alternatives` from `**Alternatives considered**`, and its `## Decision` from `**Rationale**`
- AND its status MUST be `accepted`

#### Scenario: A multi-line Choice is compressed, not truncated

- GIVEN a `**Choice**` spanning several lines
- WHEN it becomes the record's one-line description
- THEN it MUST be compressed to a single line that still states the decision
- AND the full text MUST NOT be silently cut at the first line break

#### Scenario: The added prose carries no command shape

- GIVEN the prose this change adds to the convention
- WHEN it is searched for the flag forms the write commands take — `--title`, `--description`, `--status`, `--body-file`
- THEN none MUST appear
- AND no command invocation MUST appear either

#### Scenario: A reader needing the flag is redirected, not served locally

- GIVEN a reader of the convention who needs to know which flag carries the record's description
- WHEN they read the mapping
- THEN the convention MUST leave that to the emitted capture skill, which the section's existing redirect already names
- AND the mapping MUST NOT answer it

### Requirement: The record's path is derived, and each level has a named source

The convention MUST state how `decisions/<component>/<concern>/<slug>.md` is derived: `component`
resolves from the design's `## File Changes` paths against the store's declared component owners — the
move the emitted consult skill documents; `concern` is chosen by judgement against openrecord's shipped
concern catalogue, whose per-concern topics prompt that judgement rather than serving as a lookup key;
`slug` is named freely from the decision's title. This MUST introduce no dependency beyond the two
emitted skills the convention's preamble already declares.

#### Scenario: The component comes from the paths the design changed

- GIVEN a design whose `## File Changes` name paths the store's component owners cover
- WHEN the record path is derived
- THEN `component` MUST be resolved from those paths against the declared owners
- AND the convention MUST name the emitted consult skill as where that move is documented, not restate the command

#### Scenario: The concern is a judgement, not a lookup

- GIVEN openrecord's shipped concern catalogue with its per-concern topics
- WHEN `concern` is chosen for a decision
- THEN it MUST be chosen by judgement against that catalogue
- AND the topics MUST NOT be treated as keys that mechanically select a concern

### Requirement: A missing level is not a failure

A `component`/`concern` level that does not exist yet MUST NOT stop materialization and MUST NOT be
treated as an absent store. Creating the level is a documented mechanic of the emitted capture skill,
which already instructs it before a write; apply proceeds through it and writes the record.

#### Scenario: The derived level does not exist yet

- GIVEN a store that exists but carries no level for the derived component/concern pair
- WHEN apply materializes a decision into it
- THEN it MUST create the level through the mechanic the emitted capture skill documents and write the record
- AND it MUST NOT report this as an issue or a skip

### Requirement: A store that is not on disk produces a named skip, never a bootstrap and never a silent drop

When no store exists on disk, apply MUST skip materialization, MUST name the skip in the existing
`### Issues Found` section of its return, and MUST complete the batch. It MUST NOT create a store, and
it MUST NOT pass over the skip in silence. The task's code is unaffected.

There is no declaration left for this case to turn on: openrecord is installed unconditionally, so an
absent store is a project that has not been set up rather than one that opted out, and the named skip
is the only thing that tells anyone which.

#### Scenario: The store is absent
- WHEN the batch that would materialize reaches the step
- THEN it MUST write no record and create no store
- AND it MUST name the skip in `### Issues Found` and finish the batch with its code work intact

#### Scenario: The skip does not stop the run

- GIVEN the same absent store
- WHEN the run returns
- THEN its status MUST reflect the code work it completed, not a block on the missing store
- AND the named `### Issues Found` line MUST be the whole of the reporting — no new return section is introduced

### Requirement: The pointing test governs every body section no design field feeds

The convention MUST state, as a rule with a test rather than an encouragement to be careful, that apply
writes into `## Context` and `## Consequences` **only what it can point at** — a file it changed, a
constraint it hit, a test it wrote, a line of the spec. When it can point at nothing, it MUST write the
absence into the record in one line; an empty-but-honest section is the correct result, not a defect to
fill. The same test binds `## Decision`: developing the design's `**Rationale**` MUST NOT become
extending it with reasons the design never gave.

#### Scenario: Context is written from what apply can point at

- GIVEN apply has just implemented the governing code and holds firsthand knowledge of it
- WHEN it authors `## Context`
- THEN every statement MUST trace to a file it changed, a constraint it hit, a test it wrote, or a line of the spec
- AND a statement of what "must have been true" MUST NOT be written

#### Scenario: Nothing to point at is written as an absence

- GIVEN a decision whose consequences were never established during implementation
- WHEN `## Consequences` is authored
- THEN the section MUST state that absence explicitly, in one line
- AND leaving it honestly empty MUST NOT be treated as a defect to fill with a plausible sentence

#### Scenario: "Developed" does not license invention

- GIVEN a design `**Rationale**` that gives one reason
- WHEN it becomes the record's `## Decision`, developed
- THEN the development MUST stay inside what apply can point at
- AND reasons the design never gave MUST NOT be added

#### Scenario: The instruction is checkable

- GIVEN the authorship guidance as written in the convention
- WHEN it is read
- THEN it MUST state a test a reader can apply to a sentence they just wrote
- AND it MUST NOT rest on wording that only asks the reader to be careful

### Requirement: apply's cue reuses the single existing pointer, and the small tier is untouched

`internal/assets/skills/sdd-apply/SKILL.md` MUST still carry **exactly one** literal
`openrecord-convention.md` citation after this change — the one it already has — so the procedural cue
this change adds MUST reuse it rather than add a second. The model-small tier of that file MUST be
byte-for-byte unchanged, because it carries no
openrecord participation today and the pin counts citations over the whole file.

#### Scenario: The pointer pin passes unchanged

- GIVEN `internal/assets/openrecord_pointer_test.go` as it stands on `feat/recordstore-test-defect-closure`
- WHEN `go test ./internal/assets/...` runs after the change
- THEN it MUST pass
- AND that test file MUST NOT have been edited

#### Scenario: The cue delegates instead of citing again

- GIVEN the procedural cue added to the model-capable tier
- WHEN it is read
- THEN it MUST name what it delegates so the jump to the convention is deliberate
- AND it MUST NOT contain a second literal `openrecord-convention.md` citation

#### Scenario: The small tier is byte-for-byte unchanged

- GIVEN the `<!-- section:model-small -->` tier of `sdd-apply/SKILL.md`
- WHEN it is compared against the pre-change file
- THEN it MUST be byte-for-byte identical
- AND `internal/assets/sdd_phase_rules_tier_parity_test.go` MUST pass with no edit
