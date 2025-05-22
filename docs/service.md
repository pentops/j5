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
