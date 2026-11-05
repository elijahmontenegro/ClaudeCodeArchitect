# Writing docs agents act on

Write what only you know, and let cca keep it consistent. A capable agent already knows standard practice, common tools and how to write code. What it can't know is what you decided, why, and where the edges are. Those are what a doc set is for.

## Write what only you know

- **Intent and scope.** What the system is for, what it covers, and what it deliberately leaves out. An agent meeting a case nobody wrote down leans on this.
- **Decisions, with their reasons.** A decision without its reason gets reopened, or applied where it doesn't fit.
- **Open questions, marked as open.** Keep them in a list of their own, so an agent neither treats them as settled nor reopens what is.
- **Where numbers come from.** Say whether a limit is a contract, a measurement or an estimate, and whether it's hard. That decides whether an agent can trade against it.
- **Your own vocabulary.** Define each term you use in a sense of your own, once, in one place. Otherwise an agent reads it in its everyday sense.
- **What something isn't,** where the confusion is likely.

Leave out what any capable reader knows. It adds length, and a second copy of anything drifts from the first.

## Keep one source for each fact

A capable agent acts on what the docs say, confidently. A wrong doc costs more than a missing one.

- **A value used in more than one place is an attribute**, defined once (`:retention-days: 30`) and written `{retention-days}` everywhere else. Run `cca impact retention-days` before changing it.
- **A passage that applies in several places is a tagged block**, written once and included where it applies:

  ```asciidoc
  // tag::retention[]
  Delivery records are kept {retention-days} days, then deleted.
  // end::retention[]
  ```

  ```asciidoc
  include::systems/delivery.adoc[tag=retention]
  ```

- **Where the truth lives elsewhere** (code, a dashboard, another repo), link to it rather than copy it. A copy says where it came from and as of when.
- **A fact that changes** (a price, a version, a status, an owner) carries a source or an as-of date.

## Cover your own scope

A doc set covers its own system and points to other doc sets for theirs. Repeating another system's details makes a second copy that nobody updates.

## Structure

- One concern per file, with `MANIFEST.adoc` including them in reading order. An edit then touches one file, and a diff shows one concern.
- Link with cross-references (`<<delivery>>` to a `[[delivery]]` anchor) instead of repeating content.
- Give sections titles distinct enough to name with `cca compile --section`.

## Examples

A lone example becomes the template an agent copies, even where it doesn't fit. Give several that differ, or none.

## Changing docs

- Commit one change at a time, with a message that says why. An agent can then read the change with `cca diff` and the reason from the log.
- Run `cca validate --quick` before committing. It fails on what's mechanically broken: a missing include, a reference to an undefined attribute, a cross-reference with no target, or an include from outside the docs' folder.
- Run `cca validate` for a review with Claude. By default it checks for contradictions, undefined terms, open questions written as decided (and the reverse), changing facts with no source or date, and numbers with no stated basis.

## Bring your own checklist

The review's checklist is the default above unless the docs name their own in `.cca.yaml`:

```yaml
checklist: review/checklist.md
```

The file is a list of checks in plain words, which replaces the default. That's where a methodology goes: an architecture team can check its specs for what its own process requires, and a handbook can check for what a handbook needs.
