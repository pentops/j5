package j5reflect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type tPathElement func(t *testing.T, f Field) Field

type fieldable interface {
	asDetachedField() (Field, error)
}

func (obj *objectImpl) asDetachedField() (Field, error) {
	return &objectField{
		objectImpl:    obj,
		fieldDefaults: fieldDefaults{},
	}, nil
}

func testPath(t *testing.T, wrap fieldable, elements ...tPathElement) Field {
	t.Helper()

	field, err := wrap.asDetachedField()
	if err != nil {
		t.Fatal(err)
	}

	return testFieldPath(t, field, elements...)
}

func testFieldPath(t *testing.T, field Field, elements ...tPathElement) Field {
	t.Helper()
	for _, el := range elements {
		t.Logf(".next.")
		field = el(t, field)
	}
	return field
}

func toObj(t *testing.T, f Field) ObjectField {
	t.Helper()
	if f == nil {
		t.Fatal("field is nil")
	}
	asObj, ok := f.(ObjectField)
	if !ok {
		t.Fatalf("expected ObjectField, got %T", f)
	}
	return asObj
}

func tObjectFullName(name string) tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Helper()
		t.Logf("T ObjectFullName")
		obj := toObj(t, f)
		t.Logf("       Want name %s", name)
		t.Logf("        Got name %s", obj.SchemaName())
		/*
				impl := obj.(*ObjectImpl)
				descName := impl.value.descriptor.FullName()
				t.Logf(" with descriptor %s", descName)

				t.Logf(" field wrap desc %s\n", impl.value.descriptor.FullName())

				if impl.value.parent == nil {
					t.Logf("       parent is nil")
				} else {
					t.Logf("   parentMessage %s", impl.value.parent.descriptor.FullName())
				}
				if impl.value.fieldInParent == nil {
					t.Logf("fieldInParent is nil")
				} else {
					t.Logf("fieldInPareht    %s", impl.value.fieldInParent.FullName())
				}
			if name != string(descName) {
				t.Fatalf("FATAL message desc %s, got %s", name, descName)
			}
		*/
		if name != obj.SchemaName() {
			t.Fatalf("expected %s, got %s", name, obj.SchemaName())
		}
		return f
	}
}

func tOneofFullName(name string) tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Helper()
		t.Logf("T OneofFullName %q", name)
		oneof := toOneof(t, f)
		t.Logf("       Want name %s", name)
		t.Logf("       Got  name %s", oneof.SchemaName())
		assert.Equal(t, name, oneof.SchemaName())
		return f
	}
}

func tObjectProperty(name string) tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Helper()
		t.Logf("T ObjectProperty %q", name)
		obj := toObj(t, f)
		prop, err := obj.NewValue(name)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("  found object property %s", name)

		return prop
	}
}

func toOneof(t *testing.T, f Field) Oneof {
	t.Helper()
	if f == nil {
		t.Fatal("field is nil")
	}
	asOneof, ok := f.(OneofField)
	if !ok {
		t.Fatalf("expected OneofField, got %T", f)
	}
	return asOneof
}

/*
	field := asOneof.(*oneofField)
asReal := field.value.(*realOneofField)
t.Logf("Is a OneofField")
t.Logf("field      %s", asReal.fieldInParent.FullName())
t.Logf("in parent  %s", asReal.parent.descriptor.FullName())
}*/

func tOneofProperty(name string) tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Helper()
		t.Logf("T OneofProperty")
		oneof := toOneof(t, f)
		prop, err := oneof.NewValue(name)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("  found enum property %s", name)

		return prop
	}
}

/*
func tIsScalar() tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Logf("T IsScalar")
		objField, ok := f.(ScalarField)
		if !ok {
			t.Fatalf("expected ScalarField, got %T", f)
		}
		return objField
	}
}*/

func tSetScalar(v interface{}) tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Helper()
		t.Logf("T SetScalar")
		objField, ok := f.(ScalarField)
		if !ok {
			t.Fatalf("expected ScalarField, got %T", f)
		}
		must(t, objField.SetGoValue(v))
		return f
	}
}

func tArrayElement(inArrayElement ...tPathElement) tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Helper()
		t.Logf("T ArrayElement")
		if f == nil {
			t.Fatal("field is nil")
		}

		objField, ok := f.(MutableArrayField)
		if !ok {
			t.Fatalf("expected MutableArrayField, got %T", f)
		}
		t.Logf("  new array element")
		element := objField.NewElement()
		t.Logf("  Run with new element %q", element.NameInParent())
		testFieldPath(t, element, inArrayElement...)

		// Then returns the outer array, so we can call this more than once
		return f
	}
}

func tArrayOfScalar() tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Helper()
		t.Logf("T ArrayOfScalar")
		_, ok := f.(ArrayOfScalarField)
		if !ok {
			t.Fatalf("expected ArrayOfScalarField, got %T", f)
		}
		return f
	}
}

func tAppendScalar(v interface{}) tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Helper()
		t.Logf("T AppendScalar")
		objField, ok := f.(ArrayOfScalarField)
		if !ok {
			t.Fatalf("expected ArrayOfScalarField, got %T", f)
		}
		_, err := objField.AppendGoValue(v)
		must(t, err)
		return f
	}
}

func tArrayOfEnum() tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Helper()
		t.Logf("T ArrayOfEnum")
		_, ok := f.(ArrayOfEnumField)
		if !ok {
			t.Fatalf("expected ArrayOfEnumField, got %T", f)
		}
		return f
	}
}

func tAppendEnumFromString(v string) tPathElement {
	return func(t *testing.T, f Field) Field {
		t.Helper()
		t.Logf("T AppendEnumFromString")
		objField, ok := f.(ArrayOfEnumField)
		if !ok {
			t.Fatalf("expected ArrayOfEnumField, got %T", f)
		}
		_, err := objField.AppendEnumFromString(v)
		must(t, err)
		return f
	}
}

func ptr[T any](t T) *T {
	return &t
}
