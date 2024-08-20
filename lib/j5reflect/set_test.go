package j5reflect

import (
	"testing"

	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
	"github.com/pentops/j5/gen/test/schema/v1/schema_testpb"
	"github.com/stretchr/testify/assert"
)

func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetter(t *testing.T) {
	refl := New()

	newRoot := func(t *testing.T) (obj *ObjectImpl, schema *schema_testpb.FullSchema) {
		// not setting helper, as errors here are global fault
		msg := &schema_testpb.FullSchema{}
		root, err := refl.NewObject(msg.ProtoReflect())
		if err != nil {
			t.Fatal(err)
		}
		return root, msg
	}

	t.Run("scalar", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a scalar field
		prop := root.MaybeGetProperty("sString").Field()
		if prop == nil {
			t.Fatal("missing field sString")
		}
		sString, ok := prop.(ScalarField)
		if !ok {
			t.Fatalf("Wrong type for field: %T", prop)
		}

		must(t, sString.SetGoValue("hello"))

		assert.Equal(t, "hello", msg.SString)
	})

	t.Run("nested", func(t *testing.T) {
		root, msg := newRoot(t)
		// Set a nested message field
		sBar, ok := root.MaybeGetProperty("sBar").Field().(ObjectField)
		if !ok {
			t.Fatal("missing field sBar")
		}

		asObj, err := sBar.Object()
		if err != nil {
			t.Fatalf("calling bar.Object(): %s", err)
		}

		sBarID, ok := asObj.MaybeGetProperty("barId").Field().(ScalarField)
		if !ok {
			t.Fatal("missing field ID")
		}

		must(t, sBarID.SetGoValue("123"))

		if msg.SBar == nil {
			t.Fatal("msg.SBar is nil")
		}

		assert.Equal(t, "123", msg.SBar.BarId)
	})

	t.Run("repeated scalar", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a repeated leaf
		sRepeated, ok := root.MaybeGetProperty("rString").Field().(ArrayOfScalarField)
		if !ok {
			t.Fatal("missing field")
		}

		must(t, sRepeated.AppendGoScalar("a"))

		must(t, sRepeated.AppendGoScalar("b"))

		assert.Equal(t, []string{"a", "b"}, msg.RString)
	})

	t.Run("repeated mutable", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a repeated mutable field
		prop := root.MaybeGetProperty("rBars").Field().(ArrayOfObjectField)
		if prop == nil {
			t.Fatal("missing field rBar")
		}
		sRepeated, ok := prop.(MutableArrayField)
		if !ok {
			t.Fatalf("rBar is a %T", prop)
		}

		ids := []string{"1", "2", "3"}
		for _, id := range ids {
			element := sRepeated.NewElement()
			barField, ok := element.(ObjectField)
			if !ok {
				t.Fatalf("bar is a %T", element)
			}
			barObject, err := barField.Object()
			if err != nil {
				t.Fatalf("calling barField.Object(): %s", err)
			}

			idField, ok := barObject.MaybeGetProperty("barId").Field().(ScalarField)
			if !ok {
				t.Fatal("missing field id")
			}
			must(t, idField.SetGoValue(id))
		}

		if len(msg.RBars) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(msg.RBars))
		}

		for i, id := range ids {
			assert.Equal(t, id, msg.RBars[i].BarId)
		}

	})

	t.Run("map scalar", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a map scalar field
		sMap, ok := root.MaybeGetProperty("mapStringString").Field().(MapOfScalarField)
		if !ok {
			t.Fatal("missing field mString")
		}

		must(t, sMap.SetGoScalar("key", "value"))

		assert.Equal(t, "value", msg.MapStringString["key"])
	})

	t.Run("nil optional", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a nil boolean field
		sBool, ok := root.MaybeGetProperty("oBool").Field().(ScalarField)
		if !ok {
			t.Fatal("missing field sBool")
		}

		must(t, sBool.SetGoValue(true))
		assert.Equal(t, ptr(true), msg.OBool)

		var b *bool
		must(t, sBool.SetGoValue(b))
		assert.Nil(t, msg.OBool)

		must(t, sBool.SetGoValue(true))
		assert.Equal(t, ptr(true), msg.OBool)

		must(t, sBool.SetGoValue(nil))
		assert.Nil(t, msg.OBool)
	})

	t.Run("nil required", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a nil boolean field
		sBool, ok := root.MaybeGetProperty("sBool").Field().(ScalarField)
		if !ok {
			t.Fatal("missing field sBool")
		}

		must(t, sBool.SetGoValue(true))
		assert.Equal(t, true, msg.SBool)

		var b *bool
		must(t, sBool.SetGoValue(b))
		assert.False(t, msg.SBool)

		pr := msg.ProtoReflect()
		assert.False(t, pr.Has(pr.Descriptor().Fields().ByJSONName("sBool")))

		must(t, sBool.SetGoValue(true))
		assert.Equal(t, true, msg.SBool)

		must(t, sBool.SetGoValue(nil))
		assert.False(t, msg.SBool)
	})

	t.Run("array of enum", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a repeated leaf
		sRepeated, ok := root.MaybeGetProperty("rEnum").Field().(ArrayOfEnumField)
		if !ok {
			t.Fatal("missing field")
		}

		must(t, sRepeated.AppendEnumFromString("VALUE1"))

		must(t, sRepeated.AppendEnumFromString("VALUE2"))

		assert.Equal(t, []schema_testpb.Enum{
			schema_testpb.Enum_ENUM_VALUE1,
			schema_testpb.Enum_ENUM_VALUE2,
		}, msg.REnum)
	})

	t.Run("nested flattened", func(t *testing.T) {
		root, msg := newRoot(t)

		// skips the top level message, should find in the child.
		field, ok := root.MaybeGetProperty("fieldFromFlattened").Field().(ScalarField)
		if !ok {
			t.Fatal("missing field fieldFromFlattened")
		}

		if err := field.SetGoValue("hello"); err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, msg.Flattened)
		assert.Equal(t, "hello", msg.Flattened.FieldFromFlattened)

	})

	t.Run("oneof", func(t *testing.T) {
		root, msg := newRoot(t)

		testPath(t, root.MaybeGetProperty("sImplicitOneof").Field(),
			tOneofFullName("test.schema.v1.ImplicitOneof"),
			tOneofProperty("ioBar"),
			tObjectFullName("test.schema.v1.Bar"),
			tObjectProperty("barId"),
			tSetScalar("123"),
		)

		setBar := msg.SImplicitOneof.GetIoBar()
		if setBar == nil {
			t.Fatal("missing oneof field")
		}

		assert.Equal(t, "123", setBar.BarId)

	})

	t.Run("oneof schema", func(t *testing.T) {

		msg := &sourcedef_j5pb.SourceFile{}
		root, err := refl.NewObject(msg.ProtoReflect())
		if err != nil {
			t.Fatal(err)
		}

		cb = func(name string, params ...interface{}) {
			t.Logf(name, params...)
		}

		testPath(t, root.MaybeGetProperty("entities").Field(),
			tArrayElement(),
			tObjectFullName("j5.sourcedef.v1.Entity"),
			tObjectProperty("schemas"),
			//tArrayElement(),
			tOneofFullName("j5.sourcedef.v1.RootSchema"),
			tOneofProperty("object"),
			tObjectFullName("j5.sourcedef.v1.Object"),
			tObjectProperty("def"),
			tObjectFullName("j5.schema.v1.Object"),
		)

	})

}

