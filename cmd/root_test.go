package cmd

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestRootRunEPassesConfigToSubcommandWithoutSchema(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "env")
	bin := filepath.Join(dir, "tbls-envdump")
	script := "#!/bin/sh\nenv > \"$ENVDUMP_OUT\"\n"
	if err := os.WriteFile(bin, []byte(script), 0o700); err != nil { //nolint:gosec
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, ".tbls.yml")
	dsn := "json://" + filepath.Join(dir, "missing.json")
	if err := os.WriteFile(cfgPath, []byte("dsn: "+dsn+"\ndocPath: "+filepath.Join(dir, "doc")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("ENVDUMP_OUT", out)
	// A TBLS_DSN inherited from the host would make cfg.Load skip the config file, and an
	// inherited TBLS_* would reach the subcommand even if tbls did not set it.
	for _, k := range []string{"TBLS_DSN", "TBLS_CONFIG_PATH", "TBLS_SCHEMA"} {
		t.Setenv(k, "")
		os.Unsetenv(k) //nolint:errcheck
	}
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	if err := rootCmd.RunE(rootCmd, []string{"-c", cfgPath, "envdump"}); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, l := range strings.Split(string(b), "\n") {
		if k, v, ok := strings.Cut(l, "="); ok && strings.HasPrefix(k, "TBLS_") {
			got[k] = v
		}
	}
	if got["TBLS_DSN"] != dsn {
		t.Errorf("TBLS_DSN: got %q, want %q", got["TBLS_DSN"], dsn)
	}
	if got["TBLS_CONFIG_PATH"] != cfgPath {
		t.Errorf("TBLS_CONFIG_PATH: got %q, want %q", got["TBLS_CONFIG_PATH"], cfgPath)
	}
	if v, ok := got["TBLS_SCHEMA"]; ok {
		t.Errorf("TBLS_SCHEMA: got %q, want it unset", v)
	}
	if !strings.Contains(buf.String(), "the schema could not be loaded") {
		t.Errorf("the reason was not reported, got %q", buf.String())
	}
}
