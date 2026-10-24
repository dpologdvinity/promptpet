#compdef promptpet

_promptpet() {
  local -a commands shells species pets
  commands=(
    'init:print the shell hook'
    'pet:select or list pets'
    'play:play with your pet'
    'name:rename your pet'
    'species:change species'
    'prompt:print the prompt segment'
    'completion:generate shell autocompletions'
    'gui:launch companion mini-window'
    'games:list mini-games and high scores'
    '--version:print version'
    '--help:show help'
  )
  shells=(zsh bash fish)
  species=(cat dragon blob)
  pets=(usagi kochi puff mochi spook pippin gromit boba)

  _arguments -C \
    '1: :->command' \
    '2: :->arg' \
    '*:: :->args'

  case $state in
    command)
      _describe -t commands 'command' commands
      ;;
    arg)
      case $words[2] in
        init|completion)
          _values 'shell' $shells
          ;;
        species)
          _values 'species' $species
          ;;
        pet)
          _values 'pet' $pets
          ;;
        prompt)
          _arguments '--exit[exit code of last command]:exit code:'
          ;;
      esac
      ;;
  esac
}

if type compdef >/dev/null 2>&1; then
  compdef _promptpet promptpet
fi
