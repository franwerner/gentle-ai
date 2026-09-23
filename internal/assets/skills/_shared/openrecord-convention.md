# openrecord Convention (shared across the six participating SDD skills)

This file is not self-contained. It fixes which phase does what and carries the two guarantees no
emitted skill carries; the retrieval and write **mechanics** live in the skills openrecord itself
emits — `openrecord-consult/SKILL.md` and `openrecord-capture/SKILL.md`, written as siblings of the
`_shared/` directory this file sits in, under the same skills root. Where a section below sends you to
one of them, open it and follow it there.

**If the skill a section names is not there, say so and stop that line of work.** A runtime can hold
this convention without holding those skills — it may have no skills directory at all, or it may never
have been selected for openrecord's fan-out. Never reconstruct the moves from memory or from an older
copy of this file: a remembered command shape is exactly the drift this delegation exists to end.

## Activation — always, never conditional

This convention always applies. openrecord is installed unconditionally, so there is nothing to
declare, nothing to resolve, and no value to read before you act on the sections below. Do not look
for one: there is no axis to query, no config key to check, and no status field that would tell you
whether this convention is in force. It is.

## Per-phase responsibility map

| Phase | Responsibility |
| --- | --- |
| `sdd-explore` | **Consults.** Runs the walk documented in `skills/openrecord-consult/SKILL.md` and reports how each record surfaced and which candidates it examined and ruled out. |
| `sdd-propose` | **Contradiction-checks.** Runs the "When your work contradicts an accepted record" check below before its approval gate is offered. |
| `sdd-spec` | **Contradiction-checks, and proposes behaviour.** Same check as propose. Additionally names each capability's type and writes its affected sections out whole, reading the record that already exists so nothing it held is lost. Spec never writes a record file. |
| `sdd-design` | **Consults and contradiction-checks.** Additionally proposes new records — `## Architecture Decisions` is where a new record appears as a proposal with its rationale. Design never writes a record file. |
| `sdd-apply` | **Writes, then reconciles.** The sole writer — the contract is under "Writing records" below, the commands in `skills/openrecord-capture/SKILL.md`. Runs the neighbourhood check documented in `skills/openrecord-reconcile/SKILL.md` against each record right after it lands, one anchor at a time. |
| `sdd-verify` | **Validates.** Runs `openrecord validate` for structural shape, and runs the audit documented in `skills/openrecord-audit/SKILL.md` — claim by claim, with `file:line` evidence — to check whether each record's prose still describes the system as actually built. |

`sdd-init`, `sdd-onboard`, `sdd-research`, `sdd-tasks` and `sdd-archive` never touch the record store.
The remediation loop inherits `sdd-apply`'s writer contract; `sdd-archive` writes nothing into it.

## The walk (consult, propose, spec, design)

The retrieval moves — which component owns a path, how to descend the levels, how to search literally
and by meaning, how to tell a search that ran from one that could not, and how to state what you found
and what you ruled out — are documented in `skills/openrecord-consult/SKILL.md`, beside this file under
the same skills root. Run the walk from there. This convention keeps no copy of it, precisely so that a
copy cannot fall out of step with the binary.

If that file is not there for you, say so and stop the consult, per the preamble — do not improvise the
walk.

## The audit (verify)

The claim-by-claim check — separating what a record claims from why it claims it, resolving the surface
it governs, reading the code whole rather than grepping for the record's own wording, and giving each
claim a verdict backed by `file:line` evidence — is documented in `skills/openrecord-audit/SKILL.md`,
beside this file under the same skills root. Run the audit from there. This convention keeps no copy of
it, precisely so that a copy cannot fall out of step with the binary.

If that file is not there for you, say so and stop the audit, per the preamble — do not improvise the
check.

## A record that is `pending` settles nothing

`accepted` is what a record reaches because a person put it there. `pending` is the other thing
entirely: someone left that question open, deliberately, and said so.

So a pending record constrains nothing — and it authorises nothing either. Reading one as "the
question is open, so I get to choose" invents a mandate nobody granted, and it is a tempting reading
precisely because a pending record looks like an invitation. It is a record of an open question, not
of a delegation.

