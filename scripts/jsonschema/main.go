package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/invopop/jsonschema"
	"github.com/SouhlInc/tbls/schema"
)

// Generate JSON Schema of schema.json.
func main() {
	if err := _main(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func _main() error {
	r := new(jsonschema.Reflector)
	r.Namer = func(t reflect.Type) string {
		return strings.TrimSuffix(t.Name(), "JSON")
	}
	r.KeyNamer = strcase.ToSnake
	s := r.Reflect(&schema.SchemaJSON{})
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	fmt.Print(string(b))
	return nil
}
