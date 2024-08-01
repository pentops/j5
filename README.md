J5 API Schemas
==============

As with all pentops repos, a very opinionated take.

J5 Schemas are currently specified as - and sort of similar to - proto files,
they define data shapes, methods and events to expose as an API.

The goal of the project is to define Event-Driven APIs and data schemas for storage,
which is not the goal of proto and gRPC.

Currently only JSON encoding is implemented. It does NOT follow the protojson
rules, allowing far more flexibility in designing an API for a Client.

In the future it should be possible to use any number of wire formats, including
XML, Avro... and even Proto as a full round-trip.


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

A bundle can optionally be 'published' by adding a registry config item, giving
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
