# CCA

**Compile, check and query modular AsciiDoc docs, for people and for coding agents.**

CCA turns a doc set split across AsciiDoc files into one Markdown document, with every include expanded and every shared value filled in. It answers questions about the docs: which section, which file, where a value is used, what changed. It fails loudly when the docs are broken, and its Claude Code skill tells the agent reading them what to check for. Use it for any docs an agent reads and acts on: a context repo, a glossary, a handbook, a spec.

## Install

**Requirements:** Go 1.26+, asciidoctor. Claude Code >= 2.1 for the skill.

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
cca validate                    # Missing includes, undefined attributes, dangling cross-references
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
| `cca validate` | Checks for broken docs; exit 1 if any (`--json` for CI) |
| `cca skill` | Install the Claude Code skill |

An ambiguous `--section` query fails and lists the candidates instead of picking one.

## Configuration

CCA finds the entry point by convention: `MANIFEST.adoc`, then `docs/MANIFEST.adoc`. An optional `.cca.yaml` in the docs' root can name another one:

```yaml
entry: docs/index.adoc
```

An unknown key in `.cca.yaml` is an error, so a misspelled setting can't be silently ignored.

## Checks

`cca validate` fails (exit 1) when:

- the docs don't compile;
- asciidoctor reports a warning or an error: a missing include, a reference to an undefined attribute, a section out of sequence, an include from outside the docs' folder;
- a cross-reference has no target;
- there are no sections.

`--json` prints the checks for CI.

## Reviewing What the Docs Say

Whether docs say what only their authors know is a judgment, not a mechanical check, and it belongs to the agent reading them. The skill gives that agent the checks, each beside a case that looks similar but is fine: contradictions, undefined terms, decisions and open questions confused, changing facts with no source or date, and decisions without the reasons that say where they apply. It also tells the agent to name what's missing rather than invent it.

A doc set with its own methodology states its rules in its own docs, and the skill tells agents to follow them. cca builds in none.

## Writing Docs

[docs/writing.md](docs/writing.md) is a short guide to writing docs agents act on: write what only you know, keep one source for each fact, and let cca keep it consistent. [examples/relay](examples/relay/) is a small doc set written that way.

## How It Works

1. Finds the entry point via `.cca.yaml` or convention
2. Compiles AsciiDoc to HTML with asciidoctor, keeping its warnings, and reports references to undefined attributes unless the docs set `attribute-missing` themselves
3. Converts the HTML to Markdown: the content alone, without page header or footer
4. Outputs to stdout

Key: `{retention-days}` becomes `30`. An agent sees actual values, not placeholders.

Attributes can be defined in any included file; `impact` and `--section` find them. A section compiled alone is shifted to the top level and gets the whole document's attributes. A file included whole and also by tag elsewhere counts once, where it's included whole.

Includes are confined to the directory holding the entry point (asciidoctor's SAFE mode). A document can't pull in a file from elsewhere on the machine, which matters when an agent compiles a repo it was only asked to read; such an include becomes a warning and fails `cca validate`.

### Requirements

- **asciidoctor**: AsciiDoc compilation. `npm install -g @asciidoctor/cli`

### Maintenance

CI runs vet, tests, gofmt and `go mod tidy -diff` on Go 1.26 and 1.27, on Linux and Windows, plus govulncheck. GitHub Actions are pinned to commit SHAs. Dependabot proposes Go module and action updates weekly, two weeks after each release.

---

**License:** MIT
