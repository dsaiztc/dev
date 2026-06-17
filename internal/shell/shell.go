package shell

// WrapperFunc returns the shell function that wraps the dev binary.
// The function evals stdout from cd and clone commands so they can
// affect the parent shell (e.g., change directory).
func WrapperFunc() string {
	return `dev() {
  # Pass through directly when help or version flags are present.
  local arg
  for arg in "$@"; do
    if [[ "$arg" == "--help" || "$arg" == "-h" || "$arg" == "--version" ]]; then
      command dev "$@"
      return $?
    fi
  done
  if [[ "$1" == "cd" || "$1" == "clone" || "$1" == "new" || ( ( "$1" == "wt" || "$1" == "wkt" ) && ( -z "$2" || "$2" == "-" || "$2" =~ ^(cd|new|rm)$ ) ) ]]; then
    local output
    output="$(command dev "$@")"
    local exit_code=$?
    if [[ $exit_code -eq 0 && -n "$output" ]]; then
      eval "$output"
    fi
    return $exit_code
  else
    command dev "$@"
  fi
}`
}
