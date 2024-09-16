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

	newRoot := func(t *testing.T) (obj *objectImpl, schema *schema_testpb.FullSchema) {
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
		testPath(t, root,
			tObjectProperty("sString"),
			tSetScalar("hello"),
		)

		assert.Equal(t, "hello", msg.SString)
	})

	t.Run("nested", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a nested message field
		testPath(t, root,
			tObjectProperty("sBar"),
			tObjectProperty("barId"),
			tSetScalar("123"),
		)

		if msg.SBar == nil {
			t.Fatal("msg.SBar is nil")
		}

		assert.Equal(t, "123", msg.SBar.BarId)
	})

	t.Run("repeated scalar", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a repeated leaf
		testPath(t, root,
			tObjectProperty("rString"),
			tArrayOfScalar(),
			tAppendScalar("a"),
			tAppendScalar("b"),
		)

		assert.Equal(t, []string{"a", "b"}, msg.RString)
	})

	t.Run("empty oneof object", func(t *testing.T) {
		root, msg := newRoot(t)

		testPath(t, root,
			tObjectProperty("wrappedOneof"),
			tOneofProperty("wOneofBar"),
		)

		if msg.WrappedOneof == nil {
			t.Fatal("msg.WrappedOneof is nil")
		}
		if msg.WrappedOneof.Type == nil {
			t.Fatal("msg.WrappedOneof.Type is nil")
		}
		sv, ok := msg.WrappedOneof.Type.(*schema_testpb.WrappedOneof_WOneofBar)
		if !ok {
			t.Fatalf("wrong type: %T", msg.WrappedOneof.Type)
		}
		assert.NotNil(t, sv.WOneofBar)

	})

	t.Run("repeated mutable", func(t *testing.T) {
		root, msg := newRoot(t)

		testPath(t, root,
			tObjectProperty("rBars"),
			tArrayElement(
				tObjectProperty("barId"),
				tSetScalar("1"),
			),
			tArrayElement(
				tObjectProperty("barId"),
				tSetScalar("2"),
			),
			tArrayElement(
				tObjectProperty("barId"),
				tSetScalar("3"),
			),
		)
		ids := []string{"1", "2", "3"}
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
		sMapProp, err := root.NewValue("mapStringString")
		if err != nil {
			t.Fatal(err)
		}

		sMap, ok := sMapProp.(MapOfScalarField)
		if !ok {
			t.Fatalf("Wrong type for field: %T", sMapProp)
		}

		must(t, sMap.SetGoValue("key", "value"))

		assert.Equal(t, "value", msg.MapStringString["key"])
	})

	t.Run("array of enum", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a repeated leaf
		testPath(t, root,
			tObjectProperty("rEnum"),
			tArrayOfEnum(),
			tAppendEnumFromString("VALUE1"),
			tAppendEnumFromString("VALUE2"),
		)

		assert.Equal(t, []schema_testpb.Enum{
			schema_testpb.Enum_ENUM_VALUE1,
			schema_testpb.Enum_ENUM_VALUE2,
		}, msg.REnum)
	})

	t.Run("nested flattened", func(t *testing.T) {
		root, msg := newRoot(t)

		testPath(t, root,
			tObjectProperty("fieldFromFlattened"),
			tSetScalar("hello"),
		)

		// skips the top level message, should find in the child.

		assert.NotNil(t, msg.Flattened)
		assert.Equal(t, "hello", msg.Flattened.FieldFromFlattened)

	})

	t.Run("oneof", func(t *testing.T) {
		root, msg := newRoot(t)

		testPath(t, root,
			tObjectProperty("sImplicitOneof"),
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

	t.Run("nil optional", func(t *testing.T) {
		root, msg := newRoot(t)

		// Set a nil boolean field
		sBoolProp, err := root.NewValue("oBool")
		if err != nil {
			t.Fatal("missing field sBool")
		}
		sBool, ok := sBoolProp.(ScalarField)
		if !ok {
			t.Fatalf("Wrong type for field: %T", sBoolProp)
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
		sBoolProp, err := root.NewValue("sBool")
		if err != nil {
			t.Fatal(err)
		}
		sBool, ok := sBoolProp.(ScalarField)
		if !ok {
			t.Fatalf("Wrong type for field: %T", sBoolProp)
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

}
