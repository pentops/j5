package j5grpc

import (
	"fmt"
	"sort"

	"github.com/pentops/j5/gen/j5/messaging/v1/messaging_j5pb"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type InfoProvider interface {
	GetServiceInfo() map[string]grpc.ServiceInfo
}

func PrintServerInfo(server InfoProvider) error {
	info := server.GetServiceInfo()
	subscriptions := make([]string, 0)

	paths := make([]string, 0)

	for name := range info {
		desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(name))
		if err != nil {
			return err
		}

		serviceDesc, ok := desc.(protoreflect.ServiceDescriptor)
		if !ok {
			return fmt.Errorf("not a service: %s", name)
		}

		serviceOpt := proto.GetExtension(desc.Options(), messaging_j5pb.E_Service).(*messaging_j5pb.ServiceConfig)
		if serviceOpt != nil {
			var role string
			switch rr := serviceOpt.Role.(type) {
			case *messaging_j5pb.ServiceConfig_Publish_:
				role = "publish"
			case *messaging_j5pb.ServiceConfig_Request_:
				role = "request"
			case *messaging_j5pb.ServiceConfig_Reply_:
				role = "reply"
			case *messaging_j5pb.ServiceConfig_Event_:
				role = fmt.Sprintf("event %s", rr.Event.EntityName)
			case *messaging_j5pb.ServiceConfig_Upsert_:
				role = fmt.Sprintf("upsert %s", rr.Upsert.EntityName)
			}

			var topic string
			if serviceOpt.TopicName != nil {
				topic = *serviceOpt.TopicName
			}

			subscriptions = append(subscriptions, fmt.Sprintf("  - name: \"/%s\" # %s as %s", name, role, topic))
			continue
		}

		for i := 0; i < serviceDesc.Methods().Len(); i++ {
			method := serviceDesc.Methods().Get(i)
			fmt.Printf("  %s\n", method.FullName())

			httpOpt := proto.GetExtension(method.Options(), annotations.E_Http).(*annotations.HttpRule)
			if httpOpt == nil {
				return fmt.Errorf("no http rule on %s", method.FullName())
			}

			var httpMethod string
			var httpPath string

			switch pt := httpOpt.Pattern.(type) {
			case *annotations.HttpRule_Get:
				httpMethod = "GET"
				httpPath = pt.Get
			case *annotations.HttpRule_Post:
				httpMethod = "POST"
				httpPath = pt.Post
			case *annotations.HttpRule_Put:
				httpMethod = "PUT"
				httpPath = pt.Put
			case *annotations.HttpRule_Delete:
				httpMethod = "DELETE"
				httpPath = pt.Delete
			case *annotations.HttpRule_Patch:
				httpMethod = "PATCH"
				httpPath = pt.Patch

			default:
				return fmt.Errorf("unsupported http method %T", pt)
			}

			paths = append(paths, fmt.Sprintf("%s %s", httpPath, httpMethod))
		}

	}

	sort.Strings(paths)
	for _, path := range paths {
		fmt.Println(path)
	}

	sort.Strings(subscriptions)
	fmt.Println("subscriptions:")
	for _, sub := range subscriptions {
		fmt.Println(sub)
	}

	return nil
}
