Topics
======

A topic is similar to a Service, however the endpoints do not return a response,
these are used for messaging between services, relying on the o5-messaging repo.

`topic {name} {type}`

### Publish

A publish topic has one or more message blocks, designed for simply sending a
message from one application to another.


```j5s
topic Foo publish {
  message PostFoo {
    field fooId key:id62
  }
}
```

Converts to the below, in a sub-package `foo.v1.topic`.

```proto

service FooTopic {
  option (j5.messaging.v1.service) = {
    topic_name: "foo"
    publish: {
    }
  };

  rpc PostFoo(PostFooMessage) returns (google.protobuf.Empty) {}
}

message PostFooMessage {
...
}
```

### ReqRes

A request-response topic has two messages: A request and a reply.
Both messages have the 'request metadata' field, which the requester can use to
store context, the replier must copy the request metadata from the request to
the reply.

```j5s
topic Foo reqres {
    request {
        field fooId key:id62
    }

    reply {
        field name string
    }
}
```

Converts to the below, in a sub-package `foo.v1.topic`.

Note the automatic inclusion of the Request field.

```proto
service FooRequestTopic {
  option (j5.messaging.v1.service) = {
    topic_name: "foo"
    request: {
    }
  };

  rpc FooRequest(FooRequestMessage) returns (google.protobuf.Empty) {}
}

service FooReplyTopic {
  option (j5.messaging.v1.service) = {
    topic_name: "foo"
    reply: {
    }
  };

  rpc FooReply(FooReplyMessage) returns (google.protobuf.Empty) {}
}

message FooRequestMessage {
  option (j5.ext.v1.message).object = {};

  j5.messaging.v1.RequestMetadata request = 1 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  // further fields
}

message FooReplyMessage {
  option (j5.ext.v1.message).object = {};

  j5.messaging.v1.RequestMetadata request = 1 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  // further fields

}
```

### Upsert

Services may publish an 'upsert' message, which is usually linked to a state machine, either as the full state data or a derived summary.

Upserts work like Publish, but with one single message, and an enforced field of
'upsert metadata', similar to the Request context in ReqRes.

The structure of the upsert metadata message allows consumers to update a local
database with the latest state in a generic way.

Upsert has exactly one message.

```j5s
topic Foo upsert {
    message UpsertFoo {
        field fooId key:id62
    }
}
```

Converts to the below, in a sub-package `foo.v1.topic`.

```proto
service FooUpsertTopic {
  option (j5.messaging.v1.service) = {
    topic_name: "foo"
    upsert: {
    }
  };

  rpc UpsertFoo(UpsertFooMessage) returns (google.protobuf.Empty) {}
}

message UpsertFooMessage {
  option (j5.ext.v1.message).object = {};

  j5.messaging.v1.UpsertMetadata upsert = 1 [
    (buf.validate.field).required = true,
    (j5.ext.v1.field).object = {}
  ];

  // further fields
}
```


