package codec

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pentops/flowtest/prototest"
	"github.com/pentops/j5/gen/test/schema/v1/schema_testpb"
	"github.com/pentops/j5/j5types/date_j5t"
)

func TestUnmarshal(t *testing.T) {

	testTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name         string
		wantProto    proto.Message
		json         string
		altInputJSON []string // when set, the output can be different to the input, but all outputs should be the same
	}{
		{
			name: "scalars",

			json: `{
				"sString": "nameVal",
				"oString": "otherNameVal",
				"rString": ["r1", "r2"],
				"sFloat": 1.1,
				"oFloat": 2.2,
				"rFloat": [3.3, 4.4],
				"enum": "VALUE1",
				"rEnum": ["VALUE1", "VALUE2"]
			}`,
			wantProto: &schema_testpb.FullSchema{
				SString: "nameVal",
				OString: proto.String("otherNameVal"),
				RString: []string{"r1", "r2"},

				SFloat: 1.1,
				OFloat: proto.Float32(2.2),
				RFloat: []float32{3.3, 4.4},

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
				base64.URLEncoding.EncodeToString([]byte("rBytes2")) +
				`"]
			}`,
			wantProto: &schema_testpb.FullSchema{
				SBytes: []byte("sBytes"),
				RBytes: [][]byte{[]byte("rBytes1"), []byte("rBytes2")},
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
			name: "well known types",
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
			name: "nested messages",
			json: `{
				"sBar": {
					"id": "barId"
				},
				"rBars": [{
					"id": "bar1"
				}, {
					"id": "bar2"
				}]
			}`,
			wantProto: &schema_testpb.FullSchema{
				SBar: &schema_testpb.Bar{
					Id: "barId",
				},
				RBars: []*schema_testpb.Bar{{
					Id: "bar1",
				}, {
					Id: "bar2",
				}},
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
			name: "exposed oneof",
			json: `{
				"exposedOneof": {
					"!type": "exposedString",
					"exposedString": "stringVal"
				}
			}`,
			wantProto: &schema_testpb.FullSchema{
				ExposedOneof: &schema_testpb.FullSchema_ExposedString{
					ExposedString: "stringVal",
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
			name: "exposed oneof in nested message",
			json: `{
				"nestedExposedOneof": {
					"type": {
						"!type": "de1",
						"de1": "de1Val"
					}
				}
			  }`,
			wantProto: &schema_testpb.FullSchema{
				NestedExposedOneof: &schema_testpb.NestedExposed{
					Type: &schema_testpb.NestedExposed_De1{
						De1: "de1Val",
					},
				},
			},
		}, {
			name: "lightly recursive nested oneof",
			json: `{
				"nestedExposedOneof": {
					"type": {
						"!type": "de3",
						"de3": {
							"type": {
								"!type": "de1",
								"de1": "de1Val"
							}
						}
					}
				}
			}`,
			wantProto: &schema_testpb.FullSchema{
				NestedExposedOneof: &schema_testpb.NestedExposed{
					Type: &schema_testpb.NestedExposed_De3{
						De3: &schema_testpb.NestedExposed{
							Type: &schema_testpb.NestedExposed_De1{
								De1: "de1Val",
							},
						},
					},
				},
			},
		}, {
			name: "recursive nested oneof",
			json: `{
				"nestedExposedOneof": {
					"type": {
						"!type": "de3",
						"de3": {
							"type": {
								"!type": "de3",
								"de3": {
									"type": {
										"!type": "de1",
										"de1": "de1Val"
									}
								}
							}
						}
					}
				}
			}`,
			wantProto: &schema_testpb.FullSchema{
				NestedExposedOneof: &schema_testpb.NestedExposed{
					Type: &schema_testpb.NestedExposed_De3{
						De3: &schema_testpb.NestedExposed{
							Type: &schema_testpb.NestedExposed_De3{
								De3: &schema_testpb.NestedExposed{
									Type: &schema_testpb.NestedExposed_De1{
										De1: "de1Val",
									},
								},
							},
						},
					},
				},
			},
		}, {
			name: "repeated exposed oneof in nested message",
			json: `{
				"nestedExposedOneofs": [{
					"type": {
						"!type": "de1",
						"de1": "de1Val"
					}
				}, {
					"type": {
						"!type": "de2",
						"de2": "de2Val"
					}
				}]
			  }`,
			wantProto: &schema_testpb.FullSchema{
				NestedExposedOneofs: []*schema_testpb.NestedExposed{{
					Type: &schema_testpb.NestedExposed_De1{
						De1: "de1Val",
					},
				}, {
					Type: &schema_testpb.NestedExposed_De2{
						De2: "de2Val",
					},
				}},
			},
		}} {
		t.Run(tc.name, func(t *testing.T) {

			allInputs := append(tc.altInputJSON, tc.json)

			codec := NewCodec()
			for _, input := range allInputs {
				/*
					schema, err := codec.schemaSet.SchemaObject(tc.wantProto.ProtoReflect().Descriptor())
					if err != nil {
						t.Fatal(err)
					}
					t.Logf("SCHEMA: %s", protojson.Format(schema))
				*/
				logIndent(t, "input", input)

				msg := tc.wantProto.ProtoReflect().New().Interface()
				if err := codec.JSONToProto([]byte(input), msg.ProtoReflect()); err != nil {
					t.Fatal(err)
				}

				t.Logf("got decoded proto: %s \n%v\n", msg.ProtoReflect().Descriptor().FullName(), prototext.Format(msg))

				if !proto.Equal(tc.wantProto, msg) {
					a := prototext.Format(tc.wantProto)
					t.Fatalf("expected proto %s\n%v\n", tc.wantProto.ProtoReflect().Descriptor().FullName(), string(a))
				}

				encoded, err := codec.ProtoToJSON(msg.ProtoReflect())
				if err != nil {
					t.Fatal(err)
				}

				logIndent(t, "output", string(encoded))

				CompareJSON(t, []byte(tc.json), encoded)
			}

		})
	}
}

func logIndent(t *testing.T, label, jsonStr string) {
	buffer := &bytes.Buffer{}
	if err := json.Indent(buffer, []byte(jsonStr), " | ", "  "); err != nil {
		t.Fatalf("invalid test case: %s", err)
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
		return
	}

	outputBuffer := &bytes.Buffer{}

	gotLines := strings.Split(gotStr, "\n")
	wantLines := strings.Split(wantStr, "\n")
	for i, wantLine := range wantLines {
		gotLine := ""
		if i < len(gotLines) {
			gotLine = gotLines[i]
		}

		if wantLine != gotLine {
			fmt.Fprintf(outputBuffer, "G: %s\nW: %s\n", gotLine, wantLine)
		} else {
			fmt.Fprintf(outputBuffer, " : %s\n", wantLine)
		}

	}

	t.Fatalf("JSON Mismatch: \n%s", outputBuffer.String())

}
