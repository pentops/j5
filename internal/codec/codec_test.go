package codec

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pentops/flowtest/prototest"
	"github.com/pentops/j5/internal/gen/test/schema/v1/schema_testpb"
	"github.com/pentops/j5/j5types/any_j5t"
	"github.com/pentops/j5/j5types/date_j5t"
	"github.com/pentops/j5/j5types/decimal_j5t"
	"github.com/pentops/j5/lib/j5reflect"
	"github.com/pentops/j5/lib/j5schema"

	"github.com/pentops/j5/internal/j5s/j5test"
)

func mustJ5Any(t testing.TB, msg proto.Message, asJSON []byte) *any_j5t.Any {
	a, err := any_j5t.FromProto(msg)
	if err != nil {
		t.Fatal(err)
	}
	a.J5Json = asJSON
	return a
}

func mustProtoAny(t testing.TB, msg proto.Message) *anypb.Any {
	a, err := anypb.New(msg)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

type testSchema struct {
	t      testing.TB
	schema string
}

func NewTestSchema(t testing.TB, schema string) *testSchema {
	return &testSchema{
		t:      t,
		schema: schema,
	}
}

func (dc *testSchema) Object() j5reflect.Object {
	return j5test.DynamicObject(dc.t, dc.schema)
}

func (dc *testSchema) WantJSON(j string, opts ...CodecOption) *testSchemaWithOutput {
	dc.t.Helper()

	cache := j5schema.NewSchemaCache()
	reflector := j5reflect.NewWithCache(cache)
	codec := NewCodec(append(opts, WithReflector(reflector))...)

	tco := &testSchemaWithOutput{
		testSchema: dc,
		wantOutput: []byte(j),
		codec:      codec,
	}
	// assert that the expected JSON output, when given as Input, results in itself
	tco.InputJSON(j)
	return tco
}

type testSchemaWithOutput struct {
	*testSchema
	codec      *Codec
	wantOutput []byte
}

func (dc *testSchemaWithOutput) InputJSON(jsonInput string) *testSchemaWithOutput {
	dc.t.Helper()

	parsed := dc.Object()
	err := dc.codec.JSONToReflect([]byte(jsonInput), parsed)
	if err != nil {
		dc.t.Fatalf("JSONToReflect: %s", err)
	}

	dc.t.Logf("got proto:\n%s\n", prototext.Format(parsed))

	encodedJSON, err := dc.codec.ReflectToJSON(parsed)
	if err != nil {
		dc.t.Fatalf("ReflectToJSON: %s", err)
	}
	match, diff := JSONEqual(dc.t, dc.wantOutput, encodedJSON)
	if !match {
		dc.t.Logf("For input %s", jsonInput)
		dc.t.Errorf("JSON mismatch\n%s", diff)
	}

	return dc
}

func TestDynamic(t *testing.T) {

	t.Run("string", func(t *testing.T) {
		schema := NewTestSchema(t, `
			object Foo {
			  field sString string
			}
		`)

		schema.WantJSON(`{"sString": "val"}`)
		schema.WantJSON(`{}`).
			InputJSON(`{"sString": ""}`).
			InputJSON(`{"sString": null}`)

		schema.WantJSON(`{"sString": ""}`, WithIncludeEmpty()).
			InputJSON(`{}`).
			InputJSON(`{"sString": null}`)
	})

	t.Run("nullable string", func(t *testing.T) {
		schema := NewTestSchema(t, `
			object Foo {
			  field sString ? string
			}
		`)

		schema.WantJSON(`{"sString": "val"}`)
		schema.WantJSON(`{"sString": ""}`)
		schema.WantJSON(`{}`).
			InputJSON(`{"sString": null}`)

	})

	t.Run("required string", func(t *testing.T) {
		schema := NewTestSchema(t, `
			object Foo {
			  field sString ! string
			}
		`)

		schema.WantJSON(`{"sString": "val"}`)
	})

	t.Run("integer", func(t *testing.T) {
		schema := NewTestSchema(t, `
		object FooCharacteristics {
			field weight integer:INT64 {
				listRules.filtering.filterable = true
				listRules.sorting.sortable = true
			}

			field height integer:INT64 {
				listRules.filtering.filterable = true
				listRules.sorting.sortable = true
			}

			field length integer:INT64 {
				listRules.filtering.filterable = true
				listRules.sorting.sortable = true
			}
		}
		`)

		schema.WantJSON(`{}`)

		schema.WantJSON(`{"weight": "0", "height": "0", "length": "0"}`, WithIncludeEmpty()).InputJSON(`{}`)

		schema.WantJSON(`{"weight": "0", "height": "0", "length": "1"}`, WithIncludeEmpty()).InputJSON(`{"length": "1"}`)

		empty := schema.Object()

		val, ok, err := empty.GetField("weight")
		if err != nil {
			t.Fatalf("GetField: %s", err)
		}
		if ok {
			t.Fatalf("expected weight to not be set, but it is: %s", val)
		}

		scalar, ok := val.AsScalar()
		if !ok {
			t.Fatalf("expected weight to be a scalar, but it is not: %s", val)
		}
		goValue, err := scalar.ToGoValue()
		if err != nil {
			t.Fatalf("ToGoValue: %s", err)
		}

		if goValue != int64(0) {
			t.Fatalf("expected weight to be 0, but it is: %v", goValue)
		}

	})

}

func TestUnmarshal(t *testing.T) {

	cache := j5schema.NewSchemaCache()
	reflector := j5reflect.NewWithCache(cache)
	codec := NewCodec(WithProtoToAny(), WithReflector(reflector))

	testTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name         string
		wantProto    proto.Message
		json         string
		altInputJSON []string // when set, the output can be different to the input, but all outputs should be the same

		queries []url.Values
	}{
		{
			name: "strings",

			json: `{
				"sString": "nameVal",
				"oString": "otherNameVal",
				"rString": ["r1", "r2"]
			}`,
			queries: []url.Values{{
				"sString": []string{"nameVal"},
				"oString": []string{"otherNameVal"},
				"rString": []string{"r1", "r2"},
			}},
			wantProto: &schema_testpb.FullSchema{
				SString: "nameVal",
				OString: proto.String("otherNameVal"),
				RString: []string{"r1", "r2"},
			},
		}, {
			name: "floats",
			json: `{
				"sFloat": 1.1,
				"oFloat": 2.2,
				"rFloat": [3.3, 4.4]
			}`,
			altInputJSON: []string{`{
				"sFloat": "1.1",
				"oFloat": "2.2",
				"rFloat": ["3.3", "4.4"]
			}`},
			queries: []url.Values{{
				"sFloat": []string{"1.1"},
				"oFloat": []string{"2.2"},
				"rFloat": []string{"3.3", "4.4"},
			}},
			wantProto: &schema_testpb.FullSchema{
				SFloat: 1.1,
				OFloat: proto.Float32(2.2),
				RFloat: []float32{3.3, 4.4},
			},
		}, {
			name: "integers",
			json: `{
				"sInt32": -1,
				"sUint32": 1,
				"sSint32": -1,
				"sInt64": "-1",
				"sUint64": "2"
			}`,
			altInputJSON: []string{`{
				"sInt32": "-1",
				"sUint32": "1",
				"sSint32": "-1",
				"sInt64": "-1",
				"sUint64": "2"
			}`, `{
				"sInt32": -1,
				"sUint32": 1,
				"sSint32": -1,
				"sInt64": -1,
				"sUint64": 2
			}`},

			wantProto: &schema_testpb.FullSchema{
				SInt32:  -1,
				SUint32: 1,
				SSint32: -1,
				SInt64:  -1,
				SUint64: 2,
			},
		}, {
			name: "bool",
			json: `{
				"sBool": true
			}`,
			wantProto: &schema_testpb.FullSchema{
				SBool: true,
			},
		}, {
			name: "enum",
			json: `{
				"enum": "VALUE1",
				"rEnum": ["VALUE1", "VALUE2"]
			}`,
			wantProto: &schema_testpb.FullSchema{

				Enum: schema_testpb.Enum_ENUM_VALUE1,
				REnum: []schema_testpb.Enum{
					schema_testpb.Enum_ENUM_VALUE1,
					schema_testpb.Enum_ENUM_VALUE2,
				},
			},
		}, {
			name: "long enums",

			json: `{
				"enum": "VALUE1",
				"rEnum": ["VALUE1", "VALUE2"]
			}`,
			altInputJSON: []string{` {
					"enum": "VALUE1",
					"rEnum": ["ENUM_VALUE1", "ENUM_VALUE2"]
				}`},

			wantProto: &schema_testpb.FullSchema{
				Enum: schema_testpb.Enum_ENUM_VALUE1,
				REnum: []schema_testpb.Enum{
					schema_testpb.Enum_ENUM_VALUE1,
					schema_testpb.Enum_ENUM_VALUE2,
				},
			},
		}, {
			name: "bytes",
			json: `{
				"sBytes": "` +
				base64.StdEncoding.EncodeToString([]byte("sBytes")) +
				`",
				"rBytes": ["` +
				base64.StdEncoding.EncodeToString([]byte("rBytes1")) +
				`","` +
				base64.StdEncoding.EncodeToString([]byte("rBytes2")) +
				`"]
			}`,
			wantProto: &schema_testpb.FullSchema{
				SBytes: []byte("sBytes"),
				RBytes: [][]byte{[]byte("rBytes1"), []byte("rBytes2")},
			},
		}, {
			name: "base64 diffs",
			// encoding of 0xFBFO is +/A= or -_A= depending on the method
			// The codec should accept either, with or without padding
			json: `{"sBytes": "+/A="}`,
			altInputJSON: []string{
				`{"sBytes": "-_A="}`,
				`{"sBytes": "+/A"}`,
				`{"sBytes": "-_A"}`,
			},
			wantProto: &schema_testpb.FullSchema{
				SBytes: []byte{0xfB, 0xF0},
			},
		}, {
			name: "map",
			json: `{ "mapStringString": {
				"k1": "val1"
			} }`,
			// TODO: Can only test one key this way while maps are unordered
			wantProto: &schema_testpb.FullSchema{
				MapStringString: map[string]string{
					"k1": "val1",
				},
			},
		}, {
			name: "mapObject",
			json: `{ 
				"mapStringBar": {
					"k1": {"barId": "id"}
				}
			}`,
			// TODO: Can only test one key this way while maps are unordered
			wantProto: &schema_testpb.FullSchema{
				MapStringBar: map[string]*schema_testpb.Bar{
					"k1": {
						BarId: "id",
					},
				},
			},
		}, {
			name: "timestamp",
			json: `{
				"ts": "2020-01-01T00:00:00Z",
				"rTs": ["2020-01-01T00:00:00Z", "2020-01-01T00:00:00Z"]
			}`,
			wantProto: &schema_testpb.FullSchema{
				Ts: timestamppb.New(testTime),
				RTs: []*timestamppb.Timestamp{
					timestamppb.New(testTime),
					timestamppb.New(testTime),
				},
			},
		}, {
			name: "date",
			json: `{
				"date": "2000-01-02"
			}`,
			queries: []url.Values{{
				"date": []string{"2000-01-02"},
			}},
			wantProto: &schema_testpb.FullSchema{
				Date: &date_j5t.Date{Year: 2000, Month: 1, Day: 2},
			},
		}, {
			name: "date array",

			json: `{
				"rDate": ["2001-01-02", "2002-01-02"]
			}`,
			queries: []url.Values{{
				"rDate": []string{"2001-01-02", "2002-01-02"},
			}, {
				"rDate": []string{`["2001-01-02","2002-01-02"]`},
			}},
			wantProto: &schema_testpb.FullSchema{
				RDate: []*date_j5t.Date{
					{Year: 2001, Month: 1, Day: 2},
					{Year: 2002, Month: 1, Day: 2},
				},
			},
		}, {
			name: "object",
			json: `{
				"sBar": {
					"barId": "barId",
					"barField": "field"
				}
			}`,
			queries: []url.Values{{
				"sBar.barId":    []string{"barId"},
				"sBar.barField": []string{"field"},
			}, {
				"sBar": []string{`{"barId": "barId", "barField": "field"}`},
			}, {
				"s_bar.bar_id":    []string{"barId"},
				"s_bar.bar_field": []string{"field"},
			}},
			wantProto: &schema_testpb.FullSchema{
				SBar: &schema_testpb.Bar{
					BarId:    "barId",
					BarField: "field",
				},
			},
		}, {
			name: "array of objects",
			json: `{
				"rBars": [{
					"barId": "bar1"
				}, {
					"barId": "bar2"
				}]
			}`,
			queries: []url.Values{{
				"rBars": []string{`[{"barId": "bar1"}, {"barId": "bar2"}]`},
			}, {
				"rBars": []string{`{"barId": "bar1"}`, `{"barId": "bar2"}`},
			}},
			wantProto: &schema_testpb.FullSchema{
				RBars: []*schema_testpb.Bar{{
					BarId: "bar1",
				}, {
					BarId: "bar2",
				}},
			},
		}, {
			name: "objects null",
			json: `{}`,
			altInputJSON: []string{
				`{}`,
				`{"sBar": null}`,
			},
			wantProto: &schema_testpb.FullSchema{},
		}, {
			name: "objects empty",
			json: `{ "sBar": {} }`,
			wantProto: &schema_testpb.FullSchema{
				SBar: &schema_testpb.Bar{},
			},
		}, {

			name: "flattened messages",
			json: `{
				"fieldFromFlattened": "fieldFromFlattenedVal",
				"field2FromFlattened": "field2FromFlattenedVal"
			}`,
			wantProto: &schema_testpb.FullSchema{
				Flattened: &schema_testpb.FlattenedMessage{
					FieldFromFlattened:   "fieldFromFlattenedVal",
					Field_2FromFlattened: "field2FromFlattenedVal",
				},
			},
		}, {
			name: "anon oneof",
			json: `{
				"aOneofString": "stringVal"
			}`,
			wantProto: &schema_testpb.FullSchema{
				AnonOneof: &schema_testpb.FullSchema_AOneofString{
					AOneofString: "stringVal",
				},
			},
		}, {
			name: "oneof wrapper",
			json: `{
				"wrappedOneof": {
					"!type": "wOneofString",
					"wOneofString": "Wrapped oneofStringVal"
				}
			}`,
			wantProto: &schema_testpb.FullSchema{
				WrappedOneof: &schema_testpb.WrappedOneof{
					Type: &schema_testpb.WrappedOneof_WOneofString{
						WOneofString: "Wrapped oneofStringVal",
					},
				},
			},
		}, {
			name: "decimal",
			json: `{
				"decimal": "1.1",
				"rDecimal": ["2.2", "3.3"]
			}`,
			wantProto: &schema_testpb.FullSchema{
				Decimal: &decimal_j5t.Decimal{Value: "1.1"},
				RDecimal: []*decimal_j5t.Decimal{{
					Value: "2.2",
				}, {
					Value: "3.3",
				}},
			},
		}, {
			name: "key",
			json: `{"keyString": "keyVal"}`,
			wantProto: &schema_testpb.FullSchema{
				KeyString: "keyVal",
			},
		}, {
			name: "j5any",
			json: `{
				"j5any": {
					"!type": "test.schema.v1.Bar",
					"value": {
						"barId": "barId"
					}
				}
			}`,
			wantProto: &schema_testpb.FullSchema{
				J5Any: mustJ5Any(t, &schema_testpb.Bar{
					BarId: "barId",
				}, []byte(`{"barId":"barId"}`)),
			},
		}, {
			name: "protoAny",
			json: `{
				"pbany": {
					"!type": "test.schema.v1.Bar",
					"value": {
						"barId": "barId"
					}
				}
			}`,
			wantProto: &schema_testpb.FullSchema{
				Pbany: mustProtoAny(t, &schema_testpb.Bar{
					BarId: "barId",
				}),
			},
		}, {
			name: "polymorphic",
			json: `{
				"polymorph": {
					"!type": "test.schema.v1.Bar",
					"value": {
						"barId": "barId"
					}
				}
			}`,
			wantProto: &schema_testpb.FullSchema{
				Polymorph: &schema_testpb.PolyMessage{
					Value: mustJ5Any(t, &schema_testpb.Bar{
						BarId: "barId",
					}, []byte(`{"barId":"barId"}`)),
				},
			},
		}} {
		t.Run(tc.name, func(t *testing.T) {

			allInputs := append(tc.altInputJSON, tc.json)

			for _, input := range allInputs {
				logIndent(t, "input", input)

				msg := tc.wantProto.ProtoReflect().New().Interface()
				if err := codec.JSONToProto([]byte(input), msg.ProtoReflect()); err != nil {
					t.Fatalf("Input JSONToProto: %s", err)
				}

				t.Logf("GOT proto: %s \n%v\n", msg.ProtoReflect().Descriptor().FullName(), prototext.Format(msg))

				if !proto.Equal(tc.wantProto, msg) {
					a := prototext.Format(tc.wantProto)
					t.Fatalf("FATAL: Expected proto %s\n%v\n", tc.wantProto.ProtoReflect().Descriptor().FullName(), string(a))
				}

				encoded, err := codec.ProtoToJSON(msg.ProtoReflect())
				if err != nil {
					t.Fatalf("FATAL: ProtoToJSON: %s", err)
				}

				logIndent(t, "output", string(encoded))

				CompareJSON(t, []byte(tc.json), encoded)
			}

			for _, query := range tc.queries {

				msg := tc.wantProto.ProtoReflect().New().Interface()
				if err := codec.QueryToProto(query, msg.ProtoReflect()); err != nil {
					t.Fatalf("JSONToProto: %s", err)
				}

				t.Logf("GOT proto: %s \n%v\n", msg.ProtoReflect().Descriptor().FullName(), prototext.Format(msg))

				if !proto.Equal(tc.wantProto, msg) {
					a := prototext.Format(tc.wantProto)
					t.Fatalf("FATAL: Expected proto %s\n%v\n", tc.wantProto.ProtoReflect().Descriptor().FullName(), string(a))
				}

				encoded, err := codec.ProtoToJSON(msg.ProtoReflect())
				if err != nil {
					t.Fatalf("FATAL: ProtoToJSON: %s", err)
				}

				logIndent(t, "output", string(encoded))
			}

		})
	}
}

