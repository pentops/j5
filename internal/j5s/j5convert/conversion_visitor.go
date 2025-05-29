package j5convert

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/messaging/v1/messaging_j5pb"
	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/internal/j5s/sourcewalk"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type conversionVisitor struct {
	root          *rootContext
	file          *fileContext
	parentContext parentContext
}

func (ww *conversionVisitor) _clone() *conversionVisitor {
	return &conversionVisitor{
		root:          ww.root,
		file:          ww.file,
		parentContext: ww.parentContext,
	}
}

func (rr *conversionVisitor) addErrorf(node sourcewalk.SourceNode, format string, args ...any) {
	err := fmt.Errorf(format, args...)
	rr.addError(node, err)
}

func (rr *conversionVisitor) addError(node sourcewalk.SourceNode, err error) {
	loc := node.GetPos()
	wrapped := errpos.AddPosition(err, loc)
	rr.root.errors = append(rr.root.errors, wrapped)
}

func (ww *conversionVisitor) inMessage(msg *MessageBuilder) *conversionVisitor {
	walk := ww._clone()
	walk.parentContext = msg
	return walk
}

func (ww *conversionVisitor) subPackageFile(subPackage string) *conversionVisitor {
	file := ww.root.subPackageFile(subPackage)
	walk := ww._clone()
	walk.file = file
	walk.parentContext = file
	return walk
}

func (ww *conversionVisitor) resolveType(ref *sourcewalk.RefNode) (*TypeRef, error) {
	if ref == nil {
		return nil, fmt.Errorf("missing ref")
	}

	if ref.Inline {
		// Inline conversions will already exist, they were converted from

		if ref.InlineEnum != nil {
			return enumTypeRef(ref.InlineEnum), nil
		} else if ref.InlineOneof != nil {
			return oneofTypeRef(ref.InlineOneof), nil
		} else if ref.InlineObject != nil {
			return objectTypeRef(ref.InlineObject), nil
		} else {
			return nil, fmt.Errorf("unhandled inline conversion")
		}
	}

	typeRef, err := ww.root.resolveType(ref.Ref)
	if err != nil {
		pos := ref.Source.GetPos()
		err = errpos.AddPosition(err, pos)
		return nil, err
	}

	ww.file.ensureImport(typeRef.File)
	return typeRef, nil
}

func (ww *conversionVisitor) visitFileNode(file *sourcewalk.FileNode) error {
	return file.RangeRootElements(sourcewalk.FileCallbacks{
		SchemaCallbacks: walkerSchemaVisitor(ww),
		TopicFile: func(tn *sourcewalk.TopicFileNode) error {
			subWalk := ww.subPackageFile("topic")
			return subWalk.visitTopicFileNode(tn)
		},
		ServiceFile: func(sn *sourcewalk.ServiceFileNode) error {
			subWalk := ww.subPackageFile("service")
			return subWalk.visitServiceFileNode(sn)
		},
	})
}

func walkerSchemaVisitor(ww *conversionVisitor) sourcewalk.SchemaCallbacks {
	return sourcewalk.SchemaCallbacks{
		Object: func(on *sourcewalk.ObjectNode) error {
			ww.visitObjectNode(on)
			return nil
		},
		Oneof: func(on *sourcewalk.OneofNode) error {
			ww.visitOneofNode(on)
			return nil
		},
		Enum: func(en *sourcewalk.EnumNode) error {
			ww.visitEnumNode(en)
			return nil
		},
		Polymorph: func(pn *sourcewalk.PolymorphNode) error {
			ww.visitPolymorphNode(pn)
			return nil
		},
	}
}

func (ww *conversionVisitor) visitTopicFileNode(tn *sourcewalk.TopicFileNode) error {
	return tn.Accept(sourcewalk.TopicFileCallbacks{
		Topic: func(tn *sourcewalk.TopicNode) error {
			ww.visitTopicNode(tn)
			return nil
		},
		Object: func(on *sourcewalk.ObjectNode) error {
			ww.visitObjectNode(on)
			return nil
		},
	})
}

