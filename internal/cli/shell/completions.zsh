#compdef _{{ $.Executable }} {{ $.Executable }}

_{{ $.Executable }}() {
  local curcontext="$curcontext" state len chosen args sub opts
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
          case "$len" in
            3)
            opts="{{ $.Params.Upgrade.Main }}"
            ;;
            4 | 5)
            chosen=$words[3]
            if [ "$len" -eq 5 ]; then
              sub=$words[4]
              if [ "$sub" = "--" ]; then
                sub=""
              fi
            fi
            {{ $.Params.Upgrade.Sub }}
          esac
          ;;
        "{{ $.Command.Purge }}")
          case "$len" in
            3)
            opts="{{ $.Params.Purge.Main }}"
            ;;
            4)
            chosen=$words[3]
            {{ $.Params.Purge.Sub }}
            ;;
          esac
          ;;
        *)
          return
          ;;
      esac
  esac
  if [ -n "$opts" ]; then
    for item in $(echo "$opts" | tr ' ' '\n'); do
      compadd -- "$item"
    done
  fi
}

compdef _{{ $.Executable }} {{ $.Executable }}
