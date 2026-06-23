package generator

import (
	"bytes"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

func TestHasGenerateReset(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		want bool
	}{
		{"nil", "", false},
		{"missing", "// foo", false},
		{"present", "// generate:reset", true},
		{"among others", "// foo\n// generate:reset\n// bar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cg *ast.CommentGroup

			if tt.doc != "" {
				var list []*ast.Comment
				for _, s := range strings.Split(tt.doc, "\n") {
					list = append(list, &ast.Comment{Text: s})
				}
				cg = &ast.CommentGroup{List: list}
			}

			if got := hasGenerateReset(cg); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestHasReset(t *testing.T) {
	pkg := types.NewPackage("example", "example")

	named := types.NewNamed(
		types.NewTypeName(token.NoPos, pkg, "Foo", nil),
		types.NewStruct(nil, nil),
		nil,
	)

	if hasReset(named) {
		t.Fatal("unexpected Reset")
	}

	sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)

	method := types.NewFunc(
		token.NoPos,
		pkg,
		"Reset",
		sig,
	)

	named.AddMethod(method)

	if !hasReset(named) {
		t.Fatal("expected Reset")
	}
}

func TestShouldImport(t *testing.T) {
	local := types.NewPackage("example", "example")
	foreign := types.NewPackage("other/pkg", "other")

	tests := []struct {
		name string
		pkg  *types.Package
		want bool
	}{
		{"nil", nil, false},
		{"local", local, false},
		{"foreign", foreign, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var named *types.Named

			if tt.pkg != nil {
				named = types.NewNamed(
					types.NewTypeName(token.NoPos, tt.pkg, "T", nil),
					types.NewStruct(nil, nil),
					nil,
				)
			}

			got := shouldImport(named, "example")

			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestEmitReset(t *testing.T) {
	tests := []struct {
		name string
		typ  types.Type
		want string
	}{
		{
			"string",
			types.Typ[types.String],
			`s.Name = ""`,
		},
		{
			"bool",
			types.Typ[types.Bool],
			`s.Name = false`,
		},
		{
			"int",
			types.Typ[types.Int],
			`s.Name = 0`,
		},
		{
			"slice",
			types.NewSlice(types.Typ[types.Int]),
			`s.Name = s.Name[:0]`,
		},
		{
			"map",
			types.NewMap(types.Typ[types.String], types.Typ[types.Int]),
			`clear(s.Name)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var buf bytes.Buffer

			emitReset(
				&buf,
				"s.Name",
				tt.typ,
				map[string]string{},
				"example",
			)

			if !strings.Contains(buf.String(), tt.want) {
				t.Fatalf("got:\n%s\nwant:\n%s", buf.String(), tt.want)
			}
		})
	}
}

func TestEmitResetPointer(t *testing.T) {
	st := types.NewStruct(nil, nil)

	named := types.NewNamed(
		types.NewTypeName(token.NoPos, types.NewPackage("example", "example"), "User", nil),
		st,
		nil,
	)

	var buf bytes.Buffer

	emitReset(
		&buf,
		"s.User",
		types.NewPointer(named),
		map[string]string{},
		"example",
	)

	code := buf.String()

	if !strings.Contains(code, "if s.User != nil") {
		t.Fatal(code)
	}

	if !strings.Contains(code, "*s.User = User{}") {
		t.Fatal(code)
	}
}
