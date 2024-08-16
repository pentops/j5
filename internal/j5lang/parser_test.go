package j5lang

import (
	"strings"
	"testing"
)

func TestSimple(t *testing.T) {

}

func TestBasicAssign(t *testing.T) {
	input := `
package pentops.j5lang.example
version = "v1"
number = 123
// bool = true
// float = 1.23
`

	file, err := ParseFile(input)
	if err != nil {
		t.Fatal(err)
	}

	assertOptions(t, file.Decls, []KV{
		{"package", "pentops.j5lang.example"},
		{"version", "v1"},
		{"number", "123"},
	})
}

func TestEnumDecl(t *testing.T) {
	input := strings.Join([]string{
		/*  1 */ `enum Foo {`,
		/*  2 */ `  | This is a description of Foo`,
		/*  3 */ ``,
		/*  4 */ `  GOOD`,
		/*  5 */ `  BAD | Really Really Bad`,
		/*  6 */ `  UGLY {`,
		/*  7 */ `    | This is a description of UGLY`,
		/*  8 */ `  }`,
		/*  9 */ `}`,
	}, "\n")

	file, err := ParseFile(input)
	if err != nil {
		t.Fatal(err)
	}

	if len(file.Decls) != 1 {
		t.Fatalf("expected 1 decl in file, got %d", len(file.Decls))
	}

	declEnum, ok := file.Decls[0].(EnumDecl)
	if !ok {
		t.Fatalf("expected EnumDecl, got %T", file.Decls[0])
	}

	if declEnum.Name != "Foo" {
		t.Fatalf("expected Foo, got %s", declEnum.Name)
	}

	if declEnum.Description != "This is a description of Foo" {
		t.Fatalf("got description %q", declEnum.Description)
	}

	if len(declEnum.Options) != 3 {
		t.Fatalf("expected 3 values, got %d", len(declEnum.Options))
	}

	if declEnum.Options[0].Name != "GOOD" {
		t.Fatalf("expected GOOD, got %s", declEnum.Options[0])
	}

	if declEnum.Options[1].Name != "BAD" {
		t.Fatalf("expected BAD, got %s", declEnum.Options[1])
	}

	if declEnum.Options[2].Name != "UGLY" {
		t.Fatalf("expected UGLY, got %s", declEnum.Options[2])
	}

}

func TestObjectDecl(t *testing.T) {
	input := strings.Join([]string{
		/*  1 */ `object Foo {`,
		/*  2 */ `  | This is a description of Foo`,
		/*  3 */ `  | It has a bar field and a baz field`,
		/*  4 */ `  `,
		/*  5 */ `  property = "value"`,
		/*  6 */ `  `,
		/*  7 */ `  field bar_field string {`,
		/*  8 */ `    validate.required = true`,
		/*  9 */ `  }`,
		/* 10 */ `  `,
		/* 11 */ `  field baz_field object {`,
		/* 12 */ `    ref path.to.Type`,
		/* 13 */ `  }`,
		/* 14 */ `}`,
	}, "\n")

	file, err := ParseFile(input)
	if err != nil {
		t.Fatal(err)
	}

	if len(file.Decls) != 1 {
		t.Fatalf("expected 1 decl in file, got %d", len(file.Decls))
	}

	declObj, ok := file.Decls[0].(ObjectDecl)
	if !ok {
		t.Fatalf("expected ObjectDecl, got %T", file.Decls[0])
	}

	assertOptions(t, declObj.Decls, []KV{
		{"property", "value"},
	})

	if declObj.Name != "Foo" {
		t.Fatalf("expected Foo, got %s", declObj.Name)
	}
	if declObj.Description != "This is a description of Foo\nIt has a bar field and a baz field" {
		t.Fatalf("got description %q", declObj.Description)
	}

	if len(declObj.Decls) != 3 {
		t.Fatalf("expected 3 decls, got %d", len(declObj.Decls))
	}

	fieldBar, ok := declObj.Decls[1].(FieldDecl)
	if !ok {
		t.Fatalf("expected FieldDecl, got %T", declObj.Decls[1])
	}

	assertOptions(t, fieldBar.Decls, []KV{
		{"validate.required", "true"},
	})

	fieldBaz, ok := declObj.Decls[2].(FieldDecl)
	if !ok {
		t.Fatalf("expected FieldDecl, got %T", declObj.Decls[2])
	}

	assertOptions(t, fieldBaz.Decls, []KV{
		{"ref", "path.to.Type"},
	})

}

type KV struct {
	Key   string
	Value string
}

func filterOptions(decls []Decl) []KV {
	var opts []KV
	for _, decl := range decls {
		switch d := decl.(type) {
		case SpecialDecl:
			opts = append(opts, KV{d.Key.String(), d.Value.ToString()})
		case ValueAssign:
			opts = append(opts, KV{d.Key.ToString(), d.Value.Lit})
		}
	}
	return opts
}

func assertOptions(t *testing.T, decls []Decl, expected []KV) {
	opts := filterOptions(decls)
	for idx, opt := range opts {
		if idx >= len(expected) {
			t.Errorf("unexpected option %s = %q", opt.Key, opt.Value)
			continue
		}

		expectKey := expected[idx].Key
		if opt.Key != expectKey {
			t.Errorf("expected key %q, got %q", expectKey, opt.Key)
		} else if opt.Value != expected[idx].Value {
			t.Errorf("expected value %q, got %q", expected[idx].Value, opt.Value)
		} else {
			t.Logf("OK %d: %s = %q", idx, opt.Key, opt.Value)
		}

		for _, opt := range expected[len(opts):] {
			t.Errorf("missing option %s = %q", opt.Key, opt.Value)
		}
	}
}
