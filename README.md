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
