Builds
======

## Go Mod Versioning & Tech Debt

protoc tools allow generators to generate code files only.

In go, that means only `.go` files are generated.

This applies to the 'standard' go and grpc plugins, as well as custom ones in
pentops (protostate, sugar, messaging).

The generated code has imports, but there is no mechanism in protoc to specify
and version those imports in the go.mod file of the generated code.

J5's builder has an 'outputFormat' overlay, which supports building a go.mod
file, by specifying the direct dependencies of the mod file directly.

To update those dependencies requires manual steps:

1. Publish the code as is

`j5 publish --bundle foo --dest ../tmp-api`

2. Use the standard go get command to update the go.mod file as required

`go get github.com/pentops/protostate@v0.0.0-20231001000000-abcdef123456`

3. Manually copy back the result to the `publish[].outputFormat.goProxy.deps`
   field in the j5.bundle.yaml file.

The libraries and versions will be added directly to the go.mod file as direct
dependencies.

There is no need to specify all the transitive dependencies, as the go
dependency management system will handle those automatically when the code
is imported to another project, but this may require a `go mod tidy` step in
the importing project.

This process is only required when the generated code requires features which
are not available in the existing libraries, i.e., it is used to force an update
top a minimum version, the rest is on the go tool's rules to hopefully give sane
results. At the end of the day, importing code may also manage its own
dependencies as a fallback.
