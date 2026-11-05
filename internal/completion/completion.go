package completion

// Bash returns bash completion script
func Bash() string {
	return `_cca() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="compile validate diff impact list skill version help completion"

    case "${prev}" in
        cca)
            COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
            return 0
            ;;
        validate)
            COMPREPLY=( $(compgen -W "--json" -- ${cur}) )
            return 0
            ;;
        compile)
            COMPREPLY=( $(compgen -W "--section" -- ${cur}) )
            return 0
            ;;
        skill)
            COMPREPLY=( $(compgen -W "--global -g" -- ${cur}) )
            return 0
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- ${cur}) )
            return 0
            ;;
    esac

    COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
}
complete -F _cca cca
`
}

// Zsh returns zsh completion script
func Zsh() string {
	return `#compdef cca

_cca() {
    local -a commands
    commands=(
        'compile:Compile the docs to Markdown'
        'validate:Check the docs for breakage'
        'diff:Diff compiled output'
        'impact:Show attribute impact'
        'list:List sections'
        'skill:Install Claude Code skill'
        'version:Show version'
        'help:Show help'
        'completion:Generate shell completion'
    )

    _arguments -C \
        '1: :->command' \
        '*: :->args'

    case $state in
        command)
            _describe 'command' commands
            ;;
        args)
            case $words[2] in
                validate)
                    _arguments '--json[Print the checks as JSON]'
                    ;;
                compile)
                    _arguments '--section[Compile one section or file]:section:'
                    ;;
                skill)
                    _arguments '--global[Install globally]' '-g[Install globally]'
                    ;;
                completion)
                    _arguments '1:shell:(bash zsh fish)'
                    ;;
            esac
            ;;
    esac
}

_cca "$@"
`
}

// Fish returns fish completion script
func Fish() string {
	return `complete -c cca -f

complete -c cca -n '__fish_use_subcommand' -a compile -d 'Compile the docs to Markdown'
complete -c cca -n '__fish_use_subcommand' -a validate -d 'Check the docs for breakage'
complete -c cca -n '__fish_use_subcommand' -a diff -d 'Diff compiled output'
complete -c cca -n '__fish_use_subcommand' -a impact -d 'Show attribute impact'
complete -c cca -n '__fish_use_subcommand' -a list -d 'List sections'
complete -c cca -n '__fish_use_subcommand' -a skill -d 'Install Claude Code skill'
complete -c cca -n '__fish_use_subcommand' -a version -d 'Show version'
complete -c cca -n '__fish_use_subcommand' -a help -d 'Show help'
complete -c cca -n '__fish_use_subcommand' -a completion -d 'Generate shell completion'

complete -c cca -n '__fish_seen_subcommand_from validate' -l json -d 'Print the checks as JSON'

complete -c cca -n '__fish_seen_subcommand_from compile' -l section -d 'Compile one section or file'

complete -c cca -n '__fish_seen_subcommand_from skill' -l global -s g -d 'Install globally'

complete -c cca -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'
`
}
