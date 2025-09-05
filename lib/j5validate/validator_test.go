package j5validate

import (
	"strings"
	"testing"

	"github.com/pentops/j5/internal/j5s/j5test"
	"github.com/pentops/j5/lib/j5reflect"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestValidate(t *testing.T) {

	t.Run("SimpleString", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! string {
			   rules.maxLength = 3
			   rules.minLength = 2
			}
		}`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar", "fooo").AssertInvalid("bar", "maximum length")
		schema.New().SetScalar("bar", "fo").AssertValid()
		schema.New().SetScalar("bar", "f").AssertInvalid("bar", "minimum length")
	})

	t.Run("Key", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field k1 key:id62
			field k2 key:uuid
			field k3 key:custom {
				pattern = "^[a-zA-Z0-9]{10}$"
			}
			field k4 key 
		}`)

		schema.New().SetScalar("k1", "030CwigNfSed7iOSKYfXYO").AssertValid()
		schema.New().SetScalar("k4", "asdf").AssertValid()
	})

	t.Run("NestedString", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! object {
				field baz ! string {
					rules.maxLength = 3
					rules.minLength = 2
				}

				field qux ? string
			}
		}`)
		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar.baz", "fooo").AssertInvalid("bar.baz", "maximum length")
		schema.New().SetScalar("bar.qux", "value").AssertInvalid("bar.baz", "required")
	})

	t.Run("ArrayOfString", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar array:string {
				rules.minItems = 2
				items.string.rules.maxLength = 3
			}
		}`)

		schema.New().AssertInvalid("bar", "required")

		schema.New().
			SetScalar("bar[]", "val1").
			SetScalar("bar[]", "val").
			AssertInvalid("bar[0]", "maximum length")

		schema.New().SetScalar("bar[]", "val").AssertInvalid("bar", "minimum items")

	})

	t.Run("StringPattern", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! string {
			   rules.pattern = "^[a-z]+$"
			}
		}`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar", "FOO").AssertInvalid("bar", "pattern")
		schema.New().SetScalar("bar", "foo").AssertValid()
		schema.New().SetScalar("bar", "foo123").AssertInvalid("bar", "pattern")
	})

	t.Run("OptionalString", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar string {
			   rules.maxLength = 3
			}
		}`)

		schema.New().AssertValid()
		schema.New().SetScalar("bar", "fooo").AssertInvalid("bar", "maximum length")
		schema.New().SetScalar("bar", "fo").AssertValid()
		schema.New().SetScalar("bar", "").AssertValid()
	})

	t.Run("OptionalStringWithMin", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar string {
			   rules.maxLength = 3
			   rules.minLength = 2
			}
		}`)

		schema.New().AssertValid()
		schema.New().SetScalar("bar", "fooo").AssertInvalid("bar", "maximum length")
		schema.New().SetScalar("bar", "fo").AssertValid()
		schema.New().SetScalar("bar", "f").AssertInvalid("bar", "minimum length")

		// Can't tell between "" and nil with neither ? nor ! are set
		schema.New().SetScalar("bar", "").AssertValid()
	})

	t.Run("ExplicitlyOptionalString", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ? string {
			   rules.maxLength = 3
			}
		}`)

		schema.New().AssertValid()
		schema.New().SetScalar("bar", "fooo").AssertInvalid("bar", "maximum length")
		schema.New().SetScalar("bar", "fo").AssertValid()

		// nothing prevents an explicitly empty string here
		schema.New().SetScalar("bar", "").AssertValid()
	})

	t.Run("ExplicitlyOptionalStringWithMin", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ? string {
			   rules.minLength = 3
			}
		}`)

		schema.New().SetScalar("bar", "").AssertInvalid("bar", "minimum length")
	})

	t.Run("RequiredBool", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! bool
		}`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar", true).AssertValid()

		// making required bool pretty useless
		schema.New().SetScalar("bar", false).AssertInvalid("bar", "required")
	})

	t.Run("OptionalBool", func(t *testing.T) {
		schema := newReflectCase(t, `
			object Foo {
				field bar ? bool
			}`)

		schema.New().AssertValid()
		schema.New().SetScalar("bar", true).AssertValid()
		schema.New().SetScalar("bar", false).AssertValid()

		// nothing prevents an explicitly empty value here
		schema.New().SetScalar("bar", nil).AssertValid()
	})

	t.Run("BoolConstraint", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! bool {
				rules.const = true
			}
		}`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar", true).AssertValid()
		schema.New().SetScalar("bar", false).AssertInvalid("bar")
	})

	t.Run("OptionalBoolConstraint", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ? bool {
				rules.const = true
			}
		}`)

		schema.New().AssertValid()
		schema.New().SetScalar("bar", true).AssertValid()
		schema.New().SetScalar("bar", false).AssertInvalid("bar", "must be true")

		// nothing prevents an explicitly empty value here
		schema.New().SetScalar("bar", nil).AssertValid()
	})

	t.Run("RangeInt32", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! integer:INT32 {
				rules.minimum = 10
				rules.maximum = 20
			}
		}`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar", 5).AssertInvalid("bar", "minimum")
		schema.New().SetScalar("bar", 25).AssertInvalid("bar", "maximum")
		schema.New().SetScalar("bar", 10).AssertValid()
		schema.New().SetScalar("bar", 20).AssertValid()
	})

	t.Run("RangeInt32Exclusive", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! integer:INT32 {
				rules.minimum = 10
				rules.maximum = 20
				rules.exclusiveMinimum = true
				rules.exclusiveMaximum = true
			}
		}`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar", 10).AssertInvalid("bar", "minimum")
		schema.New().SetScalar("bar", 20).AssertInvalid("bar", "maximum")
		schema.New().SetScalar("bar", 11).AssertValid()
		schema.New().SetScalar("bar", 19).AssertValid()
	})

	t.Run("RangeInt64", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! integer:INT64 {
				rules.minimum = 10
				rules.maximum = 20
			}
		}`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar", 5).AssertInvalid("bar", "minimum")
		schema.New().SetScalar("bar", 25).AssertInvalid("bar", "maximum")
		schema.New().SetScalar("bar", 10).AssertValid()
		schema.New().SetScalar("bar", 20).AssertValid()
	})

	t.Run("Oneof", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! oneof {
				option a object {
					field x ! string
				}
				option b object {
					field y ! string
				}
			}
		}`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar.a.x", "value").AssertValid()
		schema.New().SetScalar("bar.b.y", "value").AssertValid()

	})

	t.Run("Enum", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! enum:TestEnum
		}
		
		enum TestEnum {
		option A
		option B
		}
		`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar", "A").AssertValid()
	})

	t.Run("Optional Nil Object", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ? object {
				field baz ! string
				field other string
			}
		}`)

		schema.New().AssertValid()
		schema.New().SetScalar("bar", nil).AssertValid()
		schema.New().SetScalar("bar.other", "value").AssertInvalid("bar.baz", "required")
		schema.New().SetScalar("bar.baz", "value").AssertValid()
	})

	t.Run("Decimal", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! decimal {
				rules.minimum = "10"
				rules.maximum = "19.9"
			}
		}`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar", "10").AssertValid()
		schema.New().SetScalar("bar", "19.9").AssertValid()
		schema.New().SetScalar("bar", "9.9").AssertInvalid("bar", "minimum")
		schema.New().SetScalar("bar", "20").AssertInvalid("bar", "maximum")
	})

	t.Run("Decimal Exclusive", func(t *testing.T) {
		schema := newReflectCase(t, `
		object Foo {
			field bar ! decimal {
				rules.minimum = "10"
				rules.maximum = "19.9"
				rules.exclusiveMinimum = true
				rules.exclusiveMaximum = true
			}
		}`)

		schema.New().AssertInvalid("bar", "required")
		schema.New().SetScalar("bar", "10.1").AssertValid()
		schema.New().SetScalar("bar", "19.8").AssertValid()
		schema.New().SetScalar("bar", "10").AssertInvalid("bar", "minimum")
		schema.New().SetScalar("bar", "19.9").AssertInvalid("bar", "maximum")
	})
}

type reflectCase struct {
	desc      protoreflect.MessageDescriptor
	t         testing.TB
	reflector *j5reflect.Reflector
}

func newReflectCase(t *testing.T, schema string) *reflectCase {
	obj := j5test.ObjectReflect(t, schema)
	if obj == nil {
		t.Fatalf("failed to create object from schema: %s", schema)
	}

	schemaCache := j5schema.NewSchemaCache()
	reflector := j5reflect.NewWithCache(schemaCache)

	return &reflectCase{
		desc:      obj,
		t:         t,
		reflector: reflector,
	}
}

type objectCase struct {
	reflectCase
	j5reflect.Object
}

func (rc *reflectCase) New() *objectCase {
	msg := dynamicpb.NewMessage(rc.desc)

	root, err := rc.reflector.NewRoot(msg)
	if err != nil {
		rc.t.Fatalf("FATAL: Failed to create root: %s", err.Error())
	}

	rootObj, ok := root.(j5reflect.Object)
	if !ok {
		rc.t.Fatalf("FATAL: Root is not an object")
	}
	return &objectCase{
		reflectCase: *rc,
		Object:      rootObj,
	}
}

func (oc *objectCase) AssertInvalid(field string, msgContains ...string) {
	oc.t.Helper()
	msg := oc.Interface().(*dynamicpb.Message)
	oc.t.Logf("value: %s", prototext.Format(msg))

	assertInvalid(oc.t, oc.Object, field, msgContains...)
}

func (oc *objectCase) AssertValid() {
	oc.t.Helper()
	assertValid(oc.t, oc.Object)
}

func (oc *objectCase) SetScalar(field string, value any) *objectCase {
	fieldParts := strings.Split(field, ".")
	oc.t.Helper()
	err := oc.Object.SetScalar(value, fieldParts...)
	if err != nil {
		oc.t.Fatalf("unexpected error setting field %s: %v", field, err)
	}
	return oc
}

func assertValid(t testing.TB, obj j5reflect.Object) {
	t.Helper()
	val := NewValidator()
	err := val.Validate(obj)
	if err != nil {
		t.Fatalf("expected no validation error, got %v %s", err, err.Error())
	}

}

func assertInvalid(t testing.TB, obj j5reflect.Object, field string, msgContains ...string) {
	t.Helper()
	val := NewValidator()
	err := val.Validate(obj)
	if err == nil {
		t.Fatalf("expected validation error for field %s, got nil", field)
	}

	errs, ok := err.(Errors)
	if !ok {
		t.Fatalf("expected error to be of type Errors, got %T: %s", err, err.Error())
	}

	if len(errs) == 0 {
		t.Fatalf("expected validation errors, got none")
	}
	if len(errs) != 1 {
		t.Fatalf("expected exactly one validation error, got %d", len(errs))
	}
	gotErr := errs[0]
	if jp := gotErr.JSONPath(); jp != field {
		t.Fatalf("expected JSON path %q, got %q (for err %s)", field, jp, gotErr.Message)
	}

	for _, msg := range msgContains {
		if !strings.Contains(gotErr.Message, msg) {
			t.Fatalf("expected error message to contain %q, got %q", msg, gotErr.Message)
		}
	}

}
