// Code generated by protoc-gen-go-o5-messaging. DO NOT EDIT.
// versions:
// - protoc-gen-go-o5-messaging 0.0.0
// source: j5st/v1/topic/foo.p.j5s.proto

package j5st_tpb

import (
	context "context"
	messaging_pb "github.com/pentops/o5-messaging/gen/o5/messaging/v1/messaging_pb"
	o5msg "github.com/pentops/o5-messaging/o5msg"
)

// Service: FooPublishTopic
// Method: FooEvent

func (msg *FooEventMessage) O5MessageHeader() o5msg.Header {
	header := o5msg.Header{
		GrpcService:      "j5st.v1.topic.FooPublishTopic",
		GrpcMethod:       "FooEvent",
		Headers:          map[string]string{},
		DestinationTopic: "foo_publish",
	}
	header.Extension = &messaging_pb.Message_Event_{
		Event: &messaging_pb.Message_Event{
			EntityName: "j5st.v1.Foo",
		},
	}
	return header
}

type FooPublishTopicTxSender[C any] struct {
	sender o5msg.TxSender[C]
}

func NewFooPublishTopicTxSender[C any](sender o5msg.TxSender[C]) *FooPublishTopicTxSender[C] {
	sender.Register(o5msg.TopicDescriptor{
		Service: "j5st.v1.topic.FooPublishTopic",
		Methods: []o5msg.MethodDescriptor{
			{
				Name:    "FooEvent",
				Message: (*FooEventMessage).ProtoReflect(nil).Descriptor(),
			},
		},
	})
	return &FooPublishTopicTxSender[C]{sender: sender}
}

type FooPublishTopicCollector[C any] struct {
	collector o5msg.Collector[C]
}

func NewFooPublishTopicCollector[C any](collector o5msg.Collector[C]) *FooPublishTopicCollector[C] {
	collector.Register(o5msg.TopicDescriptor{
		Service: "j5st.v1.topic.FooPublishTopic",
		Methods: []o5msg.MethodDescriptor{
			{
				Name:    "FooEvent",
				Message: (*FooEventMessage).ProtoReflect(nil).Descriptor(),
			},
		},
	})
	return &FooPublishTopicCollector[C]{collector: collector}
}

type FooPublishTopicPublisher struct {
	publisher o5msg.Publisher
}

func NewFooPublishTopicPublisher(publisher o5msg.Publisher) *FooPublishTopicPublisher {
	publisher.Register(o5msg.TopicDescriptor{
		Service: "j5st.v1.topic.FooPublishTopic",
		Methods: []o5msg.MethodDescriptor{
			{
				Name:    "FooEvent",
				Message: (*FooEventMessage).ProtoReflect(nil).Descriptor(),
			},
		},
	})
	return &FooPublishTopicPublisher{publisher: publisher}
}

// Method: FooEvent

func (send FooPublishTopicTxSender[C]) FooEvent(ctx context.Context, sendContext C, msg *FooEventMessage) error {
	return send.sender.Send(ctx, sendContext, msg)
}

func (collect FooPublishTopicCollector[C]) FooEvent(sendContext C, msg *FooEventMessage) {
	collect.collector.Collect(sendContext, msg)
}

func (publish FooPublishTopicPublisher) FooEvent(ctx context.Context, msg *FooEventMessage) error {
	return publish.publisher.Publish(ctx, msg)
}

// Service: FooSummaryTopic
// Method: FooSummary

func (msg *FooSummaryMessage) O5MessageHeader() o5msg.Header {
	header := o5msg.Header{
		GrpcService:      "j5st.v1.topic.FooSummaryTopic",
		GrpcMethod:       "FooSummary",
		Headers:          map[string]string{},
		DestinationTopic: "foo_summary",
	}
	header.Extension = &messaging_pb.Message_Upsert_{
		Upsert: &messaging_pb.Message_Upsert{
			EntityName: "j5st.v1.Foo",
		},
	}
	return header
}

type FooSummaryTopicTxSender[C any] struct {
	sender o5msg.TxSender[C]
}

func NewFooSummaryTopicTxSender[C any](sender o5msg.TxSender[C]) *FooSummaryTopicTxSender[C] {
	sender.Register(o5msg.TopicDescriptor{
		Service: "j5st.v1.topic.FooSummaryTopic",
		Methods: []o5msg.MethodDescriptor{
			{
				Name:    "FooSummary",
				Message: (*FooSummaryMessage).ProtoReflect(nil).Descriptor(),
			},
		},
	})
	return &FooSummaryTopicTxSender[C]{sender: sender}
}

type FooSummaryTopicCollector[C any] struct {
	collector o5msg.Collector[C]
}

func NewFooSummaryTopicCollector[C any](collector o5msg.Collector[C]) *FooSummaryTopicCollector[C] {
	collector.Register(o5msg.TopicDescriptor{
		Service: "j5st.v1.topic.FooSummaryTopic",
		Methods: []o5msg.MethodDescriptor{
			{
				Name:    "FooSummary",
				Message: (*FooSummaryMessage).ProtoReflect(nil).Descriptor(),
			},
		},
	})
	return &FooSummaryTopicCollector[C]{collector: collector}
}

type FooSummaryTopicPublisher struct {
	publisher o5msg.Publisher
}

func NewFooSummaryTopicPublisher(publisher o5msg.Publisher) *FooSummaryTopicPublisher {
	publisher.Register(o5msg.TopicDescriptor{
		Service: "j5st.v1.topic.FooSummaryTopic",
		Methods: []o5msg.MethodDescriptor{
			{
				Name:    "FooSummary",
				Message: (*FooSummaryMessage).ProtoReflect(nil).Descriptor(),
			},
		},
	})
	return &FooSummaryTopicPublisher{publisher: publisher}
}

// Method: FooSummary

func (send FooSummaryTopicTxSender[C]) FooSummary(ctx context.Context, sendContext C, msg *FooSummaryMessage) error {
	return send.sender.Send(ctx, sendContext, msg)
}

func (collect FooSummaryTopicCollector[C]) FooSummary(sendContext C, msg *FooSummaryMessage) {
	collect.collector.Collect(sendContext, msg)
}

func (publish FooSummaryTopicPublisher) FooSummary(ctx context.Context, msg *FooSummaryMessage) error {
	return publish.publisher.Publish(ctx, msg)
}
