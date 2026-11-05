package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/compiler"
	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/completion"
	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/config"
	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/differ"
	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/impact"
	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/skill"
	"github.com/elijahmontenegro/ClaudeCodeArchitect/internal/validator"
	versionpkg "github.com/elijahmontenegro/ClaudeCodeArchitect/internal/version"
)

var version = "dev" // set via ldflags: -X main.version=

func getVersion() string {
	// ldflags takes priority (goreleaser sets this)
	if version != "dev" && version != "" {
		return version
	}
	// go install embeds version in build info
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "dev"
}

func main() {
	// Check for updates (non-blocking, cached)
	if latest := versionpkg.CheckForUpdate(getVersion()); latest != "" {
		fmt.Fprintf(os.Stderr, "cca %s available (current: %s) - go install github.com/elijahmontenegro/ClaudeCodeArchitect/cmd/cca@latest\n\n", latest, getVersion())
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Check for skill update (skip if running skill command itself)
	if command != "skill" && skill.NeedsUpdate(skill.GetProjectSkillDir()) {
		fmt.Fprintf(os.Stderr, "Skill update available — run `cca skill`\n\n")
	}

	var err error
	switch command {
	case "compile":
		err = runCompile()
	case "validate":
		err = runValidate()
	case "diff":
		err = runDiff()
	case "impact":
		err = runImpact()
	case "list":
		err = runList()
	case "skill":
		err = runSkill()
	case "completion":
		runCompletion()
		return
	case "version", "-v", "--version":
		fmt.Printf("cca %s\n", getVersion())
		return
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`cca - compile, check and query modular AsciiDoc docs

Usage:
  cca compile                      Compile the docs to Markdown (stdout); problems go to stderr
  cca compile --section <query>    Compile one section or file (title, path, or path#Title)
  cca list                         List every section with its file and line
  cca impact <attribute>           Show where an attribute is defined and used
  cca diff [commit]                Diff compiled output vs commit (default: HEAD~1)
  cca validate --quick             Structural checks: compiles, no asciidoctor warnings,
                                   cross-references resolve
  cca validate                     Structural checks, then a Claude review against a checklist
  cca validate --ultra             Three independent reviews, merged
  cca validate --yes               Skip confirmation for large docs
  cca skill                        Install/update the Claude Code skill
  cca skill --global               Install to ~/.claude/skills (all projects)
  cca completion [bash|zsh|fish]   Generate shell completion script
  cca version                      Show version
  cca help                         Show this help

Flags:
  --quick, -q     Structural checks only, no Claude review
  --ultra, -u     Three reviews in parallel, then a synthesis
  --yes, -y       Skip interactive confirmation
  --json          Output JSON (for CI, use with --quick)

Configuration:
  Optional .cca.yaml in the docs' root:
    entry: docs/index.adoc    # entry point, if not found by convention
    checklist: checklist.md   # the docs' own review checklist, in place of the default

  By convention, cca looks for:
    - MANIFEST.adoc
    - docs/MANIFEST.adoc

Examples:
  cca compile                           # The whole document
  cca compile --section "Terms"         # One section, with its subsections
  cca compile --section systems/api.adoc # One file
  cca compile --section api.adoc#Scope  # One of several sections with the same title
  cca validate --quick                  # Before committing
  cca diff HEAD~1                       # What the last commit changed
  cca impact retention-days             # Before changing a shared value
`)
}

func runCompile() error {
	entryPath, err := config.FindEntry(".")
	if err != nil {
		return err
	}

	// Check for --section flag
	sectionQuery := ""
	for i, arg := range os.Args {
		if arg == "--section" && i+1 < len(os.Args) {
			sectionQuery = os.Args[i+1]
			break
		}
		if strings.HasPrefix(arg, "--section=") {
			sectionQuery = strings.TrimPrefix(arg, "--section=")
			break
		}
	}

	var report compiler.Report
	if sectionQuery != "" {
		report, err = compiler.CompileSection(entryPath, sectionQuery)
	} else {
		report, err = compiler.CompileWithReport(entryPath)
	}

	if err != nil {
		return err
	}

	fmt.Print(report.Markdown)

	// Problems go to stderr, so the Markdown on stdout stays clean.
	for _, w := range report.Warnings {
		fmt.Fprintf(os.Stderr, "asciidoctor: %s\n", w)
	}
	for _, r := range report.BrokenRefs {
		fmt.Fprintf(os.Stderr, "broken cross-reference: <<%s>>\n", r)
	}
	return nil
}

func runValidate() error {
	// Parse flags and optional path argument
	quick := false
	opts := validator.ValidationOptions{}
	dir := "."

	for _, arg := range os.Args[2:] {
		switch arg {
		case "--quick", "-q":
			quick = true
		case "--yes", "-y":
			opts.SkipConfirm = true
		case "--ultra", "-u":
			opts.Ultra = true
		case "--json":
			opts.JSON = true
		default:
			if !strings.HasPrefix(arg, "-") {
				dir = arg
			}
		}
	}

	entryPath, err := config.FindEntry(dir)
	if err != nil {
		return err
	}

	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	opts.Checklist = cfg.ChecklistPath(dir)

	if quick {
		result, err := validator.ValidateQuick(entryPath)
		if err != nil {
			return err
		}
		if opts.JSON {
			fmt.Println(validator.FormatStructuralChecksJSON(result.StructuralChecks))
		} else {
			fmt.Print(validator.FormatStructuralChecks(result.StructuralChecks))
		}
		if !result.StructuralPassed {
			os.Exit(1)
		}
		return nil
	}

	// Full validation: structural + Claude
	result, err := validator.Validate(entryPath, os.Stdout, opts)
	if err != nil {
		return err
	}

	if !result.StructuralPassed || result.Cancelled {
		os.Exit(1)
	}

	return nil
}

func runDiff() error {
	entryPath, err := config.FindEntry(".")
	if err != nil {
		return err
	}

	// Get target commit (default: HEAD~1)
	targetCommit := "HEAD~1"
	if len(os.Args) > 2 {
		targetCommit = os.Args[2]
	}

	result, err := differ.DiffCompiled(entryPath, targetCommit)
	if err != nil {
		return err
	}

	fmt.Println(differ.FormatDiffResult(result))
	return nil
}

func runImpact() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: cca impact <attribute-name>")
	}

	attrName := os.Args[2]

	entryPath, err := config.FindEntry(".")
	if err != nil {
		return err
	}

	result, err := impact.AnalyzeAttribute(entryPath, attrName)
	if err != nil {
		return err
	}

	baseDir := filepath.Dir(entryPath)
	fmt.Println(impact.FormatImpact(result, baseDir))
	return nil
}

