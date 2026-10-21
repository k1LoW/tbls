package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/k1LoW/tbls/schema"
)

// Run each invocation in a separate process because diff exits with status 1
// when it finds a difference, and the commands share package-level flag values.
func TestDiffFilterProcess(_ *testing.T) {
	if os.Getenv("TBLS_TEST_DIFF_FILTER") != "1" {
		return
	}
	rootCmd.SetArgs(os.Args[slices.Index(os.Args, "--")+1:])
	Execute()
	os.Exit(0)
}

func TestDiffTableFilters(t *testing.T) {
	for _, flag := range []string{"--include", "--exclude"} {
		t.Run(flag, func(t *testing.T) {
			dir := t.TempDir()
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			docs := filepath.Join(dir, "docs")
			writeSchema := func(name, keptComment, omittedComment string) string {
				t.Helper()
				s := &schema.Schema{Name: "test", Tables: []*schema.Table{
					{Name: "kept", Type: "BASE TABLE", Comment: keptComment},
					{Name: "omitted", Type: "BASE TABLE", Comment: omittedComment},
				}}
				data, err := json.Marshal(s)
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(dir, name+".json")
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
				return "json://" + filepath.ToSlash(path)
			}
			a := writeSchema("a", "kept table", "old omitted table")
			b := writeSchema("b", "kept table", "changed omitted table")
			changed := writeSchema("changed", "changed kept table", "old omitted table")
			cfg := filepath.Join(dir, "config.yml")
			if err := os.WriteFile(cfg, []byte("dsn: "+a+"\ndocPath: "+filepath.ToSlash(docs)+"\ner:\n  skip: true\n"), 0600); err != nil {
				t.Fatal(err)
			}
			pattern := "kept"
			if flag == "--exclude" {
				pattern = "omitted"
			}
			run := func(wantCode int, args ...string) string {
				t.Helper()
				args = append(args, "--config", cfg, flag, pattern)
				// #nosec G204 -- Runs only this test binary with local fixture arguments.
				cmd := exec.Command(binary, append([]string{"-test.run=^TestDiffFilterProcess$", "--"}, args...)...)
				cmd.Env = append(os.Environ(), "TBLS_TEST_DIFF_FILTER=1")
				cmd.Dir = dir
				output, err := cmd.CombinedOutput()
				code := 0
				if err != nil {
					var exitErr *exec.ExitError
					if errors.As(err, &exitErr) {
						code = exitErr.ExitCode()
					} else {
						t.Fatal(err)
					}
				}
				if code != wantCode {
					t.Fatalf("%v: exit %d, want %d\n%s", args, code, wantCode, output)
				}
				return string(output)
			}
			run(0, "doc", a, docs)
			if _, err := os.Stat(filepath.Join(docs, "omitted.md")); !os.IsNotExist(err) {
				t.Fatalf("omitted table was documented: %v", err)
			}
			for _, args := range [][]string{
				{"diff"},
				{"diff", docs},
				{"diff", b},
				{"diff", a, docs},
				{"diff", a, b},
			} {
				if output := run(0, args...); output != "" {
					t.Errorf("%v: unexpected diff: %s", args, output)
				}
			}
			if output := run(1, "diff", a, changed); !strings.Contains(output, "changed kept table") {
				t.Errorf("change to included table was not reported: %s", output)
			}
		})
	}
}
