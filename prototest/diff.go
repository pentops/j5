package prototest

import (
	"bytes"
	"fmt"
	"math"
	"reflect"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type differ interface {
	add(Diff)
	child(string) differ
}

type differRoot struct {
	Diffs []Diff
}

func (d *differRoot) add(diff Diff) {
	d.Diffs = append(d.Diffs, diff)
}

func (d *differRoot) child(name string) differ {
	return &differChild{
		parent:       d,
		nameInParent: name,
	}
}

type differChild struct {
	parent       differ
	nameInParent string
}

type Diff struct {
	Path string
	A    string
	B    string
}

func (d *differChild) child(name string) differ {
	return &differChild{
		parent:       d,
		nameInParent: name,
	}
}

func (d *differChild) add(diff Diff) {
	d.parent.add(Diff{
		Path: d.nameInParent + "." + diff.Path,
		A:    diff.A,
		B:    diff.B,
	})
}

func MessageDiff(v1, v2 protoreflect.Message) []Diff {
	d := &differRoot{}
	diffMessage(d, v1, v2)
	return d.Diffs
}

func equalValue(kind protoreflect.Kind, x, y protoreflect.Value) bool {
	switch kind {
	case protoreflect.BoolKind:
		return x.Bool() == y.Bool()
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		return x.Int() == y.Int()
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		return x.Uint() == y.Uint()
	case protoreflect.FloatKind:
		return equalFloat(x.Float(), y.Float())
	case protoreflect.StringKind:
		return x.String() == y.String()
	case protoreflect.BytesKind:
		return bytes.Equal(x.Bytes(), y.Bytes())
	case protoreflect.EnumKind:
		return x.Enum() == y.Enum()
	default:
		panic(fmt.Sprintf("unknown type: %T", x))
	}
}

// equalFloat compares two floats, where NaNs are treated as equal.
func equalFloat(x, y float64) bool {
	if math.IsNaN(x) || math.IsNaN(y) {
		return math.IsNaN(x) && math.IsNaN(y)
	}
	return x == y
}

// equalMessage compares two messages.
func diffMessage(d differ, mx, my protoreflect.Message) {

	xFields := mx.Descriptor().Fields()
	yFields := my.Descriptor().Fields()

	if xFields.Len() != yFields.Len() {
		d.add(Diff{
			Path: "!Fields",
			A:    fmt.Sprintf("%v", xFields.Len()),
			B:    fmt.Sprintf("%v", yFields.Len()),
		})
		return
	}

	for i := 0; i < xFields.Len(); i++ {
		fx := xFields.Get(i)
		fy := yFields.ByNumber(fx.Number())

		if fx.Name() != fy.Name() {
			d.add(Diff{
				Path: fmt.Sprintf("!field %d.name", i),
				A:    fmt.Sprintf("%v", fx.Name()),
				B:    fmt.Sprintf("%v", fy.Name()),
			})
			return
		}

		if fx.Kind() != fy.Kind() {
			d.add(Diff{
				Path: fmt.Sprintf("!%s.Kind()", fx.Name()),
				A:    fmt.Sprintf("%v", fx.Kind()),
				B:    fmt.Sprintf("%v", fy.Kind()),
			})
			return
		}

		if fx.IsList() {
			if !fy.IsList() {
				d.add(Diff{
					Path: fmt.Sprintf("!%s.IsList()", fx.Name()),
					A:    fmt.Sprintf("%v", fx.IsList()),
					B:    fmt.Sprintf("%v", fy.IsList()),
				})
				return
			}
		}

		if fx.IsMap() {
			if !fy.IsMap() {
				d.add(Diff{
					Path: fmt.Sprintf("!%s.map", fx.Name()),
					A:    fmt.Sprintf("%v", fx.IsMap()),
					B:    fmt.Sprintf("%v", fy.IsMap()),
				})
				return
			}
		}

		fieldDiff := d.child(string(fx.Name()))

		if mx.Has(fx) != my.Has(fy) {
			if mx.Has(fx) {
				fieldDiff.add(Diff{
					A: fmt.Sprintf("%v", mx.Get(fx).Interface()),
					B: "<nil>",
				})
			} else {
				fieldDiff.add(Diff{
					A: "<nil>",
					B: fmt.Sprintf("%v", my.Get(fy).Interface()),
				})
			}
			return
		}

		fieldX := mx.Get(fx)
		fieldY := my.Get(fy)

		diffValue(fieldDiff, fx, fy, fieldX, fieldY)

	}

	if !equalUnknown(mx.GetUnknown(), my.GetUnknown()) {
		d.add(Diff{
			Path: "!Unknown",
			A:    fmt.Sprintf("%v", mx.GetUnknown()),
			B:    fmt.Sprintf("%v", my.GetUnknown()),
		})
	}
}

func diffValue(d differ, fx, fy protoreflect.FieldDescriptor, fieldX, fieldY protoreflect.Value) {
	if fx.IsList() {
		diffList(d, fx, fy, fieldX.List(), fieldY.List())
		return
	}

	if fx.IsMap() {
		diffMap(d, fx, fy, fieldX.Map(), fieldY.Map())
		return
	}

	if fx.Kind() == protoreflect.MessageKind {
		diffMessage(d, fieldX.Message(), fieldY.Message())
		return
	}

	if !equalValue(fx.Kind(), fieldX, fieldY) {
		d.add(Diff{
			A: fmt.Sprintf("%v", fieldX.Interface()),
			B: fmt.Sprintf("%v", fieldY.Interface()),
		})
	}
}

// equalList compares two lists.
func diffList(d differ, fx, fy protoreflect.FieldDescriptor, x, y protoreflect.List) {
	if x.Len() != y.Len() {
		d.add(Diff{
			Path: "!Len()",
			A:    fmt.Sprintf("%v", x.Len()),
			B:    fmt.Sprintf("%v", y.Len()),
		})
		return
	}
	for i := x.Len() - 1; i >= 0; i-- {
		valX := x.Get(i)
		valY := y.Get(i)

		if fx.Kind() == protoreflect.MessageKind {
			diffMessage(d, valX.Message(), valY.Message())
			return
		}

		if !equalValue(fx.Kind(), valX, valY) {
			d.add(Diff{
				A: fmt.Sprintf("%v", valX.Interface()),
				B: fmt.Sprintf("%v", valY.Interface()),
			})
		}
	}
}

// equalMap compares two maps.
func diffMap(d differ, fx, fy protoreflect.FieldDescriptor, x, y protoreflect.Map) {
	d.add(Diff{
		Path: "!MAP() not supported",
	})

}

// equalUnknown compares unknown fields by direct comparison on the raw bytes
// of each individual field number.
func equalUnknown(x, y protoreflect.RawFields) bool {
	if len(x) != len(y) {
		return false
	}
	if bytes.Equal([]byte(x), []byte(y)) {
		return true
	}

	mx := make(map[protoreflect.FieldNumber]protoreflect.RawFields)
	my := make(map[protoreflect.FieldNumber]protoreflect.RawFields)
	for len(x) > 0 {
		fnum, _, n := protowire.ConsumeField(x)
		mx[fnum] = append(mx[fnum], x[:n]...)
		x = x[n:]
	}
	for len(y) > 0 {
		fnum, _, n := protowire.ConsumeField(y)
		my[fnum] = append(my[fnum], y[:n]...)
		y = y[n:]
	}
	return reflect.DeepEqual(mx, my)
}