func logIndent(t *testing.T, label, jsonStr string) {
	t.Helper()
	buffer := &bytes.Buffer{}
	if err := json.Indent(buffer, []byte(jsonStr), " | ", "  "); err != nil {
		t.Log(jsonStr)
		t.Fatalf("%s - invalid JSON (for indent): %s", label, err)
	}
	t.Logf("%s \n | %s\n", label, buffer.String())
}

func TestScalars(t *testing.T) {
	// TODO: The tests above should be rewritten to use this method, then retire
	// testproto.

	type testCase struct {
		desc   protoreflect.MessageDescriptor
		asJSON string
		valMap map[string]protoreflect.Value
	}

	runTest := func(t testing.TB, tc testCase) {
		t.Helper()

		codec := NewCodec()

		msgIn := dynamicpb.NewMessage(tc.desc)

		for key, val := range tc.valMap {
			field := tc.desc.Fields().ByName(protoreflect.Name(key))
			if field == nil {
				t.Fatalf("field %s not found", key)
			}
			msgIn.Set(field, val)
		}

		asJSON, err := codec.ProtoToJSON(msgIn.ProtoReflect())
		if err != nil {
			t.Fatal(err)
		}

		CompareJSON(t, []byte(tc.asJSON), asJSON)

		t.Logf("asJSON: %s", asJSON)

		msgOut := dynamicpb.NewMessage(tc.desc)
		if err := codec.JSONToProto([]byte(tc.asJSON), msgOut.ProtoReflect()); err != nil {
			t.Fatal(err)
		}

		if !proto.Equal(msgIn, msgOut) {
			t.Fatalf("proto equal, expected %s but got %s", msgIn, msgOut)
		}

	}

	t.Run("string", func(t *testing.T) {
		runTest(t, testCase{
			desc:   prototest.SingleMessage(t, "string sString = 1;"),
			asJSON: `{"sString":"val"}`,
			valMap: map[string]protoreflect.Value{
				"sString": protoreflect.ValueOfString("val"),
			},
		})
	})

	t.Run("optional string", func(t *testing.T) {
		msg := prototest.SingleMessage(t, "optional string foo = 1;")

		runTest(t, testCase{
			desc:   msg,
			asJSON: `{"foo":"val"}`,
			valMap: map[string]protoreflect.Value{
				"foo": protoreflect.ValueOfString("val"),
			},
		})

		runTest(t, testCase{
			desc:   msg,
			asJSON: `{}`,
			valMap: map[string]protoreflect.Value{},
		})
	})

	t.Run("date", func(t *testing.T) {
		msg := prototest.SingleMessage(t,
			prototest.WithMessageImports("j5/types/date/v1/date.proto"),
			"j5.types.date.v1.Date date = 1;")
		dateVal := &date_j5t.Date{
			Year:  2020,
			Month: 12,
			Day:   30,
		}
		runTest(t, testCase{
			desc:   msg,
			asJSON: `{"date":"2020-12-30"}`,
			valMap: map[string]protoreflect.Value{
				"date": protoreflect.ValueOfMessage(dateVal.ProtoReflect()),
			},
		})
	})

}

