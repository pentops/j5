package e2e

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/pentops/flowtest/jsontest"
	"github.com/pentops/j5/gen/j5/client/v1/client_j5pb"
	"github.com/pentops/j5/lib/j5codec"
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

	codec := j5codec.NewCodec(j5codec.WithIncludeEmpty())

	baseJSON, err := codec.ProtoToJSON(ma.ProtoReflect())
	if err != nil {
		ma.t.Fatalf("FATAL: ProtoToJSON failed: %v", err)
	}

	out := bytes.NewBuffer(nil)
	if err := json.Indent(out, baseJSON, "", "  "); err != nil {
		ma.t.Fatalf("FATAL: json.Indent failed: %v", err)
	}

	return jsontest.NewTestAsserter(ma.t, out.String())
}
