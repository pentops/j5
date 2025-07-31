package psmigrate

import (
	"testing"

	"github.com/pentops/j5/lib/psm"
	"github.com/pentops/j5/lib/psm/internal/testproto/gen/test/v1/test_pb"
)

func TestBuildStateMachineOneKey(t *testing.T) {

	fooSpec, err := psm.BuildQueryTableSpec(
		(&test_pb.FooState{}).J5Object().ObjectSchema(),
		(&test_pb.FooEvent{}).J5Object().ObjectSchema(),
	)
	if err != nil {
		t.Fatal(err)
	}

	barSpec, err := psm.BuildQueryTableSpec(
		(&test_pb.BarState{}).J5Object().ObjectSchema(),
		(&test_pb.BarEvent{}).J5Object().ObjectSchema(),
	)
	if err != nil {
		t.Fatal(err)
	}

	data, err := BuildStateMachineMigrations(fooSpec, barSpec)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))
}
