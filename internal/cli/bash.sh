_{{ $.Executable }}() {
  local cur opts chosen
  cur=${COMP_WORDS[COMP_CWORD]}
  if [ "$COMP_CWORD" -eq 1 ]; then
    opts="{{ $.Command.Upgrade }} {{ $.Command.Purge }}"
  else
    chosen=${COMP_WORDS[1]}
    case "$COMP_CWORD" in
      2)
        opts="{{ $.Arg.Confirm }}"
        if [ "$chosen" = "{{ $.Command.Upgrade }}" ]; then 
          opts="$opts {{ $.Arg.Applications }} {{ $.Arg.Disable }}"
        fi
      ;;
      3)
        if [ "$chosen" = "{{ $.Command.Upgrade }}" ]; then 
          chosen=${COMP_WORDS[2]}
          case "$chosen" in
            "{{ $.Arg.Applications }}" | "{{ $.Arg.Disable }}")
              opts="{{ $.Arg.Confirm }}"
            ;;
            "{{ $.Arg.Confirm }}")
              opts="{{ $.Arg.Applications }} {{ $.Arg.Disable }}"
            ;;
          esac
        fi
      ;;
    esac
  fi
  if [ -n "$opts" ]; then
    # shellcheck disable=SC2207
    COMPREPLY=($(compgen -W "$opts" -- "$cur"))
  fi
}

complete -F _{{ $.Executable }} -o bashdefault {{ $.Executable }}
