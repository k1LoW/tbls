package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// An unrecognized flag or command has to be reported as an error, so that tbls exits
// non-zero like it does for every other failure.
func TestRootRunEReportsUnknownFlagAndCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"unknown flag", []string{"--nosuchflag"}},
		{"unknown command", []string{"nosuchcommand"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// RunE looks for a tbls-<subcmd> plugin on PATH and runs it when one exists, so
			// pin PATH to an empty directory to keep the lookup off the host.
			t.Setenv("PATH", t.TempDir())
			buf := &bytes.Buffer{}
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			t.Cleanup(func() {
				rootCmd.SetOut(nil)
				rootCmd.SetErr(nil)
			})
			if err := rootCmd.RunE(rootCmd, tt.args); err == nil {
				t.Error("got nil, want an error so that the exit status is not zero")
			}
		})
	}
}

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
