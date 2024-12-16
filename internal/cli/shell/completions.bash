_{{ $.Executable }}() {
  local cur opts chosen sub subset matched
  cur=${COMP_WORDS[COMP_CWORD]}
  if [ "$COMP_CWORD" -eq 1 ]; then
    opts="{{ $.Command.Upgrade }} {{ $.Command.Purge }} {{ $.Command.List }}"
  else
    chosen=${COMP_WORDS[1]}
    subset=""
    case "$chosen" in
      "{{ $.Command.Purge }}")
        opts="{{ $.Params.Purge }}"
        ;;
      "{{ $.Command.Upgrade }}") 
        opts="{{ $.Params.Upgrade }}"
        ;;
      "{{ $.Command.List }}") 
        opts="{{ $.Params.List }}"
        ;;
    esac
    for sub in $opts; do
      matched=0 
      for chosen in "${COMP_WORDS[@]}"; do
        if [ "$chosen" = "$sub" ]; then
          matched=1
          break
        fi
      done
      if [ "$matched" -eq 0 ]; then
        subset="$subset $sub"
      fi
    done
    opts="$subset"
  fi
  if [ -n "$opts" ]; then
    # shellcheck disable=SC2207
    COMPREPLY=($(compgen -W "$opts" -- "$cur"))
  fi
}

complete -F _{{ $.Executable }} -o bashdefault {{ $.Executable }}
