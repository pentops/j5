Custom Proto API
================

These library functions are used to gain more control over alternate encoding
methods for proto over the standard protojson encoding.

These methods deliberately break the specification of proto to JSON encoding.
The standard encoding is great for interoperability at the expense of
beautifull or elegant external API design.

The alternative here allows an API designer to think through and customize the
external JSON representation of their proto/gRPC APIs.

Whilst all comments mention JSON, these libraries should be built with XML and
even Avro in mind as alternate representations.




## Entities

## Source Image

`j5.source.v1.SourceImage`

Represents the source protobuf files of a package.

Extends google.protobuf.Descriptor with package metadata and codec
configuration.


## Schema

`j5.schema.v1.Schema`

Represents a data structure. Somewhere between JSON Schema and proto
descriptors.

## Package

`j5.schema.v1.Package`

Package namespace containing Methods, prose, Stateful Entities and published Events.

## API

`j5.schema.v1.API`

A collection of packages.

Contains Schemas at the top level to map to Swagger/OAS, but these should
probably be nested within packages.
