package sourcewalk

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/auth/v1/auth_j5pb"
	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/gen/j5/sourcedef/v1/sourcedef_j5pb"
)

type entityNode struct {
	name        string
	packageName string
	Source      SourceNode
	Schema      *sourcedef_j5pb.Entity
}

func newEntityNode(source SourceNode, packageName string, entity *sourcedef_j5pb.Entity) (*entityNode, error) {
	entityNode := &entityNode{
		name:        strcase.ToSnake(entity.Name),
		packageName: packageName,
		Schema:      entity,
		Source:      source,
	}
	return entityNode, nil
}

func (ent *entityNode) componentName(suffix string) string {
	return strcase.ToCamel(ent.Schema.Name) + strcase.ToCamel(suffix)
}

func (ent *entityNode) fullName() string {
	return fmt.Sprintf("%s.%s", ent.packageName, strcase.ToCamel(ent.Schema.Name))
}

func schemaRefField(pkg, desc string) *schema_j5pb.Field {
	return &schema_j5pb.Field{
		Type: &schema_j5pb.Field_Object{
			Object: &schema_j5pb.ObjectField{
				Schema: &schema_j5pb.ObjectField_Ref{
					Ref: &schema_j5pb.Ref{
						Package: pkg,
						Schema:  desc,
					},
				},
			},
		},
	}
}

// run converts the entity into lower level schema elements and calls the
// visitors on those.
func (ent *entityNode) run(visitor FileVisitor) error {

	if ent.Schema.BaseUrlPath == "" {
		pkgParts := strings.Split(ent.packageName, ".")
		pkgParts = append(pkgParts, strcase.ToSnake(ent.Schema.Name))
		ent.Schema.BaseUrlPath = strings.Join(pkgParts, "/")
	}

	if err := ent.acceptKeys(visitor); err != nil {
		return err
	}
	if err := ent.acceptData(visitor); err != nil {
		return err
	}
	if err := ent.acceptStatus(visitor); err != nil {
		return err
	}
	if err := ent.acceptState(visitor); err != nil {
		return err
	}
	if err := ent.acceptEventOneof(visitor); err != nil {
		return err
	}
	if err := ent.acceptEvent(visitor); err != nil {
		return err
	}
	if err := ent.acceptQuery(visitor); err != nil {
		return err
	}
	if err := ent.acceptCommands(visitor); err != nil {
		return err
	}
	if err := ent.acceptPublishTopic(visitor); err != nil {
		return err
	}
	if err := ent.acceptSummaryTopics(visitor); err != nil {
		return err
	}

	if len(ent.Schema.Schemas) > 0 {
		ss := mapNested(ent.Source, nil, ent.Schema.Schemas)
		if err := ss.RangeNestedSchemas(visitor); err != nil {
			return err
		}
	}

	return nil
}

func (ent *entityNode) acceptKeys(visitor FileVisitor) error {

	keyProps := make([]*schema_j5pb.ObjectProperty, 0, len(ent.Schema.Keys))
	for _, key := range ent.Schema.Keys {
		if key.Key == nil {
			key.Key = &schema_j5pb.EntityKey{}
		}
		key.Def.EntityKey = key.Key

		keyProps = append(keyProps, key.Def)
	}
	object, err := newVirtualObjectNode(
		ent.Source.child("keys"),
		nil,
		ent.componentName("Keys"),
		keyProps,
	)
	if err != nil {
		return wrapErr(ent.Source, err)
	}

	object.Entity = &schema_j5pb.EntityObject{
		Entity: ent.name,
		Part:   schema_j5pb.EntityPart_KEYS,
	}

	if err := visitor.VisitObject(object); err != nil {
		return wrapErr(ent.Source, err)
	}
	return nil
}

func (ent *entityNode) acceptData(visitor FileVisitor) error {

	node, err := newVirtualObjectNode(
		ent.Source.child("data"),
		nil,
		ent.componentName("Data"),
		ent.Schema.Data,
	)
	if err != nil {
		return wrapErr(ent.Source, err)
	}

	node.Entity = &schema_j5pb.EntityObject{
		Entity: ent.name,
		Part:   schema_j5pb.EntityPart_DATA,
	}

	return visitor.VisitObject(node)
}

