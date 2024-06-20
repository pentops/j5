package codec

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pentops/flowtest/prototest"
	"github.com/pentops/j5/gen/test/foo/v1/foo_testpb"
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
			wantProto: &foo_testpb.PostFooRequest{
				SString: "nameVal",
				OString: proto.String("otherNameVal"),
				RString: []string{"r1", "r2"},

				SFloat: 1.1,
				OFloat: proto.Float32(2.2),
				RFloat: []float32{3.3, 4.4},

				Enum: foo_testpb.Enum_ENUM_VALUE1,
				REnum: []foo_testpb.Enum{
					foo_testpb.Enum_ENUM_VALUE1,
					foo_testpb.Enum_ENUM_VALUE2,
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

			wantProto: &foo_testpb.PostFooRequest{
				Enum: foo_testpb.Enum_ENUM_VALUE1,
				REnum: []foo_testpb.Enum{
					foo_testpb.Enum_ENUM_VALUE1,
					foo_testpb.Enum_ENUM_VALUE2,
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
			wantProto: &foo_testpb.PostFooRequest{
				SBytes: []byte("sBytes"),
				RBytes: [][]byte{[]byte("rBytes1"), []byte("rBytes2")},
			},
		}, {
			name: "map",
			json: `{ "mapStringString": {
				"k1": "val1"
			} }`,
			// TODO: Can only test one key this way while maps are unordered
			wantProto: &foo_testpb.PostFooRequest{
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
			wantProto: &foo_testpb.PostFooRequest{
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
					"id": "barId",
					"name": "barName",
					"field": "barField"
				},
				"rBars": [{
					"id": "bar1"
				}, {
					"id": "bar2"
				}]
			}`,
			wantProto: &foo_testpb.PostFooRequest{
				SBar: &foo_testpb.Bar{
					Id:    "barId",
					Name:  "barName",
					Field: "barField",
				},
				RBars: []*foo_testpb.Bar{{
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
			wantProto: &foo_testpb.PostFooRequest{
				Flattened: &foo_testpb.FlattenedMessage{
					FieldFromFlattened:   "fieldFromFlattenedVal",
					Field_2FromFlattened: "field2FromFlattenedVal",
				},
			},
		}, {
			name: "anon oneof",
			json: `{
				"oneofString": "oneofStringVal"
			}`,
			wantProto: &foo_testpb.PostFooRequest{
				AnonOneof: &foo_testpb.PostFooRequest_OneofString{
					OneofString: "oneofStringVal",
				},
			},
		}, {
			name: "exposed oneof",
			json: `{
				"exposedOneof": {
					"!type": "exposedString",
					"exposedString": "oneofStringVal"
				}
			}`,
			wantProto: &foo_testpb.PostFooRequest{
				ExposedOneof: &foo_testpb.PostFooRequest_ExposedString{
					ExposedString: "oneofStringVal",
				},
			},
		}, {
			name: "oneof wrapper",
			json: `{
				"wrappedOneof": {
					"!type": "oneofString",
					"oneofString": "oneofStringVal"
				}
			}`,
			wantProto: &foo_testpb.PostFooRequest{
				WrappedOneof: &foo_testpb.WrappedOneof{
					Type: &foo_testpb.WrappedOneof_OneofString{
						OneofString: "oneofStringVal",
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
			wantProto: &foo_testpb.PostFooRequest{
				NestedExposedOneof: &foo_testpb.NestedExposed{
					Type: &foo_testpb.NestedExposed_De1{
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
			wantProto: &foo_testpb.PostFooRequest{
				NestedExposedOneof: &foo_testpb.NestedExposed{
					Type: &foo_testpb.NestedExposed_De3{
						De3: &foo_testpb.NestedExposed{
							Type: &foo_testpb.NestedExposed_De1{
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
			wantProto: &foo_testpb.PostFooRequest{
				NestedExposedOneof: &foo_testpb.NestedExposed{
					Type: &foo_testpb.NestedExposed_De3{
						De3: &foo_testpb.NestedExposed{
							Type: &foo_testpb.NestedExposed_De3{
								De3: &foo_testpb.NestedExposed{
									Type: &foo_testpb.NestedExposed_De1{
										De1: "de1Val",
									},
								},
							},
						},
					},
				},
			},
		}, {
			name: "backport test from proxy",
			json: `{
				"filters":[{
					"type":{
						"!type": "field",
						"field": {
							"name":"idVal",
							"type":{
								"!type": "value",
								"value":"f481d62c-72ff-487b-ba03-50a4a6da83b7"
							}
						}
					}
				}]
			}`,
			altInputJSON: []string{
				`{"filters":[{"type":{"field":{"name":"idVal","type":{"value":"f481d62c-72ff-487b-ba03-50a4a6da83b7"}}}}]}`,
			},
			wantProto: &foo_testpb.QueryRequest{
				Filters: []*foo_testpb.QueryRequest_Filter{{
					Type: &foo_testpb.QueryRequest_Filter_Field{
						Field: &foo_testpb.QueryRequest_Field{
							Name: "idVal",
							Type: &foo_testpb.QueryRequest_Field_Value{
								Value: "f481d62c-72ff-487b-ba03-50a4a6da83b7",
							},
						},
					},
				}},
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
			wantProto: &foo_testpb.PostFooRequest{
				NestedExposedOneofs: []*foo_testpb.NestedExposed{{
					Type: &foo_testpb.NestedExposed_De1{
						De1: "de1Val",
					},
				}, {
					Type: &foo_testpb.NestedExposed_De2{
						De2: "de2Val",
					},
				}},
			},
		}} {
		t.Run(tc.name, func(t *testing.T) {

			allInputs := append(tc.altInputJSON, tc.json)

			codec := NewCodec()
			for _, input := range allInputs {

				buffer := &bytes.Buffer{}
				if err := json.Indent(buffer, []byte(input), "", "  "); err != nil {
					t.Fatalf("invalid test case: %s", err)
				}

				t.Log(input)
				msg := tc.wantProto.ProtoReflect().New().Interface()
				if err := codec.JSONToProto([]byte(input), msg.ProtoReflect()); err != nil {
					t.Fatal(err)
				}

				t.Logf("protojson format: \n%v\n", protojson.Format(msg))

				if !proto.Equal(tc.wantProto, msg) {
					a := protojson.Format(tc.wantProto)
					b := protojson.Format(msg)
					t.Fatalf("expected \n%v but got\n%v", string(a), string(b))
				}

				encoded, err := codec.ProtoToJSON(msg.ProtoReflect())
				if err != nil {
					t.Fatal(err)
				}

				CompareJSON(t, []byte(tc.json), encoded)
			}

		})
	}
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
