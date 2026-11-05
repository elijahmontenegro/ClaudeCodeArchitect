# CCA

**Compile, check and query modular AsciiDoc docs, for people and for coding agents.**

CCA turns a doc set split across AsciiDoc files into one Markdown document, with every include expanded and every shared value filled in. It answers questions about the docs: which section, which file, where a value is used, what changed. And it fails loudly when the docs are broken. Use it for any docs an agent reads: a context repo, a glossary, a handbook, an architecture spec.

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
cd your-docs
cca list                       # Every section, with its file and line
cca compile --section "Intro"  # One section, with its subsections
cca validate --quick           # Missing includes, undefined attributes, dangling cross-references
cca skill                      # Install the Claude Code skill
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
| `cca validate` | Structural checks, plus the profile's semantic checklist |
| `cca skill` | Install the Claude Code skill |

An ambiguous `--section` query fails and lists the candidates instead of picking one.

## Configuration

CCA finds the entry point by convention: `MANIFEST.adoc`, then `docs/MANIFEST.adoc`. An optional `.cca.yaml` in the docs' root names another one:

```yaml
entry: docs/index.adoc
```

An unknown key in `.cca.yaml` is an error, so a misspelled setting can't be silently ignored.

## Checks

`cca validate --quick` fails (exit 1) when:

- the docs don't compile;
- asciidoctor reports a warning or an error: a missing include, a reference to an undefined attribute, a section out of sequence;
- a cross-reference has no target;
- there are no sections.

`--json` prints the checks for CI.

`cca validate` runs the same checks, then the profile's semantic checklist with Claude. The one profile is `architecture` (below). Without a profile, `cca validate` is the structural checks alone.

## How It Works

1. Finds the entry point via `.cca.yaml` or convention
2. Compiles AsciiDoc to HTML with asciidoctor, keeping its warnings, and reports references to undefined attributes unless the docs set `attribute-missing` themselves
3. Converts the HTML to Markdown: the content alone, without page header or footer
4. Outputs to stdout

Key: `{api-p99-latency}` becomes `100ms`. An agent sees actual values, not placeholders.

Attributes can be defined in any included file; `impact` and `--section` find them. A section compiled alone is shifted to the top level and gets the whole document's attributes.

Includes are confined to the directory holding the entry point (asciidoctor's SAFE mode). A document can't pull in a file from elsewhere on the machine, which matters when an agent compiles a repo it was only asked to read; such an include becomes a warning and fails `validate --quick`.

### Requirements

- **asciidoctor**: AsciiDoc compilation. `npm install -g @asciidoctor/cli`
- **Claude CLI**: the architecture profile's semantic checklist only

### Maintenance

CI runs vet, tests and `go mod tidy -diff` on Go 1.26 and 1.27, on Linux and Windows, plus govulncheck. GitHub Actions are pinned to commit SHAs. Dependabot proposes Go module and action updates weekly, two weeks after each release.

## Architecture Specs (`profile: architecture`)

CCA started as a system for writing architecture specifications that AI can implement correctly on the first try. That methodology lives on as a profile: `cca validate` checks a spec against the [19-point completeness checklist](docs/validation.md). A spec that wants agents to follow an implementation protocol states it in its own text.

### Why Use This?

**Bad spec:** "Use a database. Make it fast. Add caching if needed."

**Result:** AI asks questions, makes wrong choices, misses requirements.

**Good spec:** `PostgreSQL 16. P99 latency <100ms. Cache with 5min TTL (Redis 7).`

**Result:** AI implements exactly what you need, first time.

The difference is completeness, decision-making, and quantification.

### Quick Example

#### Before (Incomplete)
```
Build a user API.
- Handle authentication
- Store users
- Make it scalable
```

**Problems:** What auth? Which database? How scalable?

#### After (Complete)
```
== Context

Identity: User Authentication API
Paradigm: Stateless JWT authentication

Abstract: REST API for user registration, authentication, and
profile management.

Approach: Issue JWT tokens on login, services validate locally
without round-trip to auth service.

Scope In: Registration, login, token refresh
Scope Out: OAuth/social login, MFA (phase 2)

Stack: Go 1.21, PostgreSQL 16, Redis 7

== Implementation

Types: User struct (id uuid, email string unique, password_hash bcrypt)
Database: users table with indexes on email/username
API: POST /login, GET /users/:id, PUT /users/:id, DELETE /users/:id
Performance: P99 <100ms, support 1000 req/sec
Rate limit: 100 requests/min per IP
```

**Result:** AI knows WHAT to build AND WHY it exists. Can make correct judgment calls for unlisted edge cases.

### Core Principles

1. **Decided, Not Deciding** - Every choice made before writing spec
2. **Zero Conditionals** - No "if/when/maybe", single implementation path
3. **Complete Types** - Every field, every type, all constraints
4. **Quantified Performance** - Numbers, not adjectives (`P99 <100ms` not "fast")
5. **Mathematical Derivation** - Why every constant has its value
6. **Modular Structure** - File-per-concern using AsciiDoc includes

### Writing Specifications

Learn the methodology:

1. **[Principles](docs/principles.md)** - Decided not deciding, zero conditionals, completeness
2. **[Structure](docs/structure.md)** - File organization, types, algorithms, data formats
3. **[Frameworks](docs/frameworks.md)** - When to be explicit vs let AI decide
4. **[Anti-Patterns](docs/anti-patterns.md)** - Common mistakes to avoid
5. **[Validation](docs/validation.md)** - 19-point completeness check

See it in practice: **[Simple API Example](examples/simple-api/)**

### Example Structure

```
myproject/
├── MANIFEST.adoc              # Entry point, composes all parts
├── core/
│   ├── metadata.adoc          # System context
│   └── types.adoc             # Complete type definitions
├── concerns/
│   ├── performance.adoc       # Quantified requirements
│   └── security.adoc          # Security constraints
├── interfaces/
│   └── api.adoc               # Complete API spec
├── operations/
│   ├── deployment.adoc        # Deployment config
│   └── testing.adoc           # Test requirements
└── (implementation files)     # AI implements here
```

### Full Documentation

- **[Principles](docs/principles.md)** - Core concepts
- **[Structure](docs/structure.md)** - Modular organization
- **[Frameworks](docs/frameworks.md)** - Decision-making
- **[Anti-Patterns](docs/anti-patterns.md)** - What not to do
- **[Workflow](docs/workflow.md)** - Git-based evolution
- **[Testing](docs/testing.md)** - Test decision framework
- **[Validation](docs/validation.md)** - Completeness checking
- **[Implementation](docs/implementation.md)** - What to specify, what to omit

---

**License:** MIT
