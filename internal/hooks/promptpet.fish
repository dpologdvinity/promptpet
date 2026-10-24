# promptpet hook for fish: promptpet init fish | source
# broadcasts command exit status to promptpet for companion event streaming
if not functions -q _promptpet_orig_prompt
    functions -q fish_prompt; and functions -c fish_prompt _promptpet_orig_prompt
    function _promptpet_status
        return $argv[1]
    end
    function fish_prompt
        set -l __pp_exit $status
        set -l __pp_seg (command promptpet prompt --exit $__pp_exit 2>/dev/null)
        test -n "$__pp_seg"; and printf '%s ' $__pp_seg
        if functions -q _promptpet_orig_prompt
            _promptpet_status $__pp_exit  # restore $status for the original prompt
            _promptpet_orig_prompt
        end
    end
end
