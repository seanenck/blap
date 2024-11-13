_{{ $.Executable }}() {
  local cur opts chosen sub
  cur=${COMP_WORDS[COMP_CWORD]}
  if [ "$COMP_CWORD" -eq 1 ]; then
    opts="{{ $.Command.Upgrade }} {{ $.Command.Purge }}"
  else
    chosen=${COMP_WORDS[1]}
    case "$COMP_CWORD" in
      2)
        case "$chosen" in
          "{{ $.Command.Upgrade }}")
            opts="{{ $.Params.Upgrade.Main }}"
            ;;
          "{{ $.Command.Purge }}")
            opts="{{ $.Params.Purge.Main }}"
            ;;
        esac
      ;;
      3 | 4)
        case "$chosen" in
          "{{ $.Command.Purge }}")
            if [ "$COMP_CWORD" -eq 3 ]; then
              chosen=${COMP_WORDS[2]}
              {{ $.Params.Purge.Sub }}
            fi
            ;;
          "{{ $.Command.Upgrade }}") 
            chosen=${COMP_WORDS[2]}
            if [ "$COMP_CWORD" -eq 4 ]; then
              sub=${COMP_WORDS[3]}
            fi
            {{ $.Params.Upgrade.Sub }}
            ;;
        esac
      ;;
    esac
  fi
  if [ -n "$opts" ]; then
    # shellcheck disable=SC2207
    COMPREPLY=($(compgen -W "$opts" -- "$cur"))
  fi
}

complete -F _{{ $.Executable }} -o bashdefault {{ $.Executable }}
