case "$chosen" in
  "{{ $.Arg.Applications }}" | "{{ $.Arg.Disable }}")
    if [ -n "$sub" ]; then
      case "$sub" in
        "{{ $.Arg.Include }}")
          opts="{{ $.Arg.Confirm }}"
        ;;
        "{{ $.Arg.Confirm }}")
          opts="{{ $.Arg.Include }}"
        ;;
      esac
    else
      opts="{{ $.Arg.Confirm }} {{ $.Arg.Include }}"
    fi
  ;;
  "{{ $.Arg.Confirm }}")
    if [ -n "$sub" ]; then
      case "$sub" in
        "{{ $.Arg.Include }}")
          opts="{{ $.Arg.Applications }} {{ $.Arg.Disable }}"
        ;;
        "{{ $.Arg.Applications }}" | "{{ $.Arg.Disable }}")
          opts="{{ $.Arg.Include }}"
        ;;
      esac
    else
      opts="{{ $.Arg.Applications }} {{ $.Arg.Disable }} {{ $.Arg.Include }}"
    fi
  ;;
  "{{ $.Arg.Include }}")
    if [ -n "$sub" ]; then
      case "$sub" in
        "{{ $.Arg.Confirm }}")
          opts="{{ $.Arg.Applications }} {{ $.Arg.Disable }}"
        ;;
        "{{ $.Arg.Applications }}" | "{{ $.Arg.Disable }}")
          opts="{{ $.Arg.Confirm }}"
        ;;
      esac
    else
      opts="{{ $.Arg.Applications }} {{ $.Arg.Disable }} {{ $.Arg.Confirm }}"
    fi
  ;;
esac