func runList() error {
	entryPath, err := config.FindEntry(".")
	if err != nil {
		return err
	}

	sections, err := compiler.ListSections(entryPath)
	if err != nil {
		return err
	}

	fmt.Print("Sections:\n\n")
	fmt.Print(compiler.FormatSectionList(sections))
	return nil
}

func runSkill() error {
	// Parse flags
	global := false
	for _, arg := range os.Args[2:] {
		if arg == "--global" || arg == "-g" {
			global = true
		}
	}

	// Determine target directory
	var skillDir string
	var suffix string
	if global {
		var err error
		skillDir, err = skill.GetGlobalSkillDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		suffix = " (global)"
	} else {
		skillDir = skill.GetProjectSkillDir()
		suffix = ""
	}

	// Check if already installed
	if skill.IsInstalled(skillDir) {
		installed, _ := skill.GetInstalledContent(skillDir)
		if installed == skill.GetEmbeddedContent() {
			fmt.Printf("Skill up to date%s\n", suffix)
			return nil
		}
		if err := skill.Install(skillDir); err != nil {
			return err
		}
		fmt.Printf("Skill updated%s\n", suffix)
		return nil
	}

	// Fresh install
	if err := skill.Install(skillDir); err != nil {
		return err
	}
	fmt.Printf("Skill installed%s\n", suffix)
	return nil
}

func runCompletion() {
	shell := "bash"
	if len(os.Args) > 2 {
		shell = os.Args[2]
	}

	switch shell {
	case "bash":
		fmt.Print(completion.Bash())
	case "zsh":
		fmt.Print(completion.Zsh())
	case "fish":
		fmt.Print(completion.Fish())
	default:
		fmt.Fprintf(os.Stderr, "Unknown shell: %s (supported: bash, zsh, fish)\n", shell)
		os.Exit(1)
	}
}
