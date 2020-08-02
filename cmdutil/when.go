package cmdutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/antonmedv/expr"
)

func IsAllowedToExecute(when string) (bool, error) {
	if when == "" {
		return true, nil
	}
	ropts := []string{}
	em := envMap()
	for k := range em {
		ropts = append(ropts, fmt.Sprintf("$%s", k), fmt.Sprintf("Env.%s", k))
	}
	r := strings.NewReplacer(ropts...)
	when = r.Replace(when)
	got, err := expr.Eval(fmt.Sprintf("(%s) == true", when), struct {
		Env map[string]string
	}{
		Env: em,
	})
	if err != nil {
		return false, err
	}
	return got.(bool), nil
}

func envMap() map[string]string {
	m := map[string]string{}
	for _, kv := range os.Environ() {
		if !strings.Contains(kv, "=") {
			continue
		}
		parts := strings.SplitN(kv, "=", 2)
		k := parts[0]
		if len(parts) < 2 {
			m[k] = ""
			continue
		}
		m[k] = parts[1]
	}
	return m
}