func (ww *conversionVisitor) visitTopicNode(tn *sourcewalk.TopicNode) {
	desc := &descriptorpb.ServiceDescriptorProto{
		Name:    gl.Ptr(tn.Name),
		Options: &descriptorpb.ServiceOptions{},
	}

	proto.SetExtension(desc.Options, messaging_j5pb.E_Service, tn.ServiceConfig)

	for _, method := range tn.Methods {
		rpcDesc := &descriptorpb.MethodDescriptorProto{
			Name:       gl.Ptr(method.Name),
			OutputType: gl.Ptr(googleProtoEmptyType),
			InputType:  gl.Ptr(method.Request),
		}
		desc.Method = append(desc.Method, rpcDesc)
	}

	ww.file.ensureImport(messagingAnnotationsImport)
	ww.file.ensureImport(googleProtoEmptyImport)
	ww.file.addService(&serviceBuilder{
		desc: desc,
	})
}

func (ww *conversionVisitor) visitObjectNode(node *sourcewalk.ObjectNode) {

	message := blankMessage(node.Name)

	if node.Entity != nil {
		ww.file.ensureImport(j5ExtImport)
		proto.SetExtension(message.descriptor.Options, ext_j5pb.E_Psm, &ext_j5pb.PSMOptions{
			EntityName: node.Entity.Entity,
			EntityPart: node.Entity.Part.Enum(),
		})
	}

	objectType := &ext_j5pb.ObjectMessageOptions{}

	fqn := ww.file.fullyQualifiedName(node)

	for _, pm := range node.PolymorphMember {
		tt, err := ww.resolveType(pm)
		if err != nil {
			ww.addError(pm.Source, err)
			continue
		}

		if tt.Polymorph == nil {
			ww.addErrorf(pm.Source, "type %q is not a polymorph", tt.debugName())
			continue
		}

		if !slices.Contains(tt.Polymorph.Members, fqn) {
			names := make([]string, len(tt.Polymorph.Members))
			for idx, member := range tt.Polymorph.Members {
				names[idx] = fmt.Sprintf("%q", member)
			}
			ww.addErrorf(pm.Source, "type %q is not a polymorph member of %s.%s have %s", fqn, tt.Package, tt.Name, strings.Join(names, ", "))
			continue
		}

	}

	ww.file.ensureImport(j5ExtImport)
	ext := &ext_j5pb.MessageOptions{
		Type: &ext_j5pb.MessageOptions_Object{
			Object: objectType,
		},
	}
	proto.SetExtension(message.descriptor.Options, ext_j5pb.E_Message, ext)

	message.comment([]int32{}, node.Description)

	inMessageWalker := ww.inMessage(message)

	err := node.RangeProperties(&sourcewalk.PropertyCallbacks{
		SchemaVisitor: walkerSchemaVisitor(inMessageWalker),
		Property: func(node *sourcewalk.PropertyNode) error {

			propertyDesc, err := buildProperty(inMessageWalker, node)
			if err != nil {
				ww.addError(node.Source, err)
			}

			// Take the index (prior to append len == index), not the field number
			locPath := []int32{2, int32(len(message.descriptor.Field))}
			message.comment(locPath, node.Schema.Description)
			message.descriptor.Field = append(message.descriptor.Field, propertyDesc)

			return nil
		},
	})
	if err != nil {
		ww.addError(node.Source, err)
	}

	if node.HasNestedSchemas() {
		subContext := ww.inMessage(message)
		if err := node.RangeNestedSchemas(walkerSchemaVisitor(subContext)); err != nil {
			ww.addError(node.Source, err)
		}
	}

	ww.parentContext.addMessage(message)
}

func (ww *conversionVisitor) visitOneofNode(node *sourcewalk.OneofNode) {
	schema := node.Schema
	if schema.Name == "" {
		ww.addErrorf(node.Source, "missing object name")
	}

	message := blankMessage(schema.Name)
	message.descriptor.OneofDecl = []*descriptorpb.OneofDescriptorProto{{
		Name: gl.Ptr("type"),
	}}
	message.comment([]int32{}, schema.Description)

	oneofType := &ext_j5pb.OneofMessageOptions{}

	ww.file.ensureImport(j5ExtImport)
	ext := &ext_j5pb.MessageOptions{
		Type: &ext_j5pb.MessageOptions_Oneof{
			Oneof: oneofType,
		},
	}
	proto.SetExtension(message.descriptor.Options, ext_j5pb.E_Message, ext)

	err := node.RangeProperties(&sourcewalk.PropertyCallbacks{
		SchemaVisitor: walkerSchemaVisitor(ww.inMessage(message)),
		Property: func(node *sourcewalk.PropertyNode) error {
			schema := node.Schema
			schema.ProtoField = []int32{node.Number}

			propertyDesc, err := buildProperty(ww, node)
			if err != nil {
				ww.addError(node.Source, err)
				return nil
			}
			propertyDesc.OneofIndex = gl.Ptr(int32(0))

			// Take the index (prior to append len == index), not the field number
			locPath := []int32{2, int32(len(message.descriptor.Field))}
			message.comment(locPath, schema.Description)
			message.descriptor.Field = append(message.descriptor.Field, propertyDesc)
			return nil
		},
	})
	if err != nil {
		ww.addError(node.Source, err)
	}

	if node.HasNestedSchemas() {
		subContext := ww.inMessage(message)
		if err := node.RangeNestedSchemas(walkerSchemaVisitor(subContext)); err != nil {
			ww.addError(node.Source, err)
		}
	}

	ww.parentContext.addMessage(message)
}

