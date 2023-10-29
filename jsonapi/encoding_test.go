package jsonapi

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
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pentops/custom-proto-api/testproto/gen/testpb"
)

func TestUnmarshal(t *testing.T) {

	testTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name      string
		wantProto proto.Message
		json      string
		options   Options
	}{
		{
			name: "scalars",
			options: Options{
				ShortEnums: &ShortEnumsOption{},
			},
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
			wantProto: &testpb.PostFooRequest{
				SString: "nameVal",
				OString: proto.String("otherNameVal"),
				RString: []string{"r1", "r2"},

				SFloat: 1.1,
				OFloat: proto.Float32(2.2),
				RFloat: []float32{3.3, 4.4},

				Enum: testpb.Enum_ENUM_VALUE1,
				REnum: []testpb.Enum{
					testpb.Enum_ENUM_VALUE1,
					testpb.Enum_ENUM_VALUE2,
				},
			},
		}, {
			name: "long enums",
			options: Options{
				ShortEnums: nil,
			},
			json: `{
				"enum": "ENUM_VALUE1",
				"rEnum": ["ENUM_VALUE1", "ENUM_VALUE2"]
			}`,
			wantProto: &testpb.PostFooRequest{
				Enum: testpb.Enum_ENUM_VALUE1,
				REnum: []testpb.Enum{
					testpb.Enum_ENUM_VALUE1,
					testpb.Enum_ENUM_VALUE2,
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
			wantProto: &testpb.PostFooRequest{
				SBytes: []byte("sBytes"),
				RBytes: [][]byte{[]byte("rBytes1"), []byte("rBytes2")},
			},
		}, {
			name: "map",
			json: `{ "mapStringString": {
				"k1": "val1"
			} }`,
			// TODO: Can only test one key this way while maps are unordered
			wantProto: &testpb.PostFooRequest{
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
			wantProto: &testpb.PostFooRequest{
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
			wantProto: &testpb.PostFooRequest{
				SBar: &testpb.Bar{
					Id:    "barId",
					Name:  "barName",
					Field: "barField",
				},
				RBars: []*testpb.Bar{{
					Id: "bar1",
				}, {
					Id: "bar2",
				}},
			},
		}, {
			name: "naked oneof",
			json: `{
				"oneofString": "oneofStringVal"
			}`,
			wantProto: &testpb.PostFooRequest{
				NakedOneof: &testpb.PostFooRequest_OneofString{
					OneofString: "oneofStringVal",
				},
			},
		}, {
			name: "wrap naked oneof",
			json: `{
				"nakedOneof": {
					"oneofString": "oneofStringVal"
				}
			}`,
			options: Options{
				WrapOneof: true,
			},
			wantProto: &testpb.PostFooRequest{
				NakedOneof: &testpb.PostFooRequest_OneofString{
					OneofString: "oneofStringVal",
				},
			},
		}, {
			name: "no double wrap oneof",
			json: `{
				"wrappedOneof": {
					"oneofString": "oneofStringVal"
				}
			}`,
			options: Options{
				WrapOneof: true,
			},
			wantProto: &testpb.PostFooRequest{
				WrappedOneof: &testpb.WrappedOneof{
					Type: &testpb.WrappedOneof_OneofString{
						OneofString: "oneofStringVal",
					},
				},
			},
		}} {
		t.Run(tc.name, func(t *testing.T) {

			msg := &testpb.PostFooRequest{}
			if err := Decode(tc.options, []byte(tc.json), msg.ProtoReflect()); err != nil {
				t.Fatal(err)
			}

			t.Logf("protojson format: \n%v\n", protojson.Format(msg))

			if !proto.Equal(tc.wantProto, msg) {
				a := protojson.Format(tc.wantProto)
				b := protojson.Format(msg)
				t.Fatalf("expected \n%v but got\n%v", string(a), string(b))
			}

			encoded, err := Encode(tc.options, msg.ProtoReflect())
			if err != nil {
				t.Fatal(err)
			}

			CompareJSON(t, []byte(tc.json), encoded)

		})
	}
}

func CompareJSON(t testing.TB, wantSRC, gotSRC []byte) {
	wantBuff := &bytes.Buffer{}
	if err := json.Indent(wantBuff, wantSRC, "", "  "); err != nil {
		t.Fatalf("want json was invalid: %v", err)
	}

	wantStr := wantBuff.String()

	gotBuff := &bytes.Buffer{}
	if err := json.Indent(gotBuff, gotSRC, "", "  "); err != nil {
		t.Log(string(gotSRC))
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
