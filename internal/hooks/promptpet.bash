# promptpet hook for bash: eval "$(promptpet init bash)"
# broadcasts command exit status and wall time to promptpet for companion event streaming
_promptpet_precmd() {
  local __pp_exit=$? __pp_dur=
  if [ -n "${_promptpet_start-}" ] && [ -n "${EPOCHREALTIME-}" ]; then
    __pp_dur="--duration-ms=$(( (${EPOCHREALTIME//[!0-9]/} - _promptpet_start) / 1000 ))"
  fi
  unset _promptpet_start  # empty Enter expands no PS0: report no duration
  PROMPTPET_SEGMENT="$(command promptpet prompt --exit "$__pp_exit" $__pp_dur 2>/dev/null)"
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
# Start timer: PS0 is expanded just before each command runs (bash >= 4.4).
# The arithmetic subscript stores $EPOCHREALTIME (bash >= 5) as integer
# microseconds and expands to nothing; older bash simply reports no duration.
if [ -n "${EPOCHREALTIME-}" ]; then
  case "${PS0-}" in
    *_promptpet_start*) ;;
    *) PS0='${_promptpet_none[_promptpet_start=${EPOCHREALTIME//[!0-9]/}]-}'"${PS0-}" ;;
  esac
fi
