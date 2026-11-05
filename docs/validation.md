# Completeness Checklist

This is the checklist `cca validate` runs when `.spec.yaml` sets `profile: architecture`. Docs without a profile get the structural checks only.

## Validation Requirements

Before submitting specification to AI, **verify**:

* [ ] All library dependencies include exact versions
* [ ] All "or" choices resolved (single option specified)
* [ ] Zero conditional logic ("if X then Y" eliminated)
* [ ] Zero optional sections (include or exclude, not "maybe")
* [ ] Complete file tree provided (local to component)
* [ ] All type definitions complete (every field, every type)
* [ ] Example data formats included (actual examples, not schemas)
* [ ] Database schema complete (tables, columns, indexes, constraints)
* [ ] API routes fully specified (method, path, params, responses)
* [ ] Performance requirements quantified (numbers, not adjectives)
* [ ] All numeric constants have derivation or rationale
* [ ] Data formats include complete examples
* [ ] Error handling specified (not "handle appropriately")
* [ ] Concurrency model specified (if applicable)
* [ ] Persistence format specified (structure, not "save to disk")
* [ ] Deployment configuration complete (specific platform/approach)
* [ ] Constants/config/secrets properly separated (if applicable)
* [ ] No weak obligation language ("should", "could", "might")
* [ ] Testing requirements specified (what to test, coverage targets)

**Pass rate:** `19/19` = ***Ready for one-shot implementation***

**Pass rate:** `<19/19` = ***Incomplete specification, resolve gaps***

## Specification Completeness Test

### Method

Give specification to someone unfamiliar with project AND domain.

Ask: "Can you implement this without asking questions?"

* If answer is YES: Specification is complete
* If answer is NO: Document all questions, resolve in spec

### Target

Zero questions needed for 95% of implementation.

Only acceptable questions: Clarifying intent, not filling gaps.

### Note

Test with someone from DIFFERENT domain.

Ensures spec doesn't assume domain knowledge.

# Common Incompleteness Patterns

## Introduction

Passing checklist does not guarantee actual completeness.

Learn to recognize subtle incompleteness patterns.

## Deferral Language

### Description

Specification contains language that defers to future implementation.

### Recognition Indicators

* Qualifiers suggesting incompleteness ("simplified", "basic", "rough")
* References to undefined content ("see X", "defined elsewhere")
* Promises of detail later ("TBD", "during implementation")

### Reasoning

If spec references something not present in spec, spec is incomplete.

Implementation cannot proceed without that information.

### Self-Check

Scan specification for deferral phrases.

Result should be: None found.

Every reference resolved within specification itself.

## Type Reference Without Definition

### Description

Type or structure mentioned but fields/properties not enumerated.

### Recognition

Type name appears in spec.

Complete structure definition missing.

### Reasoning

AI cannot generate code for type without knowing its structure.

Field names, types, sizes must all be specified.

### Self-Check

List all type names mentioned in specification.

Verify each has complete definition with all fields.

100% coverage required.

## Algorithm Without Enumerated Steps

### Description

Algorithm named but procedure not specified step-by-step.

### Recognition

Operation described in prose.

Numbered steps or explicit formula missing.

### Reasoning

AI needs precise procedure to implement correctly.

Prose descriptions leave too much interpretation.

### Self-Check

List all algorithms/operations in specification.

Verify each has numbered steps OR complete formula.

No prose-only descriptions.

## Referenced Undefined Operations

### Description

Methods, functions, or operations called but behavior never defined.

### Recognition

Code snippet or description shows call to operation.

Operation's signature or behavior unspecified elsewhere in spec.

### Reasoning

Cannot implement caller without knowing what called operation does.

Every operation must have signature + behavior specification.

### Self-Check

List all function/method calls in specification.

Verify each has corresponding definition.

No dangling references.

## Quantification Gaps

### Description

Numeric values or performance requirements missing concrete numbers.

### Recognition

Adjectives instead of numbers ("fast", "large", "efficient").

Ranges without bounds ("support many users").

Constraints without values ("reasonable timeout").

### Reasoning

Without numbers, AI cannot make appropriate trade-offs.

"Fast" is subjective. "P99 <100ms" is measurable.

### Self-Check

List all performance/scale descriptions.

Verify each has concrete numbers.

No unmeasurable adjectives.

## Ambiguous Intent

### Description

Specification provides complete implementation details but leaves unstated priorities. AI must guess when resolving edge cases.

### Recognition

* Spec jumps directly into types/schemas without establishing purpose
* Trade-off decisions have no stated rationale
* Scope boundaries are implicit (what's NOT included is unclear)
* Performance numbers lack justification (why 100ms, not 50ms?)

### Reasoning

An AI implementing this spec can produce syntactically correct code but may make wrong judgment calls for unlisted scenarios.

Same spec with different intent contexts produces radically different implementations:
* "Public API where abuse is common" → aggressive rate limiting
* "Internal API for trusted services" → lenient limits

The implementation details alone don't reveal which is correct.

### Self-Check

Ask: Could an AI infer the system's purpose from this spec?

If the AI encountered an edge case not explicitly covered, would it know which direction to lean?

If answers are no, intent is ambiguous. Add Abstract/Approach/Scope to Context section.

## Self-Validation Process

### Procedure

After writing specification, apply systematic review:

**1. Deferral language audit**

Read specification looking for references to external/future content.

Pass criterion: Zero deferrals found.

**2. Type completeness verification**

List all type names → verify each has complete structure definition.

Pass criterion: 100% of types fully defined.

**3. Algorithm enumeration check**

List all operations → verify precise procedure exists for each.

Pass criterion: 100% have steps or formulas.

**4. Operation closure verification**

List all method/function calls → verify behavior specified.

Pass criterion: 100% of operations defined.

**5. Quantification audit**

List all constraints/requirements → verify numbers present.

Pass criterion: All requirements quantified.

**6. Checklist literal compliance**

Execute 19-point checklist exactly as written.

Pass criterion: 19/19, not "mostly" or "close enough".

### Interpretation

Passing all checks necessary but not sufficient for completeness.

Checks catch mechanical gaps.

Still requires judgment about whether specification is truly complete.

If uncertain: More detail better than less.

## Recognition Training

### Developing Intuition

Develop sense for incompleteness through practice.

When reading specification, notice internal responses:

* "I would need to ask a question to implement this"
* "This part feels vague"
* "What does this actually mean operationally?"
* "How would I implement this concretely?"

These responses indicate incompleteness.

Resolve before considering specification done.

### Completeness Sensation

Complete specifications feel like executable pseudocode.

**Reading experience:**

* Implementation path is clear
* No questions arise
* Details are concrete
* Ambiguities don't exist

If reading generates questions, specification needs work.

### Iterative Refinement

First drafts are rarely complete.

**Process:**

1. Write initial specification
2. Apply self-validation process
3. Identify gaps through systematic checks
4. Resolve gaps with specific details
5. Repeat until no gaps found

Expect 2-3 refinement passes for complex systems.

Completeness is achieved, not written on first attempt.
