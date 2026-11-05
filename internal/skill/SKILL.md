---
name: adoc
description: Read, check and change modular AsciiDoc docs (a MANIFEST.adoc and the files it includes) with the cca CLI. Use whenever you read, write, edit or check a doc set made of .adoc files, such as a context repo, a handbook or a spec. Cues are a MANIFEST.adoc, a .cca.yaml, or include:: lines.
---

# cca: modular AsciiDoc docs

A doc set is a `MANIFEST.adoc` (or the file `.cca.yaml` names as `entry:`) and every file it includes. cca shows you what the docs say; the `.adoc` files are where you change it.

If `cca` isn't installed: `go install github.com/elijahmontenegro/ClaudeCodeArchitect/cmd/cca@latest`. It needs asciidoctor (`npm install -g @asciidoctor/cli` or `gem install asciidoctor`).

## Reading

The docs are all you have to weigh about the system they describe. A raw `.adoc` file is a fragment: it shows `{name}` where a value goes and `include::` where content goes. An answer built on fragments looks just as sure as one built on the whole, so read through cca:

| Command | Use |
|---|---|
| `cca compile` | The whole document. |
| `cca compile --section "<title>"` | One section with its subsections. Also takes a file path (`systems/api.adoc`), or `path#Title` when two sections share a title. |
| `cca list` | Every section with its file and line. |
| `cca impact <attribute>` | Where a shared value is defined, and every place that uses it. |
| `cca diff [commit]` | What changed in the compiled text since a commit. |

An ambiguous `--section` query fails and lists the candidates. Pick one of them; don't guess.

## Changing

- Find the place with `cca list` and edit the `.adoc` source. Then read the result with `cca compile --section`: the source isn't what the next reader sees.
- A value written as an attribute (`:name: value`, used as `{name}`) has one definition, and it means the value as it stands now. Where the text means the current value, write `{name}`, never the value itself. Where it means a past one ("shipped in v2"), write the literal: the attribute will change under it. Change an attribute at its definition, after `cca impact <name>` shows what else it moves.
- Link to a section with a cross-reference (`<<id>>` to an `[[id]]` anchor) rather than repeating it.
- Run `cca validate` before committing and fix what it reports. It exits 1 on a missing include, a reference to an undefined attribute, any other asciidoctor warning, a cross-reference with no target, or an include from outside the docs' folder.

## What only the owners know

Docs exist to carry what a capable reader can't work out alone: intent and scope, decisions and their reasons, where numbers come from, the docs' own terms. Those come from the people who own the system, not from you. Before you write new docs or restructure existing ones, read `writing.md` in this skill's folder.

When you write or change docs, check what you wrote for these. Each check comes with a case that looks similar but is fine:

- **Contradiction.** Two places say different things about the same thing. *Fires:* one page says loans run 21 days, another says 14. *Fine:* "Loans run 21 days; reference books, 7." That's one rule with its exception stated.
- **Undefined term.** A word used in a meaning these docs give it, and never defined. *Fires:* "Holds lapse after the grace window," with the grace window defined nowhere. *Fine:* "a REST API," which means here what it means everywhere.
- **Decision and question confused.** An open question written as settled, or a settled decision written as open. *Fires:* "E-book lending starts in March," while the open questions still ask whether to offer it. *Fine:* a decision stated once, with its date, and absent from the open questions.
- **Changing fact with no source or date.** A statement that becomes false as time passes, though nobody edits it. *Fires:* "The catalog holds 40,000 titles." *Fine:* "Loans run 21 days." A rule stays true until someone changes it.
- **Decision without its reason.** A decision, rule or number an agent may have to apply to a case nobody named, with nothing saying why. The reason is what tells whether it reaches the new case. If the docs already apply it beyond its literal words, a reason is at work and should be written down. *Fires:* "Fines cap at $10," and nothing more: a policy choice or a legal ceiling? *Fine:* "The branch has 3 floors," a count, not a choice.

Name what's missing; don't supply it. A reason, a number, a date or a definition the owners didn't give is theirs to give: ask them for it, or add it to the open questions. Don't invent constraints, numbers or questions to make the docs look complete. Raise an absence only when the docs' own content calls for it (a term they use, a system they list), not because you'd expect it from somewhere else.

When you report on docs, say what you checked them against. Docs can agree with themselves and still be wrong about the world.

## Rules in the docs

A doc set can state its own rules for agents, such as how changes are made or what a decision record must hold. They're part of the docs: follow them. cca adds none of its own.
