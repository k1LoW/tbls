package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// -h and --help are advertised in the usage text, so they have to print help and
// succeed rather than be reported as unknown flags.
func TestRootRunEHandlesHelpFlags(t *testing.T) {
	for _, flag := range []string{"-h", "--help"} {
		t.Run(flag, func(t *testing.T) {
			buf := &bytes.Buffer{}
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			t.Cleanup(func() {
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
			})
			if err := rootCmd.RunE(rootCmd, []string{flag}); err != nil {
				t.Errorf("got %v, want nil", err)
			}
			if !strings.Contains(buf.String(), "Usage:") {
				t.Errorf("help was not printed, got %q", buf.String())
			}
		})
	}
}
