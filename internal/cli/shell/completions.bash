_{{ $.Executable }}() {
  local cur opts chosen sub
  cur=${COMP_WORDS[COMP_CWORD]}
  if [ "$COMP_CWORD" -eq 1 ]; then
    opts="{{ $.Command.Upgrade }} {{ $.Command.Purge }}"
  else
    chosen=${COMP_WORDS[1]}
    case "$COMP_CWORD" in
      2)
        opts="{{ $.Arg.Confirm }}"
        case "$chosen" in
          "{{ $.Command.Upgrade }}")
            opts="$opts {{ $.Arg.Applications }} {{ $.Arg.Disable }} {{ $.Arg.Include }}"
            ;;
          "{{ $.Command.Purge }}")
            opts="$opts {{ $.Arg.CleanDirs }}"
            ;;
        esac
      ;;
      3 | 4)
        case "$chosen" in
          "{{ $.Command.Purge }}")
            if [ "$COMP_CWORD" -eq 3 ]; then
              chosen=${COMP_WORDS[2]}
              {{ $.Params.Purge }}
            fi
            ;;
          "{{ $.Command.Upgrade }}") 
            chosen=${COMP_WORDS[2]}
            if [ "$COMP_CWORD" -eq 4 ]; then
              sub=${COMP_WORDS[3]}
            fi
            {{ $.Params.Upgrade }}
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
