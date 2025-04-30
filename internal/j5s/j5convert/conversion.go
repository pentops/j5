package j5convert

import (
	"fmt"
	"log"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/messaging/v1/messaging_j5pb"
	"github.com/pentops/j5/internal/bcl/errpos"
	"github.com/pentops/j5/internal/j5s/sourcewalk"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// parentContext is a file's root, or message, which can hold messages and
// enums. Implemented by FileBuilder and MessageBuilder.
type parentContext interface {
	addMessage(*MessageBuilder)
	addEnum(*enumBuilder)
}

type fieldContext struct {
	name string
}

type conversionVisitor struct {
	root          *rootContext
	file          *fileContext
	field         *fieldContext
	parentContext parentContext
}

func (ww *conversionVisitor) _clone() *conversionVisitor {
	return &conversionVisitor{
		root:          ww.root,
		file:          ww.file,
		field:         ww.field,
		parentContext: ww.parentContext,
	}
}

func (rr *conversionVisitor) addErrorf(node sourcewalk.SourceNode, format string, args ...any) {
	err := fmt.Errorf(format, args...)
	rr.addError(node, err)
}

func (rr *conversionVisitor) addError(node sourcewalk.SourceNode, err error) {
	loc := node.GetPos()
	if loc != nil {
		err = errpos.AddPosition(err, *loc)
	}
	log.Printf("walker error at %s: %v", strings.Join(node.Path, "."), err)
	rr.root.errors = append(rr.root.errors, err)
}

func (ww *conversionVisitor) inMessage(msg *MessageBuilder) *conversionVisitor {
	walk := ww._clone()
	walk.parentContext = msg
	walk.field = nil
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

	typeRef, err := ww.root.resolveTypeNoImport(ref.Ref)
	if err != nil {
		pos := ref.Source.GetPos()
		if pos != nil {
			err = errpos.AddPosition(err, *pos)
		}
		log.Printf("resolveType error at %s: %v", strings.Join(ref.Source.Path, "."), err)
		return nil, err
	}

	ww.file.ensureImport(typeRef.File)
	return typeRef, nil
}

func (ww *conversionVisitor) visitFileNode(file *sourcewalk.FileNode) error {
	return file.RangeRootElements(sourcewalk.FileCallbacks{
		SchemaCallbacks: sourcewalk.SchemaCallbacks{
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
		},
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

func walkerSchemaVisitor(ww *conversionVisitor) sourcewalk.SchemaVisitor {
	return &sourcewalk.SchemaCallbacks{
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
	if node.AnyMember != nil {
		objectType.AnyMember = node.AnyMember
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
		if ww.field == nil {
			ww.addErrorf(node.Source, "missing object name")
		}
		schema.Name = strcase.ToCamel(ww.field.name)
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

func (ww *conversionVisitor) visitEnumNode(node *sourcewalk.EnumNode) {

	prefix := node.Schema.Prefix
	if prefix == "" {
		prefix = strcase.ToScreamingSnake(node.Schema.Name) + "_"
	}
	eb := &enumBuilder{
		prefix: prefix,
		desc: &descriptorpb.EnumDescriptorProto{
			Name: gl.Ptr(node.Schema.Name),
			Value: []*descriptorpb.EnumValueDescriptorProto{{
				Name:   gl.Ptr(fmt.Sprintf("%sUNSPECIFIED", prefix)),
				Number: gl.Ptr(int32(0)),
			}},
		},
	}

	if node.Schema.Description != "" {
		eb.comment([]int32{}, node.Schema.Description)
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

		eb.desc.Options = &descriptorpb.EnumOptions{}
		proto.SetExtension(eb.desc.Options, ext_j5pb.E_Enum, ext)
	}

	optionsToSet := node.Schema.Options
	if len(optionsToSet) > 0 && optionsToSet[0].Number == 0 && strings.HasSuffix(optionsToSet[0].Name, "UNSPECIFIED") {
		eb.addValue(0, optionsToSet[0])
		optionsToSet = optionsToSet[1:]
	}

	for idx, value := range optionsToSet {
		eb.addValue(int32(idx+1), value)
	}

	ww.parentContext.addEnum(eb)
}
