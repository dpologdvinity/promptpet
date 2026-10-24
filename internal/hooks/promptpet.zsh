# promptpet hook for zsh: eval "$(promptpet init zsh)"
_promptpet_precmd() {
  local __pp_exit=$?
  PROMPTPET_SEGMENT="$(command promptpet prompt --exit "$__pp_exit" 2>/dev/null)"
  return $__pp_exit
}
autoload -Uz add-zsh-hook
add-zsh-hook precmd _promptpet_precmd
setopt PROMPT_SUBST
case "$PROMPT" in
  *PROMPTPET_SEGMENT*) ;;
  *) PROMPT='${PROMPTPET_SEGMENT:+$PROMPTPET_SEGMENT }'"$PROMPT" ;;
esac
