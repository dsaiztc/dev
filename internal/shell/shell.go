package shell

// WrapperFunc returns the shell function that wraps the dev binary.
// The function evals stdout from cd and clone commands so they can
// affect the parent shell (e.g., change directory).
func WrapperFunc() string {
	return `dev() {
  if [[ "$1" == "cd" || "$1" == "clone" ]]; then
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
