package anydesc

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/pentops/jsonapi/codec"
	"github.com/pentops/jsonapi/gen/test/bar/v1/bar_testpb"
	"github.com/pentops/jsonapi/gen/test/baz/v1/baz_testpb"
	"github.com/pentops/jsonapi/gen/test/foo/v1/foo_testpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestReflection(t *testing.T) {
	tt := time.Now()
	msg := &baz_testpb.Baz{
		Field: "value",
		Bar: &bar_testpb.Bar{
			Field: "bar",
		},
		BarEnum:   bar_testpb.BarEnum_BAR_ENUM_FOO,
		Timestamp: timestamppb.New(tt),
	}

	flatDesc, err := BuildAny(FlattenOptions{}, msg)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Flat Descriptor: %v", prototext.Format(flatDesc))
	descBytes, err := proto.Marshal(flatDesc)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Flat Descriptor has %d bytes", len(descBytes))

	ff, err := protodesc.NewFile(flatDesc.FileDescriptor, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}

	rootMsg := ff.Messages().ByName("ROOT")
	dyn := dynamicpb.NewMessage(rootMsg)
	if err := proto.Unmarshal(flatDesc.Value, dyn); err != nil {
		t.Fatal(err.Error())
	}

	encoded, err := codec.Encode(codec.Options{}, dyn)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(encoded))
	t.Logf("Encoded has %d bytes", len(encoded))

	decoded := map[string]interface{}{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "value", decoded["field"])
	assert.Equal(t, "bar", decoded["bar"].(map[string]interface{})["field"])
	assert.Equal(t, "BAR_ENUM_FOO", decoded["barEnum"])
	assert.Equal(t, tt.In(time.UTC).Format(time.RFC3339Nano), decoded["timestamp"])
	if decoded["field"] != "value" {
		t.Fatalf("expected field to be 'value', got %v", decoded["field"])
	}

}

func TestJ5AnnotationsPass(t *testing.T) {

	msg := &foo_testpb.PostFooRequest{
		SString: "value",
		Flattened: &foo_testpb.FlattenedMessage{
			FieldFromFlattened: "flattened",
		},
	}

	flatDesc, err := BuildAny(FlattenOptions{}, msg)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Flat Descriptor: %v", prototext.Format(flatDesc))
	descBytes, err := proto.Marshal(flatDesc)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Flat Descriptor has %d bytes", len(descBytes))

	aa, err := NewFromAny(flatDesc)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := codec.Encode(codec.Options{}, aa)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(encoded))

	decoded := map[string]interface{}{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "flattened", decoded["fieldFromFlattened"])
	assert.Equal(t, "value", decoded["sString"])

}