When your work depends on a pending question being settled, say that out loud rather than settling it
in passing. The person reading your return is the one who left it open; that it now blocks something
is exactly what they need to hear, and it is more useful to them than a choice made on their behalf
and reported afterwards.

This is not the contradiction contract below, and does not fire it: nothing is being contradicted.
A pending record is not an obstacle you have to route around — it is a question with your work now
attached to it.

## When your work contradicts an accepted record

This is the one thing that stops.

Stop that line of work — not everything, only what depends on the conflict — and surface it with both
sides in view:

```
STOPPED: <the specific piece of work>

What the record settles:
  decisions/<component>/<concern>/<slug>.md   [accepted]
  <the record's own claim, verbatim>
  Its reason: <why the record settled it that way>

What the work requires:
  <the concrete requirement that conflicts>

Why both cannot hold:
  <the actual contradiction>

Ways forward:
  A. <option, and what it costs>
  B. <option, and what it costs>
  C. <option, and what it costs>

Recommendation: <one of the above>
```

Arriving with "this conflicts, what do I do?" puts the whole problem back on the person. Arriving with
both sides and their reasons lets them decide in one step.

**Never resolve it on your own.** A record was ratified by a person; overriding it silently makes the
store a lie, and everyone downstream keeps trusting it. Resolving it alone and reporting afterwards is
not a consultation — it is a decision taken and then announced. If the record turns out to be wrong or
stale, that is a fine outcome; it is just not yours to conclude alone.

## Writing records — apply is the sole writer

Only `sdd-apply` writes into the record store, in the same step that implements the governing code —
never a separate pass, and never design or archive.

The write commands — `openrecord level add`, `openrecord record write --body-file`, `openrecord record
edit --section --body-file`, and which verb applies to a record that does not exist yet versus one that
already does — are documented in `skills/openrecord-capture/SKILL.md`, beside this file under the same
skills root.

**Turning a design decision into a record.** Each `### Decision: {Title}` entry under the design
artifact's `## Architecture Decisions` becomes exactly one record. The mapping is fixed, and it is
stated in record-field terms: which flag carries each field is the capture skill's to say, and this
file keeps no copy of it.

| Design field (`## Architecture Decisions`) | Record input |
| --- | --- |
| `### Decision: {Title}` | the record's title |
| `**Choice**` | the record's one-line description — compressed to one line when the design's runs longer, never cut at the first line break |
| `**Alternatives considered**` | the body's `## Alternatives` |
| `**Rationale**` | the body's `## Decision`, developed |
| *(no source field)* | the body's `## Context` — authored under the pointing test below |
| *(no source field)* | the body's `## Consequences` — authored under the pointing test below |
| *(no source field)* | the status — always `accepted` |

The status is not a judgement call here. `sdd-apply` writes after the governing code has landed, so the
proposal-stage value cannot arise on this path at all.

**Where the record goes.** `decisions/<component>/<concern>/<slug>.md`. `component` is resolved from
the design's own `## File Changes` paths against the store's declared component owners — the move
`skills/openrecord-consult/SKILL.md` documents; run it from there. `concern` is chosen by judgement
against openrecord's shipped concern catalogue: its per-concern topics are there to prompt that
judgement, not to serve as keys that pick a concern for you. `slug` is named freely from the decision's
title.

A `component`/`concern` level that does not exist yet is neither a failure nor an absent store.
Creating it is one of the moves the skill named just above documents, and that skill already instructs
it before a write. Create the level, write the record, and report nothing about it.

**Turning a capability spec into a record.** Each capability the change's spec artifact names becomes
exactly one record. That artifact also carries the capability's type — an operation an actor triggers,
a rule holding across several of them, an entity's states and what moves between them, or work an event
sets off rather than a person — because the phase that wrote the behaviour is the one that knows which
it is. `sdd-apply` never re-derives it from the prose.

| Spec artifact | Record input |
| --- | --- |
| the capability's name | the record's title |
| what the capability is for, in one line | the record's description |
| the surfaces the design's `## File Changes` paths resolve to | the surfaces the record declares — mandatory for behaviour, and a record declaring none crosses with no path and drops out of every scope check silently |
| `## Scenarios`, and the sections the declared type asks for | the body sections of those same names |
| *(no source field)* | the status — always `accepted` |

