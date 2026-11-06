# promptpet hook for zsh: eval "$(promptpet init zsh)"
# broadcasts command exit status and wall time to promptpet for companion event streaming
zmodload zsh/datetime 2>/dev/null  # $EPOCHREALTIME, no `date` fork
_promptpet_preexec() {
  _promptpet_start=$EPOCHREALTIME
}
_promptpet_precmd() {
  local __pp_exit=$? __pp_dur=
  if [[ -n $_promptpet_start && -n $EPOCHREALTIME ]]; then
    local -i __pp_ms
    (( __pp_ms = (EPOCHREALTIME - _promptpet_start) * 1000 ))
    __pp_dur="--duration-ms=$__pp_ms"
  fi
  unset _promptpet_start  # empty Enter runs no preexec: report no duration
  PROMPTPET_SEGMENT="$(command promptpet prompt --exit "$__pp_exit" $__pp_dur 2>/dev/null)"
  return $__pp_exit
}
autoload -Uz add-zsh-hook
add-zsh-hook preexec _promptpet_preexec
add-zsh-hook precmd _promptpet_precmd
setopt PROMPT_SUBST
case "$PROMPT" in
  *PROMPTPET_SEGMENT*) ;;
  *) PROMPT='${PROMPTPET_SEGMENT:+$PROMPTPET_SEGMENT }'"$PROMPT" ;;
esac