func (ent *entityNode) acceptStatus(visitor FileVisitor) error {
	entity := ent.Schema
	status := &schema_j5pb.Enum{
		Name:    ent.componentName("Status"),
		Options: entity.Status,
		Prefix:  strcase.ToScreamingSnake(entity.Name) + "_STATUS_",
	}

	node, err := newEnumNode(ent.Source.child("status"), nil, status)
	if err != nil {
		return wrapErr(ent.Source, err)
	}

	return visitor.VisitEnum(node)
}

func (ent *entityNode) innerRef(name string) *schema_j5pb.Field {
	return schemaRefField("", ent.componentName(name))
}

func (ent *entityNode) findStatus(end string) (string, bool) {
	for _, status := range ent.Schema.Status {
		if status.Name == end {
			return fmt.Sprintf("%s_STATUS_%s",
				strcase.ToScreamingSnake(ent.Schema.Name),
				strcase.ToScreamingSnake(status.Name),
			), true
		}
	}
	return "", false
}

func (ent *entityNode) acceptState(visitor FileVisitor) error {
	entity := ent.Schema

	objKeys := schemaRefField("", ent.componentName("Keys"))
	objKeys.GetObject().Flatten = true

	statusEnumField := &schema_j5pb.EnumField{
		Schema: &schema_j5pb.EnumField_Ref{
			Ref: &schema_j5pb.Ref{
				Schema: ent.componentName("Status"),
			},
		},
		ListRules: &list_j5pb.EnumRules{
			Filtering: &list_j5pb.FilteringConstraint{
				Filterable: true,
			},
		},
	}

	if ent.Schema.Query != nil {
		filters := make([]string, 0, len(ent.Schema.Query.DefaultStatusFilter))
		for _, filter := range ent.Schema.Query.DefaultStatusFilter {
			status, ok := ent.findStatus(filter)
			if !ok {
				return walkerErrorf("status %q not found in entity %q", filter, entity.Name)
			}
			filters = append(filters, status)
		}
		if len(ent.Schema.Query.DefaultStatusFilter) > 0 {
			statusEnumField.ListRules.Filtering.DefaultFilters = filters
		}
	}

	statusField := &schema_j5pb.ObjectProperty{
		Name:       "status",
		ProtoField: []int32{4},
		Required:   true,
		Schema: &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Enum{
				Enum: statusEnumField,
			},
		},
	}

	state := &schema_j5pb.Object{
		Name: strcase.ToCamel(entity.Name + "State"),
		Entity: &schema_j5pb.EntityObject{
			Entity: ent.name,
			Part:   schema_j5pb.EntityPart_STATE,
		},
		Properties: []*schema_j5pb.ObjectProperty{{
			Name:       "metadata",
			ProtoField: []int32{1},
			Required:   true,
			Schema:     schemaRefField("j5.state.v1", "StateMetadata"),
		}, {
			Name:       "keys",
			ProtoField: []int32{2},
			Required:   true,
			Schema:     objKeys,
		}, {
			Name:       "data",
			ProtoField: []int32{3},
			Required:   true,
			Schema:     ent.innerRef("Data"),
		},
			statusField,
		}}

	node, err := newObjectSchemaNode(ent.Source.child("state"), nil, state)
	if err != nil {
		return wrapErr(ent.Source, err)
	}
	return visitor.VisitObject(node)
}

