package codec

import (
	"testing"

	"github.com/pentops/j5/gen/test/foo/v1/foo_testpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestEnumPrefix(t *testing.T) {

	descriptor := foo_testpb.Enum(0).Descriptor()
	prefix := enumPrefix(descriptor)

	assert.Equal(t, "ENUM_", prefix)

	val, err := (&decoder{}).decodeEnum(descriptor, "VALUE1")
	assert.NoError(t, err)
	assert.Equal(t, protoreflect.EnumNumber(1), val)

	encoded, err := (&encoder{}).encodeEnum(descriptor, protoreflect.EnumNumber(1))
	assert.NoError(t, err)
	assert.Equal(t, "VALUE1", encoded)

}
