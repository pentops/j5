package e2e

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/pentops/flowtest/jsontest"
	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/lib/j5codec"
	"google.golang.org/protobuf/proto"
)

type ClientAsserter struct {
	t testing.TB
	*client_j5pb.API
}

func NewClientExplorer(t testing.TB, api *client_j5pb.API) *ClientAsserter {
	return &ClientAsserter{
		API: api,
		t:   t,
	}
}

func (ca *ClientAsserter) GetPackage(name string) *PackageAsserter {
	ca.t.Helper()
	for _, pkg := range ca.Packages {
		if pkg.Name == name {
			return &PackageAsserter{
				t:       ca.t,
				Package: pkg,
			}
		}
	}
	ca.t.Fatalf("FATAL: Package %s not found in client API", name)
	return nil
}

type PackageAsserter struct {
	t testing.TB
	*client_j5pb.Package
}

func (pa *PackageAsserter) GetMethod(name string) *MethodAsserter {
	pa.t.Helper()
	for _, service := range pa.Services {
		for _, method := range service.Methods {
			if method.Method.Name == name || method.Method.FullGrpcName == name {
				return &MethodAsserter{
					t:      pa.t,
					Method: method,
				}
			}
		}
	}
	pa.t.Fatalf("FATAL: Method %s not found in package %s", name, pa.Name)
	return nil
}

type MethodAsserter struct {
	t testing.TB
	*client_j5pb.Method
}

func (ma *MethodAsserter) JSONAsserter() *jsontest.TestAsserter {
	return toJSONAsserter(ma.t, ma.Method)
}

type ObjectAsserter struct {
	t      testing.TB
	Object *schema_j5pb.Object
}

func (pa *PackageAsserter) GetObject(name string) *ObjectAsserter {
	pa.t.Helper()
	schema, ok := pa.Schemas[name]
	if !ok {
		pa.t.Fatalf("FATAL: Object %s not found in package %s", name, pa.Name)
	}
	asObject := schema.GetObject()
	if asObject == nil {
		pa.t.Fatalf("FATAL: Object %s in package %s is not an object schema", name, pa.Name)
	}

	return &ObjectAsserter{
		t:      pa.t,
		Object: asObject,
	}
}

func (oa *ObjectAsserter) JSONAsserter() *jsontest.TestAsserter {
	return toJSONAsserter(oa.t, oa.Object)
}

func toJSONAsserter(t testing.TB, v proto.Message) *jsontest.TestAsserter {
	codec := j5codec.NewCodec(j5codec.WithIncludeEmpty())
	baseJSON, err := codec.ProtoToJSON(v.ProtoReflect())
	if err != nil {
		t.Fatalf("FATAL: ProtoToJSON failed: %v", err)
	}

	out := bytes.NewBuffer(nil)
	if err := json.Indent(out, baseJSON, "", "  "); err != nil {
		t.Fatalf("FATAL: json.Indent failed: %v", err)
	}

	return jsontest.NewTestAsserter(t, out.String())
}
