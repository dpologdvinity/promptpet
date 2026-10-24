# bash completion for promptpet
_promptpet_completions() {
  local cur prev commands shells species
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  commands="init pet play name species prompt completion --version --help"
  shells="zsh bash fish"
  species="cat dragon blob"

  case "$prev" in
    init|completion)
      COMPREPLY=( $(compgen -W "$shells" -- "$cur") )
      return 0
      ;;
    species)
      COMPREPLY=( $(compgen -W "$species" -- "$cur") )
      return 0
      ;;
    prompt)
      COMPREPLY=( $(compgen -W "--exit" -- "$cur") )
      return 0
      ;;
  esac

  if [ "$COMP_CWORD" -eq 1 ]; then
    COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
    return 0
  fi
}
complete -F _promptpet_completions promptpet
