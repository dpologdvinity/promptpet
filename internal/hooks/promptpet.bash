# promptpet hook for bash: eval "$(promptpet init bash)"
# broadcasts command exit status to promptpet for companion event streaming
_promptpet_precmd() {
  local __pp_exit=$?
  PROMPTPET_SEGMENT="$(command promptpet prompt --exit "$__pp_exit" 2>/dev/null)"
  return $__pp_exit
}
case ";${PROMPT_COMMAND:-};" in
  *";_promptpet_precmd;"*) ;;
  *) PROMPT_COMMAND="_promptpet_precmd${PROMPT_COMMAND:+;$PROMPT_COMMAND}" ;;
esac
case "$PS1" in
  *PROMPTPET_SEGMENT*) ;;
  *) PS1='${PROMPTPET_SEGMENT:+$PROMPTPET_SEGMENT }'"$PS1" ;;
esac
