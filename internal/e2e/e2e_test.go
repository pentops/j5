package e2e

import (
	"testing"

	"github.com/pentops/j5/gen/j5/list/v1/list_j5pb"
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

func TestInt(t *testing.T) {

	fb := NewFileBuilder("test/v1/test.j5s")
	foo := fb.Object("Foo")
	foo.Def.Properties = append(foo.Def.Properties, &schema_j5pb.ObjectProperty{
		Name:     "int",
		Required: true,
		Schema: &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Integer{
				Integer: &schema_j5pb.IntegerField{
					Format: schema_j5pb.IntegerField_FORMAT_INT64,
					ListRules: &list_j5pb.IntegerRules{
						Filtering: &list_j5pb.FilteringConstraint{
							Filterable: true,
						},
					},
				},
			},
		},
	})

	fb.ListObjectMethod("FooQuery", "ListFoos", &schema_j5pb.Ref{Schema: "Foo"})

	client := NewClientExplorer(t, fb.BuildClientAPI(t))

	method := client.GetPackage("test.v1").GetMethod("ListFoos").JSONAsserter()
	method.Print()

	method.AssertEqual("request.list.filterableFields.0.name", "int")
}

func TestDate(t *testing.T) {

	fb := NewFileBuilder("test/v1/test.j5s")
	foo := fb.Object("Foo")
	foo.Def.Properties = append(foo.Def.Properties, &schema_j5pb.ObjectProperty{
		Name:     "date",
		Required: true,
		Schema: &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Date{
				Date: &schema_j5pb.DateField{
					ListRules: &list_j5pb.DateRules{
						Filtering: &list_j5pb.FilteringConstraint{
							Filterable: true,
						},
					},
				},
			},
		},
	})
	fb.ListObjectMethod("FooQuery", "ListFoos", &schema_j5pb.Ref{Schema: "Foo"})

	client := NewClientExplorer(t, fb.BuildClientAPI(t))

	pkg := client.GetPackage("test.v1")

	fooSchema := pkg.GetObject("Foo").JSONAsserter()
	fooSchema.AssertEqual("properties.0.name", "date")
	fooSchema.AssertEqual("properties.0.schema.date.listRules.filtering.filterable", true)

	method := pkg.GetMethod("ListFoos").JSONAsserter()
	method.Print()

	method.AssertEqual("request.list.filterableFields.0.name", "date")

}

func TestOneof(t *testing.T) {

	fb := NewFileBuilder("test/v1/test.j5s")
	foo := fb.Object("Foo")
	foo.Def.Properties = append(foo.Def.Properties, &schema_j5pb.ObjectProperty{
		Name:     "oneof",
		Required: true,
		Schema: &schema_j5pb.Field{
			Type: &schema_j5pb.Field_Oneof{
				Oneof: &schema_j5pb.OneofField{
					Schema: &schema_j5pb.OneofField_Ref{
						Ref: &schema_j5pb.Ref{
							Schema: "OneofType",
						},
					},
					ListRules: &list_j5pb.OneofRules{
						Filtering: &list_j5pb.FilteringConstraint{
							Filterable: true,
						},
					},
				},
			},
		},
	})

	oneof := fb.Oneof("OneofType")
	oneof.Def.Properties = append(oneof.Def.Properties, &schema_j5pb.ObjectProperty{
		Name: "typeA",
		Schema: &schema_j5pb.Field{

			Type: &schema_j5pb.Field_Object{
				Object: &schema_j5pb.ObjectField{
					Schema: &schema_j5pb.ObjectField_Object{
						Object: &schema_j5pb.Object{
							Name: "TypeA",
						},
					},
				},
			},
		},
	})

	fb.ListObjectMethod("FooQuery", "ListFoos", &schema_j5pb.Ref{Schema: "Foo"})

	client := NewClientExplorer(t, fb.BuildClientAPI(t))

	pkg := client.GetPackage("test.v1")

	fooSchema := pkg.GetObject("Foo").JSONAsserter()
	fooSchema.Print()
	fooSchema.AssertEqual("properties.0.name", "oneof")
	fooSchema.AssertEqual("properties.0.schema.oneof.listRules.filtering.filterable", true)

	method := client.GetPackage("test.v1").GetMethod("ListFoos").JSONAsserter()
	method.Print()

	method.AssertEqual("request.list.filterableFields.0.name", "oneof.!type")
}