func toObj(t *testing.T, f Field) Object {
	if f == nil {
		t.Fatal("field is nil")
	}
	asObj, ok := f.(ObjectField)
	if !ok {
		t.Fatalf("expected ObjectField, got %T", f)
	}
	obj, err := asObj.Object()
	if err != nil {
		t.Fatal(err)
	}
	return obj
}

func tObjectFullName(name string) tPathElement {
	return func(t *testing.T, f Field) Field {
		obj := toObj(t, f)
		t.Logf("Assert object.Name()")
		t.Logf("       Want name %s", name)
		t.Logf("        Got name %s", obj.Name())
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

		if name != obj.Name() {
			t.Fatalf("expected %s, got %s", name, obj.Name())
		}
		if name != string(descName) {
			t.Fatalf("FATAL message desc %s, got %s", name, descName)
		}
		return f
	}
}

func tOneofFullName(name string) tPathElement {
	return func(t *testing.T, f Field) Field {
		oneof := toOneof(t, f)
		t.Logf("Assert oneof.Name()")
		t.Logf("       Want name %s", name)
		t.Logf("       Got  name %s", oneof.Name())
		impl := oneof.(*OneofImpl)
		descName := impl.value.descriptor.FullName()
		t.Logf(" with descriptor %s", descName)
		assert.Equal(t, name, oneof.Name())
		return f
	}
}

func tObjectProperty(name string) tPathElement {
	return func(t *testing.T, f Field) Field {
		obj := toObj(t, f)
		prop, err := obj.GetProperty(name)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("found object property %s", name)
		return prop.Field()
	}
}

func toOneof(t *testing.T, f Field) Oneof {
	if f == nil {
		t.Fatal("field is nil")
	}
	asOneof, ok := f.(OneofField)
	if !ok {
		t.Fatalf("expected OneofField, got %T", f)
	}
	field := asOneof.(*oneofField)
	asReal := field.value.(*realProtoMessageField)
	t.Logf("Is a OneofField")
	t.Logf("field      %s", asReal.fieldInParent.FullName())
	t.Logf("in parent  %s", asReal.parent.descriptor.FullName())
	obj, err := asOneof.Oneof()
	if err != nil {
		t.Fatal(err)
	}
	return obj
}

func tOneofProperty(name string) tPathElement {
	return func(t *testing.T, f Field) Field {
		oneof := toOneof(t, f)
		prop, err := oneof.GetProperty(name)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("found enum property %s", name)

		return prop.Field()
	}
}

func tIsScalar() tPathElement {
	return func(t *testing.T, f Field) Field {
		objField, ok := f.(ScalarField)
		if !ok {
			t.Fatalf("expected ScalarField, got %T", f)
		}
		return objField
	}
}

func tSetScalar(v string) tPathElement {
	return func(t *testing.T, f Field) Field {
		objField, ok := f.(ScalarField)
		if !ok {
			t.Fatalf("expected ScalarField, got %T", f)
		}
		must(t, objField.SetGoValue(v))
		return f
	}
}

func tArrayElement() tPathElement {
	return func(t *testing.T, f Field) Field {
		if f == nil {
			t.Fatal("field is nil")
		}
		objField, ok := f.(MutableArrayField)
		if !ok {
			t.Fatalf("expected MutableArrayField, got %T", f)
		}
		t.Logf("new element")
		return objField.NewElement()
	}
}

type tPathElement func(t *testing.T, f Field) Field

func testPath(t *testing.T, field Field, elements ...tPathElement) Field {
	t.Helper()
	for _, el := range elements {
		field = el(t, field)
	}
	return field
}

func ptr[T any](t T) *T {
	return &t
}
