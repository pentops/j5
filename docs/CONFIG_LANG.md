J5 Specification Language
=========================

## Introduction

The J5 Spec language is designed to be a human format for the j5-schema.

It takes the place of .proto files, JSON Schema files, and other formats.

Goals:
 - All about the humans reading and writing.
 - Ability to extend the language over time.
 - Support J5 state entities and other conventions.
 - Support Documentation inline with the schema.
 - Familiar, small, easy to learn and all those good things we always try to
   achieve.


### Why this and not X?

The big ticket item is that we want the J5 conventions to be built-in to
something, and that documentation is top of mind.

J5 was first designed to describe APIs to consumers, but then got into list
query and state management, there is a lot of content in these files.

- PROTO is cool, and obviously the base for J5, but the extensions are
  getting out of hand. The 'defaults' we want to define on all fields are all
  extension options, so we would end up writing a validator anyway. Proto itself
  is still used, but the config files are getting a bit much.

- HCL would suit well, but I don't want to get involved in whatever copyleft
  and right problem is happening there.

- JSON Schema is good, and close-ish to what we want for the schema definitions.
  Swagger/OAS, however, is too based in REST and ends up being an exercise in
  duck punching.

- JSON - no comments, no set structure

- JSON + Schema + extensions for comments: we are still inventing a 'language',
  but without control over the syntax. This applies for YAML, XML as well

- YAML - I love yaml, but my eyes are getting dim, I can't figure out how nested
  I am.

- TOML - Great for configuration, but it is really only designed for a few
  levels of nesting at best.

- PKL - looks interesting, but is typed in the config file itself rather than
  filling in a pre-defined schema.

- In-Code - I did consider this in two ways, either defining the schema as
  source code in a central repo, let's say it uses javascript, python or
  something, which is quite flexible and can be made into whatever structure is
  required for the task. Problem is it isn't uniform, and so it doesn't end up
  looking like a schema... it looks like procedural code... because it is.
  Other option was to emit the schemas from the code which implements, as in not
  doing schema driven development. That's a fine idea but not the goal of J5.

## Status

This is very early days. To keep from making a complete mess, I am starting with
a transpiler, much like the schema did, to convert between .proto files and .j5s
files.

That allows automated migration of the files.

Currently the parser is very basic and rigid, but there is a generalized case in
the works, which takes the basic syntax and removes the specific 'types', i.e.
this becomes a file format for specifying data in any J5 schema, including the
schema itself, config files etc.

The file extension is `.j5s` because:

- `.j5` looks like `.js` in some fonts (`5S`)
- `s` actively stands out as different
- `s` could stand for 'schema' or 'spec'?

The internal package structure of j5 is a mess, that's why it's internal, some
of it should be public.


## File Structure

Similar to Go and Buf-Proto, the directory of a file specifies the package.

Like Go, but not proto, the file name does not matter, you can freely move
content between files without changing the result. Imports import the package,
not the file, and there is no 'index' file like js modules.

## Syntax

The base syntax structures are as follows:

### Base elements

An `ident` starts with a unicode letter character, upper and lower, followed
by letters or numbers. `[a-zA-Z][a-zA-Z0-9]*`

A `reference` is a series of `ident` separated by periods. `ident(.ident)*`

A `literal` is a string, number, or boolean.
 - Strings, quoted with ""
 - Numbers, integers or floats specified 1.1 or 1
 - Booleans, true or false with no quotes
 - null, for un-setting things like partial overrides

