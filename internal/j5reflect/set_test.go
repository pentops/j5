package j5reflect

import (
	"testing"

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
		prop := root.GetProperty("sString").Field()
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
		sBar, ok := root.GetProperty("sBar").Field().(ObjectField)
		if !ok {
			t.Fatal("missing field sBar")
		}

		asObj, err := sBar.Object()
		if err != nil {
			t.Fatalf("calling bar.Object(): %s", err)
		}

		sBarID, ok := asObj.GetProperty("id").Field().(ScalarField)
		if !ok {
			t.Fatal("missing field ID")
		}

		must(t, sBarID.SetGoValue("123"))

		if msg.SBar == nil {
			t.Fatal("msg.SBar is nil")
		}

		assert.Equal(t, "123", msg.SBar.Id)
	})

	t.Run("repeated scalar", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a repeated leaf
		sRepeated, ok := root.GetProperty("rString").Field().(ArrayOfScalarField)
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
		prop := root.GetProperty("rBars").Field().(ArrayOfObjectField)
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

			idField, ok := barObject.GetProperty("id").Field().(ScalarField)
			if !ok {
				t.Fatal("missing field id")
			}
			must(t, idField.SetGoValue(id))
		}

		if len(msg.RBars) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(msg.RBars))
		}

		for i, id := range ids {
			assert.Equal(t, id, msg.RBars[i].Id)
		}

	})

	t.Run("map scalar", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a map scalar field
		sMap, ok := root.GetProperty("mapStringString").Field().(MapOfScalarField)
		if !ok {
			t.Fatal("missing field mString")
		}

		must(t, sMap.SetGoScalar("key", "value"))

		assert.Equal(t, "value", msg.MapStringString["key"])
	})

	t.Run("nil optional", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a nil boolean field
		sBool, ok := root.GetProperty("oBool").Field().(ScalarField)
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
		sBool, ok := root.GetProperty("sBool").Field().(ScalarField)
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
		sRepeated, ok := root.GetProperty("rEnum").Field().(ArrayOfEnumField)
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
		field, ok := root.GetProperty("fieldFromFlattened").Field().(ScalarField)
		if !ok {
			t.Fatal("missing field fieldFromFlattened")
		}

		if err := field.SetGoValue("hello"); err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, msg.Flattened)
		assert.Equal(t, "hello", msg.Flattened.FieldFromFlattened)

	})

}

func ptr[T any](t T) *T {
	return &t
}
