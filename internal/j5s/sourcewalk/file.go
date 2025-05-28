package sourcewalk

import (
	"fmt"
	"strconv"

	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
)

type FileVisitor interface {
	SchemaVisitor
	VisitTopicFile(*TopicFileNode) error
	VisitServiceFile(*ServiceFileNode) error
}

type FileCallbacks struct {
	SchemaCallbacks
	TopicFile   func(*TopicFileNode) error
	ServiceFile func(*ServiceFileNode) error
}

func (fc FileCallbacks) VisitTopicFile(tfn *TopicFileNode) error {
	return fc.TopicFile(tfn)
}

func (fc FileCallbacks) VisitServiceFile(sfn *ServiceFileNode) error {
	return fc.ServiceFile(sfn)
}

var _ FileVisitor = FileCallbacks{}

type FileNode struct {
	*sourcedef_j5pb.SourceFile
	Source SourceNode
}

func wrapErr(source SourceNode, err error) error {
	return fmt.Errorf("at %s: %w", source.PathString(), err)
}

func (fn *FileNode) RangeRootElements(visitor FileVisitor) error {
	for idx, element := range fn.Elements {
		source := fn.Source.child("elements", strconv.Itoa(idx))
		switch element := element.Type.(type) {
		case *sourcedef_j5pb.RootElement_Object:
			source := source.child("object")
			objectNode, err := newObjectNode(source, nil, element.Object)
			if err != nil {
				return wrapErr(source, err)
			}
			if err := visitor.VisitObject(objectNode); err != nil {
				return wrapErr(source, err)
			}

		case *sourcedef_j5pb.RootElement_Oneof:
			oneofNode, err := newOneofNode(source.child("oneof"), nil, element.Oneof)
			if err != nil {
				return wrapErr(source, err)
			}
			if err := visitor.VisitOneof(oneofNode); err != nil {
				return err
			}

		case *sourcedef_j5pb.RootElement_Enum:
			enumNode, err := newEnumNode(source.child("enum"), nil, element.Enum)
			if err != nil {
				return wrapErr(source, err)
			}
			if err := visitor.VisitEnum(enumNode); err != nil {
				return err
			}

		case *sourcedef_j5pb.RootElement_Polymorph:
			polymorphNode, err := newPolymorphNode(source.child("polymorph"), nil, element.Polymorph.Def, element.Polymorph.Includes)
			if err != nil {
				return wrapErr(source, err)
			}
			if err := visitor.VisitPolymorph(polymorphNode); err != nil {
				return err
			}

		case *sourcedef_j5pb.RootElement_Entity:
			entityNode, err := newEntityNode(source.child("entity"), fn.Package.Name, element.Entity)
			if err != nil {
				return wrapErr(source, err)
			}
			if err := entityNode.run(visitor); err != nil {
				return err
			}
			// Entity is converted on-the-fly to root schemas, and uses the file
			// callbacks for the elements it creates.

		case *sourcedef_j5pb.RootElement_Topic:
			topic := element.Topic
			topicFileNode := &TopicFileNode{
				topics: []*topicRef{{
					schema: topic,
					source: source.child("topic"),
				}},
			}
			if err := visitor.VisitTopicFile(topicFileNode); err != nil {
				return err
			}

		case *sourcedef_j5pb.RootElement_Service:
			node, err := newServiceRef(source.child("service"), element.Service)
			if err != nil {
				return wrapErr(source, err)
			}
			serviceFileNode := &ServiceFileNode{
				services: []*serviceBuilder{node},
			}
			if err := visitor.VisitServiceFile(serviceFileNode); err != nil {
				return err
			}

		default:
			return walkerErrorf("unknown root element in FileNode %T", element)
		}
	}
	return nil
}
