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

	newRoot := func(t *testing.T) (obj *Object, schema *schema_testpb.FullSchema) {
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
		prop := root.GetProperty("sString")
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
		sBar, ok := root.GetProperty("sBar").(ObjectField)
		if !ok {
			t.Fatal("missing field sBar")
		}

		asObj, err := sBar.Object()
		if err != nil {
			t.Fatalf("calling bar.Object(): %s", err)
		}

		sBarID, ok := asObj.GetProperty("id").(ScalarField)
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
		sRepeated, ok := root.GetProperty("rString").(LeafArrayField)
		if !ok {
			t.Fatal("missing field")
		}

		must(t, sRepeated.AppendGoValue("a"))

		must(t, sRepeated.AppendGoValue("b"))

		assert.Equal(t, []string{"a", "b"}, msg.RString)
	})

	t.Run("repeated mutable", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a repeated mutable field
		prop := root.GetProperty("rBars")
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

			idField, ok := barObject.GetProperty("id").(ScalarField)
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
		sMap, ok := root.GetProperty("mapStringString").(LeafMapField)
		if !ok {
			t.Fatal("missing field mString")
		}

		must(t, sMap.SetGoValue("key", "value"))

		assert.Equal(t, "value", msg.MapStringString["key"])
	})

}