Context defines the type of the literal, so 1 and 1.1 are both valid for floats
(i.e. you don't have to write 1.0)

Strings may span multiple lines, by escaping the end of line, and may also
escape quotes.

```j5
key = "This is a string"
key = "This \
is a string"
key = "This is a "string""
```
### Comments

Comments are C-style, `//` for single line, `/* */` for multi-line.

### Assignment

```j5
key = value
```

Keys are 'reference' type.
Values are 'literal' type.

### Directive

```j5
keyword value
package foo.bar
field foo string
```

The available keywords for directives are context dependent, and built in.
The context and keyword defines the number and type of the arguments.

Keys are 'reference' type.
Values are 'literal' type.

### Element

Elements are directives which accept a body.

### Body

The root of the document is a body, which is a series of Assignments,
Directives, comments and Definitions.

Elements use curly braces `{}` to define the body.

```j5
field foo string{
  // ... body elements
}
```

Skipping the body is equivalent to an empty body, if there is nothing to define
the body you can just skip it.

```j5
field foo string
field foo string {
}
```

### Description

Descriptions are like multi-line comments, but specifically used to describe
elements for documentation in generated code and schemas, rather than comment
about the code.

Descriptions are valid for all Elements elements and can be specified in two ways:

```j5
field foo string | Inline Description on a single line
```

```j5
field foo string {
  | Description in the body, which can span multiple lines
  | but must be specified at the start of the block.
}
```

Inline descriptions don't work with body blocks,
i.e. the following is **not valid**

```j5
field foo string { | Inline
    // ... body elements
}
```

Descriptions are not valid at the file level, use PROSE.md in the directory to
document the package.

The descriptions are to be interpreted as markdown, but start carefully until we
can build some validation rules:

- Using **bold**, *italic* and `code`: Good to go.
- Headings: Avoid until we figure out what the nesting would be. It certainly
  wouldn't make sense to have a full document with headings in the description
  of an Enum Option, for example.
- Paragraphs: In the right context, paragraphs are fine, using a blank like.
- Block-Quotes, Block Code: Not yet
- Lists: Use sparingly in the right context, nest freely
- Tables: Not yet, except for the special emum case.

Links... Are going to be a whole thing. Linking to wikipedia is fine, but the
path structure relative to the docs is not yet defined.
A special syntax for named links is pending, so when we refer to an [Account] in
a [Transaction] we can link to the definition of Account automatically, need to
consider how the scoping works and if it follows the same import rules.

Sequence and State Diagrams in mermaid are coming.

## Types and file structure

### package

At the top level, the file declares its package in full from the root.

TODO: This might not be necessary, as the file structure defines the package

```j5
package foo.bar
```

### version

```j5
version = "1.0"
```

### object

Defines a j5.schema.v1.Object

```j5
object foo {
  field bar string
}
```

### field

Defines a j5.schema.v1.ObjectProperty

```j5
field foo string
```

The keyword `field`, the field name, and the type.

```j5
field <name> <type>
```

Fields may have bodies to define rules etc

```j5
field foo string {
  required
  validate.regex = "^[a-z]+$"
}
```

'object', 'enum' and 'oneof' are valid field types.

```j5
object foo {
    field bar object {
        field baz string
    }
}
```

The `ref` directive can be used on those three types to reference the element
rather than define it inline.

```j5
object foo {
    field bar object {
        ref baz
    }
}

object baz {
    field baz string
}
```

'baz' is not a 'type', it can't be used in place of the 'object' type, so in
cross-package (or cross file) references we don't have to check the underlying
type... but also makes the parser easier.

A shorthand exists to the ref directive: `field <name> object:<type>`

```j5
object foo {
    field bar object:baz
}
```

This can be used with or without body blocks.

```j5
object foo {
    field bar object:baz {
        | Bar is a reference to Baz
    }
}
```

### enum

Defines a j5.schema.v1.Enum

```j5
enum foo {
  BAR
  BAZ
}
```

The values for the enum are elements, so they allow descriptions in either
format.

```j5
enum foo {
  option BAR | Description
  option BAZ {
    | Longer Description
  }
}
```

And a special directive for table-type enums in documentation.
(really this is just a description directive which contains a '|')

```j5
enum foo {
  doctable   | Name | Suffix
  option BAR | bar  | B
  option BAZ | baz  | Z
}
```

### oneof

Defines a j5.schema.v1.OneOf

```j5
oneof foo {
  option bar string
  option baz string
}
```

Option works like field in objects.

### service and method

Defines a j5.source.v1.Service and j5.source.v1.Method

```j5
service foo {
    path = "/foo"
    method bar {
        // Method body
        request {
            // Object body
        }
        response {
            // Object body
        }
    }
}
```

### topic and message

Defines a j5.source.v1.Topic and j5.source.v1.TopicMessage

```j5
topic foo {
    message bar {
        // Object body
    }
}
```

### entity

Defines a protostate entity, j5.source.v1.Entity (note source, not schema)

```j5
entity foo {
  field bar string

  data {
    // Object body
  }

  key foo string {
    // Field types, merged into the 'key' object
  }

  event eventName {
    // Object body defining the fields for the event
  }
  event event2 {
    // each event is a separate object
  }

  status PENDING // Enum Options for the status enum
  status ACTIVE | And descriptions work

  command create {
    // Works like an endpoint
  }
}
```


### partial and include

Defines a partial entity, which can be merged into another entity of the same
type.

```j5
partial field cusip {
    required
    validate.regex = "^[A-Z0-9]{9}$"
}

field foo string {
    include cusip
}

field bar string {
    include cusip
    validate.regex = null
}
```

The resulting Foo is required with the regex.

Bar will still be required (as useless as that is) but will not have the regex.

There is no syntax to nullify a directive. // TODO: 'unset' directive?


### import

Imports all exported elements from another file under the given namespace.

```j5
import foo.bar
import foo.bar as baz
```

The exported elements are available as `bar.element` or `baz.element` respectively, rather than requiring the full namespace to be repeated.

### export

Exports all elements for use by other namespaces.
By default, elements defined within a namespace are available to that namespace
and its children, but not to other namespaces.

```j5

// namespace/foo.j5
object foo {
  export
  field bar string
}

partial object baseline {
  field createdAt timestamp
}

// other/bar.js
import namespace.foo as baz

object qux {
  include baz.baseline

  field bar_ref {
    ref baz.foo
  }
}
```

