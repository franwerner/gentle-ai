# engram-decision-carveout Specification

## Purpose

Define the observable behavior of the memory-save contract gentle-ai ships to agents — the injected
Engram protocol and the shared SDD skill assets — with respect to where a decision is written. An
architecture or design decision has exactly one destination, the record store, rather than being
written twice in two wordings with nothing saying which is authoritative. In-flight SDD phase
artifacts are memory, not records, and keep writing to Engram.

This capability was previously named `recordstore-conditional-memory` and described the same carve-out
as conditional on a declared `sdd.record_store` axis. That axis is gone: openrecord is installed
unconditionally, so there is no mode in which a decision belongs in Engram, and nothing left for a
reader to resolve before obeying the contract. The renaming is part of the behavior — a capability
promising conditionality is a promise the shipped assets no longer keep.

## Requirements

### Requirement: Every rendered protocol variant that orders a decision save carries the carve-out

Every rendered Engram protocol variant that orders the agent to save an architecture or design
decision MUST carve that order out; no such order MAY remain in force. Every variant that carries the
carve-out MUST name the SAME decision class, so that no variant silences more decisions than another.

#### Scenario: The full variant carries the carve-out

- GIVEN a rendered protocol variant that lists "Architecture or design decision made" as a proactive save trigger
- WHEN that variant is rendered
- THEN it MUST state that the trigger does not apply
- AND the carve-out MUST be stated once, as its own passage, rather than as a modification of an existing trigger item

#### Scenario: The slim variant carries the carve-out

- GIVEN a rendered protocol variant whose save rule covers decisions in coarser wording than an itemised trigger list
- WHEN that variant is rendered
- THEN it MUST also carry the carve-out

#### Scenario: Full and slim exempt the same decision class

- GIVEN the full variant and the slim variant are both rendered
- WHEN the decision class each one exempts is compared
- THEN both MUST name architecture and design decisions specifically
- AND neither MAY exempt a broader class of saves than the other

---

### Requirement: The carve-out instructs no check of any kind

The rendered protocol is baked into a global, often cross-project system prompt at setup time. It
therefore MUST NOT instruct the agent to determine anything before obeying the carve-out: no config
key to read, no directory to look for on disk, no status field to project, and no injected value to
consult. The carve-out states what does not go to Engram and stops there.

This inverts an earlier requirement rather than relaxing it. The protocol used to carry a read-time
self-check precisely because it could not receive a per-launch fact; with nothing left to decide, a
surviving check would be an instruction to resolve a question that has one answer — and an agent that
performed it and got the answer wrong would silently reinstate the double write.

#### Scenario: No detection language survives in any variant

- GIVEN every rendered protocol variant that carries the carve-out
- WHEN each is read after the change
- THEN none MUST name a config key, a store declaration, an on-disk directory, or a status field
- AND none MUST describe a state in which the carve-out does not apply

#### Scenario: The carve-out applies without the agent establishing anything

- GIVEN an agent reading the rendered protocol in any workspace
- WHEN it reaches the decision-save trigger
- THEN the carve-out MUST already be in force
- AND the agent MUST NOT be required to have read anything outside the protocol for it to be in force

---

### Requirement: The Non-SDD save-type enum does not offer `decision`

`internal/assets/skills/_shared/persistence-contract.md` MUST NOT offer `decision` in the Non-SDD
sub-agent `type:` enum. The remaining values (`bugfix`, `discovery`, `pattern`) MUST be unchanged, and
the in-flight SDD artifact save instructions in the same file MUST be unaffected.

#### Scenario: `decision` is not an offered save type

- GIVEN a sub-agent reading the Non-SDD persistence instruction
- WHEN it selects a `type:` for what it is about to save
- THEN `decision` MUST NOT be among the values it is offered

#### Scenario: The remaining values and the SDD instructions are untouched

- GIVEN the same instruction
- WHEN its other values and the file's in-flight SDD artifact instructions are compared against the pre-carve-out file
- THEN `bugfix`, `discovery` and `pattern` MUST remain available
- AND the SDD artifact save instructions MUST be unchanged

---

### Requirement: SDD phase artifacts keep writing to Engram

`internal/assets/skills/_shared/engram-convention.md` MUST state explicitly that Engram
`type: architecture` carries two distinct senses, and that they have different destinations: SDD phase
artifacts (`explore` through `archive-report`) MUST continue to be saved to Engram with
`type: architecture`, while the human architecture- or design-decision sense of that same type MUST
NOT be written to Engram at all.

The split MUST be stated, not left to be inferred from the type. A reader who meets only the carve-out
could otherwise conclude that phase artifacts stop being persisted, which would break the bus every
headless phase depends on to reach the next one.

#### Scenario: A phase artifact save is unaffected

- GIVEN an SDD phase persisting its own artifact
- WHEN it saves
- THEN it MUST still save to Engram with `type: architecture`

#### Scenario: A human decision is not written to Engram

- GIVEN an agent that would otherwise save an architecture or design decision with `type: architecture`
- WHEN it reaches the save
- THEN that save MUST NOT go to Engram
- AND the convention MUST state this split explicitly rather than leave it to be inferred from the type

---

### Requirement: Surfaces outside the carve-out are unchanged

The carve-out MUST be contained to the protocol variants and shared assets named above. Protocol
variants that carry no decision-save trigger MUST remain byte-identical, and the sibling
decision-shaped triggers that the carve-out does not name MUST keep their save order.

#### Scenario: Untriggered protocol variants do not move

- GIVEN the protocol variants that carry no decision-save trigger (`section:passive-capture`, `section:compact`)
- WHEN the protocol asset is rendered
- THEN those variants MUST be byte-identical to their pre-carve-out output
- AND their byte-exact fixtures MUST pass without being edited

#### Scenario: Sibling decision-shaped triggers keep their save order

- GIVEN the full variant's other decision-shaped triggers — tool or library choice with tradeoffs, team convention documented, workflow change agreed, pattern established
- WHEN the carve-out is rendered
- THEN the carve-out MUST name only the architecture-or-design-decision trigger it exempts
- AND those four triggers MUST keep ordering their saves