**Whole sections, never a patch.** A record holds a document rather than a history of edits, and a
section is the smallest thing that can be replaced — so handing over only this change's part of a
section deletes everything that section already held. `sdd-spec` therefore reads the record that exists
and writes each affected section out whole, already carrying both what was there and what this change
adds; `sdd-apply` ships what it was handed and merges nothing. A capability with no record yet is
written whole from the same sections, for the same reason.

**Where a capability spec goes.** `specs/<type>/<slug>.md` — `type` as the spec artifact declared it,
`slug` named freely from the capability's name. A level that does not exist yet is created exactly as a
decision's is, by the move the skill named above documents.

**When.** Inside an `applyState: ready` run, after the tasks are marked complete and before the return,
and only when that batch leaves no pending task in the tasks artifact. A `## Architecture Decisions`
entry governs the design as a whole rather than one task, so the step that implements it is the step
that finishes the change's implementation. A capability the spec artifact names is change-wide in the
same way, and lands under the same gate, in the same run. Under a multi-batch apply the batch clearing the last
pending task materializes; every earlier batch writes nothing and says nothing about it, because a
later batch owning the write is the normal case, not an issue.

**A record store that is not on disk.** Write nothing and create nothing — never bootstrap a store —
name the skip in the `### Issues Found` section the return already carries, and finish the batch: the
task's code is unaffected. Do not let the skip pass in silence either. A project whose record store is
missing is a configuration defect.

That line is where the skip is **named**, not where it is **caught**. `### Issues Found` gates nothing,
so a skip reported only there repeats every run while the reasoning it dropped stays in one place that
was never written. `sdd-verify` is what catches it: a design carrying `## Architecture Decisions`
entries whose reasoning reached no record is a CRITICAL finding there, not a note. Apply reports;
verify escalates.

**Reconciling after each write.** Immediately after `record write` or `record edit` lands for a record,
run the neighbourhood check documented in `skills/openrecord-reconcile/SKILL.md`, beside this file under
the same skills root, with that record as the anchor. The anchor mechanism is single-record by design,
so a batch that materializes several records runs this check that many times — once per record, never
once for the whole batch. Do this per record as it lands, before the design-artifact rewrite below,
which stays a once-per-batch pass.

If that file is not there for you, say so and stop the reconcile, per the preamble — do not improvise
the neighbourhood check.

**Reconcile's findings.** `skills/openrecord-reconcile/SKILL.md` reports four categories, and they do
not all belong in the same lane. A direct contradiction and a spec asking for what a decision forbids
are both cases of two `accepted` records — or a spec and a decision — that cannot both hold right now;
they stop under the discipline of "When your work contradicts an accepted
record" above — stop only the line of work that depends on the conflict, both sides in view, never
resolved here — but take their shape from `skills/openrecord-reconcile/SKILL.md`, which pairs the
anchor against the other record rather than against the work that section's block is written for. Duplication and a broken dependency are store hygiene,
not a work-stopping contradiction today, and the `STOPPED` block is not built for either of them; name
them in the `### Issues Found` section the return already carries instead, the same lane a missing
record store is named in above.

Reconcile's third neighbourhood source depends on `openrecord search`'s semantic pass, which can run,
degrade to a lexical-only search, or come back unavailable. Report which of the three happened, exactly
as `skills/openrecord-reconcile/SKILL.md` itself requires, rather than letting a thinner reconcile — one
that saw fewer neighbours because the meaning-based pass did not run — read the same as a clean one. A
degraded pass does not block `sdd-apply`; failing to report the degradation does.

**After a record exists, the design artifact stops carrying its reasoning.** `sdd-design` writes each
decision's `**Choice**`, `**Alternatives considered**` and `**Rationale**` in full, and must keep doing
so — that artifact is the only channel the reasoning has to reach `sdd-apply`. Once the record holds
it, though, a second durable copy is the duplication this convention exists to remove. So after every
record this batch materializes has landed, and before the return, rewrite the persisted design artifact
so each materialized entry keeps its `### Decision: {Title}` heading and one `**Record**:` line naming
the record's path, and nothing else. No status word joins that line: the record carries the status, and
a copy here would be a second thing to keep in step.

