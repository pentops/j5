Entities
========

An entity is a shortcut for a number of pre-existing elements, but enforcing
strict convention.

Entities work with the pentops/protostate repo to implement event-driven state
machines.

An entity consists of:

### Keys

Primary, foreign and natural, all immutable.

Specified as the collection of `key` fields in the entity definition
entity.

These become the `FooKeys` message.

### Data

An object representing the mutable data in the entity.

Specified as the collection of `data` fields in the entity definition.

These become the `FooData` message.

### Status

An Enum representing the fixed statuses the entity can be in.

This becomes the `FooStatus` enum.

```j5s
entity Foo {
  ...
  status ACTIVE
  status INACTIVE
}
```

```proto
enum FooStatus {
  FOO_STATUS_UNSPECIFIED = 0;
  FOO_STATUS_ACTIVE = 1;
  FOO_STATUS_INACTIVE = 2;
}
```

### Events

The event shapes which modify the state.

Each event is specified as an object.

```j5s
event Create {
  field name string
}
```

The event types are represented as a single oneof 'FooEventType'.


### Service

The entity definition will automatically produce the 'query' service, which has
a standard 'Get', 'List' and 'ListEvents' endpoint for interacting with the entity.

In addition, 'command' or alternate query services can be attached to the
entity. The commands and queries defined within the entity should only interact
with the entity itself, for any cross-cutting concerns, use a detached service.
Attached services should be seen as 'methods' of the entity.


### Topic

The entity definition will automatically produce a 'Publish' topic, which
publishes all events (state transitions) for the entity.

Custom topics can not be defined within an entity block (yet).


### Foo Example

```j5s
package foo.v1

entity Foo {
	| Foo is lorem ipsum

	key fooId key:id62

	data name string

	status ACTIVE
	status INACTIVE

	event Create {
		field name string
	}

	event Archive {
	}
}
```

Produces the working objects:

#### FooKeys

The wrapper for all key fields

```proto
message FooKeys {
  option (j5.ext.v1.psm) = {
    entity_name: "foo"
    entity_part: ENTITY_PART_KEYS
  };

  option (j5.ext.v1.message).object = {};

  string foo_id = 1 [
    (buf.validate.field).string.pattern = "^[0-9A-Za-z]{22}$",
    (j5.ext.v1.field).key = {}
  ];
}
```

#### FooData

The mutable data fields

```proto
message FooData {
  option (j5.ext.v1.psm) = {
    entity_name: "foo"
    entity_part: ENTITY_PART_DATA
  };

  option (j5.ext.v1.message).object = {};

  string name = 1 [(j5.ext.v1.field).string = {}];
}
```

#### FooState

The wrapper state message

```proto
message FooState {
  option (j5.ext.v1.psm) = {
    entity_name: "foo"
    entity_part: ENTITY_PART_STATE
  };

  option (j5.ext.v1.message).object = {};

  j5.state.v1.StateMetadata metadata = 1 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  FooKeys keys = 2 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object.flatten = true
  ];

  FooData data = 3 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  FooStatus status = 4 [
    (buf.validate.field) = {
      required: true
      enum: {
        defined_only: true
      }
    },
    (j5.ext.v1.field).enum = {}
  ];
}
```

#### FooEventType

The oneof wrapper for all events

```proto
message FooEventType {
  option (j5.ext.v1.message).oneof = {};

  oneof type {
    Create create = 1 [(j5.ext.v1.field).object = {}];

    Archive archive = 2 [(j5.ext.v1.field).object = {}];
  }

  message Create {
    option (j5.ext.v1.message).object = {};

    string name = 1 [(j5.ext.v1.field).string = {}];
  }

  message Archive {
    option (j5.ext.v1.message).object = {};
  }
}
```

#### FooEvent

A wrapper for the event itself, taking the type and metadata.

```proto
message FooEvent {
  option (j5.ext.v1.psm) = {
    entity_name: "foo"
    entity_part: ENTITY_PART_EVENT
  };

  option (j5.ext.v1.message).object = {};

  j5.state.v1.EventMetadata metadata = 1 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  FooKeys keys = 2 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object.flatten = true
  ];

  FooEventType event = 3 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).oneof = {}
  ];
}
```

#### FooStatus

The enum to store the status.


```proto
enum FooStatus {
  FOO_STATUS_UNSPECIFIED = 0;
  FOO_STATUS_ACTIVE = 1;
  FOO_STATUS_INACTIVE = 2;
}
```

#### FooQueryService

The query service, with Get, List and ListEvents methods, produced into the
sub-package `foo.v1.service`.

```proto

service FooQueryService {
  option (j5.ext.v1.service).state_query.entity = "foo";

  rpc FooGet(FooGetRequest) returns (FooGetResponse) {
    option (google.api.http) = {get: "/foo/v1/foo/q"};
    option (j5.ext.v1.method).state_query.get = true;
  }

  rpc FooList(FooListRequest) returns (FooListResponse) {
    option (google.api.http) = {get: "/foo/v1/foo/q"};
    option (j5.ext.v1.method).state_query.list = true;
  }

  rpc FooEvents(FooEventsRequest) returns (FooEventsResponse) {
    option (google.api.http) = {get: "/foo/v1/foo/q/events"};
    option (j5.ext.v1.method).state_query.list_events = true;
  }
}

//... plus implementation messages

```

#### FooPublishTopic

The publish topic, for sending events to the entity, produced into the
sub-package `foo.v1.topic`.

```proto
service FooPublishTopic {
  option (j5.messaging.v1.service) = {
    topic_name: "foo_publish"
    publish: {
    }
  };

  rpc FooEvent(FooEventMessage) returns (google.protobuf.Empty) {}

}

message FooEventMessage {
  option (j5.ext.v1.message).object = {};

  j5.state.v1.EventPublishMetadata metadata = 1 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  foo.v1.FooKeys keys = 2 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  foo.v1.FooEventType event = 3 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).oneof = {}
  ];

  foo.v1.FooData data = 4 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  foo.v1.FooStatus status = 5 [
    (buf.validate.field) = {
      required: true
      enum: {
        defined_only: true
      }
    },
    (j5.ext.v1.field).enum = {}
  ];
}
```

