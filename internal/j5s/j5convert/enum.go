package j5convert

import (
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/internal/j5s/sourcewalk"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type enumBuilder struct {
	desc   *descriptorpb.EnumDescriptorProto
	prefix string

	commentSet
}

func (e *enumBuilder) addValue(schema sourcewalk.EnumOption) {
	value := &descriptorpb.EnumValueDescriptorProto{
		Name:   gl.Ptr(schema.Name),
		Number: gl.Ptr(schema.Number),
	}

	if len(schema.Info) > 0 {
		value.Options = &descriptorpb.EnumValueOptions{}
		proto.SetExtension(value.Options, ext_j5pb.E_EnumValue, &ext_j5pb.EnumValueOptions{
			Info: schema.Info,
		})
	}

	if schema.Number == 0 {
		e.desc.Value[0] = value
	} else {
		e.desc.Value = append(e.desc.Value, value)
	}
	if schema.Description != "" {
		e.comment([]int32{2, schema.Number}, schema.Description)
	}

}