func (ent *entityNode) acceptEventOneof(visitor FileVisitor) error {

	if len(ent.Schema.Events) == 0 {
		return wrapErr(ent.Source, errors.New("entity has no events"))
	}

	entity := ent.Schema
	eventOneof := &schema_j5pb.Oneof{
		Name:       strcase.ToCamel(entity.Name + "EventType"),
		Properties: make([]*schema_j5pb.ObjectProperty, 0, len(entity.Events)),
	}

	nestedNodes := make([]*nestedNode, 0, len(entity.Events))

	for idx, eventObjectSchema := range entity.Events {

		nestedName := eventObjectSchema.Def.Name

		nested := &sourcedef_j5pb.NestedSchema_Object{Object: eventObjectSchema}
		nestedNodes = append(nestedNodes, &nestedNode{
			schema: nested,
			source: ent.Source.child("events", strconv.Itoa(idx)),
		})

		propSchema := &schema_j5pb.ObjectProperty{
			Name:       strcase.ToLowerCamel(eventObjectSchema.Def.Name),
			ProtoField: []int32{int32(idx + 1)},
			Schema: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_Object{
					Object: &schema_j5pb.ObjectField{
						Schema: &schema_j5pb.ObjectField_Ref{
							Ref: &schema_j5pb.Ref{
								Package: "",
								Schema: fmt.Sprintf("%s.%s",
									eventOneof.Name,
									nestedName,
								),
							},
						},
					},
				},
			},
		}

		eventOneof.Properties = append(eventOneof.Properties, propSchema)
	}

	node, err := newOneofSchemaNode(ent.Source.child(virtualPathNode, "event_type"), nil, eventOneof)
	if err != nil {
		return wrapErr(ent.Source, err)
	}

	node.nestedSet = nestedSet{
		children: nestedNodes,
		parent:   node,
	}

	return visitor.VisitOneof(node)
}

func (ent *entityNode) acceptEvent(visitor FileVisitor) error {
	entity := ent.Schema

	eventKeys := ent.innerRef("Keys")
	eventKeys.GetObject().Flatten = true

	eventObject := &schema_j5pb.Object{
		Name: strcase.ToCamel(entity.Name + "Event"),
		Entity: &schema_j5pb.EntityObject{
			Entity: ent.name,
			Part:   schema_j5pb.EntityPart_EVENT,
		},
		Properties: []*schema_j5pb.ObjectProperty{{
			Name:       "metadata",
			ProtoField: []int32{1},
			Required:   true,
			Schema:     schemaRefField("j5.state.v1", "EventMetadata"),
		}, {
			Name:       "keys",
			ProtoField: []int32{2},
			Required:   true,
			Schema:     eventKeys,
		}, {
			Name:       "event",
			ProtoField: []int32{3},
			Required:   true,
			Schema: &schema_j5pb.Field{
				Type: &schema_j5pb.Field_Oneof{
					Oneof: &schema_j5pb.OneofField{
						Schema: &schema_j5pb.OneofField_Ref{
							Ref: &schema_j5pb.Ref{
								Schema: ent.componentName("EventType"),
							},
						},
						ListRules: &list_j5pb.OneofRules{
							Filtering: &list_j5pb.FilteringConstraint{
								Filterable: true,
							},
						},
					},
				},
			},
		}},
	}

	node, err := newObjectSchemaNode(ent.Source.child("event"), nil, eventObject)
	if err != nil {
		return wrapErr(ent.Source, err)
	}
	return visitor.VisitObject(node)
}

func (ent *entityNode) acceptCommands(visitor FileVisitor) error {
	entity := ent.Schema

	services := make([]*serviceBuilder, 0)
	for idx, serviceSrc := range entity.Commands {

		var serviceName string
		var servicePath string

		if serviceSrc.Name != nil {
			serviceName = *serviceSrc.Name
			if !strings.HasSuffix(serviceName, "Command") {
				serviceName += "Command"
			}
		} else {
			serviceName = fmt.Sprintf("%sCommand", strcase.ToCamel(ent.Schema.Name))
		}

		if serviceSrc.BasePath != nil {
			servicePath = fmt.Sprintf("/%s/%s", entity.BaseUrlPath, *serviceSrc.BasePath)
		} else {
			servicePath = fmt.Sprintf("/%s/c", entity.BaseUrlPath)
		}

		var auth *auth_j5pb.MethodAuthType
		if serviceSrc.Options != nil {
			auth = serviceSrc.Options.DefaultAuth
		}

		service := &sourcedef_j5pb.Service{
			Name:        &serviceName,
			BasePath:    &servicePath,
			Description: serviceSrc.Description,
			Methods:     serviceSrc.Methods,
			Options: &ext_j5pb.ServiceOptions{
				Type: &ext_j5pb.ServiceOptions_StateCommand_{
					StateCommand: &ext_j5pb.ServiceOptions_StateCommand{
						Entity: ent.name,
					},
				},
				DefaultAuth: auth,
			},
		}

		source := ent.Source.child("commands", strconv.Itoa(idx))
		node, err := newServiceRef(source, service)
		if err != nil {
			return wrapErr(source, err)
		}
		services = append(services, node)
	}

	return visitor.VisitServiceFile(&ServiceFileNode{
		services: services,
	})
}

