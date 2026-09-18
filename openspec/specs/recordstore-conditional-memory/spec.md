# recordstore-conditional-memory Specification

## Purpose

Define the observable behavior of the memory-save contract that gentle-ai ships to agents — the
injected Engram protocol and the shared SDD skill assets — once the resolved record store is taken
into account. When a workspace declares `openrecord` as its record store, an architecture or design
decision has exactly one destination instead of being written twice, in two wordings, with nothing
saying which is authoritative. In-flight SDD phase artifacts are memory, not records, and keep writing
to Engram in every mode.

## Requirements

### Requirement: Record-store carve-out in every rendered protocol variant that orders a decision save

Every rendered Engram protocol variant that orders the agent to save an architecture or design
decision MUST state a record-store condition on that order; no such order MAY remain unconditional.
Every variant that carries the carve-out MUST name the SAME decision class, so that no variant
silences more decisions than another.

#### Scenario: The full variant carries the carve-out

- GIVEN a rendered protocol variant that lists "Architecture or design decision made" as a proactive save trigger
- WHEN that variant is rendered
- THEN it MUST state the record-store condition under which that trigger does not apply
- AND the condition MUST be stated once, as its own passage, rather than as a modification of an existing trigger item

#### Scenario: The slim variant carries the carve-out

- GIVEN a rendered protocol variant whose save rule covers decisions in coarser wording than an itemised trigger list
- WHEN that variant is rendered
- THEN it MUST also state the record-store condition

#### Scenario: Full and slim exempt the same decision class

- GIVEN the full variant and the slim variant are both rendered
- WHEN the decision class each one exempts is compared
- THEN both MUST name architecture and design decisions specifically
- AND neither MAY exempt a broader class of saves than the other

---

### Requirement: The protocol's record-store check is declared, never presence-detected

The rendered protocol is baked into a global, often cross-project system prompt at setup time and
receives no per-launch resolved record-store fact, so its carve-out MUST instruct the agent to perform
the check itself at read time. The check MUST treat the store as `openrecord` only when the
workspace's `openspec/config.yaml` (or `.yml`) declares `sdd.record_store: openrecord`. The presence
of an `.openrecord/` directory MUST NOT activate the carve-out. The wording MUST reuse the phrasing
already shipped in `openrecord-convention.md` rather than introduce new detection language.

#### Scenario: A declaring workspace activates the carve-out

- GIVEN a workspace whose `openspec/config.yaml` declares the record store as `openrecord`
- WHEN an agent reads the rendered protocol and reaches the decision-save trigger
- THEN the carve-out MUST tell it not to write that decision to Engram

#### Scenario: A directory on disk does not activate the carve-out

- GIVEN a workspace with an `.openrecord/` directory but no `openrecord` declaration in `openspec/config.yaml`
- WHEN an agent reads the rendered protocol and reaches the decision-save trigger
- THEN the carve-out MUST NOT apply
- AND the unconditional save order MUST remain in force

---

### Requirement: The Non-SDD save-type enum is conditional on the resolved record store

`internal/assets/skills/_shared/persistence-contract.md` MUST offer `decision` in the Non-SDD
sub-agent `type:` enum only when the resolved record store is NOT `openrecord`. The remaining values
(`bugfix`, `discovery`, `pattern`) MUST be unchanged in both modes, and the in-flight SDD artifact
save instructions in the same file MUST be unaffected.

#### Scenario: Under openrecord, decision is not an offered save type

- GIVEN a launch whose status reports the resolved record store as `openrecord`
- WHEN a sub-agent reads the Non-SDD persistence instruction
- THEN `decision` MUST NOT be among the `type:` values it is offered

#### Scenario: Outside openrecord, the enum is unchanged

- GIVEN a launch whose status reports any resolved record store other than `openrecord`
- WHEN a sub-agent reads the Non-SDD persistence instruction
- THEN the `type:` enum MUST be identical to the one shipped before this change
- AND `bugfix`, `discovery` and `pattern` MUST remain available in both modes

---

### Requirement: SDD phase artifacts keep writing to Engram in every record-store mode

`internal/assets/skills/_shared/engram-convention.md` MUST state explicitly that Engram
`type: architecture` carries two distinct senses, and that only one of them is conditional: SDD phase
artifacts (`explore` through `archive-report`) MUST continue to be saved to Engram with
`type: architecture` regardless of the resolved record store, while the human architecture- or
design-decision sense of that same type MUST NOT be written to Engram under `openrecord`.

#### Scenario: A phase artifact save is unaffected under openrecord

- GIVEN a launch whose resolved record store is `openrecord`
- WHEN an SDD phase persists its own artifact
- THEN it MUST still save to Engram with `type: architecture`

#### Scenario: A human decision is not written to Engram under openrecord

- GIVEN a launch whose resolved record store is `openrecord`
- WHEN an agent would otherwise save an architecture or design decision with `type: architecture`
- THEN that save MUST NOT go to Engram
- AND the convention MUST state this split explicitly rather than leave it to be inferred from the type

---

### Requirement: Surfaces outside the carve-out are unchanged

The change MUST be contained to the protocol variants and shared assets named in Scope. Protocol
variants that carry no decision-save trigger MUST remain byte-identical, and the sibling
decision-shaped triggers that the carve-out does not name MUST keep their unconditional order.

#### Scenario: Untriggered protocol variants do not move

- GIVEN the protocol variants that carry no decision-save trigger (`section:passive-capture`, `section:compact`)
- WHEN the protocol asset is rendered after this change
- THEN those variants MUST be byte-identical to their pre-change output
- AND their byte-exact fixtures MUST pass without being edited

#### Scenario: Sibling decision-shaped triggers keep their unconditional order

- GIVEN the full variant's other decision-shaped triggers — tool or library choice with tradeoffs, team convention documented, workflow change agreed, pattern established
- WHEN the carve-out is rendered
- THEN the carve-out MUST name only the architecture-or-design-decision trigger it exempts
- AND those four triggers MUST remain unconditional
