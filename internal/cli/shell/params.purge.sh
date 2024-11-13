case "$chosen" in
  "{{ $.Arg.CleanDirs }}")
    opts="{{ $.Arg.Confirm }}"
  ;;
  "{{ $.Arg.Confirm }}")
    opts="{{ $.Arg.CleanDirs }}"
  ;;
esac
