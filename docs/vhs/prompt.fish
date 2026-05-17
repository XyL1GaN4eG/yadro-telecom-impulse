function fish_prompt
    set -l last_status $status
    set -l duration "$CMD_DURATION$cmd_duration"

    starship prompt \
        --terminal-width="$COLUMNS" \
        --status=$last_status \
        --cmd-duration=$duration
end

function fish_right_prompt
end

function fish_mode_prompt
end
