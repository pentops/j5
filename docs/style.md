Style Guide
===========

## Identifier Names

Limited by Protobuf and JSON, only ASCII characters are permitted in identifiers.

In all cases, Initialism / abbreviations are single words:
    - `DNS_REQUEST`
    - `DnsRequest` (not DNSRequest)
    - `dnsRequest`
    - `FOO_ID`
    - `FooId` (not `FooID`)
    - `fooId` (not `fooID`)

### Field Names

Field names are `camelCase`

- Must begin with a lowercase letter
- uppercase letters indicate a new word
- `[a-z][a-zA-Z0-9]*`

### Schema names

Object, Enum, Oneof, Entity and Polymorph types are all schemas.

Schema names are `PascalCase` (aka `UpperSnakeCase`)

- Must begin with an uppercase letter
- uppercase letters indicate a new word
- `[A-Z][a-zA-Z0-9]*`

### Enum values

Enum values are `UPPER_SNAKE_CASE`

- words separated by underscore
- words contain uppercase letters and numbers
- words can begin with any allowed character
- words must have at least one character
- `[A-Z0-9]+(_[A-Z0-9]+)*`

### Package names

Package names are `lower_snake_case` separated by dots.

The last segment must be a short version number, e.g. `v1`, `v2`, etc.,
matching `V[0-9]+`.

Sub-packages are not allowed in `.j5s` files. service and topic definitions
translate to sub-packages in `.proto` files.

For each segment:

- words separated by underscore
- words contain lowercase letters and numbers
- words can begin with any allowed character
- words must have at least one character
- `[a-z0-9]+(_[a-z0-9]+)*`


### Topic and Service Names

Both the Topic and Service itself, and the methods within them, are
`PascalCase`.

- Must begin with an uppercase letter
- uppercase letters indicate a new word
- `[A-Z][a-zA-Z0-9]*`
