J5 API Definitions
==================

J5 is a data schema structure, language and toolkit for defining Event-Driven APIs.

As with all pentops tools, it is very opinionated. J5 definitions are intended to work within the wider pentops ecosystem, especially for [Messaging](https://github.com/pentops/o5-messaging) and [State Machines](https://github.com/pentops/protostate)

J5 wraps a **subset** of the data types and schemas of [Protocol Buffers](https://protobuf.dev/). The goal of the project is different to Protocol Buffers, and different trade-offs are made. J5 schemas are based on [proto descriptors](https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/descriptor.proto), 
and can be fully represented in .proto files, (leaning heavily on annotations), but not all Protocol Buffer structures can be represented in J5.

Files
=====

J5 APIs can be defined in either  `.proto` files using [proto3 syntax](https://protobuf.dev/programming-guides/proto3/), or in `.j5s` files, based on [bcl](https://github.com/pentops/bcl.go), a config format designed specifically for defining J5 schemas and configuration.

J5 does not support all proto structures, so .proto files must be structured according to the rigid rules of J5, including annotations.


### Descriptions

The pipe syntax is used for descriptions, and occurs in many schema
elements. Descriptions must be the first element in the body, and may span
multiple lines. Where there is no body, a single line description can be added
to the end of the definition line:

```j5s
field name string {
    | Rather long
    | Description
    | 
    | With a newline
}
```

```j5s
field name string | Name of the object
```

### Packages and Imports

JS files exist in a 'package', which must match their path from the root.
The package declaration should be the first non-comment line in the file.

```j5s
package foo.bar.v1
```

The file should be in the directory `/foo/bar/v1/` relative to the 'bundle
root'.

An Import declaration at the top of a file names an entire package and brings
the package into scope by either the package name ('bar' not 'v1') or by the
alias name.

`import <package>:<alias>`


```j5s
package foo.bar.v1

import foo.baz.v1:baz

object Foo {
    field bar baz.Bar
}
```

The prefix 'baz' can then be used to refer to Objects, Oneof and Enums in the
baz package, regardless of the file name.

A j5s source can import a proto source, and v/v. In proto, imports specify the filename of the schema. These are converted on the fly. The filename for a j5s file, when imported from proto, is the full j5s filename followed by `.proto`, e.g. `/foo/v1/bar.j5s` can be imported as `import "/foo/v1/bar.j5s.proto".

Schemas
=======

J5 schemas define objects and fields, similar to JSON Schema and 'pure proto' (not gRPC).

The meta-structure of schemas is defined as proto files in the [j5.schema.v1](https://github.com/pentops/j5/blob/main/proto/j5/j5/schema/v1/schema.proto) package.

### Object

A set of named fields (properties), each with its own data type and other annotations.

```j5s
object Foo {
  field fooId key:id62 {
    | The primary key of Foo
    required = true
  }

  field name string
}
```

```proto
message Foo {
  string foo_id = 1 [
    (buf.validate.field) = {
      required: true
      string: {
        pattern: "^[0-9A-Za-z]{22}$"
      }
    },
    (j5.ext.v1.field).key = {}
  ];

  string name = 2;
}

```

#### Object Field

Fields in an object are defined by `j5.schema.v1.ObjectProperty`.

In J5s files, each field must have a name and type, defined by the tags in the
syntax.

```j5s
field name string
```

Some types require further qualification, such as the `key` type, which requires
a key type, such as `id62` or `uuid`, or integer types, which require a bit
width and signedness, and object, oneof and enum require the type name.

```j5s
field fooId key:id62
field age integer:INT32
field bar object:Bar
```

For further attributes, the field can have a 'body', using curly braces:

```j5s
field name string {
  required = true
}
```

All object properties have two common attributes - `required` and
`explicitlyOptional`. A shortcut for these is to add a `!` or `?` respectively before the
data type in the definition:

```j5s
field name ! string
```

is equivalent to

```j5s
field name string {
  required = true
}
```

All other attributes are defined by the specific data type of the field, for
example, the `key` type has `primary` (bool) and `foreign` (string) attributes.

```j5s
field fooId key:id62 {
  required = true
  foreign = "bar.v1.Bar"
}
```

When a field is an object type, the field can have the `flatten = true`
attribute, which makes the JSON encoding of the inner object act as fields of
the outer object. This can be used to create an 'extends' sort of pattern.

#### Inline Types

When a field has a type of `object`, `oneof` or `enum`, the type can be defined
either as a reference `object:Bar` or inline:

```j5s
object Foo {
  field bar object {
    field barId key:id62
  }
}
```

```proto
message Foo {
  Bar bar = 1;
  message Bar {
  ...
  }
}
```

This can also be used when the type is an array or map of an object, oneof or
enum.

```j5s
object Foo {
  field bars array {
    field barId key:id62
  }
}
```

```proto
message Foo {
  repeated Bar bars = 1;
  message Bar {
  ...
  }
}
```

The inline type by default will take the name of the field, but can be
overridden by directly accessing the name property:

```j5s
object Foo {
  field bars array:object {
    object.name = "Bar"
    field barId key:id62
  }
}
```

```proto
message Foo {
  repeated Bar bars = 1;
  message Bar {
  ...
  }
}
```

### Oneof

Like an object, but at most one key can be set at a time, and all of the
properties must be objects.

```j5s
oneof Foo {
  option bar object {
    field barId key:id62
  }
  option baz object {
    field bazId key:id62
  }
}
```

```proto
message Foo {
  oneof type {
    Bar bar = 1;
    Baz baz = 2;
  }

  message Bar {
    ...
  }

  message Baz {
    ...
  }
}
```

A 'oneof' notation in proto is a validation rule, the fields belong on the
parent message, where this proto structure uses a message to wrap oneof into
being a unique type, allowing it to be reused, and for code generation to add
methods to it.

Oneofs are encoded as a JSON object, with a `!type` field, and the single key
matching that type as a sub-object.

```json
{
  "!type": "bar",
  "bar": {
    "barId": "123"
  }
}
```

### Enum

A set of named values.

```j5s
enum Status {
  option ACTIVE
  option INACTIVE
}
```

These map to proto enums following the Buf rules.

```proto
enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_INACTIVE = 2;
}
```

The first proto enum value will always be `{prefix}_UNSPECIFIED`, it can be
omitted from the source, or explicitly included (as `UNSPECIFIED`) to add
descriptions or other extensions to the option.

```j5s
enum Status {
  option UNSPECIFIED | Initial Status
  option ACTIVE
}
```

```proto
enum Status {
  // Initial Status
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
}
```

JSON encoding uses the shorter string (e.g. `ACTIVE` rather than
`STATUS_ACTIVE`) but will decode either form.

TODO: Clarify unspecified as '', null or omitted, in both the implementation and
the docs

### Map and Array

Arrays and Maps are defined as fields with a type of `array` or `map` as a
prefix to the sub-type.

- The sub-type can be anything other than a map or array. (constraint carried
  over from proto)
- Map keys are strings, as in JSON.

```j5s
field names array:string
field ages map:integer:INT32
```

```proto
repeated string names = 1;
map<string, int32> ages = 2;
```


### Scalar Types


| J5 Type         | Proto Type                  | JSON Type           | 
| --------------- | --------------------------- | ------------------- | 
| string          | string                      | string              |
| bool            | bool                        | bool (true,false)   |
| integer:INT32   | int32                       | unquoted literal    |
| integer:INT64   | int64                       | quoted string       |
| integer:UINT32  | uint32                      | unquoted literal    |
| integer:UINT64  | uint64                      | quoted string       |
| float:FLOAT32   | float                       | unquoted literal    |
| float:FLOAT64   | double                      | unquoted literal    |
| bytes           | bytes                       | base64 std string   |
| timestamp       | google.protobuf.Timestamp   | RFC3339 string      |
| date            | j5.types.date.v1.Date       | string "YYYY-MM-DD" |
| decimal         | j5.types.decimal.v1.Decimal | quoted string       |
| key             | string (with annotation)    | string              |


'quoted literal' means a numerical string with quotes, e.g. `"123"`, and
'unquoted literal' means a numerical string without quotes, e.g. `123`.

The J5 Codec translates between JSON and Proto representations. It produces the
representations above, but accepts a more flexible range of inputs:

- All number types (ints, floats, decimal) can be represented as a quoted or unquoted
- Base64 can be encoded with either URL or Standard encoding, with or without
  padding


Services
========

A service is a collection synchronous Request-Response endpoints, mapped as JSON
over HTTP requests from the outside, and to gRPC calls internally.


```j5s
package foo.v1;

service Foo {
  basePath = "/foo/v1"
  method Bar {
    httpMethod = "GET"
    httpPath = "/bar"

    request {
    }

    response {
      field name string
    }
  }
}
```

Generates the proto in a sub-package, foo.v1.service, with the following
structure:

```proto

service FooService {
  rpc Bar(BarRequest) returns (BarResponse) {
    option (google.api.http) = {get: "/foo/v1/bar"};
  }
}

message BarRequest {
}

message BarResponse {
  string name = 1;
}
```

Topics
======

A topic is similar to a Service, however the endpoints do not return a response,
these are used for messaging between services, relying on the o5-messaging repo.

`topic {name} {type}`

### Publish

A publish topic has one or more message blocks, designed for simply sending a
message from one application to another.


```j5s
topic Foo publish {
  message PostFoo {
    field fooId key:id62
  }
}
```

Converts to the below, in a sub-package `foo.v1.topic`.

```proto

service FooTopic {
  option (j5.messaging.v1.service) = {
    topic_name: "foo"
    publish: {
    }
  };

  rpc PostFoo(PostFooMessage) returns (google.protobuf.Empty) {}
}

message PostFooMessage {
...
}
```

### ReqRes

A request-response topic has two messages: A request and a reply.
Both messages have the 'request metadata' field, which the requester can use to
store context, the replier must copy the request metadata from the request to
the reply.

```j5s
topic Foo reqres {
    request {
        field fooId key:id62
    }

    reply {
        field name string
    }
}
```

Converts to the below, in a sub-package `foo.v1.topic`.

Note the automatic inclusion of the Request field.

```proto
service FooRequestTopic {
  option (j5.messaging.v1.service) = {
    topic_name: "foo"
    request: {
    }
  };

  rpc FooRequest(FooRequestMessage) returns (google.protobuf.Empty) {}
}

service FooReplyTopic {
  option (j5.messaging.v1.service) = {
    topic_name: "foo"
    reply: {
    }
  };

  rpc FooReply(FooReplyMessage) returns (google.protobuf.Empty) {}
}

message FooRequestMessage {
  option (j5.ext.v1.message).object = {};

  j5.messaging.v1.RequestMetadata request = 1 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  // further fields
}

message FooReplyMessage {
  option (j5.ext.v1.message).object = {};

  j5.messaging.v1.RequestMetadata request = 1 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  // further fields

}
```

### Upsert

Services may publish an 'upsert' message, which is usually linked to a state machine, either as the full state data or a derived summary.

Upserts work like Publish, but with one single message, and an enforced field of
'upsert metadata', similar to the Request context in ReqRes.

The structure of the upsert metadata message allows consumers to update a local
database with the latest state in a generic way.

Upsert has exactly one message.

```j5s
topic Foo upsert {
    message UpsertFoo {
        field fooId key:id62
    }
}
```

Converts to the below, in a sub-package `foo.v1.topic`.

```proto
service FooUpsertTopic {
  option (j5.messaging.v1.service) = {
    topic_name: "foo"
    upsert: {
    }
  };

  rpc UpsertFoo(UpsertFooMessage) returns (google.protobuf.Empty) {}
}

message UpsertFooMessage {
  option (j5.ext.v1.message).object = {};

  j5.messaging.v1.UpsertMetadata upsert = 1 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  // further fields
}
```

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

JSON Codec
==========

Currently only JSON to Proto encoding is implemented. In the future it should be possible to use any number of wire formats, including
XML, Avro... and even Proto as a full round-trip.

J5 does NOT follow the protojson rules, focusing on the client-side experience and conventions of JSON driven rest-like APIs. 




Configuration Files
-------------------

### Repo root

A Repo is more or less a git repo, but doesn't strictly have to be.

A repo can store source files, generated output, or both.

The root of a repo is marked with a `j5.yaml` file, which is the entry point for
all configuration.

The repo config file is deifned at `j5.config.v1.RepoConfigFile`.

### Package

A Package is a versioned namespace for source files. The name of the package is
any number of dot-separated strings ending in a version 'v1'. e.g. 'foo.v1' or
'foo.bar.v1' etc.

Schemas are defined in the package root.

Methods and Topics use gRPC service notations, and are defined in
'sub-packages', which are a single name under the root of the package. e.g.
`foo.v1.service` or `foo.v1.topic`.

The sub-package types are defined at the bundle level, in the bundle's config file.


### Bundle

A Bundle is a collection of packages and their source files.

Each bundle has its own `j5.yaml` file defined at `j5.config.v1.BundleConfigFile`

A bundle can optionally be 'published' by adding a registry config item, giv
ing
it an org/name structure similar to github. When a bundle has a publish config,
it can be pushed to a registry server, implemented at `github.com/pentops/registry`.

There is no central registry, and a registry is not strictly required, as
imports can also use git repositories.


### Generate

In the Repo config file, a `generate` section can be defined, which is a list of
code generation targets for the repo. Each target defines one or more inputs
which relate to bundles, an optput path and a list of plugins to run.

Each Plugin is either a PLUGIN_PROTO - meaning a protoc plugin, or J5_CLIENT
which is j5's own version of protoc, taking the a J5 schema instead.