Three things bind that rewrite. It happens **once per batch, after the last record lands** — never
interleaved between writes, so an interruption leaves records written and reasoning still present,
which is the direction that loses nothing. It **never happens when nothing was written**: the skip
above, or a batch still holding pending tasks, means the reasoning is the only copy there is and
removing it destroys it. And it replaces the artifact **whole** — re-read the persisted document, swap
that one section inside the text you read, and write the entire document back. The memory store's
update replaces a whole observation, so handing it the pruned section alone deletes every other section
the design wrote.

An entry that already carries a `**Record**:` line was materialized by an earlier run: skip it
entirely — no record, no rewrite, nothing reported. That marker is the test, and it is read per entry
rather than per section, because a design re-run after a rewrite can add a fresh entry beside pruned
ones.

**The pointing test.** `sdd-apply` did the work, so it holds genuine knowledge of most of what
`## Context` needs — which is exactly why the line between what it recalls and what it infers is
hardest to see here, and why a confident sentence is cheapest to write. So the test runs per sentence,
before you write it: **can you point at the thing that makes it true** — a file you changed, a
constraint you hit, a test you wrote, a line of the spec? Point at it, or do not write it, however
plausible it reads.

When a section has nothing you can point at, write that absence into it in one line — `not established
during implementation` — and move on. That is the section's correct content, not a defect to be filled
in later. The same test binds `## Decision`: developing the design's `**Rationale**` means explaining
the reason the design gave, and never adding a reason it did not.

**Never touch a record file with a file-editing tool.** A record is a `.md`, and a phase carrying Edit
and Write tools will reach for them exactly as it would for any other file — but every guard openrecord
performs lives on the `record write` / `record edit` path, and an edit that goes around that path goes
around all of them. This binds `sdd-apply` hardest, since it is the sole writer: create with
`openrecord record write`, update with `openrecord record edit`, always — never a direct Edit or Write
call against a file under `decisions/` or `specs/`. It also binds `sdd-design`, whose `## Architecture
Decisions` proposes a record without ever writing one — the proposal is prose in the design artifact,
never a patch to the record file itself. openrecord does not enforce this: the binary does not police
anyone's editor, so carrying this guard is this convention's job, not the tool's.

`sdd-verify` runs `openrecord validate` for structural shape (frontmatter, template sections, branch
anchors) and now also for `body-hash-mismatch` — a record whose body no longer matches the hash the
tool stamped when it was last written or edited through the binary, meaning something changed it
outside `record write` / `record edit`. That check is detection, not prevention: it finds the edit
after the fact, it does not stop it from happening.

Separately, `sdd-verify` runs the audit documented in `skills/openrecord-audit/SKILL.md` against each
record's prose, claim by claim, with `file:line` evidence — a judgement the binary cannot make on its
own. Audit only detects, by its own hard limit: it never edits the record, never fixes the code, and
never changes `status`, not even when the fix looks small. Detecting a contradiction is a disciplined
judgement backed by mandatory evidence, which an SDD phase can do unattended; resolving one — deciding
which side is wrong — is never inferred, and is always a person's call.

That split matters most exactly here, because `sdd-verify` runs immediately after `sdd-apply`, and
`sdd-apply` wrote both the code and the record in that same cycle. A contradiction the audit finds at
this point is most likely apply having implemented something different from what it recorded, not a
record that aged out on its own. Rewriting the record to match the code would do two harmful things at
once: it would hide that implementation defect, and it would stamp the defect into the store as a
ratified decision that later readers treat as settled.

So a contradiction the audit finds stops that line of work under the discipline of "When your work
contradicts an accepted record" above, and surfaces in the `STOPPED` block `skills/openrecord-audit/SKILL.md`
carries for it — which sets the record against the code, not against the work that section's own block is
written for. The remediation loop never resolves it, and never rewrites the record on the audit's say-so. Accept the consequence this carries: a contradiction here
means the SDD cycle stops and asks, every time. `sdd-archive` writes nothing into the record store,
ever.