func (ent *entityNode) acceptSummaryTopics(visitor FileVisitor) error {

	topics := make([]*topicRef, 0)

	names := make(map[string]bool)
	for idx, summary := range ent.Schema.Summaries {
		source := ent.Source.child("summaries", strconv.Itoa(idx))

		if names[summary.Name] {
			return walkerErrorf("duplicate summary name %q", summary.Name)
		}
		names[summary.Name] = true

		var name string
		if summary.Name == "" {
			name = fmt.Sprintf("%sSummary", strcase.ToCamel(ent.Schema.Name))
		} else {
			name = fmt.Sprintf("%s%s", strcase.ToCamel(ent.Schema.Name), strcase.ToCamel(summary.Name))
		}

		topicDef := &sourcedef_j5pb.Topic{
			Name: name,
			Type: &sourcedef_j5pb.TopicType{
				Type: &sourcedef_j5pb.TopicType_Upsert_{
					Upsert: &sourcedef_j5pb.TopicType_Upsert{
						EntityName: ent.fullName(),
						Message: &sourcedef_j5pb.TopicMethod{
							Name:        gl.Ptr(name),
							Description: fmt.Sprintf("Publishes summary output of state for the %s entity", ent.Schema.Name),
							Fields:      summary.Fields,
						},
					},
				},
			},
		}

		topics = append(topics, &topicRef{
			schema: topicDef,
			source: source,
		})

	}

	return visitor.VisitTopicFile(&TopicFileNode{
		topics: topics,
	})
}

func (ent *entityNode) acceptPublishTopic(visitor FileVisitor) error {

	source := ent.Source.child(virtualPathNode, "publish")

	properties := []*schema_j5pb.ObjectProperty{{
		Name:       "metadata",
		ProtoField: []int32{1},
		Required:   true,
		Schema:     schemaRefField("j5.state.v1", "EventPublishMetadata"),
	}, {
		Name:       "keys",
		ProtoField: []int32{2},
		Required:   true,
		Schema:     schemaRefField("", ent.componentName("Keys")),
	}, {
		Name:       "event",
		ProtoField: []int32{3},
		Required:   true,
		Schema: &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Oneof{
				Oneof: &schema_j5pb.OneofField{
					Schema: &schema_j5pb.OneofField_Ref{
						Ref: &schema_j5pb.Ref{
							Schema: ent.componentName("EventType"),
						},
					},
				},
			},
		},
	}, {
		Name:       "data",
		ProtoField: []int32{3},
		Required:   true,
		Schema:     ent.innerRef("Data"),
	}, {
		Name:       "status",
		ProtoField: []int32{4},
		Required:   true,
		Schema: &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Enum{
				Enum: &schema_j5pb.EnumField{
					Schema: &schema_j5pb.EnumField_Ref{
						Ref: &schema_j5pb.Ref{
							Schema: ent.componentName("Status"),
						},
					},
				},
			},
		},
	}}

	publishTopic := &sourcedef_j5pb.Topic{
		Name: fmt.Sprintf("%sPublish", strcase.ToCamel(ent.Schema.Name)),
		Type: &sourcedef_j5pb.TopicType{
			Type: &sourcedef_j5pb.TopicType_Event_{
				Event: &sourcedef_j5pb.TopicType_Event{
					EntityName: ent.fullName(),
					Message: &sourcedef_j5pb.TopicMethod{
						Name:        gl.Ptr(fmt.Sprintf("%sEvent", strcase.ToCamel(ent.Schema.Name))),
						Description: fmt.Sprintf("Publishes all events for the %s entity", ent.Schema.Name),
						Fields:      properties,
					},
				},
			},
		},
	}

	return visitor.VisitTopicFile(&TopicFileNode{
		topics: []*topicRef{{
			schema: publishTopic,
			source: source,
		}},
	})
}