func CompareJSON(t testing.TB, wantSRC, gotSRC []byte) {
	match, diff := JSONEqual(t, wantSRC, gotSRC)
	if !match {
		t.Fatalf("JSON Mismatch: \n%s", diff)
	}
}

func JSONEqual(t testing.TB, wantSRC, gotSRC []byte) (bool, string) {
	t.Helper()
	wantBuff := &bytes.Buffer{}
	if err := json.Indent(wantBuff, wantSRC, "", "  "); err != nil {
		t.Fatalf("want json was invalid: %v", err)
	}

	wantStr := wantBuff.String()

	gotBuff := &bytes.Buffer{}
	if err := json.Indent(gotBuff, gotSRC, "", "  "); err != nil {
		t.Logf("Raw Got String: %s", string(gotSRC))
		t.Fatalf("got json was invalid: %v", err)
	}

	gotStr := gotBuff.String()

	if wantStr == gotStr {
		return true, ""
	}

	outputBuffer := &bytes.Buffer{}

	gotLines := strings.Split(gotStr, "\n")
	wantLines := strings.Split(wantStr, "\n")
	for i := range max(len(wantLines), len(gotLines)) {
		gotLine := ""
		if i < len(gotLines) {
			gotLine = gotLines[i]
		}
		wantLine := ""
		if i < len(wantLines) {
			wantLine = wantLines[i]
		}

		if wantLine != gotLine {
			fmt.Fprintf(outputBuffer, "G: %s\nW: %s\n", gotLine, wantLine)
		} else {
			fmt.Fprintf(outputBuffer, " : %s\n", wantLine)
		}

	}

	return false, outputBuffer.String()

}
