package j5convert

import (
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pentops/golib/gl"
	"github.com/pentops/j5/gen/j5/ext/v1/ext_j5pb"
	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5s/sourcewalk"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type serviceBuilder struct {
	desc *descriptorpb.ServiceDescriptorProto
	commentSet
}

func blankService(name string) *serviceBuilder {
	return &serviceBuilder{
		desc: &descriptorpb.ServiceDescriptorProto{
			Name: gl.Ptr(name),
		},
	}
}

type MethodBuilder struct {
	desc *descriptorpb.MethodDescriptorProto
	commentSet
}

func blankMethod(name string) *MethodBuilder {
	return &MethodBuilder{
		desc: &descriptorpb.MethodDescriptorProto{
			Name:    gl.Ptr(name),
			Options: &descriptorpb.MethodOptions{},
		},
	}
}

func (ww *conversionVisitor) visitServiceFileNode(sn *sourcewalk.ServiceFileNode) error {
	return sn.Accept(sourcewalk.ServiceFileCallbacks{
		Service: func(sn *sourcewalk.ServiceNode) error {
			ww.visitServiceNode(sn)
			return nil
		},
		Object: func(on *sourcewalk.ObjectNode) error {
			ww.visitObjectNode(on)
			return nil
		},
	})
}

func (ww *conversionVisitor) visitServiceNode(node *sourcewalk.ServiceNode) {
	serviceWalker := ww.subPackageFile("service")

	service := blankService(node.Name)

	for _, method := range node.Methods {
		ww.visitServiceMethodNode(service, method)
	}

	if node.ServiceOptions != nil {
		service.desc.Options = &descriptorpb.ServiceOptions{}
		proto.SetExtension(service.desc.Options, ext_j5pb.E_Service, node.ServiceOptions)
	}

	serviceWalker.file.addService(service)
}

func (ww *conversionVisitor) visitServiceMethodNode(service *serviceBuilder, node *sourcewalk.ServiceMethodNode) {
	method := node.Schema
	methodBuilder := blankMethod(method.Name)
	methodBuilder.comment([]int32{}, method.Description)
	ww.file.ensureImport(googleApiAnnotationsImport)

	methodBuilder.desc.InputType = gl.Ptr(node.InputType)
	methodBuilder.desc.OutputType = gl.Ptr(node.OutputType)

	if node.OutputType == "google.api.HttpBody" {
		ww.file.ensureImport(googleApiHttpBodyImport)
	}

	annotation := &annotations.HttpRule{}
	reqPath := node.ResolvedPath

	if method.Request != nil {
		reqPathParts := strings.Split(node.ResolvedPath, "/")
		for idx, part := range reqPathParts {
			if strings.HasPrefix(part, ":") {
				var field *schema_j5pb.ObjectProperty
				found := make([]string, 0)
				for _, search := range method.Request.Properties {
					found = append(found, search.Name)
					if search.Name == part[1:] {
						field = search
						break
					}
				}
				if field == nil {
					ww.addErrorf(node.Source, "field %s from request path not found in %s/%s, have %s", part[1:], *service.desc.Name, method.Name, strings.Join(found, ", "))
				}

				fieldName := strcase.ToSnake(part[1:])
				reqPathParts[idx] = "{" + fieldName + "}"

			}
		}

		reqPath = strings.Join(reqPathParts, "/")
	}

	switch method.HttpMethod {
	case schema_j5pb.HTTPMethod_GET:
		annotation.Pattern = &annotations.HttpRule_Get{
			Get: reqPath,
		}
	case schema_j5pb.HTTPMethod_POST:
		annotation.Pattern = &annotations.HttpRule_Post{
			Post: reqPath,
		}
		annotation.Body = "*"

	case schema_j5pb.HTTPMethod_DELETE:
		annotation.Pattern = &annotations.HttpRule_Delete{
			Delete: reqPath,
		}
		annotation.Body = "*"

	case schema_j5pb.HTTPMethod_PATCH:
		annotation.Pattern = &annotations.HttpRule_Patch{
			Patch: reqPath,
		}
		annotation.Body = "*"

	case schema_j5pb.HTTPMethod_PUT:
		annotation.Pattern = &annotations.HttpRule_Put{
			Put: reqPath,
		}
		annotation.Body = "*"

	default:
		ww.addErrorf(node.Source, "unsupported http method %s", method.HttpMethod)
		return
	}

	proto.SetExtension(methodBuilder.desc.Options, annotations.E_Http, annotation)

	if method.Options != nil {
		proto.SetExtension(methodBuilder.desc.Options, ext_j5pb.E_Method, method.Options)
	}

	if method.ListRequest != nil {
		proto.SetExtension(methodBuilder.desc.Options, list_j5pb.E_ListRequest, method.ListRequest)
	}
	service.desc.Method = append(service.desc.Method, methodBuilder.desc)
}
