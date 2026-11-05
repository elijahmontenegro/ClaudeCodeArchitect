# CCA

**Compile, check and query modular AsciiDoc docs, for people and for coding agents.**

CCA turns a doc set split across AsciiDoc files into one Markdown document, with every include expanded and every shared value filled in. It answers questions about the docs: which section, which file, where a value is used, what changed. It fails loudly when the docs are broken, and it can have Claude review them against a checklist. Use it for any docs an agent reads and acts on: a context repo, a glossary, a handbook, a spec.

## Install

**Requirements:** Go 1.26+, asciidoctor. Claude Code >= 2.1 for the skill and the review.

```bash
npm install -g @asciidoctor/cli
go install github.com/elijahmontenegro/ClaudeCodeArchitect/cmd/cca@latest
```

Or download a binary from [Releases](https://github.com/elijahmontenegro/ClaudeCodeArchitect/releases). Each archive has a SHA-256 in `checksums.txt` and a signed build provenance attestation:

```bash
gh attestation verify cca_linux_amd64.tar.gz --repo elijahmontenegro/ClaudeCodeArchitect
```

## Try It

```bash
cd examples/relay
cca list                        # Every section, with its file and line
cca compile --section Delivery  # One section, with its subsections
cca impact retention-days       # Where a shared value is defined and used
cca validate --quick            # Missing includes, undefined attributes, dangling cross-references
cca skill                       # Install the Claude Code skill
```

## Commands

| Command | Purpose |
|---------|---------|
| `cca compile` | The whole document as Markdown. Problems go to stderr. |
| `cca compile --section <query>` | One section with its subsections, or one file. The query is a title, a file path, or `path#Title`. |
| `cca list` | Every section with its file and line |
| `cca impact <attr>` | Where an attribute is defined and every place that uses it |
| `cca diff [commit]` | Compiled output diff between commits |
| `cca validate --quick` | Structural checks (fast, no Claude) |
| `cca validate` | Structural checks, then a Claude review against a checklist |
| `cca validate --ultra` | Three independent reviews, merged by how many found each problem |
| `cca skill` | Install the Claude Code skill |

An ambiguous `--section` query fails and lists the candidates instead of picking one.

## Configuration

CCA finds the entry point by convention: `MANIFEST.adoc`, then `docs/MANIFEST.adoc`. An optional `.cca.yaml` in the docs' root can name another one, and the docs' own review checklist:

```yaml
entry: docs/index.adoc        # entry point, if not found by convention
checklist: review/checks.md   # replaces the default review checklist
```

An unknown key in `.cca.yaml` is an error, so a misspelled setting can't be silently ignored.

## Checks

`cca validate --quick` fails (exit 1) when:

- the docs don't compile;
- asciidoctor reports a warning or an error: a missing include, a reference to an undefined attribute, a section out of sequence, an include from outside the docs' folder;
- a cross-reference has no target;
- there are no sections.

`--json` prints the checks for CI.

## Review

`cca validate` runs the same checks, then sends the compiled docs to Claude with a checklist. The default checklist covers what a capable agent can't work out for itself and what docs most often get wrong:

1. **contradiction**: two places give different values, rules or descriptions for the same thing.
2. **undefined-term**: a term used in a sense of the docs' own is never defined.
3. **decision-status**: an open question is written as decided, or a settled decision as open.
4. **undated-fact**: a fact that changes over time has no source and no as-of date.
5. **unexplained-number**: a number an agent may have to trade against says nothing about where it comes from or whether it's a hard limit.

A doc set with its own methodology names its own checklist in `.cca.yaml`. The file is a list of checks in plain words, and it replaces the default, so a team can review its docs for whatever its process requires.

## Writing Docs

[docs/writing.md](docs/writing.md) is a short guide to writing docs agents act on: write what only you know, keep one source for each fact, and let cca keep it consistent. [examples/relay](examples/relay/) is a small doc set written that way.

## How It Works

1. Finds the entry point via `.cca.yaml` or convention
2. Compiles AsciiDoc to HTML with asciidoctor, keeping its warnings, and reports references to undefined attributes unless the docs set `attribute-missing` themselves
3. Converts the HTML to Markdown: the content alone, without page header or footer
4. Outputs to stdout

Key: `{retention-days}` becomes `30`. An agent sees actual values, not placeholders.

Attributes can be defined in any included file; `impact` and `--section` find them. A section compiled alone is shifted to the top level and gets the whole document's attributes. A file included whole and also by tag elsewhere counts once, where it's included whole.

Includes are confined to the directory holding the entry point (asciidoctor's SAFE mode). A document can't pull in a file from elsewhere on the machine, which matters when an agent compiles a repo it was only asked to read; such an include becomes a warning and fails `validate --quick`.

### Requirements

- **asciidoctor**: AsciiDoc compilation. `npm install -g @asciidoctor/cli`
- **Claude CLI**: the review (`cca validate` without `--quick`)

### Maintenance

CI runs vet, tests, gofmt and `go mod tidy -diff` on Go 1.26 and 1.27, on Linux and Windows, plus govulncheck. GitHub Actions are pinned to commit SHAs. Dependabot proposes Go module and action updates weekly, two weeks after each release.

---

**License:** MIT
