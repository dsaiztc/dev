package fuzzy

import (
	"strings"
	"testing"
)

func TestDisableColor(t *testing.T) {
	// Before disabling, styles may or may not emit ANSI depending on the test
	// terminal — we just verify DisableColor makes them not emit escapes.
	DisableColor()
	out := selectedStyle.Render("hello")
	if strings.Contains(out, "\x1b") {
		t.Errorf("DisableColor: output still contains ANSI escape: %q", out)
	}
	if out != "hello" {
		t.Errorf("DisableColor: output = %q, want %q", out, "hello")
	}
}
