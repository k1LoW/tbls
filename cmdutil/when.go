package cmdutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/k1LoW/errors"
)

// AST walker which replaces `$IDENTIFIER` with `Env.IDENTIFIER` member lookup expressions.
type EnvPatcher struct{}

func (EnvPatcher) Visit(node *ast.Node) {
	if n, ok := (*node).(*ast.IdentifierNode); ok && n.Value[0] == '$' && n.Value != "$env" {
		ast.Patch(
			node,
			&ast.MemberNode{
				Node:     &ast.IdentifierNode{Value: "Env"},
				Property: &ast.StringNode{Value: n.Value[1:]},
			},
		)
	}
}

// The predefined variables of a when expression.
type WhenEnv struct {
	Env map[string]string
}

var NewWhenEnv = func() *WhenEnv {
	return &WhenEnv{Env: envMap()}
}

func IsAllowedToExecute(when string) (bool, error) {
	if when == "" {
		return true, nil
	}

	whenEnv := NewWhenEnv()
	// when expressions must produce a boolean result
	program, err := expr.Compile(when, expr.Patch(&EnvPatcher{}), expr.AsBool(), expr.Env(whenEnv))
	if err != nil {
		return false, errors.WithStack(err)
	}
	result, err := expr.Run(program, whenEnv)
	if err != nil {
		return false, errors.WithStack(err)
	}
	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expected bool, but got %T", result)
	}
	return boolResult, nil
}

func envMap() map[string]string {
	m := map[string]string{}
	for _, kv := range os.Environ() {
		if k, v, ok := strings.Cut(kv, "="); ok {
			m[k] = v
		}
	}
	return m
}
