---
name: adoc
description: Read, check and update modular AsciiDoc docs with the cca CLI. Use when a repo has .adoc files, a MANIFEST.adoc or a .cca.yaml, or when asked to read, edit or validate AsciiDoc documentation.
---

# cca: modular AsciiDoc docs

`cca` compiles a doc set into Markdown, with every `include::` expanded and every `{attribute}` resolved, and checks it for breakage. A doc set is a `MANIFEST.adoc` (or the file `.cca.yaml` names as `entry:`) and the files it includes.

If `cca` isn't installed: `go install github.com/elijahmontenegro/ClaudeCodeArchitect/cmd/cca@latest`. It needs asciidoctor (`npm install -g @asciidoctor/cli` or `gem install asciidoctor`).

## Reading

Read through `cca` rather than the raw files. A raw `.adoc` shows `{name}` placeholders and `include::` lines, not what they resolve to.

| Command | Use |
|---|---|
| `cca compile` | The whole document. |
| `cca compile --section "<title>"` | One section with its subsections. Also takes a file path (`systems/api.adoc`), or `path#Title` when two sections share a title. |
| `cca list` | Every section with its file and line: the map for finding what to read or edit. |
| `cca impact <attribute>` | Where a shared value is defined, and every place that uses it. |
| `cca diff [commit]` | What changed in the compiled text since a commit. |

An ambiguous `--section` query fails and lists the candidates. Pick one of them; don't guess.

## Editing

- Edit the `.adoc` source files. `cca list` gives the file and line of each section.
- A value defined as an attribute (`:name: value`, used as `{name}`) is defined once. Change it at its definition, after `cca impact <name>` shows what else it changes. Don't retype the value anywhere; write `{name}`.
- Link to a section with a cross-reference (`<<id>>` to an `[[id]]` anchor) instead of repeating its content.
- Run `cca validate --quick` before committing, and fix what it reports. It fails on a missing include, a reference to an undefined attribute, any other asciidoctor warning, a cross-reference with no target, and an include from outside the docs' folder.

## Writing

Write what only the authors know, and leave out what any capable reader knows:

- intent, scope, and what's deliberately left out;
- decisions with their reasons, and open questions kept in a list of their own;
- where a number comes from, and whether it's a hard limit;
- terms used in a sense of their own, defined once;
- a source or an as-of date on any fact that changes.

Keep one source for each fact: an attribute for a repeated value, a tagged block for a repeated passage, a link where the truth lives elsewhere. The full guide is `docs/writing.md` in the cca repository.

## Rules in the docs

A doc set can state rules for agents, such as how changes are made or what order work happens in. Those rules are part of the docs: follow them. cca adds none of its own.

## Review

`cca validate` without `--quick` runs the structural checks, then has Claude review the compiled docs against a checklist. The default checklist looks for contradictions, undefined terms, open questions written as decided (and the reverse), changing facts with no source or date, and numbers with no stated basis. A doc set can name its own checklist in `.cca.yaml` (`checklist: path/to/checklist.md`), which replaces the default.
