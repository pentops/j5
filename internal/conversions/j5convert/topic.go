package j5convert

import (
	"github.com/pentops/j5/gen/j5/messaging/v1/messaging_j5pb"
	"github.com/pentops/j5build/internal/sourcewalk"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func convertTopic(ww *walkContext, tn *sourcewalk.TopicFileNode) {

	tn.Accept(sourcewalk.TopicFileCallbacks{
		Topic: func(tn *sourcewalk.TopicNode) {
			convertTopicNode(ww, tn)
		},
		Object: func(on *sourcewalk.ObjectNode) {
			convertObject(ww, on)
		},
	})
}

func convertTopicNode(ww *walkContext, tn *sourcewalk.TopicNode) {
	desc := &descriptorpb.ServiceDescriptorProto{
		Name:    ptr(tn.Name + "Topic"),
		Options: &descriptorpb.ServiceOptions{},
	}

	proto.SetExtension(desc.Options, messaging_j5pb.E_Service, tn.ServiceConfig)

	for _, method := range tn.Methods {
		rpcDesc := &descriptorpb.MethodDescriptorProto{
			Name:       ptr(method.Name),
			OutputType: ptr(googleProtoEmptyType),
			InputType:  ptr(method.Request),
		}
		desc.Method = append(desc.Method, rpcDesc)
	}

	ww.file.ensureImport(messagingAnnotationsImport)
	ww.file.ensureImport(googleProtoEmptyImport)
	ww.file.addService(&ServiceBuilder{
		desc: desc,
	})
}
