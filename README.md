# J5 API Definitions

J5 is a data schema structure, language and toolkit for defining Event-Driven APIs.

As with all pentops tools, it is very opinionated. J5 definitions are intended to work within the wider pentops ecosystem, especially for [Messaging](https://github.com/pentops/o5-messaging) and [State Machines](https://github.com/pentops/protostate)

J5 wraps a **subset** of the data types and schemas of [Protocol Buffers](https://protobuf.dev/). The goal of the project is different to Protocol Buffers, and different trade-offs are made. J5 schemas are based on [proto descriptors](https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/descriptor.proto), and can be fully represented in .proto files, (leaning heavily on annotations), but not all Protocol Buffer structures can be represented in J5.

---

## J5 Tool

**Installation**:

```bash
go install github.com/pentops/j5/cmd/j5@latest
```

`j5 --help` for brief usage

---

## Files

J5 APIs can be defined in either `.proto` files using [proto3 syntax](https://protobuf.dev/programming-guides/proto3/), or in `.j5s` files, based on [bcl](https://github.com/pentops/bcl.go), a config format designed specifically for defining J5 schemas and configuration.

J5 does not support all proto structures, so .proto files must be structured according to the rigid rules of J5, including annotations.

J5 files follow the style guide defined in [docs/style.md](docs/style.md) and can be checked and formatted using `j5 j5s fmt` and `j5 j5s lint`.

### Descriptions

The pipe syntax is used for descriptions, and occurs in many schema elements. Descriptions must be the first element in the body, and may span multiple lines. Where there is no body, a single line description can be added to the end of the definition line:

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

JS files exist in a 'package', which must match their path from the root. The package declaration should be the first non-comment line in the file.

```j5s
package foo.bar.v1
```

The file should be in the directory `/foo/bar/v1/` relative to the 'bundle root'.

An Import declaration at the top of a file names an entire package and brings the package into scope by either the package name ('bar' not 'v1') or by the alias name.

`import <package>:<alias>`

```j5s
package foo.bar.v1

import foo.baz.v1:baz

object Foo {
    field bar baz.Bar
}
```

The prefix 'baz' can then be used to refer to Objects, Oneof and Enums in the baz package, regardless of the file name.

A j5s source can import a proto source, and v/v. In proto, imports specify the filename of the schema. These are converted on the fly. The filename for a j5s file, when imported from proto, is the full j5s filename followed by `.proto`, e.g. `/foo/v1/bar.j5s` can be imported as `import "/foo/v1/bar.j5s.proto"`.

---

## Schemas

J5 schemas define objects and fields, similar to JSON Schema and 'pure proto' (not gRPC).

The meta-structure of schemas is defined as proto files in the [j5.schema.v1](https://github.com/pentops/j5/blob/main/proto/j5/j5/schema/v1/schema.proto) package.

See [docs/schema.md](docs/schema.md) for the basic schema types.

Within schema files, the following elements can also be defined, similar to gRPC definitions in proto files.

### Services

A service is a collection synchronous Request-Response endpoints, mapped as JSON over HTTP requests from the outside, and to gRPC calls internally.

See [docs/service.md](docs/service.md) for more info.

### Topics

A topic is similar to a Service, however the endpoints do not return a response, these are used for messaging between services, relying on the o5-messaging repo.

See [docs/topic.md](docs/topic.md) for more info.

### Entities

An entity is a shortcut for a number of pre-existing elements, but enforcing strict convention.

See [docs/entity.md](docs/entity.md) for more info.

---

## JSON Codec

Currently only JSON to Proto encoding is implemented. In the future it should be possible to use any number of wire formats, including XML, Avro... and even Proto as a full round-trip.

J5 does NOT follow the protojson rules, focusing on the client-side experience and conventions of JSON driven rest-like APIs.

---

## Configuration

### Repo root

A Repo is more or less a git repo, but doesn't strictly have to be.

A repo can store source files, generated output, or both.

The root of a repo is marked with a `j5.yaml` file, which is the entry point for all configuration.

The repo config file is defined at `j5.config.v1.RepoConfigFile`.

### Package

A Package is a versioned namespace for source files. The name of the package is any number of dot-separated strings ending in a version 'v1'. e.g. 'foo.v1' or 'foo.bar.v1' etc.

Schemas are defined in the package root.

Methods and Topics use gRPC service notations, and are defined in 'sub-packages', which are a single name under the root of the package. e.g. `foo.v1.service` or `foo.v1.topic`.

The sub-package types are defined at the bundle level, in the bundle's config file.

### Bundle

A Bundle is a collection of packages and their source files.

Each bundle has its own `j5.yaml` file defined at `j5.config.v1.BundleConfigFile`

A bundle can optionally be 'published' by adding a registry config item, giving it an org/name structure similar to github. When a bundle has a publish config, it can be pushed to a registry server, implemented at `github.com/pentops/registry`.

There is no central registry, and a registry is not strictly required, as imports can also use git repositories.

### Generate

In the Repo config file, a `generate` section can be defined, which is a list of code generation targets for the repo. Each target defines one or more inputs which relate to bundles, an optput path and a list of plugins to run.

Each Plugin is either a PLUGIN_PROTO - meaning a protoc plugin, or J5_CLIENT which is j5's own version of protoc, taking the a J5 schema instead.