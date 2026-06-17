package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// captureStdout runs fn and returns whatever it wrote to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRunCD_PreviousDir(t *testing.T) {
	out := captureStdout(t, func() {
		if err := runCD(cdCmd, []string{"-"}); err != nil {
			t.Errorf("runCD(-): %v", err)
		}
	})
	if out != "cd -\n" {
		t.Errorf("runCD(-) stdout = %q, want %q", out, "cd -\n")
	}
}

func TestRunWtCd_PreviousDir(t *testing.T) {
	out := captureStdout(t, func() {
		if err := runWtCd(wtCdCmd, []string{"-"}); err != nil {
			t.Errorf("runWtCd(-): %v", err)
		}
	})
	if out != "cd -\n" {
		t.Errorf("runWtCd(-) stdout = %q, want %q", out, "cd -\n")
	}
}