func (ww *conversionVisitor) resolvePolymorphIncludes(includes []*sourcewalk.RefNode) ([]string, error) {
	members := make([]string, 0)
	for _, include := range includes {
		typeRef, err := ww.root.resolveType(include.Ref)
		if err != nil {
			return nil, err
		}

		if typeRef.Polymorph == nil {
			return nil, fmt.Errorf("type %q is not a polymorph", typeRef.debugName())
		}

		members = append(members, typeRef.Polymorph.Members...)

		subIncluded, err := ww.resolvePolymorphIncludes(typeRef.Polymorph.Includes)
		if err != nil {
			return nil, err
		}
		members = append(members, subIncluded...)
	}

	return members, nil
}

func (ww *conversionVisitor) visitPolymorphNode(node *sourcewalk.PolymorphNode) {
	message := blankMessage(node.Name)

	message.comment([]int32{}, node.Description)

	valueField := &descriptorpb.FieldDescriptorProto{
		Name:     proto.String("value"),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
		Number:   proto.Int32(1),
		TypeName: proto.String(".j5.types.any.v1.Any"),
		JsonName: proto.String("value"),
	}
	message.descriptor.Field = []*descriptorpb.FieldDescriptorProto{valueField}
	pmm := &ext_j5pb.PolymorphMessageOptions{}

	extraMembers, err := ww.resolvePolymorphIncludes(node.Includes)
	if err != nil {
		ww.addError(node.Source, err)
	} else {
		pmm.Members = append(node.Members, extraMembers...)
	}

	ww.file.ensureImport(j5AnyImport)
	ww.file.ensureImport(j5ExtImport)

	ext := &ext_j5pb.MessageOptions{
		Type: &ext_j5pb.MessageOptions_Polymorph{
			Polymorph: pmm,
		},
	}
	proto.SetExtension(message.descriptor.Options, ext_j5pb.E_Message, ext)
	ww.parentContext.addMessage(message)

}

func (ww *conversionVisitor) visitEnumNode(node *sourcewalk.EnumNode) {

	desc := &descriptorpb.EnumDescriptorProto{
		Name: gl.Ptr(node.Schema.Name),
		Value: []*descriptorpb.EnumValueDescriptorProto{{
			Name:   gl.Ptr(fmt.Sprintf("%sUNSPECIFIED", node.Prefix)),
			Number: gl.Ptr(int32(0)),
		}},
	}

	var comments commentSet

	if node.Schema.Description != "" {
		comments.comment([]int32{}, node.Schema.Description)
	}

	if node.Schema.Info != nil {
		ext := &ext_j5pb.EnumOptions{}

		for _, field := range node.Schema.Info {
			ext.InfoFields = append(ext.InfoFields, &ext_j5pb.EnumInfoField{
				Name:        field.Name,
				Label:       field.Label,
				Description: field.Description,
			})
		}

		desc.Options = &descriptorpb.EnumOptions{}
		proto.SetExtension(desc.Options, ext_j5pb.E_Enum, ext)
	}

	for _, src := range node.Options {
		value := &descriptorpb.EnumValueDescriptorProto{
			Name:   gl.Ptr(src.Name),
			Number: gl.Ptr(src.Number),
		}

		if len(src.Info) > 0 {
			value.Options = &descriptorpb.EnumValueOptions{}
			proto.SetExtension(value.Options, ext_j5pb.E_EnumValue, &ext_j5pb.EnumValueOptions{
				Info: src.Info,
			})
		}

		if src.Number == 0 {
			desc.Value[0] = value
		} else {
			desc.Value = append(desc.Value, value)
		}
		if src.Description != "" {
			comments.comment([]int32{2, src.Number}, src.Description)
		}

	}

	ww.parentContext.addEnum(desc, comments)
}