func (ent *entityNode) acceptQuery(visitor FileVisitor) error {

	entity := ent.Schema
	name := ent.name

	getKeys := make([]*schema_j5pb.ObjectProperty, 0, len(ent.Schema.Keys))
	httpPath := []string{}

	listKeys := make([]*schema_j5pb.ObjectProperty, 0)
	listHttpPath := []string{}

	baseURLParts := strings.Split(entity.BaseUrlPath, "/")
	baseURLFields := map[string]bool{}
	for _, part := range baseURLParts {
		if !strings.HasPrefix(part, ":") {
			continue
		}
		baseURLFields[part[1:]] = true
	}
	for _, key := range ent.Schema.Keys {
		isPathKey := key.Key.ShardKey || baseURLFields[key.Def.Name]
		if key.Key.Primary || key.Def.EntityKey != nil && key.Def.EntityKey.Primary {
			// The field is a Primary Key of the entity
			getKeys = append(getKeys, key.Def)
			if !baseURLFields[key.Def.Name] {
				httpPath = append(httpPath, fmt.Sprintf(":%s", key.Def.Name))
			}

			if isPathKey {
				// primary and shard.
				listKeys = append(listKeys, key.Def)
				if !baseURLFields[key.Def.Name] {
					listHttpPath = append(listHttpPath, fmt.Sprintf(":%s", key.Def.Name))
				}
			}
		} else {
			if isPathKey {
				// just shard, not primary - still part of the URL

				listKeys = append(listKeys, key.Def)
				getKeys = append(getKeys, key.Def)

				if !baseURLFields[key.Def.Name] {
					listHttpPath = append(listHttpPath, fmt.Sprintf(":%s", key.Def.Name))
					httpPath = append(httpPath, fmt.Sprintf(":%s", key.Def.Name))
				}
			}
		}

	}
	getMethod := &sourcedef_j5pb.APIMethod{
		Name:       fmt.Sprintf("%sGet", strcase.ToCamel(name)),
		HttpPath:   strings.Join(httpPath, "/"),
		HttpMethod: client_j5pb.HTTPMethod_GET,
		Request: &sourcedef_j5pb.AnonymousObject{
			Properties: getKeys,
		},

		Response: &sourcedef_j5pb.AnonymousObject{
			Properties: []*schema_j5pb.ObjectProperty{{
				Name:       strcase.ToLowerCamel(name),
				ProtoField: []int32{1},
				Schema:     ent.innerRef("State"),
				Required:   true,
			}},
		},
		Options: &ext_j5pb.MethodOptions{
			StateQuery: &ext_j5pb.StateQueryMethodOptions{
				Get: true,
			},
		},
	}

	listMethod := &sourcedef_j5pb.APIMethod{
		Name:       fmt.Sprintf("%sList", strcase.ToCamel(name)),
		HttpPath:   strings.Join(listHttpPath, "/"),
		HttpMethod: client_j5pb.HTTPMethod_GET,
		Request: &sourcedef_j5pb.AnonymousObject{
			Properties: append(listKeys, &schema_j5pb.ObjectProperty{
				Name:       "page",
				ProtoField: []int32{100},
				Schema:     schemaRefField("j5.list.v1", "PageRequest"),
			}, &schema_j5pb.ObjectProperty{
				Name:       "query",
				ProtoField: []int32{101},
				Schema:     schemaRefField("j5.list.v1", "QueryRequest"),
			}),
		},
		Response: &sourcedef_j5pb.AnonymousObject{
			Properties: []*schema_j5pb.ObjectProperty{{
				Name:       strcase.ToLowerCamel(name),
				ProtoField: []int32{1},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Array{
						Array: &schema_j5pb.ArrayField{
							Items: ent.innerRef("State"),
						},
					},
				},
			}, {
				Name:       "page",
				ProtoField: []int32{100},
				Schema:     schemaRefField("j5.list.v1", "PageResponse"),
			}},
		},
		Options: &ext_j5pb.MethodOptions{
			StateQuery: &ext_j5pb.StateQueryMethodOptions{
				List: true,
			},
		},
	}

	eventsMethod := &sourcedef_j5pb.APIMethod{
		Name:       fmt.Sprintf("%sEvents", strcase.ToCamel(name)),
		HttpPath:   strings.Join(append(httpPath, "events"), "/"),
		HttpMethod: client_j5pb.HTTPMethod_GET,
		Request: &sourcedef_j5pb.AnonymousObject{
			Properties: append(getKeys, &schema_j5pb.ObjectProperty{
				Name:       "page",
				ProtoField: []int32{100},
				Schema:     schemaRefField("j5.list.v1", "PageRequest"),
			}, &schema_j5pb.ObjectProperty{
				Name:       "query",
				ProtoField: []int32{101},
				Schema:     schemaRefField("j5.list.v1", "QueryRequest"),
			}),
		},
		Response: &sourcedef_j5pb.AnonymousObject{
			Properties: []*schema_j5pb.ObjectProperty{{
				Name:       "events",
				ProtoField: []int32{1},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Array{
						Array: &schema_j5pb.ArrayField{
							Items: ent.innerRef("Event"),
						},
					},
				},
			}, {
				Name:       "page",
				ProtoField: []int32{100},
				Schema:     schemaRefField("j5.list.v1", "PageResponse"),
			}},
		},
		Options: &ext_j5pb.MethodOptions{
			StateQuery: &ext_j5pb.StateQueryMethodOptions{
				ListEvents: true,
			},
		},
	}

	var auth *auth_j5pb.MethodAuthType
	if ent.Schema.Query != nil {
		if ent.Schema.Query.EventsInGet {
			getMethod.Response.Properties = append(getMethod.Response.Properties, &schema_j5pb.ObjectProperty{
				Name:       "events",
				ProtoField: []int32{2},
				Schema: &schema_j5pb.Field{
					Type: &schema_j5pb.Field_Array{
						Array: &schema_j5pb.ArrayField{
							Items: ent.innerRef("Event"),
						},
					},
				},
			})
		}

		if ent.Schema.Query.ListRequest != nil {
			listMethod.ListRequest = ent.Schema.Query.ListRequest
		}

		if ent.Schema.Query.EventsListRequest != nil {
			eventsMethod.ListRequest = ent.Schema.Query.EventsListRequest
		}

		auth = ent.Schema.Query.Auth
	}

	query := &sourcedef_j5pb.Service{
		BasePath: gl.Ptr(fmt.Sprintf("/%s/q", entity.BaseUrlPath)),
		Name:     gl.Ptr(fmt.Sprintf("%sQuery", strcase.ToCamel(name))),
		Methods: []*sourcedef_j5pb.APIMethod{
			getMethod,
			listMethod,
			eventsMethod,
		},
		Options: &ext_j5pb.ServiceOptions{
			Type: &ext_j5pb.ServiceOptions_StateQuery_{
				StateQuery: &ext_j5pb.ServiceOptions_StateQuery{
					Entity: name,
				},
			},
			DefaultAuth: auth,
		},
	}

	serviceNode, err := newServiceRef(ent.Source.child(virtualPathNode, "query"), query)
	if err != nil {
		return wrapErr(ent.Source, err)
	}

	return visitor.VisitServiceFile(&ServiceFileNode{
		services: []*serviceBuilder{serviceNode},
	})

}
