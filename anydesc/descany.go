package anydesc

import (
	"fmt"
	"strings"

	"github.com/pentops/jsonapi/codec"
	"github.com/pentops/jsonapi/gen/j5/anydesc/v1/anydesc_j5pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type FlattenOptions struct {
	ExcludePrefixes []string
}

func BuildAny(opts FlattenOptions, msg proto.Message) (*anydesc_j5pb.Any, error) {
	refl := msg.ProtoReflect().Descriptor()
	flat, err := FlattenDescriptor(opts, refl)
	if err != nil {
		return nil, err
	}

	protoData, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return &anydesc_j5pb.Any{
		TypeUrl:        fmt.Sprintf("type.googleapis.com/%s", refl.FullName()),
		Value:          protoData,
		FileDescriptor: flat,
	}, nil

}

func NewFromAny(flatDesc *anydesc_j5pb.Any) (*dynamicpb.Message, error) {

	ff, err := protodesc.NewFile(flatDesc.FileDescriptor, protoregistry.GlobalFiles)
	if err != nil {
		return nil, err
	}

	rootMsg := ff.Messages().ByName("ROOT")
	dyn := dynamicpb.NewMessage(rootMsg)

	return dyn, proto.Unmarshal(flatDesc.Value, dyn)
}

func MarshalAnyToJSON(flatDesc *anydesc_j5pb.Any, opts codec.Options) ([]byte, error) {

	dyn, err := NewFromAny(flatDesc)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(flatDesc.Value, dyn); err != nil {
		return nil, err
	}

	return codec.Encode(opts, dyn)
}

func FlattenDescriptor(opts FlattenOptions, msg protoreflect.MessageDescriptor) (*descriptorpb.FileDescriptorProto, error) {
	if opts.ExcludePrefixes == nil {
		opts.ExcludePrefixes = []string{
			"google.protobuf.",
		}
	}

	ff := &flattener{
		opts: &opts,
		file: &descriptorpb.FileDescriptorProto{
			Name:    proto.String("anydesc"),
			Package: proto.String("anydesc"),
			Syntax:  proto.String("proto3"),
		},
		seen: make(map[protoreflect.FullName]string),
	}

	ff.addMessage(msg, true)

	return ff.file, nil
}

type flattener struct {
	opts *FlattenOptions

	// is a protoreflect message of a file descriptor. :-S
	file *descriptorpb.FileDescriptorProto
	seen map[protoreflect.FullName]string
}

func (ff *flattener) newType(thing protoreflect.Descriptor) (string, *string) {
	name := thing.FullName()
	if name, ok := ff.seen[name]; ok {
		return name, nil
	}

	for _, prefix := range ff.opts.ExcludePrefixes {
		// don't alias or include these, return the real path and include the
		// dependency import
		if strings.HasPrefix(string(name), prefix) {
			filePath := thing.ParentFile().Path()
			found := false
			for _, dep := range ff.file.Dependency {
				if dep == filePath {
					found = true
					break
				}
			}
			if !found {
				ff.file.Dependency = append(ff.file.Dependency, filePath)
			}

			importName := "." + string(name)
			ff.seen[name] = importName
			return importName, nil
		}
	}

	newName := strings.ReplaceAll(string(name), ".", "_")
	importName := ".anydesc." + newName
	ff.seen[name] = importName
	return importName, &newName
}

func (ff *flattener) addEnum(enum protoreflect.EnumDescriptor) string {

	if newName, ok := ff.seen[enum.FullName()]; ok {
		return newName
	}

	importAs, newName := ff.newType(enum)
	if newName == nil {
		return importAs
	}

	descriptor := protodesc.ToEnumDescriptorProto(enum)
	descriptor.Name = newName

	ff.file.EnumType = append(ff.file.EnumType, descriptor)

	return importAs
}

func (ff *flattener) addMessage(msg protoreflect.MessageDescriptor, root bool) string {
	var newName *string
	var importAs string
	if root {
		if importSeen, ok := ff.seen[msg.FullName()]; ok {
			return importSeen
		}
		rootName := "ROOT"
		newName = &rootName
		importAs = ".anydesc.ROOT"
		ff.seen[msg.FullName()] = importAs

	} else {
		importAs, newName = ff.newType(msg)
		if newName == nil {
			return importAs
		}
	}

	descriptor := ff.buildMessage(msg, *newName)

	ff.file.MessageType = append(ff.file.MessageType, descriptor)

	return importAs
}

func (ff *flattener) buildMessage(msg protoreflect.MessageDescriptor, newName string) *descriptorpb.DescriptorProto {

	fields := msg.Fields()

	descriptor := protodesc.ToDescriptorProto(msg)
	descriptor.Name = &newName

	// NestedTypes will be included only if they are used.
	descriptor.NestedType = nil

	for idx, field := range descriptor.Field {
		// fields without a type name are not messages or enums, and don't need
		// to be flattened.
		if field.TypeName == nil {
			continue
		}

		fieldRefl := fields.Get(idx)
		switch fieldRefl.Kind() {
		case protoreflect.MessageKind:

			if fieldRefl.IsMap() {
				mapRefl := fieldRefl.Message()
				mapMsg := ff.buildMessage(mapRefl, string(mapRefl.Name()))

				mapName := newName + "." + string(mapRefl.Name())
				field.TypeName = proto.String(string(mapName))
				descriptor.NestedType = append(descriptor.NestedType, mapMsg)

			} else {

				newName := ff.addMessage(fieldRefl.Message(), false)
				field.TypeName = &newName
			}

		case protoreflect.EnumKind:
			newName := ff.addEnum(fieldRefl.Enum())
			field.TypeName = &newName

		}

		// not really important given the context
		field.JsonName = nil

	}

	return descriptor

}
