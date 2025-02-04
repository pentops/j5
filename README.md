J5 API Definitions
==================

J5 is a data schema structure, language and toolkit for defining Event-Driven APIs.

As with all pentops tools, it is very opinionated. J5 definitions are intended to work within the wider pentops ecosystem, especially for [Messaging](https://github.com/pentops/o5-messaging) and [State Machines](https://github.com/pentops/protostate)

J5 wraps a **subset** of the data types and schemas of [Protocol Buffers](https://protobuf.dev/). The goal of the project is different to Protocol Buffers, and different trade-offs are made. J5 schemas are based on [proto descriptors](https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/descriptor.proto), (leaning heavily on annotations), and can be fully represented in .proto files, but not all Protocol Buffer structures can be represented in J5.

Files
=====

J5 APIs can be defined in either  `.proto` files using [proto3 syntax](https://protobuf.dev/programming-guides/proto3/), or in `.j5s` files, a config format designed specifically for designing J5 schemas and configuration.

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

### Oneof

Like an object, but at most one key can be set at a time, and all of the
properties must be objects.

```j5s
oneof Foo {
  option bar object {
    field barId key:id62
  }
  option baz object:Baz {
    field bazId key:id62
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



JSON Codec
----------

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

