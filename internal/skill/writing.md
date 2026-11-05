# Writing docs agents act on

Read this before you write new docs or restructure existing ones. You're writing for the people who own the system, and for the agents who will read your docs and act on them.

The next agent can weigh only what the docs give it. Whatever they leave out isn't weighed and set aside; it never enters, and that agent's answer looks just as sure either way. It already knows standard practice, common tools and how to write code. What it can't know is what the owners decided, why, and where the edges are. You don't know those either, unless the owners told you or pointed you to a source. They come from the owners, through the conversation or the sources they name, never from your own assumptions. Get them, write them, let cca keep them consistent, and leave the rest out.

## Write what the next step depends on

Include what the next agent's work depends on, nothing more and nothing less:

- **Intent and scope.** What the system is for, what it covers, and what it leaves out on purpose. An agent meeting a case nobody wrote down leans on this.
- **The owners' own terms.** Define each term the docs use in a sense of their own, once, in one place. Otherwise the next agent reads it in its everyday sense.
- **Open questions,** in a list of their own, so the next agent neither treats them as settled nor reopens what is.
- **Where facts come from.** A fact that goes stale (a count, a version, a status, an owner) carries a source or an as-of date.

Leave out what any capable reader knows. It costs the reader's attention, and a second copy of anything drifts from the first.

## Decisions

A decision binds every agent that reads it, including on cases nobody imagined when it was made. For each one, write down:

- **What it decides.**
- **Why.** The reason is what tells the next agent whether the decision reaches a case the docs don't name. Without it, that agent settles the question from its own habits, and nothing in its answer shows it did. If the docs already apply a rule beyond its literal words, a reason is at work. Ask the owners for it.
- **What was allowed to count.** The options that were on the table and the criteria that weighed. A criterion ruled out matters as much as one ruled in: "cost is never a criterion" changes every decision downstream of it. Decision records usually keep the reasoning and drop this, though it settled the outcome before any reasoning ran.
- **What it leaves out on purpose.**

A number is a decision too. Record whether a limit comes from a contract, a measurement or an estimate, and whether it's hard. That decides whether the next agent can trade against it.

When the owners haven't given you one of these, ask. If they can't answer yet, put the question in the open list. Don't fill the gap with something plausible: a guessed reason reads exactly like a real one.

## Keep one source for each fact

The next agent acts on what the docs say, and confidently. A wrong doc costs more than a missing one.

- **A value used in more than one place is an attribute**, defined once (`:retention-days: 30`) and written `{retention-days}` wherever the text means the current value. An attribute is the value as it stands now, so a statement about the past keeps the literal ("records were kept 14 days until March"). Run `cca impact retention-days` before changing it.
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

One source is for facts. A concept the docs depend on is different: illustrate it where it matters, each time with a different example, rather than restating it. A concept met several ways in several places holds; the same sentence repeated adds nothing.

## Draw the edge of a rule

A rule that applies in some cases and not others needs its edge shown, not just its center. When the owners give you a rule, find its edge: ask about the closest case you can think of that it might not cover.

- **Put the case it covers beside the closest case it doesn't.** "Fines cap at $10" beside "lost-item charges aren't fines, and have no cap."
- **State the principle that makes a case belong,** not a surface sign of it. A rule keyed to a word fails on cases that share the principle but not the word.
- **Say what to do, and what to do when the rule doesn't apply,** so the next agent has a way out other than guessing.

## Examples

Examples are the strongest thing in a doc an agent reads, and a lone one becomes the template it copies, even where it doesn't fit. Give several that differ, and spend them on the cases that are easy to confuse.

Invent examples freely, as long as they're plainly illustrations; never invent facts. And don't lift an example from a case you'll later check an agent against: that turns the check into recall.

## When an agent misreads the docs

Assume the agent followed the docs, so the misreading is in the docs. More prose rarely fixes it:

- **Ask first whether the rule is missing, or present and not reaching.** A missing rule has somewhere to land. A rule that isn't reaching won't be rescued by a second sentence arguing for it.
- **Change the frame before the wording.** A narrower, concrete statement added alongside works better than a better-worded version of the broad one.
- **Keep a qualifying clause away from the exclusion it qualifies.** Written a line or two apart, the qualifier tends to lose to the exclusion. (Edge examples are different: they belong beside the rule.) Often the better repair is one principle that covers both cases.
- **A confident wrong reading isn't moved by prose.** Make the point structural instead: an attribute, a check, a test, or code that decides rather than a sentence that asks.
- **Check a fix by watching an agent work from the docs** on a case they don't already contain.

These held up under measurement while tuning a production prompt; treat them as strong defaults for docs.

## Stay within the system's scope

A doc set covers its own system and points to other doc sets for theirs. Repeating another system's details makes a second copy that nobody updates.

## Structure

- One concern per file, with `MANIFEST.adoc` including them in reading order. An edit then touches one file, and a diff shows one concern.
- Link with cross-references (`<<delivery>>` to a `[[delivery]]` anchor) instead of repeating content.
- Give sections titles distinct enough to name with `cca compile --section`.

## Rules for agents

When the owners have rules of their own for agents working on the docs (what a decision record must hold, how changes are made, what to check before committing), write them in the docs. The cca skill tells agents to follow them.

## Changing docs

- Make one change per commit, with a message that gives the owners' reason for it. The next agent can then read the change with `cca diff`, and the reason in the log.
- Run `cca validate` before committing. It fails on what's mechanically broken: a missing include, a reference to an undefined attribute, a cross-reference with no target, an include from outside the docs' folder.
- Read the change back the way the next agent will, through `cca compile`, and check it against the sections above. cca can tell you the docs are intact. Whether they say what the owners decided is for you to check against what the owners told you, and to ask about where you can't.
