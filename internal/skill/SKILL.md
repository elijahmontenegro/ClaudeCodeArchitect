---
name: adoc
description: Read, check and update modular AsciiDoc docs with the cca CLI. Use when a repo has .adoc files, a MANIFEST.adoc or a .spec.yaml, or when asked to read, edit or validate AsciiDoc documentation.
---

# cca: modular AsciiDoc docs

`cca` compiles a doc set into Markdown, with every `include::` expanded and every `{attribute}` resolved, and checks it for breakage. A doc set is a `MANIFEST.adoc` (or the file `.spec.yaml` names) and the files it includes.

If `cca` isn't installed: `go install github.com/emontenegr/ClaudeCodeArchitect/cmd/cca@latest`. It needs asciidoctor (`npm install -g asciidoctor` or `gem install asciidoctor`).

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
- Run `cca validate --quick` before committing, and fix what it reports. It fails on a missing include, a reference to an undefined attribute, any other asciidoctor warning, and a cross-reference with no target.

## Rules in the docs

A doc set can state rules for agents, such as how changes are made or what order work happens in. Those rules are part of the docs: follow them. cca adds none of its own.

## Validation profiles

`cca validate` without `--quick` runs the structural checks, then a semantic checklist if `.spec.yaml` names a profile. The one profile is `architecture`: a completeness checklist for architecture specs, run with Claude. Without a profile, `cca validate` is the structural checks alone.
