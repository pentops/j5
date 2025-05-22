# Schema

J5 schemas define objects and fields, similar to JSON Schema and 'pure proto' (not gRPC).

The meta-structure of schemas is defined as proto files in the [j5.schema.v1](https://github.com/pentops/j5/blob/main/proto/j5/j5/schema/v1/schema.proto) package.

## Root Types

Schema definitions which can be registered directly to a package, independently to an object
or wrapper type.

Root types may also be nested within Objects and Oneofs, and inline with field
definitions.


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

### Polymorph

A polymorph is a merge between a oneof and an any field.

The value of the any is constrained to the listed types on the polymorph.

It allows for reverse-imports where types are known in advance, but the
members are in packages which import and use the definition.

The proto implementation is a message with a single 'value' field, which is a
j5.types.any.v1

The listed types must be fully qualified type names, they may be in the same
package or in a separate package, however if they are in a separate package, the
packages they belong to cannot be imported directly or indirectly by the
package containing the polymorph.

The implementations of the polymorph must also mark themselves as such, which
does require an import.

The implementations are checked against the polymorph's type list to ensure they
are listed.

```j5s
package foo.v1

polymorph FooMorph {
   member bar.v1.Bar
   member baz.v1.Baz
}

```

```j5s

package bar.v1

import foo.v1 as foo

object Bar {
  polymorphMember foo.FooMorph

  field barId key:id62
}
```



## Field Types

Fields in an object or oneof are defined by `j5.schema.v1.ObjectProperty`.

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

### Inline Types

When a field references a root type, the type can be defined
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



### Map and Array

Arrays and Maps are defined as fields with a type of `array` or `map` as a
prefix to the sub-type.

- The sub-type can be anything other than a map or array. (constraint carried over from proto)
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

