package j5lang

import (
	"github.com/pentops/j5/gen/j5/source/v1/source_j5pb"
	"github.com/pentops/j5/internal/j5reflect"
)

const (
	VersionKey      = "version"
	FieldTag        = "field"
	ObjectTag       = "object"
	EntityTag       = "entity"
	EnumTag         = "enum"
	EnumOptionTag   = "option"
	EnumPrefixTag   = "prefix"
	OneofOptionTag  = "option"
	EntityDataTag   = "data"
	EntityKeyTag    = "key"
	EntityStatusTag = "status"
	EntityEventTag  = "event"
)

type Tag struct {
	Path        string
	Split       []string
	IsGroup     bool
	RightToLeft bool
}

type ASTFile interface {
	ApplySchema(schema *ObjectNode) error
}

func ConvertTreeToSource(f ASTFile) (*source_j5pb.SourceFile, error) {

	schemaSet := j5reflect.New()

	base := &source_j5pb.SourceFile{}
	refl, err := schemaSet.NewObject(base.ProtoReflect())
	if err != nil {
		return nil, err
	}

	fieldNode := &ObjectNode{
		Tags: []Tag{{Path: "name"}, {Path: "schema", IsGroup: true}},
		Attributes: map[string]string{
			"required": "required",
		},
		Auto: true,
	}
	fieldNode.Blocks = map[string]*BlockDef{
		"ref": {
			Property: "ref",
			ObjectNode: &ObjectNode{
				Tags: []Tag{{Split: []string{"schema", "package"}, RightToLeft: true}},
				Auto: true,
			},
		},
		"field": {
			Property:   "object.properties",
			ObjectNode: fieldNode,
		},
	}

	objectNode := &ObjectNode{
		Tags: []Tag{{Path: "name"}},
		Blocks: map[string]*BlockDef{
			"field": {
				Property:   "properties",
				ObjectNode: fieldNode,
			},
		},
		Auto: true,
	}

	schema := &ObjectNode{
		PropSets: []PropertySet{refl},
		Attributes: map[string]string{
			"version": "version",
		},
		Description: "description",
		Blocks: map[string]*BlockDef{
			"entity": {
				Property: "entities",
				ObjectNode: &ObjectNode{
					Tags: []Tag{{Path: "name"}},
				},
			},
			"enum": {
				Property: "schemas.enum",
				ObjectNode: &ObjectNode{
					Tags: []Tag{{Path: "name"}},
					Auto: true,
					Blocks: map[string]*BlockDef{
						"option": {
							Property: "options",
							ObjectNode: &ObjectNode{
								Auto: true,
							},
						},
					},
				},
			},
			"object": {
				Property:   "schemas.object",
				ObjectNode: objectNode,
			},
			"oneof": {
				Property: "schemas.oneof",
				ObjectNode: &ObjectNode{
					Tags: []Tag{{Path: "name"}},
					Blocks: map[string]*BlockDef{
						"option": {
							Property:   "properties",
							ObjectNode: fieldNode,
						},
					},
				},
			},
		},
	}

	if err := f.ApplySchema(schema); err != nil {
		return nil, err
	}

	return base, nil
}

type BlockDef struct {
	Property string
	*ObjectNode
}

type ObjectNode struct {
	Description string
	PropSets    []PropertySet
	Attributes  map[string]string
	Blocks      map[string]*BlockDef
	Auto        bool
	Tags        []Tag
	TagGroups   int
}

type PropertySet interface {
	GetPropertyOrError(name string) (j5reflect.Property, error)
	GetProperty(name string) j5reflect.Property
	Name() string
}
