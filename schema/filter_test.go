package schema

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSchema_SeparateFunctionsThatAreIncludedOrNot(t *testing.T) {
	tests := []struct {
		name         string
		schema       *Schema
		opt          *FilterOption
		wantIncludes []string
		wantExcludes []string
	}{
		{
			name: "include specific functions",
			schema: &Schema{
				Functions: []*Function{
					{Name: "func1"},
					{Name: "func2"},
					{Name: "other_func"},
				},
			},
			opt: &FilterOption{
				Include: []string{"func*"},
			},
			wantIncludes: []string{"func1", "func2"},
			wantExcludes: []string{"other_func"},
		},
		{
			name: "exclude specific functions",
			schema: &Schema{
				Functions: []*Function{
					{Name: "func1"},
					{Name: "func2"},
					{Name: "other_func"},
				},
			},
			opt: &FilterOption{
				Exclude: []string{"other_*"},
			},
			wantIncludes: []string{"func1", "func2"},
			wantExcludes: []string{"other_func"},
		},
		{
			name: "include and exclude with more specific include",
			schema: &Schema{
				Functions: []*Function{
					{Name: "public.func1"},
					{Name: "public.func2"},
					{Name: "private.func1"},
				},
			},
			opt: &FilterOption{
				Include: []string{"public.func1"},
				Exclude: []string{"public.*"},
			},
			wantIncludes: []string{"public.func1"},
			wantExcludes: []string{"public.func2", "private.func1"},
		},
		{
			name: "include all when no filters",
			schema: &Schema{
				Functions: []*Function{
					{Name: "func1"},
					{Name: "func2"},
				},
			},
			opt:          &FilterOption{},
			wantIncludes: []string{"func1", "func2"},
			wantExcludes: []string{},
		},
		{
			name: "include all with schema prefix",
			schema: &Schema{
				Functions: []*Function{
					{Name: "public.func1"},
					{Name: "public.func2"},
					{Name: "private.func1"},
				},
			},
			opt: &FilterOption{
				Include: []string{"public.*"},
			},
			wantIncludes: []string{"public.func1", "public.func2"},
			wantExcludes: []string{"private.func1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			includes, excludes, err := tt.schema.SeparateFunctionsThatAreIncludedOrNot(tt.opt)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			gotIncludeNames := getFunctionNames(includes)
			gotExcludeNames := getFunctionNames(excludes)

			if diff := cmp.Diff(tt.wantIncludes, gotIncludeNames); diff != "" {
				t.Errorf("includes mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantExcludes, gotExcludeNames); diff != "" {
				t.Errorf("excludes mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func getFunctionNames(functions []*Function) []string {
	names := make([]string, len(functions))
	for i, f := range functions {
		names[i] = f.Name
	}
	return names
}
