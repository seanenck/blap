#compdef _{{ $.Executable }} {{ $.Executable }}

_{{ $.Executable }}() {
  local curcontext="$curcontext" state len chosen args sub opts subset matched
  typeset -A opt_args

  _arguments \
    '1: :->main'\
    '*: :->args'

  len=${#words[@]}
  opts=""
  case $state in
    main)
      args="{{ $.Command.Upgrade }} {{ $.Command.Purge }} {{ $.Command.List }}"
      _arguments "1:main:($args)"
    ;;
    *)
      if [ "$len" -lt 2 ]; then
        return
      fi
      chosen=$words[2]
      case "$chosen" in
        "{{ $.Command.Upgrade }}")
            opts=({{ $.Params.Upgrade }})
            ;;
        "{{ $.Command.Purge }}")
            opts=({{ $.Params.Purge }})
            ;;
        "{{ $.Command.List }}")
            opts=({{ $.Params.List }})
            ;;
      esac
      subset=""
      for sub in "${opts[@]}"; do
        matched=0
        for chosen in "${words[@]}"; do
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
  esac
  if [ -n "$opts" ]; then
    for item in $(echo "$opts" | tr ' ' '\n'); do
      compadd -- "$item"
    done
  fi
}

compdef _{{ $.Executable }} {{ $.Executable }}
