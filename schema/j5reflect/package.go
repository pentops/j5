package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

type API struct {
	Packages []*Package
	Metadata *schema_j5pb.Metadata
}

type Package struct {
	Name     string
	Label    string
	Services []*Service
	Events   []*Event
}

type Event struct {
	Package *Package
	Name    string
	Schema  RootSchema
}

type Service struct {
	Package *Package
	Name    string
	Methods []*Method
}

func (ss *Service) ToJ5Proto() (*schema_j5pb.Service, error) {
	service := &schema_j5pb.Service{
		Name:    ss.Name,
		Methods: make([]*schema_j5pb.Method, len(ss.Methods)),
	}

	for i, method := range ss.Methods {
		m, err := method.ToJ5Proto()
		if err != nil {
			return nil, err
		}
		service.Methods[i] = m
	}

	return service, nil
}

type Method struct {
	GRPCMethodName string
	HTTPPath       string
	HTTPMethod     schema_j5pb.HTTPMethod

	HasBody bool

	Request  RootSchema
	Response RootSchema

	Service *Service
}

func (mm *Method) ToJ5Proto() (*schema_j5pb.Method, error) {

	requestSchema, ok := mm.Request.(*ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("request should be an object, got %T", mm.Request)
	}

	requestBody, err := requestSchema.ToJ5Proto()
	if err != nil {
		return nil, err
	}

	responseSchema, ok := mm.Response.(*ObjectSchema)
	if !ok {
		return nil, fmt.Errorf("response schema was not an object: %T", mm.Response)
	}

	responseBody, err := responseSchema.ToJ5Proto()
	if err != nil {
		return nil, err
	}

	return &schema_j5pb.Method{
		FullGrpcName: fmt.Sprintf("/%s.%s/%s", mm.Service.Package.Name, mm.Service.Name, mm.GRPCMethodName),
		Name:         mm.GRPCMethodName,
		HttpMethod:   mm.HTTPMethod,
		HttpPath:     mm.HTTPPath,
		ResponseBody: responseBody,
		RequestBody:  requestBody,
	}, nil
}
